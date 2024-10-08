// Code generated by MockGen. DO NOT EDIT.
// Source: role.go

// Package mock is a generated GoMock package.
package mock

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockRoleService is a mock of RoleService interface.
type MockRoleService struct {
	ctrl     *gomock.Controller
	recorder *MockRoleServiceMockRecorder
}

// MockRoleServiceMockRecorder is the mock recorder for MockRoleService.
type MockRoleServiceMockRecorder struct {
	mock *MockRoleService
}

// NewMockRoleService creates a new mock instance.
func NewMockRoleService(ctrl *gomock.Controller) *MockRoleService {
	mock := &MockRoleService{ctrl: ctrl}
	mock.recorder = &MockRoleServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockRoleService) EXPECT() *MockRoleServiceMockRecorder {
	return m.recorder
}

// BulkAddSubjects mocks base method.
func (m *MockRoleService) BulkAddSubjects(roleType, system string, subjectPKs []int64) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BulkAddSubjects", roleType, system, subjectPKs)
	ret0, _ := ret[0].(error)
	return ret0
}

// BulkAddSubjects indicates an expected call of BulkAddSubjects.
func (mr *MockRoleServiceMockRecorder) BulkAddSubjects(roleType, system, subjectPKs interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BulkAddSubjects", reflect.TypeOf((*MockRoleService)(nil).BulkAddSubjects), roleType, system, subjectPKs)
}

// BulkDeleteSubjects mocks base method.
func (m *MockRoleService) BulkDeleteSubjects(roleType, system string, subjectPKs []int64) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BulkDeleteSubjects", roleType, system, subjectPKs)
	ret0, _ := ret[0].(error)
	return ret0
}

// BulkDeleteSubjects indicates an expected call of BulkDeleteSubjects.
func (mr *MockRoleServiceMockRecorder) BulkDeleteSubjects(roleType, system, subjectPKs interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BulkDeleteSubjects", reflect.TypeOf((*MockRoleService)(nil).BulkDeleteSubjects), roleType, system, subjectPKs)
}

// ListSystemIDBySubjectPK mocks base method.
func (m *MockRoleService) ListSystemIDBySubjectPK(pk int64) ([]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListSystemIDBySubjectPK", pk)
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListSystemIDBySubjectPK indicates an expected call of ListSystemIDBySubjectPK.
func (mr *MockRoleServiceMockRecorder) ListSystemIDBySubjectPK(pk interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListSystemIDBySubjectPK", reflect.TypeOf((*MockRoleService)(nil).ListSystemIDBySubjectPK), pk)
}
