package actions

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"net/http"
	"net/http/httptest"
	"net/url"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/spf13/viper"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
	wazerosys "github.com/tetratelabs/wazero/sys"
	whost "github.com/treeverse/lakefs/pkg/actions/wasm/host"
	"github.com/treeverse/lakefs/pkg/api/apiutil"
	"github.com/treeverse/lakefs/pkg/auth"
	"github.com/treeverse/lakefs/pkg/auth/model"
	"github.com/treeverse/lakefs/pkg/graveler"
	"github.com/treeverse/lakefs/pkg/logging"
	"github.com/treeverse/lakefs/pkg/stats"

	luahook "github.com/treeverse/lakefs/pkg/actions/lua/hook"
	"github.com/treeverse/lakefs/pkg/actions/lua/lakefs"
)

const wasmRemoteAddr = "[internal (wasm)]"

type WasmHook struct {
	HookBase
	ModulePath string
	Args       map[string]any
	collector  stats.Collector
}

type wasmInput struct {
	Action   map[string]any `json:"action"`
	Args     map[string]any `json:"args"`
	Username string         `json:"username"`
}

func (h *WasmHook) Run(ctx context.Context, record graveler.HookRecord, buf *bytes.Buffer) error {
	if !h.Config.Wasm.Enabled {
		return fmt.Errorf("wasm hooks are disabled: %w", ErrInvalidAction)
	}

	user, err := auth.GetUser(ctx)
	if err != nil {
		return err
	}

	module, err := h.loadObjectByPath(ctx, record, user, h.ModulePath, "module_path")
	if err != nil {
		return err
	}

	payload := wasmInput{
		Action:   buildHookAction(h.ActionName, h.ID, record),
		Args:     h.Args,
		Username: user.Username,
	}

	return h.runModule(ctx, buf, user, module, payload)
}

func (h *WasmHook) runModule(ctx context.Context, buf *bytes.Buffer, user *model.User, module []byte, payload wasmInput) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	runCtx := ctx
	if h.Config.Wasm.Timeout > 0 {
		var cancel context.CancelFunc
		runCtx, cancel = context.WithTimeout(ctx, h.Config.Wasm.Timeout)
		defer cancel()
	}

	runtimeCfg := wazero.NewRuntimeConfig().WithCloseOnContextDone(true)
	if h.Config.Wasm.MemoryLimitPages > 0 {
		runtimeCfg = runtimeCfg.WithMemoryLimitPages(h.Config.Wasm.MemoryLimitPages)
	}
	runtime := wazero.NewRuntimeWithConfig(runCtx, runtimeCfg)
	defer runtime.Close(context.Background())

	if _, err := wasi_snapshot_preview1.Instantiate(runCtx, runtime); err != nil {
		return fmt.Errorf("instantiate wasi: %w", err)
	}

	var hookFailure string
	state := whost.NewState(whost.Config{
		OnHookFailure: func(message string) {
			hookFailure = message
		},
		LakeFSCall: func(callCtx context.Context, operation string, request []byte) ([]byte, error) {
			return h.lakefsCall(callCtx, user, operation, request)
		},
	})
	if err := whost.Instantiate(runCtx, runtime, state); err != nil {
		return fmt.Errorf("instantiate host module: %w", err)
	}

	out := &loggingBuffer{buf: buf, ctx: ctx}
	moduleCfg := wazero.NewModuleConfig().
		WithName("").
		WithStdin(bytes.NewReader(payloadBytes)).
		WithStdout(out).
		WithStderr(out).
		WithStartFunctions("_start")

	wasmModule, err := runtime.InstantiateWithConfig(runCtx, module, moduleCfg)
	if err == nil {
		_ = wasmModule.Close(context.Background())
	}
	if hookFailure != "" {
		return luahook.ErrHookFailure(hookFailure)
	}
	if err == nil {
		return nil
	}
	if state.IsHookFailure(err) {
		return luahook.ErrHookFailure("hook failed")
	}
	var exitErr *wazerosys.ExitError
	if errors.As(err, &exitErr) {
		if exitErr.ExitCode() == 0 {
			return nil
		}
		if buf.Len() > 0 {
			return fmt.Errorf("wasm module exited with code %d: %s", exitErr.ExitCode(), bytes.TrimSpace(buf.Bytes()))
		}
		return fmt.Errorf("wasm module exited with code %d", exitErr.ExitCode())
	}
	if buf.Len() > 0 {
		return fmt.Errorf("%w: %s", err, bytes.TrimSpace(buf.Bytes()))
	}
	return err
}

