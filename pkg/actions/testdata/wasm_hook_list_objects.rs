#[link(wasm_import_module = "lakefs_actions_v1")]
extern "C" {
    fn lakefs_call(op_ptr: u32, op_len: u32, req_ptr: u32, req_len: u32) -> u64;
    fn result_len(handle: u64) -> i32;
    fn result_read(handle: u64, out_ptr: u32) -> i32;
    fn result_free(handle: u64);
    fn hook_fail(ptr: u32, len: u32);
}

fn main() {
    let op = b"list_objects";
    let req = br#"{"repo":"example123","ref":"abc123"}"#;

    unsafe {
        let handle = lakefs_call(
            op.as_ptr() as u32,
            op.len() as u32,
            req.as_ptr() as u32,
            req.len() as u32,
        );
        let len = result_len(handle);
        if len <= 0 {
            let msg = b"invalid result length";
            hook_fail(msg.as_ptr() as u32, msg.len() as u32);
            return;
        }

        let mut buf = vec![0u8; len as usize];
        let read = result_read(handle, buf.as_mut_ptr() as u32);
        if read != len {
            let msg = b"failed reading host result";
            hook_fail(msg.as_ptr() as u32, msg.len() as u32);
            return;
        }

        result_free(handle);
        println!("{}", String::from_utf8_lossy(&buf));
    }
}
