// Code generated by mockery v2.51.0. DO NOT EDIT.

package mocks

import (
	context "context"

	entity "github.com/nathakusuma/astungkara/domain/entity"

	mock "github.com/stretchr/testify/mock"

	uuid "github.com/google/uuid"
)

// MockIAuthRepository is an autogenerated mock type for the IAuthRepository type
type MockIAuthRepository struct {
	mock.Mock
}

type MockIAuthRepository_Expecter struct {
	mock *mock.Mock
}

func (_m *MockIAuthRepository) EXPECT() *MockIAuthRepository_Expecter {
	return &MockIAuthRepository_Expecter{mock: &_m.Mock}
}

// CreateSession provides a mock function with given fields: ctx, session
func (_m *MockIAuthRepository) CreateSession(ctx context.Context, session *entity.Session) error {
	ret := _m.Called(ctx, session)

	if len(ret) == 0 {
		panic("no return value specified for CreateSession")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *entity.Session) error); ok {
		r0 = rf(ctx, session)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockIAuthRepository_CreateSession_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CreateSession'
type MockIAuthRepository_CreateSession_Call struct {
	*mock.Call
}

// CreateSession is a helper method to define mock.On call
//   - ctx context.Context
//   - session *entity.Session
func (_e *MockIAuthRepository_Expecter) CreateSession(ctx interface{}, session interface{}) *MockIAuthRepository_CreateSession_Call {
	return &MockIAuthRepository_CreateSession_Call{Call: _e.mock.On("CreateSession", ctx, session)}
}

func (_c *MockIAuthRepository_CreateSession_Call) Run(run func(ctx context.Context, session *entity.Session)) *MockIAuthRepository_CreateSession_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*entity.Session))
	})
	return _c
}

func (_c *MockIAuthRepository_CreateSession_Call) Return(_a0 error) *MockIAuthRepository_CreateSession_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockIAuthRepository_CreateSession_Call) RunAndReturn(run func(context.Context, *entity.Session) error) *MockIAuthRepository_CreateSession_Call {
	_c.Call.Return(run)
	return _c
}

// DeleteSession provides a mock function with given fields: ctx, userID
func (_m *MockIAuthRepository) DeleteSession(ctx context.Context, userID uuid.UUID) error {
	ret := _m.Called(ctx, userID)

	if len(ret) == 0 {
		panic("no return value specified for DeleteSession")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) error); ok {
		r0 = rf(ctx, userID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockIAuthRepository_DeleteSession_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DeleteSession'
type MockIAuthRepository_DeleteSession_Call struct {
	*mock.Call
}

// DeleteSession is a helper method to define mock.On call
//   - ctx context.Context
//   - userID uuid.UUID
func (_e *MockIAuthRepository_Expecter) DeleteSession(ctx interface{}, userID interface{}) *MockIAuthRepository_DeleteSession_Call {
	return &MockIAuthRepository_DeleteSession_Call{Call: _e.mock.On("DeleteSession", ctx, userID)}
}

func (_c *MockIAuthRepository_DeleteSession_Call) Run(run func(ctx context.Context, userID uuid.UUID)) *MockIAuthRepository_DeleteSession_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(uuid.UUID))
	})
	return _c
}

func (_c *MockIAuthRepository_DeleteSession_Call) Return(_a0 error) *MockIAuthRepository_DeleteSession_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockIAuthRepository_DeleteSession_Call) RunAndReturn(run func(context.Context, uuid.UUID) error) *MockIAuthRepository_DeleteSession_Call {
	_c.Call.Return(run)
	return _c
}

// DeleteUserRegisterOTP provides a mock function with given fields: ctx, email
func (_m *MockIAuthRepository) DeleteUserRegisterOTP(ctx context.Context, email string) error {
	ret := _m.Called(ctx, email)

	if len(ret) == 0 {
		panic("no return value specified for DeleteUserRegisterOTP")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, email)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockIAuthRepository_DeleteUserRegisterOTP_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DeleteUserRegisterOTP'
type MockIAuthRepository_DeleteUserRegisterOTP_Call struct {
	*mock.Call
}

// DeleteUserRegisterOTP is a helper method to define mock.On call
//   - ctx context.Context
//   - email string
func (_e *MockIAuthRepository_Expecter) DeleteUserRegisterOTP(ctx interface{}, email interface{}) *MockIAuthRepository_DeleteUserRegisterOTP_Call {
	return &MockIAuthRepository_DeleteUserRegisterOTP_Call{Call: _e.mock.On("DeleteUserRegisterOTP", ctx, email)}
}

func (_c *MockIAuthRepository_DeleteUserRegisterOTP_Call) Run(run func(ctx context.Context, email string)) *MockIAuthRepository_DeleteUserRegisterOTP_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *MockIAuthRepository_DeleteUserRegisterOTP_Call) Return(_a0 error) *MockIAuthRepository_DeleteUserRegisterOTP_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockIAuthRepository_DeleteUserRegisterOTP_Call) RunAndReturn(run func(context.Context, string) error) *MockIAuthRepository_DeleteUserRegisterOTP_Call {
	_c.Call.Return(run)
	return _c
}

