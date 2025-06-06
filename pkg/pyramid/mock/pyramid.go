// Code generated by MockGen. DO NOT EDIT.
// Source: pyramid.go

// Package mock is a generated GoMock package.
package mock

import (
	context "context"
	os "os"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	pyramid "github.com/treeverse/lakefs/pkg/pyramid"
)

// MockFS is a mock of FS interface.
type MockFS struct {
	ctrl     *gomock.Controller
	recorder *MockFSMockRecorder
}

// MockFSMockRecorder is the mock recorder for MockFS.
type MockFSMockRecorder struct {
	mock *MockFS
}

// NewMockFS creates a new mock instance.
func NewMockFS(ctrl *gomock.Controller) *MockFS {
	mock := &MockFS{ctrl: ctrl}
	mock.recorder = &MockFSMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockFS) EXPECT() *MockFSMockRecorder {
	return m.recorder
}

// Create mocks base method.
func (m *MockFS) Create(ctx context.Context, storageID, namespace string) (pyramid.StoredFile, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Create", ctx, storageID, namespace)
	ret0, _ := ret[0].(pyramid.StoredFile)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Create indicates an expected call of Create.
func (mr *MockFSMockRecorder) Create(ctx, storageID, namespace interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockFS)(nil).Create), ctx, storageID, namespace)
}

// Exists mocks base method.
func (m *MockFS) Exists(ctx context.Context, storageID, namespace, filename string) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Exists", ctx, storageID, namespace, filename)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Exists indicates an expected call of Exists.
func (mr *MockFSMockRecorder) Exists(ctx, storageID, namespace, filename interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Exists", reflect.TypeOf((*MockFS)(nil).Exists), ctx, storageID, namespace, filename)
}

// GetRemoteURI mocks base method.
func (m *MockFS) GetRemoteURI(ctx context.Context, filename string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetRemoteURI", ctx, filename)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetRemoteURI indicates an expected call of GetRemoteURI.
func (mr *MockFSMockRecorder) GetRemoteURI(ctx, filename interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetRemoteURI", reflect.TypeOf((*MockFS)(nil).GetRemoteURI), ctx, filename)
}

// Open mocks base method.
func (m *MockFS) Open(ctx context.Context, storageID, namespace, filename string) (pyramid.File, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Open", ctx, storageID, namespace, filename)
	ret0, _ := ret[0].(pyramid.File)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Open indicates an expected call of Open.
func (mr *MockFSMockRecorder) Open(ctx, storageID, namespace, filename interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Open", reflect.TypeOf((*MockFS)(nil).Open), ctx, storageID, namespace, filename)
}

// MockFile is a mock of File interface.
type MockFile struct {
	ctrl     *gomock.Controller
	recorder *MockFileMockRecorder
}

// MockFileMockRecorder is the mock recorder for MockFile.
type MockFileMockRecorder struct {
	mock *MockFile
}

// NewMockFile creates a new mock instance.
func NewMockFile(ctrl *gomock.Controller) *MockFile {
	mock := &MockFile{ctrl: ctrl}
	mock.recorder = &MockFileMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockFile) EXPECT() *MockFileMockRecorder {
	return m.recorder
}

// Close mocks base method.
func (m *MockFile) Close() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close.
func (mr *MockFileMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockFile)(nil).Close))
}

