// Code generated by MockGen. DO NOT EDIT.
// Source: crypto.go
//
// Generated by this command:
//
//	mockgen -package=cryptocoremocks -source=crypto.go -destination=./mocks/crypto.go
//

// Package cryptocoremocks is a generated GoMock package.
package cryptocoremocks

import (
	reflect "reflect"

	cryptocore "github.com/alvii147/nymphadora-api/pkg/cryptocore"
	gomock "go.uber.org/mock/gomock"
)

// MockCrypto is a mock of Crypto interface.
type MockCrypto struct {
	ctrl     *gomock.Controller
	recorder *MockCryptoMockRecorder
	isgomock struct{}
}

// MockCryptoMockRecorder is the mock recorder for MockCrypto.
type MockCryptoMockRecorder struct {
	mock *MockCrypto
}

// NewMockCrypto creates a new mock instance.
func NewMockCrypto(ctrl *gomock.Controller) *MockCrypto {
	mock := &MockCrypto{ctrl: ctrl}
	mock.recorder = &MockCryptoMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockCrypto) EXPECT() *MockCryptoMockRecorder {
	return m.recorder
}

// CheckPassword mocks base method.
func (m *MockCrypto) CheckPassword(hashedPassword, password string) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CheckPassword", hashedPassword, password)
	ret0, _ := ret[0].(bool)
	return ret0
}

// CheckPassword indicates an expected call of CheckPassword.
func (mr *MockCryptoMockRecorder) CheckPassword(hashedPassword, password any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CheckPassword", reflect.TypeOf((*MockCrypto)(nil).CheckPassword), hashedPassword, password)
}

// CreateAPIKey mocks base method.
func (m *MockCrypto) CreateAPIKey() (string, string, string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateAPIKey")
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(string)
	ret2, _ := ret[2].(string)
	ret3, _ := ret[3].(error)
	return ret0, ret1, ret2, ret3
}

// CreateAPIKey indicates an expected call of CreateAPIKey.
func (mr *MockCryptoMockRecorder) CreateAPIKey() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateAPIKey", reflect.TypeOf((*MockCrypto)(nil).CreateAPIKey))
}

// CreateActivationJWT mocks base method.
func (m *MockCrypto) CreateActivationJWT(userUUID string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateActivationJWT", userUUID)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateActivationJWT indicates an expected call of CreateActivationJWT.
func (mr *MockCryptoMockRecorder) CreateActivationJWT(userUUID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateActivationJWT", reflect.TypeOf((*MockCrypto)(nil).CreateActivationJWT), userUUID)
}

// CreateAuthJWT mocks base method.
func (m *MockCrypto) CreateAuthJWT(userUUID string, tokenType cryptocore.JWTType) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateAuthJWT", userUUID, tokenType)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateAuthJWT indicates an expected call of CreateAuthJWT.
func (mr *MockCryptoMockRecorder) CreateAuthJWT(userUUID, tokenType any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateAuthJWT", reflect.TypeOf((*MockCrypto)(nil).CreateAuthJWT), userUUID, tokenType)
}

// CreateCodeSpaceInvitationJWT mocks base method.
func (m *MockCrypto) CreateCodeSpaceInvitationJWT(userUUID, inviteeEmail string, codeSpaceID int64, accessLevel int) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateCodeSpaceInvitationJWT", userUUID, inviteeEmail, codeSpaceID, accessLevel)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateCodeSpaceInvitationJWT indicates an expected call of CreateCodeSpaceInvitationJWT.
func (mr *MockCryptoMockRecorder) CreateCodeSpaceInvitationJWT(userUUID, inviteeEmail, codeSpaceID, accessLevel any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateCodeSpaceInvitationJWT", reflect.TypeOf((*MockCrypto)(nil).CreateCodeSpaceInvitationJWT), userUUID, inviteeEmail, codeSpaceID, accessLevel)
}

// HashPassword mocks base method.
func (m *MockCrypto) HashPassword(password string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "HashPassword", password)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// HashPassword indicates an expected call of HashPassword.
func (mr *MockCryptoMockRecorder) HashPassword(password any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HashPassword", reflect.TypeOf((*MockCrypto)(nil).HashPassword), password)
}

// ParseAPIKey mocks base method.
func (m *MockCrypto) ParseAPIKey(key string) (string, string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ParseAPIKey", key)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(string)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// ParseAPIKey indicates an expected call of ParseAPIKey.
func (mr *MockCryptoMockRecorder) ParseAPIKey(key any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ParseAPIKey", reflect.TypeOf((*MockCrypto)(nil).ParseAPIKey), key)
}

// ValidateActivationJWT mocks base method.
func (m *MockCrypto) ValidateActivationJWT(token string) (*cryptocore.ActivationJWTClaims, bool) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ValidateActivationJWT", token)
	ret0, _ := ret[0].(*cryptocore.ActivationJWTClaims)
	ret1, _ := ret[1].(bool)
	return ret0, ret1
}

// ValidateActivationJWT indicates an expected call of ValidateActivationJWT.
func (mr *MockCryptoMockRecorder) ValidateActivationJWT(token any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ValidateActivationJWT", reflect.TypeOf((*MockCrypto)(nil).ValidateActivationJWT), token)
}

// ValidateAuthJWT mocks base method.
func (m *MockCrypto) ValidateAuthJWT(token string, tokenType cryptocore.JWTType) (*cryptocore.AuthJWTClaims, bool) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ValidateAuthJWT", token, tokenType)
	ret0, _ := ret[0].(*cryptocore.AuthJWTClaims)
	ret1, _ := ret[1].(bool)
	return ret0, ret1
}

// ValidateAuthJWT indicates an expected call of ValidateAuthJWT.
func (mr *MockCryptoMockRecorder) ValidateAuthJWT(token, tokenType any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ValidateAuthJWT", reflect.TypeOf((*MockCrypto)(nil).ValidateAuthJWT), token, tokenType)
}

// ValidateCodeSpaceInvitationJWT mocks base method.
func (m *MockCrypto) ValidateCodeSpaceInvitationJWT(token string) (*cryptocore.CodeSpaceInvitationJWTClaims, bool) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ValidateCodeSpaceInvitationJWT", token)
	ret0, _ := ret[0].(*cryptocore.CodeSpaceInvitationJWTClaims)
	ret1, _ := ret[1].(bool)
	return ret0, ret1
}

// ValidateCodeSpaceInvitationJWT indicates an expected call of ValidateCodeSpaceInvitationJWT.
func (mr *MockCryptoMockRecorder) ValidateCodeSpaceInvitationJWT(token any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ValidateCodeSpaceInvitationJWT", reflect.TypeOf((*MockCrypto)(nil).ValidateCodeSpaceInvitationJWT), token)
}
