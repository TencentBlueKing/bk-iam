// Code generated by MockGen. DO NOT EDIT.
// Source: group_alter_event.go

// Package mock is a generated GoMock package.
package mock

import (
	types "iam/pkg/service/types"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockGroupAlterEventService is a mock of GroupAlterEventService interface.
type MockGroupAlterEventService struct {
	ctrl     *gomock.Controller
	recorder *MockGroupAlterEventServiceMockRecorder
}

// MockGroupAlterEventServiceMockRecorder is the mock recorder for MockGroupAlterEventService.
type MockGroupAlterEventServiceMockRecorder struct {
	mock *MockGroupAlterEventService
}

// NewMockGroupAlterEventService creates a new mock instance.
func NewMockGroupAlterEventService(ctrl *gomock.Controller) *MockGroupAlterEventService {
	mock := &MockGroupAlterEventService{ctrl: ctrl}
	mock.recorder = &MockGroupAlterEventServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockGroupAlterEventService) EXPECT() *MockGroupAlterEventServiceMockRecorder {
	return m.recorder
}

// CreateByGroupAction mocks base method.
func (m *MockGroupAlterEventService) CreateByGroupAction(groupPK int64, actionPKs []int64) (types.GroupAlterEvent, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateByGroupAction", groupPK, actionPKs)
	ret0, _ := ret[0].(types.GroupAlterEvent)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateByGroupAction indicates an expected call of CreateByGroupAction.
func (mr *MockGroupAlterEventServiceMockRecorder) CreateByGroupAction(groupPK, actionPKs interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateByGroupAction", reflect.TypeOf((*MockGroupAlterEventService)(nil).CreateByGroupAction), groupPK, actionPKs)
}

// CreateByGroupSubject mocks base method.
func (m *MockGroupAlterEventService) CreateByGroupSubject(groupPK int64, subjectPKs []int64) (types.GroupAlterEvent, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateByGroupSubject", groupPK, subjectPKs)
	ret0, _ := ret[0].(types.GroupAlterEvent)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateByGroupSubject indicates an expected call of CreateByGroupSubject.
func (mr *MockGroupAlterEventServiceMockRecorder) CreateByGroupSubject(groupPK, subjectPKs interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateByGroupSubject", reflect.TypeOf((*MockGroupAlterEventService)(nil).CreateByGroupSubject), groupPK, subjectPKs)
}

// ListUncheckedByGroup mocks base method.
func (m *MockGroupAlterEventService) ListUncheckedByGroup(groupPK int64) ([]types.GroupAlterEvent, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListUncheckedByGroup", groupPK)
	ret0, _ := ret[0].([]types.GroupAlterEvent)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListUncheckedByGroup indicates an expected call of ListUncheckedByGroup.
func (mr *MockGroupAlterEventServiceMockRecorder) ListUncheckedByGroup(groupPK interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListUncheckedByGroup", reflect.TypeOf((*MockGroupAlterEventService)(nil).ListUncheckedByGroup), groupPK)
}