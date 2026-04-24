#[link(wasm_import_module = "lakefs_actions_v1")]
extern "C" {
    fn hook_fail(ptr: u32, len: u32);
}

fn main() {
    let msg = b"wasm hook failed as requested";
    unsafe {
        hook_fail(msg.as_ptr() as u32, msg.len() as u32);
    }
}
