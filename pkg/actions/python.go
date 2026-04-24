package actions

import (
	"bytes"
	"context"
	"fmt"
	"maps"
	"net/http"

	"github.com/spf13/viper"
	"github.com/treeverse/lakefs/pkg/actions/pythonrunner"
	"github.com/treeverse/lakefs/pkg/auth"
	"github.com/treeverse/lakefs/pkg/graveler"
	"github.com/treeverse/lakefs/pkg/stats"
)

type PythonHook struct {
	HookBase
	ScriptPath string
	Script     string
	Args       map[string]any
	collector  stats.Collector
}

func NewPythonHook(h ActionHook, action *Action, cfg Config, e *http.Server, _ string, collector stats.Collector) (Hook, error) {
	if !cfg.Wasm.Enabled {
		return nil, fmt.Errorf("python hooks require wasm hooks enabled: %w", ErrInvalidAction)
	}

	args := make(map[string]any)
	argsVal, hasArgs := h.Properties["args"]
	if hasArgs {
		switch typedArgs := argsVal.(type) {
		case Properties:
			maps.Copy(args, typedArgs)
		case map[string]any:
			args = typedArgs
		default:
			return nil, fmt.Errorf("'args' should be a map: %w", errWrongValueType)
		}
	}
	parsedArgs, err := DescendArgs(args, &EnvironmentVariableGetter{
		Enabled: cfg.Env.Enabled,
		Prefix:  viper.GetEnvPrefix(),
	})
	if err != nil {
		return nil, fmt.Errorf("error parsing args: %w", err)
	}
	args, ok := parsedArgs.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("error parsing args, got wrong type: %T: %w", parsedArgs, ErrInvalidAction)
	}

	script := ""
	if scriptVal, ok := h.Properties["script"]; ok {
		v, ok := scriptVal.(string)
		if !ok {
			return nil, fmt.Errorf("'script' should be string: %w", errWrongValueType)
		}
		script = v
	}
	scriptPath := ""
	if scriptPathVal, ok := h.Properties["script_path"]; ok {
		v, ok := scriptPathVal.(string)
		if !ok {
			return nil, fmt.Errorf("'script_path' should be string: %w", errWrongValueType)
		}
		scriptPath = v
	}

	return &PythonHook{
		HookBase: HookBase{
			ID:         h.ID,
			ActionName: action.Name,
			Config:     cfg,
			Endpoint:   e,
		},
		ScriptPath: scriptPath,
		Script:     script,
		Args:       args,
		collector:  collector,
	}, nil
}

func (h *PythonHook) Run(ctx context.Context, record graveler.HookRecord, buf *bytes.Buffer) error {
	user, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}

	script := h.Script
	if script == "" {
		loader := &WasmHook{HookBase: h.HookBase}
		content, err := loader.loadObjectByPath(ctx, record, user, h.ScriptPath, "script_path")
		if err != nil {
			return err
		}
		script = string(content)
	}

	args := make(map[string]any, len(h.Args)+1)
	maps.Copy(args, h.Args)
	args["script"] = script

	payload := wasmInput{
		Action:   buildHookAction(h.ActionName, h.ID, record),
		Args:     args,
		Username: user.Username,
	}
	exec := &WasmHook{HookBase: h.HookBase}
	err = exec.runModule(ctx, buf, user, pythonrunner.RunnerWASM, payload)
	if err != nil {
		return fmt.Errorf("python runner failed: %w", err)
	}
	return nil
}
