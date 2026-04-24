use std::io::{self, Read};

use rustpython::vm::{PyResult, VirtualMachine};
use rustpython::{InterpreterBuilder, InterpreterBuilderExt};
use serde_json::Value;

#[link(wasm_import_module = "lakefs_actions_v1")]
unsafe extern "C" {
    fn lakefs_call(op_ptr: u32, op_len: u32, req_ptr: u32, req_len: u32) -> u64;
    fn result_len(handle: u64) -> i32;
    fn result_read(handle: u64, out_ptr: u32) -> i32;
    fn result_free(handle: u64);
    fn hook_fail(ptr: u32, len: u32);
}

const PYTHON_PRELUDE: &str = r#"
import json

class _LakeFS:
    def _call(self, operation, request):
        raw_response = _lakefs_call_raw(operation, json.dumps(request))
        response = json.loads(raw_response)
        status = response.get("status", 500)
        if status >= 400:
            raise Exception(f"lakefs call failed ({status}): {response}")
        return response.get("body")

    def list_objects(self, repo, ref, **kwargs):
        req = {"repo": repo, "ref": ref}
        req.update(kwargs)
        return self._call("list_objects", req)

    def stat_object(self, repo, ref, path, **kwargs):
        req = {"repo": repo, "ref": ref, "path": path}
        req.update(kwargs)
        return self._call("stat_object", req)

    def diff_refs(self, repo, left_ref, right_ref, **kwargs):
        req = {"repo": repo, "left_ref": left_ref, "right_ref": right_ref}
        req.update(kwargs)
        return self._call("diff_refs", req)

    def create_tag(self, repo, ref, id):
        return self._call("create_tag", {"repo": repo, "ref": ref, "id": id})

    def update_object_user_metadata(self, repo, branch, path, set):
        return self._call("update_object_user_metadata", {"repo": repo, "branch": branch, "path": path, "set": set})

    def fail(self, message):
        _lakefs_hook_fail(str(message))

    def log(self, *args, sep=" ", end="\n"):
        print(*args, sep=sep, end=end, flush=True)

lakefs = _LakeFS()
log = lakefs.log

__lakefs_input = json.loads(__lakefs_input_json)
action = __lakefs_input.get("action", {})
args = __lakefs_input.get("args", {})
username = __lakefs_input.get("username", "")
"#;

fn host_lakefs_call(
    operation: String,
    request_json: String,
    vm: &VirtualMachine,
) -> PyResult<String> {
    let op = operation.as_bytes();
    let req = request_json.as_bytes();
    let handle = unsafe {
        lakefs_call(
            op.as_ptr() as u32,
            op.len() as u32,
            req.as_ptr() as u32,
            req.len() as u32,
        )
    };

    let response_len = unsafe { result_len(handle) };
    if response_len < 0 {
        unsafe { result_free(handle) };
        return Err(vm.new_runtime_error("invalid response handle"));
    }
    let mut response = vec![0_u8; response_len as usize];
    let read = unsafe { result_read(handle, response.as_mut_ptr() as u32) };
    unsafe { result_free(handle) };
    if read != response_len {
        return Err(vm.new_runtime_error("invalid response read size"));
    }

    String::from_utf8(response).map_err(|err| vm.new_runtime_error(err.to_string()))
}

fn host_hook_fail(message: String, vm: &VirtualMachine) -> PyResult<()> {
    let msg = message.as_bytes();
    unsafe {
        hook_fail(msg.as_ptr() as u32, msg.len() as u32);
    }
    Err(vm.new_runtime_error("hook_fail returned unexpectedly"))
}

fn main() {
    let mut stdin = String::new();
    if io::stdin().read_to_string(&mut stdin).is_err() {
        eprintln!("failed reading stdin");
        std::process::exit(1);
    }

    let input: Value = match serde_json::from_str(&stdin) {
        Ok(v) => v,
        Err(err) => {
            eprintln!("invalid input json: {err}");
            std::process::exit(1);
        }
    };

    let script = input
        .get("args")
        .and_then(|a| a.get("script"))
        .and_then(|s| s.as_str())
        .unwrap_or("");
    if script.is_empty() {
        eprintln!("missing args.script");
        std::process::exit(1);
    }

    let interp = InterpreterBuilder::new().init_stdlib().interpreter();
    let exit_code = interp.run(|vm| {
        let scope = vm.new_scope_with_main()?;

        let lakefs_call_fn = vm.new_function("_lakefs_call_raw", host_lakefs_call);
        scope
            .globals
            .set_item("_lakefs_call_raw", lakefs_call_fn.into(), vm)?;
        let hook_fail_fn = vm.new_function("_lakefs_hook_fail", host_hook_fail);
        scope
            .globals
            .set_item("_lakefs_hook_fail", hook_fail_fn.into(), vm)?;
        scope.globals.set_item(
            "__lakefs_input_json",
            vm.ctx.new_str(stdin.as_str()).into(),
            vm,
        )?;

        vm.run_string(scope.clone(), PYTHON_PRELUDE, "<lakefs-prelude>".to_owned())?;
        vm.run_string(scope, script, "<lakefs-action>".to_owned())
            .map(|_| ())
    });
    if exit_code != 0 {
        std::process::exit(exit_code as i32);
    }
}
