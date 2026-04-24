package pythonrunner

import _ "embed"

// RunnerWASM is a reusable WASI module that executes Python code using RustPython.
//
//go:embed python_hook_runner.wasm
var RunnerWASM []byte
