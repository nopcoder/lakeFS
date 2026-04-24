package actions_test

import (
	"net/http"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/treeverse/lakefs/pkg/actions"
)

func newWasmActionHook(t *testing.T, endpoint *http.Server, modulePath string) actions.Hook {
	t.Helper()

	mockStatsCollector := NewActionStatsMockCollector()
	actionConfig := actions.Config{Enabled: true}
	actionConfig.Wasm.Enabled = true
	actionConfig.Wasm.MemoryLimitPages = 64

	h, err := actions.NewWasmHook(
		actions.ActionHook{
			ID:          "hook-" + t.Name(),
			Type:        actions.HookTypeWASM,
			Description: t.Name() + " hook description",
			Properties: map[string]any{
				"module_path": modulePath,
			},
		},
		&actions.Action{},
		actionConfig,
		endpoint,
		"",
		&mockStatsCollector,
	)
	require.NoError(t, err)
	return h
}

func TestWasmRun_LakeFSCall(t *testing.T) {
	t.Parallel()

	module, err := os.ReadFile(path.Join("testdata", "wasm_hook_list_objects.wasm"))
	require.NoError(t, err)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Query().Get("path") == "hooks/wasm_hook_list_objects.wasm":
			_, _ = w.Write(module)
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/objects/ls"):
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"results":[{"path":"a.txt"}],"pagination":{"has_more":false,"next_offset":""}}`))
		default:
			http.NotFound(w, r)
		}
	})

	h := newWasmActionHook(t, &http.Server{Handler: handler}, "hooks/wasm_hook_list_objects.wasm")
	output, err := runHook(t.Context(), h)
	require.NoError(t, err)
	require.Contains(t, output, `"status":200`)
	require.Contains(t, output, "a.txt")
}

func TestWasmRun_HookFail(t *testing.T) {
	t.Parallel()

	module, err := os.ReadFile(path.Join("testdata", "wasm_hook_fail.wasm"))
	require.NoError(t, err)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Query().Get("path") == "hooks/wasm_hook_fail.wasm" {
			_, _ = w.Write(module)
			return
		}
		http.NotFound(w, r)
	})

	h := newWasmActionHook(t, &http.Server{Handler: handler}, "hooks/wasm_hook_fail.wasm")
	_, err = runHook(t.Context(), h)
	require.Error(t, err)
	require.Contains(t, err.Error(), "wasm hook failed as requested")
}