func (h *WasmHook) lakefsCall(ctx context.Context, user *model.User, operation string, request []byte) ([]byte, error) {
	type response struct {
		Status int `json:"status"`
		Body   any `json:"body,omitempty"`
	}

	var (
		status int
		body   any
		err    error
	)

	switch operation {
	case "list_objects":
		status, body, err = h.opListObjects(ctx, user, request)
	case "stat_object":
		status, body, err = h.opStatObject(ctx, user, request)
	case "diff_refs":
		status, body, err = h.opDiffRefs(ctx, user, request)
	case "create_tag":
		status, body, err = h.opCreateTag(ctx, user, request)
	case "update_object_user_metadata":
		status, body, err = h.opUpdateObjectUserMetadata(ctx, user, request)
	default:
		return nil, fmt.Errorf("unknown operation: %s", operation)
	}
	if err != nil {
		return nil, err
	}
	return json.Marshal(response{Status: status, Body: body})
}

func buildHookAction(actionName, hookID string, record graveler.HookRecord) map[string]any {
	parents := make([]string, len(record.Commit.Parents))
	for i := 0; i < len(record.Commit.Parents); i++ {
		parents[i] = string(record.Commit.Parents[i])
	}
	metadata := make(map[string]string, len(record.Commit.Metadata))
	maps.Copy(metadata, record.Commit.Metadata)
	return map[string]any{
		"action_name":       actionName,
		"hook_id":           hookID,
		"run_id":            record.RunID,
		"pre_run_id":        record.PreRunID,
		"event_type":        string(record.EventType),
		"commit_id":         record.CommitID.String(),
		"branch_id":         record.BranchID.String(),
		"source_ref":        record.SourceRef.String(),
		"tag_id":            record.TagID.String(),
		"merge_source":      record.MergeSource.String(),
		"repository_id":     record.Repository.RepositoryID.String(),
		"storage_namespace": record.Repository.StorageNamespace.String(),
		"commit": map[string]any{
			"message":       record.Commit.Message,
			"meta_range_id": record.Commit.MetaRangeID.String(),
			"creation_date": record.Commit.CreationDate.Format(time.RFC3339),
			"version":       int(record.Commit.Version),
			"metadata":      metadata,
			"parents":       parents,
		},
	}
}

func (h *WasmHook) loadObjectByPath(ctx context.Context, record graveler.HookRecord, user *model.User, objectPath, propertyName string) ([]byte, error) {
	if h.Endpoint == nil {
		return nil, fmt.Errorf("no endpoint configured, cannot request object: %s: %w", objectPath, ErrInvalidAction)
	}
	reqURL, err := url.JoinPath(apiutil.BaseURL,
		"repositories", string(record.Repository.RepositoryID), "refs", string(record.SourceRef), "objects")
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(auth.WithUser(req.Context(), user))
	req = req.WithContext(logging.AddFields(req.Context(), getAllowedFields(logging.GetFieldsFromContext(ctx))))
	req.RemoteAddr = wasmRemoteAddr
	q := req.URL.Query()
	q.Add("path", objectPath)
	req.URL.RawQuery = q.Encode()
	rr := httptest.NewRecorder()
	h.Endpoint.Handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		return nil, fmt.Errorf("could not load %s %s: HTTP %d: %w", propertyName, objectPath, rr.Code, ErrInvalidAction)
	}
	return rr.Body.Bytes(), nil
}

