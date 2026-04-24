# Python WASM Runner

This directory contains the reusable WASI module used by `type: python` hooks.

The embedded binary lives at `pkg/actions/pythonrunner/python_hook_runner.wasm` and is embedded into Go using `runner.go`.

## Rebuild

From this directory:

```bash
rustup target add wasm32-wasip1
cargo build --release --target wasm32-wasip1
cp target/wasm32-wasip1/release/rustpython-wasi-runner.wasm python_hook_runner.wasm
```

## Python runtime contract

The Go side passes JSON on stdin:

```json
{"action": {...}, "args": {...}, "username": "..."}
```

The runner injects these Python globals before your script runs:

- `action`
- `args`
- `username`
- `lakefs` helper object

Available `lakefs` helper methods:

- `lakefs.list_objects(repo, ref, **kwargs)`
- `lakefs.stat_object(repo, ref, path, **kwargs)`
- `lakefs.diff_refs(repo, left_ref, right_ref, **kwargs)`
- `lakefs.create_tag(repo, ref, id)`
- `lakefs.update_object_user_metadata(repo, branch, path, set)`
- `lakefs.log(*args, sep=" ", end="\n")`
- `lakefs.fail(message)`
