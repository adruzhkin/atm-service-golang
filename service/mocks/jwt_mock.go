// Code generated by MockGen. DO NOT EDIT.
// Source: jwt/jwt.go

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	jwt "github.com/adruzhkin/atm-service-golang/service/jwt"
	gomock "github.com/golang/mock/gomock"
)

// MockJWT is a mock of JWT interface.
type MockJWT struct {
	ctrl     *gomock.Controller
	recorder *MockJWTMockRecorder
}

// MockJWTMockRecorder is the mock recorder for MockJWT.
type MockJWTMockRecorder struct {
	mock *MockJWT
}

// NewMockJWT creates a new mock instance.
func NewMockJWT(ctrl *gomock.Controller) *MockJWT {
	mock := &MockJWT{ctrl: ctrl}
	mock.recorder = &MockJWTMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockJWT) EXPECT() *MockJWTMockRecorder {
	return m.recorder
}

// Generate mocks base method.
func (m *MockJWT) Generate(cus, acc int) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Generate", cus, acc)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Generate indicates an expected call of Generate.
func (mr *MockJWTMockRecorder) Generate(cus, acc interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Generate", reflect.TypeOf((*MockJWT)(nil).Generate), cus, acc)
}

// Verify mocks base method.
func (m *MockJWT) Verify(strToken string) (*jwt.Claims, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Verify", strToken)
	ret0, _ := ret[0].(*jwt.Claims)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Verify indicates an expected call of Verify.
func (mr *MockJWTMockRecorder) Verify(strToken interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Verify", reflect.TypeOf((*MockJWT)(nil).Verify), strToken)
}