func (h *WasmHook) opCreateTag(ctx context.Context, user *model.User, request []byte) (int, any, error) {
	type req struct {
		Repo string `json:"repo"`
		Ref  string `json:"ref"`
		ID   string `json:"id"`
	}
	var in req
	if err := json.Unmarshal(request, &in); err != nil {
		return 0, nil, err
	}
	data, err := json.Marshal(map[string]string{"ref": in.Ref, "id": in.ID})
	if err != nil {
		return 0, nil, err
	}
	reqURL, err := url.JoinPath("/repositories", in.Repo, "tags")
	if err != nil {
		return 0, nil, err
	}
	return h.jsonRequest(ctx, user, http.MethodPost, reqURL, data, nil)
}

func (h *WasmHook) opUpdateObjectUserMetadata(ctx context.Context, user *model.User, request []byte) (int, any, error) {
	type req struct {
		Repo   string            `json:"repo"`
		Branch string            `json:"branch"`
		Path   string            `json:"path"`
		Set    map[string]string `json:"set"`
	}
	var in req
	if err := json.Unmarshal(request, &in); err != nil {
		return 0, nil, err
	}
	data, err := json.Marshal(map[string]any{"set": in.Set})
	if err != nil {
		return 0, nil, err
	}
	reqURL, err := url.JoinPath("/repositories", in.Repo, "branches", in.Branch, "objects/stat/user_metadata")
	if err != nil {
		return 0, nil, err
	}
	query := map[string]string{"path": in.Path}
	return h.jsonRequest(ctx, user, http.MethodPut, reqURL, data, query)
}

func (h *WasmHook) opListObjects(ctx context.Context, user *model.User, request []byte) (int, any, error) {
	type req struct {
		Repo         string `json:"repo"`
		Ref          string `json:"ref"`
		After        string `json:"after,omitempty"`
		Prefix       string `json:"prefix,omitempty"`
		Delimiter    string `json:"delimiter,omitempty"`
		Amount       int    `json:"amount,omitempty"`
		UserMetadata *bool  `json:"user_metadata,omitempty"`
	}
	var in req
	if err := json.Unmarshal(request, &in); err != nil {
		return 0, nil, err
	}
	reqURL, err := url.JoinPath("/repositories", in.Repo, "refs", in.Ref, "objects/ls")
	if err != nil {
		return 0, nil, err
	}
	query := map[string]string{}
	if in.After != "" {
		query["after"] = in.After
	}
	if in.Prefix != "" {
		query["prefix"] = in.Prefix
	}
	if in.Delimiter != "" {
		query["delimiter"] = in.Delimiter
	}
	if in.Amount > 0 {
		query["amount"] = fmt.Sprintf("%d", in.Amount)
	}
	if in.UserMetadata != nil {
		query["user_metadata"] = fmt.Sprintf("%t", *in.UserMetadata)
	}
	return h.jsonRequest(ctx, user, http.MethodGet, reqURL, nil, query)
}

func (h *WasmHook) opStatObject(ctx context.Context, user *model.User, request []byte) (int, any, error) {
	type req struct {
		Repo         string `json:"repo"`
		Ref          string `json:"ref"`
		Path         string `json:"path"`
		UserMetadata *bool  `json:"user_metadata,omitempty"`
	}
	var in req
	if err := json.Unmarshal(request, &in); err != nil {
		return 0, nil, err
	}
	reqURL, err := url.JoinPath("/repositories", in.Repo, "refs", in.Ref, "objects", "stat")
	if err != nil {
		return 0, nil, err
	}
	query := map[string]string{"path": in.Path}
	if in.UserMetadata != nil {
		query["user_metadata"] = fmt.Sprintf("%t", *in.UserMetadata)
	}
	return h.jsonRequest(ctx, user, http.MethodGet, reqURL, nil, query)
}

