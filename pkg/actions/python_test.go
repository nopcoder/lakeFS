package actions_test

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/treeverse/lakefs/pkg/actions"
)

func newPythonActionHook(t *testing.T, endpoint *http.Server, properties map[string]any) actions.Hook {
	t.Helper()

	mockStatsCollector := NewActionStatsMockCollector()
	actionConfig := actions.Config{Enabled: true}
	actionConfig.Wasm.Enabled = true

	h, err := actions.NewPythonHook(
		actions.ActionHook{
			ID:          "hook-" + t.Name(),
			Type:        actions.HookTypePython,
			Description: t.Name() + " hook description",
			Properties:  properties,
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

func TestPythonRun_InlineScript(t *testing.T) {
	t.Parallel()

	h := newPythonActionHook(t, nil, map[string]any{
		"script": `print("hello from python")`,
	})

	output, err := runHook(t.Context(), h)
	require.NoError(t, err)
	require.Contains(t, output, "hello from python")
}

func TestPythonRun_ScriptPath(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Query().Get("path") == "hooks/check.py" {
			_, _ = fmt.Fprint(w, `print("python from path")`)
			return
		}
		http.NotFound(w, r)
	})

	h := newPythonActionHook(t, &http.Server{Handler: handler}, map[string]any{
		"script_path": "hooks/check.py",
	})

	output, err := runHook(t.Context(), h)
	require.NoError(t, err)
	require.Contains(t, output, "python from path")
}

func TestPythonRun_ScriptError(t *testing.T) {
	t.Parallel()

	h := newPythonActionHook(t, nil, map[string]any{
		"script": `raise Exception("boom")`,
	})

	_, err := runHook(t.Context(), h)
	require.Error(t, err)
	require.Contains(t, err.Error(), "python runner failed")
	require.Contains(t, err.Error(), "boom")
}

func TestPythonRun_LakeFSHelpers(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/objects/ls") {
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprint(w, `{"results":[{"path":"from-helper.txt"}],"pagination":{"has_more":false,"next_offset":""}}`)
			return
		}
		http.NotFound(w, r)
	})

	h := newPythonActionHook(t, &http.Server{Handler: handler}, map[string]any{
		"script": `
result = lakefs.list_objects("example123", "abc123")
print(result["results"][0]["path"])
lakefs.log("logged via helper")
`,
	})

	output, err := runHook(t.Context(), h)
	require.NoError(t, err)
	require.Contains(t, output, "from-helper.txt")
	require.Contains(t, output, "logged via helper")
}