// Read mocks base method.
func (m *MockFile) Read(p []byte) (int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Read", p)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Read indicates an expected call of Read.
func (mr *MockFileMockRecorder) Read(p interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Read", reflect.TypeOf((*MockFile)(nil).Read), p)
}

// ReadAt mocks base method.
func (m *MockFile) ReadAt(p []byte, off int64) (int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ReadAt", p, off)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ReadAt indicates an expected call of ReadAt.
func (mr *MockFileMockRecorder) ReadAt(p, off interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ReadAt", reflect.TypeOf((*MockFile)(nil).ReadAt), p, off)
}

// Stat mocks base method.
func (m *MockFile) Stat() (os.FileInfo, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Stat")
	ret0, _ := ret[0].(os.FileInfo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Stat indicates an expected call of Stat.
func (mr *MockFileMockRecorder) Stat() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Stat", reflect.TypeOf((*MockFile)(nil).Stat))
}

// Sync mocks base method.
func (m *MockFile) Sync() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Sync")
	ret0, _ := ret[0].(error)
	return ret0
}

// Sync indicates an expected call of Sync.
func (mr *MockFileMockRecorder) Sync() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Sync", reflect.TypeOf((*MockFile)(nil).Sync))
}

// Write mocks base method.
func (m *MockFile) Write(p []byte) (int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Write", p)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Write indicates an expected call of Write.
func (mr *MockFileMockRecorder) Write(p interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Write", reflect.TypeOf((*MockFile)(nil).Write), p)
}

// MockStoredFile is a mock of StoredFile interface.
type MockStoredFile struct {
	ctrl     *gomock.Controller
	recorder *MockStoredFileMockRecorder
}

// MockStoredFileMockRecorder is the mock recorder for MockStoredFile.
type MockStoredFileMockRecorder struct {
	mock *MockStoredFile
}

// NewMockStoredFile creates a new mock instance.
func NewMockStoredFile(ctrl *gomock.Controller) *MockStoredFile {
	mock := &MockStoredFile{ctrl: ctrl}
	mock.recorder = &MockStoredFileMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockStoredFile) EXPECT() *MockStoredFileMockRecorder {
	return m.recorder
}

// Abort mocks base method.
func (m *MockStoredFile) Abort(ctx context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Abort", ctx)
	ret0, _ := ret[0].(error)
	return ret0
}

// Abort indicates an expected call of Abort.
func (mr *MockStoredFileMockRecorder) Abort(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Abort", reflect.TypeOf((*MockStoredFile)(nil).Abort), ctx)
}

// Close mocks base method.
func (m *MockStoredFile) Close() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close.
func (mr *MockStoredFileMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockStoredFile)(nil).Close))
}

// Read mocks base method.
func (m *MockStoredFile) Read(p []byte) (int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Read", p)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Read indicates an expected call of Read.
func (mr *MockStoredFileMockRecorder) Read(p interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Read", reflect.TypeOf((*MockStoredFile)(nil).Read), p)
}

// ReadAt mocks base method.
func (m *MockStoredFile) ReadAt(p []byte, off int64) (int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ReadAt", p, off)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ReadAt indicates an expected call of ReadAt.
func (mr *MockStoredFileMockRecorder) ReadAt(p, off interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ReadAt", reflect.TypeOf((*MockStoredFile)(nil).ReadAt), p, off)
}

// Stat mocks base method.
func (m *MockStoredFile) Stat() (os.FileInfo, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Stat")
	ret0, _ := ret[0].(os.FileInfo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Stat indicates an expected call of Stat.
func (mr *MockStoredFileMockRecorder) Stat() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Stat", reflect.TypeOf((*MockStoredFile)(nil).Stat))
}

// Store mocks base method.
func (m *MockStoredFile) Store(ctx context.Context, filename string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Store", ctx, filename)
	ret0, _ := ret[0].(error)
	return ret0
}

// Store indicates an expected call of Store.
func (mr *MockStoredFileMockRecorder) Store(ctx, filename interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Store", reflect.TypeOf((*MockStoredFile)(nil).Store), ctx, filename)
}

// Sync mocks base method.
func (m *MockStoredFile) Sync() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Sync")
	ret0, _ := ret[0].(error)
	return ret0
}

// Sync indicates an expected call of Sync.
func (mr *MockStoredFileMockRecorder) Sync() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Sync", reflect.TypeOf((*MockStoredFile)(nil).Sync))
}

// Write mocks base method.
func (m *MockStoredFile) Write(p []byte) (int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Write", p)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Write indicates an expected call of Write.
func (mr *MockStoredFileMockRecorder) Write(p interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Write", reflect.TypeOf((*MockStoredFile)(nil).Write), p)
}