func (h *WasmHook) opDiffRefs(ctx context.Context, user *model.User, request []byte) (int, any, error) {
	type req struct {
		Repo      string `json:"repo"`
		LeftRef   string `json:"left_ref"`
		RightRef  string `json:"right_ref"`
		After     string `json:"after,omitempty"`
		Prefix    string `json:"prefix,omitempty"`
		Delimiter string `json:"delimiter,omitempty"`
		Amount    int    `json:"amount,omitempty"`
	}
	var in req
	if err := json.Unmarshal(request, &in); err != nil {
		return 0, nil, err
	}
	reqURL, err := url.JoinPath("/repositories", in.Repo, "refs", in.LeftRef, "diff", in.RightRef)
	if err != nil {
		return 0, nil, err
	}
	query := map[string]string{}
	if in.After != "" {
		query["after"] = in.After
	}
	if in.Prefix != "" {
		query["prefix"] = in.Prefix
	}
	if in.Delimiter != "" {
		query["delimiter"] = in.Delimiter
	}
	if in.Amount > 0 {
		query["amount"] = fmt.Sprintf("%d", in.Amount)
	}
	return h.jsonRequest(ctx, user, http.MethodGet, reqURL, nil, query)
}

func (h *WasmHook) jsonRequest(ctx context.Context, user *model.User, method, reqURL string, body []byte, query map[string]string) (int, any, error) {
	req, err := h.newLakeFSJSONRequest(ctx, user, method, reqURL, body)
	if err != nil {
		return 0, nil, err
	}
	q := req.URL.Query()
	for k, v := range query {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()
	rr := httptest.NewRecorder()
	h.Endpoint.Handler.ServeHTTP(rr, req)
	if rr.Body.Len() == 0 {
		return rr.Code, nil, nil
	}
	var output any
	if err := json.Unmarshal(rr.Body.Bytes(), &output); err != nil {
		return rr.Code, rr.Body.String(), nil
	}
	return rr.Code, output, nil
}

func (h *WasmHook) newLakeFSJSONRequest(ctx context.Context, user *model.User, method, reqURL string, data []byte) (*http.Request, error) {
	if !h.Config.Wasm.Enabled {
		return nil, fmt.Errorf("wasm hooks are disabled: %w", ErrInvalidAction)
	}
	if !bytes.HasPrefix([]byte(reqURL), []byte("/api/")) {
		var err error
		reqURL, err = url.JoinPath(apiutil.BaseURL, reqURL)
		if err != nil {
			return nil, err
		}
	}
	ctx = context.WithValue(ctx, chi.RouteCtxKey, nil)
	ctx = auth.WithUser(ctx, user)
	ctx = logging.AddFields(ctx, getAllowedFields(logging.GetFieldsFromContext(ctx)))
	var body bytes.Reader
	if data == nil {
		body = *bytes.NewReader(nil)
	} else {
		body = *bytes.NewReader(data)
	}
	req, err := http.NewRequestWithContext(ctx, method, reqURL, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", lakefs.LuaClientUserAgent)
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = wasmRemoteAddr
	return req, nil
}

func NewWasmHook(h ActionHook, action *Action, cfg Config, e *http.Server, _ string, collector stats.Collector) (Hook, error) {
	if !cfg.Wasm.Enabled {
		return nil, fmt.Errorf("wasm hooks are disabled: %w", ErrInvalidAction)
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

	modulePath, err := h.Properties.getRequiredProperty("module_path")
	if err != nil {
		return nil, err
	}

	return &WasmHook{
		HookBase: HookBase{
			ID:         h.ID,
			ActionName: action.Name,
			Config:     cfg,
			Endpoint:   e,
		},
		ModulePath: modulePath,
		Args:       args,
		collector:  collector,
	}, nil
}
