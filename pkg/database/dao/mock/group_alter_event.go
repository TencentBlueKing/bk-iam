// Code generated by MockGen. DO NOT EDIT.
// Source: group_alter_event.go

// Package mock is a generated GoMock package.
package mock

import (
	dao "iam/pkg/database/dao"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	sqlx "github.com/jmoiron/sqlx"
)

// MockGroupAlterEventManager is a mock of GroupAlterEventManager interface.
type MockGroupAlterEventManager struct {
	ctrl     *gomock.Controller
	recorder *MockGroupAlterEventManagerMockRecorder
}

// MockGroupAlterEventManagerMockRecorder is the mock recorder for MockGroupAlterEventManager.
type MockGroupAlterEventManagerMockRecorder struct {
	mock *MockGroupAlterEventManager
}

// NewMockGroupAlterEventManager creates a new mock instance.
func NewMockGroupAlterEventManager(ctrl *gomock.Controller) *MockGroupAlterEventManager {
	mock := &MockGroupAlterEventManager{ctrl: ctrl}
	mock.recorder = &MockGroupAlterEventManagerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockGroupAlterEventManager) EXPECT() *MockGroupAlterEventManagerMockRecorder {
	return m.recorder
}

// BulkCreateWithTx mocks base method.
func (m *MockGroupAlterEventManager) BulkCreateWithTx(tx *sqlx.Tx, groupAlterEvents []dao.GroupAlterEvent) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BulkCreateWithTx", tx, groupAlterEvents)
	ret0, _ := ret[0].(error)
	return ret0
}

// BulkCreateWithTx indicates an expected call of BulkCreateWithTx.
func (mr *MockGroupAlterEventManagerMockRecorder) BulkCreateWithTx(tx, groupAlterEvents interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BulkCreateWithTx", reflect.TypeOf((*MockGroupAlterEventManager)(nil).BulkCreateWithTx), tx, groupAlterEvents)
}

// Create mocks base method.
func (m *MockGroupAlterEventManager) Create(groupAlterEvent dao.GroupAlterEvent) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Create", groupAlterEvent)
	ret0, _ := ret[0].(error)
	return ret0
}

// Create indicates an expected call of Create.
func (mr *MockGroupAlterEventManagerMockRecorder) Create(groupAlterEvent interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockGroupAlterEventManager)(nil).Create), groupAlterEvent)
}

// DeleteWithTx mocks base method.
func (m *MockGroupAlterEventManager) DeleteWithTx(tx *sqlx.Tx, pk int64) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteWithTx", tx, pk)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteWithTx indicates an expected call of DeleteWithTx.
func (mr *MockGroupAlterEventManagerMockRecorder) DeleteWithTx(tx, pk interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteWithTx", reflect.TypeOf((*MockGroupAlterEventManager)(nil).DeleteWithTx), tx, pk)
}

// Get mocks base method.
func (m *MockGroupAlterEventManager) Get(pk int64) (dao.GroupAlterEvent, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", pk)
	ret0, _ := ret[0].(dao.GroupAlterEvent)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get.
func (mr *MockGroupAlterEventManagerMockRecorder) Get(pk interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockGroupAlterEventManager)(nil).Get), pk)
}

// ListByGroupStatus mocks base method.
func (m *MockGroupAlterEventManager) ListByGroupStatus(groupPK, status int64) ([]dao.GroupAlterEvent, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListByGroupStatus", groupPK, status)
	ret0, _ := ret[0].([]dao.GroupAlterEvent)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListByGroupStatus indicates an expected call of ListByGroupStatus.
func (mr *MockGroupAlterEventManagerMockRecorder) ListByGroupStatus(groupPK, status interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListByGroupStatus", reflect.TypeOf((*MockGroupAlterEventManager)(nil).ListByGroupStatus), groupPK, status)
}

// UpdateStatus mocks base method.
func (m *MockGroupAlterEventManager) UpdateStatus(pk, fromStatus, toStatus int64) (int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateStatus", pk, fromStatus, toStatus)
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpdateStatus indicates an expected call of UpdateStatus.
func (mr *MockGroupAlterEventManagerMockRecorder) UpdateStatus(pk, fromStatus, toStatus interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateStatus", reflect.TypeOf((*MockGroupAlterEventManager)(nil).UpdateStatus), pk, fromStatus, toStatus)
}