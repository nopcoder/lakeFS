package host

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
)

const ModuleName = "lakefs_actions_v1"

var errHookFailed = errors.New("hook failed")

type Config struct {
	MaxRequestBytes uint32
	MaxResultBytes  uint32
	MaxResults      int
	OnHookFailure   func(message string)
	LakeFSCall      func(ctx context.Context, operation string, request []byte) ([]byte, error)
}

type State struct {
	cfg        Config
	mu         sync.Mutex
	results    map[uint64][]byte
	nextHandle uint64
}

func NewState(cfg Config) *State {
	if cfg.MaxRequestBytes == 0 {
		cfg.MaxRequestBytes = 1 << 20
	}
	if cfg.MaxResultBytes == 0 {
		cfg.MaxResultBytes = 4 << 20
	}
	if cfg.MaxResults == 0 {
		cfg.MaxResults = 128
	}
	return &State{
		cfg:        cfg,
		results:    make(map[uint64][]byte),
		nextHandle: 1,
	}
}

func (s *State) IsHookFailure(err error) bool {
	return errors.Is(err, errHookFailed)
}

func Instantiate(ctx context.Context, r wazero.Runtime, state *State) error {
	b := r.NewHostModuleBuilder(ModuleName)
	b.NewFunctionBuilder().WithFunc(state.hookFail).Export("hook_fail")
	b.NewFunctionBuilder().WithFunc(state.lakefsCall).Export("lakefs_call")
	b.NewFunctionBuilder().WithFunc(state.resultLen).Export("result_len")
	b.NewFunctionBuilder().WithFunc(state.resultRead).Export("result_read")
	b.NewFunctionBuilder().WithFunc(state.resultFree).Export("result_free")
	_, err := b.Instantiate(ctx)
	return err
}

func (s *State) hookFail(_ context.Context, mod api.Module, messagePtr, messageLen uint32) {
	messageBytes, err := readGuestBytes(mod, messagePtr, messageLen, s.cfg.MaxRequestBytes)
	if err != nil {
		panic(err)
	}
	if s.cfg.OnHookFailure != nil {
		s.cfg.OnHookFailure(string(messageBytes))
	}
	panic(errHookFailed)
}

func (s *State) lakefsCall(ctx context.Context, mod api.Module, opPtr, opLen, reqPtr, reqLen uint32) uint64 {
	if s.cfg.LakeFSCall == nil {
		panic("lakefs callback not configured")
	}

	opBytes, err := readGuestBytes(mod, opPtr, opLen, s.cfg.MaxRequestBytes)
	if err != nil {
		panic(err)
	}
	reqBytes, err := readGuestBytes(mod, reqPtr, reqLen, s.cfg.MaxRequestBytes)
	if err != nil {
		panic(err)
	}

	result, err := s.cfg.LakeFSCall(ctx, string(opBytes), reqBytes)
	if err != nil {
		panic(err)
	}
	if uint32(len(result)) > s.cfg.MaxResultBytes {
		panic(fmt.Errorf("result too large: %d", len(result)))
	}

	h, err := s.storeResult(result)
	if err != nil {
		panic(err)
	}
	return h
}

func (s *State) resultLen(_ context.Context, _ api.Module, handle uint64) int32 {
	s.mu.Lock()
	defer s.mu.Unlock()
	b, ok := s.results[handle]
	if !ok {
		return -1
	}
	return int32(len(b))
}

func (s *State) resultRead(_ context.Context, mod api.Module, handle uint64, outPtr uint32) int32 {
	s.mu.Lock()
	b, ok := s.results[handle]
	s.mu.Unlock()
	if !ok {
		return -1
	}
	memory := mod.Memory()
	if memory == nil {
		return -1
	}
	if !memory.Write(outPtr, b) {
		return -1
	}
	return int32(len(b))
}

func (s *State) resultFree(_ context.Context, _ api.Module, handle uint64) {
	s.mu.Lock()
	delete(s.results, handle)
	s.mu.Unlock()
}

func (s *State) storeResult(result []byte) (uint64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.results) >= s.cfg.MaxResults {
		return 0, fmt.Errorf("too many outstanding results")
	}
	h := s.nextHandle
	s.nextHandle++
	b := make([]byte, len(result))
	copy(b, result)
	s.results[h] = b
	return h, nil
}

func readGuestBytes(mod api.Module, ptr, size, max uint32) ([]byte, error) {
	if size > max {
		return nil, fmt.Errorf("buffer too large: %d", size)
	}
	memory := mod.Memory()
	if memory == nil {
		return nil, fmt.Errorf("module memory not available")
	}
	b, ok := memory.Read(ptr, size)
	if !ok {
		return nil, fmt.Errorf("invalid memory range ptr=%d len=%d", ptr, size)
	}
	out := make([]byte, len(b))
	copy(out, b)
	return out, nil
}
