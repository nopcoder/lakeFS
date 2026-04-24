package host

import "testing"

func TestStateStoreResultLimit(t *testing.T) {
	t.Parallel()

	state := NewState(Config{MaxResults: 1})
	if _, err := state.storeResult([]byte("first")); err != nil {
		t.Fatalf("store first result: %v", err)
	}
	if _, err := state.storeResult([]byte("second")); err == nil {
		t.Fatal("expected error when results exceed max")
	}
}

func TestStateResultLifecycle(t *testing.T) {
	t.Parallel()

	state := NewState(Config{MaxResults: 2})
	h, err := state.storeResult([]byte("payload"))
	if err != nil {
		t.Fatalf("store result: %v", err)
	}

	state.mu.Lock()
	_, ok := state.results[h]
	state.mu.Unlock()
	if !ok {
		t.Fatalf("expected result for handle %d", h)
	}

	state.resultFree(nil, nil, h)

	state.mu.Lock()
	_, ok = state.results[h]
	state.mu.Unlock()
	if ok {
		t.Fatalf("expected result for handle %d to be released", h)
	}
}
