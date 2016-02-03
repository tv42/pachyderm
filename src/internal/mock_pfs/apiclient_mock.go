// Automatically generated by MockGen. DO NOT EDIT!
// Source: github.com/pachyderm/pachyderm/src/pfs (interfaces: APIClient)

package mock_pfs

import (
	gomock "github.com/golang/mock/gomock"
	pfs "github.com/pachyderm/pachyderm/src/pfs"
	protobuf "go.pedge.io/pb/go/google/protobuf"
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Mock of APIClient interface
type MockAPIClient struct {
	ctrl     *gomock.Controller
	recorder *_MockAPIClientRecorder
}

// Recorder for MockAPIClient (not exported)
type _MockAPIClientRecorder struct {
	mock *MockAPIClient
}

func NewMockAPIClient(ctrl *gomock.Controller) *MockAPIClient {
	mock := &MockAPIClient{ctrl: ctrl}
	mock.recorder = &_MockAPIClientRecorder{mock}
	return mock
}

func (_m *MockAPIClient) EXPECT() *_MockAPIClientRecorder {
	return _m.recorder
}

func (_m *MockAPIClient) CreateRepo(_param0 context.Context, _param1 *pfs.CreateRepoRequest, _param2 ...grpc.CallOption) (*protobuf.Empty, error) {
	_s := []interface{}{_param0, _param1}
	for _, _x := range _param2 {
		_s = append(_s, _x)
	}
	ret := _m.ctrl.Call(_m, "CreateRepo", _s...)
	ret0, _ := ret[0].(*protobuf.Empty)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockAPIClientRecorder) CreateRepo(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	_s := append([]interface{}{arg0, arg1}, arg2...)
	return _mr.mock.ctrl.RecordCall(_mr.mock, "CreateRepo", _s...)
}

func (_m *MockAPIClient) DeleteCommit(_param0 context.Context, _param1 *pfs.DeleteCommitRequest, _param2 ...grpc.CallOption) (*protobuf.Empty, error) {
	_s := []interface{}{_param0, _param1}
	for _, _x := range _param2 {
		_s = append(_s, _x)
	}
	ret := _m.ctrl.Call(_m, "DeleteCommit", _s...)
	ret0, _ := ret[0].(*protobuf.Empty)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockAPIClientRecorder) DeleteCommit(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	_s := append([]interface{}{arg0, arg1}, arg2...)
	return _mr.mock.ctrl.RecordCall(_mr.mock, "DeleteCommit", _s...)
}

func (_m *MockAPIClient) DeleteFile(_param0 context.Context, _param1 *pfs.DeleteFileRequest, _param2 ...grpc.CallOption) (*protobuf.Empty, error) {
	_s := []interface{}{_param0, _param1}
	for _, _x := range _param2 {
		_s = append(_s, _x)
	}
	ret := _m.ctrl.Call(_m, "DeleteFile", _s...)
	ret0, _ := ret[0].(*protobuf.Empty)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockAPIClientRecorder) DeleteFile(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	_s := append([]interface{}{arg0, arg1}, arg2...)
	return _mr.mock.ctrl.RecordCall(_mr.mock, "DeleteFile", _s...)
}

func (_m *MockAPIClient) DeleteRepo(_param0 context.Context, _param1 *pfs.DeleteRepoRequest, _param2 ...grpc.CallOption) (*protobuf.Empty, error) {
	_s := []interface{}{_param0, _param1}
	for _, _x := range _param2 {
		_s = append(_s, _x)
	}
	ret := _m.ctrl.Call(_m, "DeleteRepo", _s...)
	ret0, _ := ret[0].(*protobuf.Empty)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockAPIClientRecorder) DeleteRepo(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	_s := append([]interface{}{arg0, arg1}, arg2...)
	return _mr.mock.ctrl.RecordCall(_mr.mock, "DeleteRepo", _s...)
}

func (_m *MockAPIClient) FinishCommit(_param0 context.Context, _param1 *pfs.FinishCommitRequest, _param2 ...grpc.CallOption) (*protobuf.Empty, error) {
	_s := []interface{}{_param0, _param1}
	for _, _x := range _param2 {
		_s = append(_s, _x)
	}
	ret := _m.ctrl.Call(_m, "FinishCommit", _s...)
	ret0, _ := ret[0].(*protobuf.Empty)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockAPIClientRecorder) FinishCommit(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	_s := append([]interface{}{arg0, arg1}, arg2...)
	return _mr.mock.ctrl.RecordCall(_mr.mock, "FinishCommit", _s...)
}

func (_m *MockAPIClient) GetFile(_param0 context.Context, _param1 *pfs.GetFileRequest, _param2 ...grpc.CallOption) (pfs.API_GetFileClient, error) {
	_s := []interface{}{_param0, _param1}
	for _, _x := range _param2 {
		_s = append(_s, _x)
	}
	ret := _m.ctrl.Call(_m, "GetFile", _s...)
	ret0, _ := ret[0].(pfs.API_GetFileClient)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockAPIClientRecorder) GetFile(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	_s := append([]interface{}{arg0, arg1}, arg2...)
	return _mr.mock.ctrl.RecordCall(_mr.mock, "GetFile", _s...)
}

func (_m *MockAPIClient) InspectCommit(_param0 context.Context, _param1 *pfs.InspectCommitRequest, _param2 ...grpc.CallOption) (*pfs.CommitInfo, error) {
	_s := []interface{}{_param0, _param1}
	for _, _x := range _param2 {
		_s = append(_s, _x)
	}
	ret := _m.ctrl.Call(_m, "InspectCommit", _s...)
	ret0, _ := ret[0].(*pfs.CommitInfo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockAPIClientRecorder) InspectCommit(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	_s := append([]interface{}{arg0, arg1}, arg2...)
	return _mr.mock.ctrl.RecordCall(_mr.mock, "InspectCommit", _s...)
}

func (_m *MockAPIClient) InspectFile(_param0 context.Context, _param1 *pfs.InspectFileRequest, _param2 ...grpc.CallOption) (*pfs.FileInfo, error) {
	_s := []interface{}{_param0, _param1}
	for _, _x := range _param2 {
		_s = append(_s, _x)
	}
	ret := _m.ctrl.Call(_m, "InspectFile", _s...)
	ret0, _ := ret[0].(*pfs.FileInfo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockAPIClientRecorder) InspectFile(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	_s := append([]interface{}{arg0, arg1}, arg2...)
	return _mr.mock.ctrl.RecordCall(_mr.mock, "InspectFile", _s...)
}

func (_m *MockAPIClient) InspectRepo(_param0 context.Context, _param1 *pfs.InspectRepoRequest, _param2 ...grpc.CallOption) (*pfs.RepoInfo, error) {
	_s := []interface{}{_param0, _param1}
	for _, _x := range _param2 {
		_s = append(_s, _x)
	}
	ret := _m.ctrl.Call(_m, "InspectRepo", _s...)
	ret0, _ := ret[0].(*pfs.RepoInfo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockAPIClientRecorder) InspectRepo(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	_s := append([]interface{}{arg0, arg1}, arg2...)
	return _mr.mock.ctrl.RecordCall(_mr.mock, "InspectRepo", _s...)
}

func (_m *MockAPIClient) ListCommit(_param0 context.Context, _param1 *pfs.ListCommitRequest, _param2 ...grpc.CallOption) (*pfs.CommitInfos, error) {
	_s := []interface{}{_param0, _param1}
	for _, _x := range _param2 {
		_s = append(_s, _x)
	}
	ret := _m.ctrl.Call(_m, "ListCommit", _s...)
	ret0, _ := ret[0].(*pfs.CommitInfos)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockAPIClientRecorder) ListCommit(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	_s := append([]interface{}{arg0, arg1}, arg2...)
	return _mr.mock.ctrl.RecordCall(_mr.mock, "ListCommit", _s...)
}

func (_m *MockAPIClient) ListFile(_param0 context.Context, _param1 *pfs.ListFileRequest, _param2 ...grpc.CallOption) (*pfs.FileInfos, error) {
	_s := []interface{}{_param0, _param1}
	for _, _x := range _param2 {
		_s = append(_s, _x)
	}
	ret := _m.ctrl.Call(_m, "ListFile", _s...)
	ret0, _ := ret[0].(*pfs.FileInfos)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockAPIClientRecorder) ListFile(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	_s := append([]interface{}{arg0, arg1}, arg2...)
	return _mr.mock.ctrl.RecordCall(_mr.mock, "ListFile", _s...)
}

func (_m *MockAPIClient) ListRepo(_param0 context.Context, _param1 *pfs.ListRepoRequest, _param2 ...grpc.CallOption) (*pfs.RepoInfos, error) {
	_s := []interface{}{_param0, _param1}
	for _, _x := range _param2 {
		_s = append(_s, _x)
	}
	ret := _m.ctrl.Call(_m, "ListRepo", _s...)
	ret0, _ := ret[0].(*pfs.RepoInfos)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockAPIClientRecorder) ListRepo(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	_s := append([]interface{}{arg0, arg1}, arg2...)
	return _mr.mock.ctrl.RecordCall(_mr.mock, "ListRepo", _s...)
}

func (_m *MockAPIClient) PutFile(_param0 context.Context, _param1 ...grpc.CallOption) (pfs.API_PutFileClient, error) {
	_s := []interface{}{_param0}
	for _, _x := range _param1 {
		_s = append(_s, _x)
	}
	ret := _m.ctrl.Call(_m, "PutFile", _s...)
	ret0, _ := ret[0].(pfs.API_PutFileClient)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockAPIClientRecorder) PutFile(arg0 interface{}, arg1 ...interface{}) *gomock.Call {
	_s := append([]interface{}{arg0}, arg1...)
	return _mr.mock.ctrl.RecordCall(_mr.mock, "PutFile", _s...)
}

func (_m *MockAPIClient) StartCommit(_param0 context.Context, _param1 *pfs.StartCommitRequest, _param2 ...grpc.CallOption) (*pfs.Commit, error) {
	_s := []interface{}{_param0, _param1}
	for _, _x := range _param2 {
		_s = append(_s, _x)
	}
	ret := _m.ctrl.Call(_m, "StartCommit", _s...)
	ret0, _ := ret[0].(*pfs.Commit)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockAPIClientRecorder) StartCommit(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	_s := append([]interface{}{arg0, arg1}, arg2...)
	return _mr.mock.ctrl.RecordCall(_mr.mock, "StartCommit", _s...)
}