// GetSessionByToken provides a mock function with given fields: ctx, token
func (_m *MockIAuthRepository) GetSessionByToken(ctx context.Context, token string) (*entity.Session, error) {
	ret := _m.Called(ctx, token)

	if len(ret) == 0 {
		panic("no return value specified for GetSessionByToken")
	}

	var r0 *entity.Session
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*entity.Session, error)); ok {
		return rf(ctx, token)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *entity.Session); ok {
		r0 = rf(ctx, token)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*entity.Session)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, token)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockIAuthRepository_GetSessionByToken_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetSessionByToken'
type MockIAuthRepository_GetSessionByToken_Call struct {
	*mock.Call
}

// GetSessionByToken is a helper method to define mock.On call
//   - ctx context.Context
//   - token string
func (_e *MockIAuthRepository_Expecter) GetSessionByToken(ctx interface{}, token interface{}) *MockIAuthRepository_GetSessionByToken_Call {
	return &MockIAuthRepository_GetSessionByToken_Call{Call: _e.mock.On("GetSessionByToken", ctx, token)}
}

func (_c *MockIAuthRepository_GetSessionByToken_Call) Run(run func(ctx context.Context, token string)) *MockIAuthRepository_GetSessionByToken_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *MockIAuthRepository_GetSessionByToken_Call) Return(_a0 *entity.Session, _a1 error) *MockIAuthRepository_GetSessionByToken_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockIAuthRepository_GetSessionByToken_Call) RunAndReturn(run func(context.Context, string) (*entity.Session, error)) *MockIAuthRepository_GetSessionByToken_Call {
	_c.Call.Return(run)
	return _c
}

// GetUserRegisterOTP provides a mock function with given fields: ctx, email
func (_m *MockIAuthRepository) GetUserRegisterOTP(ctx context.Context, email string) (string, error) {
	ret := _m.Called(ctx, email)

	if len(ret) == 0 {
		panic("no return value specified for GetUserRegisterOTP")
	}

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (string, error)); ok {
		return rf(ctx, email)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) string); ok {
		r0 = rf(ctx, email)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, email)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockIAuthRepository_GetUserRegisterOTP_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetUserRegisterOTP'
type MockIAuthRepository_GetUserRegisterOTP_Call struct {
	*mock.Call
}

// GetUserRegisterOTP is a helper method to define mock.On call
//   - ctx context.Context
//   - email string
func (_e *MockIAuthRepository_Expecter) GetUserRegisterOTP(ctx interface{}, email interface{}) *MockIAuthRepository_GetUserRegisterOTP_Call {
	return &MockIAuthRepository_GetUserRegisterOTP_Call{Call: _e.mock.On("GetUserRegisterOTP", ctx, email)}
}

func (_c *MockIAuthRepository_GetUserRegisterOTP_Call) Run(run func(ctx context.Context, email string)) *MockIAuthRepository_GetUserRegisterOTP_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *MockIAuthRepository_GetUserRegisterOTP_Call) Return(_a0 string, _a1 error) *MockIAuthRepository_GetUserRegisterOTP_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockIAuthRepository_GetUserRegisterOTP_Call) RunAndReturn(run func(context.Context, string) (string, error)) *MockIAuthRepository_GetUserRegisterOTP_Call {
	_c.Call.Return(run)
	return _c
}

// SetUserRegisterOTP provides a mock function with given fields: ctx, email, otp
func (_m *MockIAuthRepository) SetUserRegisterOTP(ctx context.Context, email string, otp string) error {
	ret := _m.Called(ctx, email, otp)

	if len(ret) == 0 {
		panic("no return value specified for SetUserRegisterOTP")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) error); ok {
		r0 = rf(ctx, email, otp)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MockIAuthRepository_SetUserRegisterOTP_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SetUserRegisterOTP'
type MockIAuthRepository_SetUserRegisterOTP_Call struct {
	*mock.Call
}

// SetUserRegisterOTP is a helper method to define mock.On call
//   - ctx context.Context
//   - email string
//   - otp string
func (_e *MockIAuthRepository_Expecter) SetUserRegisterOTP(ctx interface{}, email interface{}, otp interface{}) *MockIAuthRepository_SetUserRegisterOTP_Call {
	return &MockIAuthRepository_SetUserRegisterOTP_Call{Call: _e.mock.On("SetUserRegisterOTP", ctx, email, otp)}
}

func (_c *MockIAuthRepository_SetUserRegisterOTP_Call) Run(run func(ctx context.Context, email string, otp string)) *MockIAuthRepository_SetUserRegisterOTP_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(string))
	})
	return _c
}

func (_c *MockIAuthRepository_SetUserRegisterOTP_Call) Return(_a0 error) *MockIAuthRepository_SetUserRegisterOTP_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockIAuthRepository_SetUserRegisterOTP_Call) RunAndReturn(run func(context.Context, string, string) error) *MockIAuthRepository_SetUserRegisterOTP_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockIAuthRepository creates a new instance of MockIAuthRepository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockIAuthRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockIAuthRepository {
	mock := &MockIAuthRepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
