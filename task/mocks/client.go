// Code generated by mockery v2.35.2. DO NOT EDIT.

package mocks

import (
	http "net/http"

	mock "github.com/stretchr/testify/mock"
)

// Client is an autogenerated mock type for the Client type
type Client struct {
	mock.Mock
}

type Client_Expecter struct {
	mock *mock.Mock
}

func (_m *Client) EXPECT() *Client_Expecter {
	return &Client_Expecter{mock: &_m.Mock}
}

// Send provides a mock function with given fields: r, respPayload
func (_m *Client) Send(r *http.Request, respPayload interface{}) error {
	ret := _m.Called(r, respPayload)

	var r0 error
	if rf, ok := ret.Get(0).(func(*http.Request, interface{}) error); ok {
		r0 = rf(r, respPayload)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Client_Send_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Send'
type Client_Send_Call struct {
	*mock.Call
}

// Send is a helper method to define mock.On call
//   - r *http.Request
//   - respPayload interface{}
func (_e *Client_Expecter) Send(r interface{}, respPayload interface{}) *Client_Send_Call {
	return &Client_Send_Call{Call: _e.mock.On("Send", r, respPayload)}
}

func (_c *Client_Send_Call) Run(run func(r *http.Request, respPayload interface{})) *Client_Send_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*http.Request), args[1].(interface{}))
	})
	return _c
}

func (_c *Client_Send_Call) Return(_a0 error) *Client_Send_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Client_Send_Call) RunAndReturn(run func(*http.Request, interface{}) error) *Client_Send_Call {
	_c.Call.Return(run)
	return _c
}

// NewClient creates a new instance of Client. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewClient(t interface {
	mock.TestingT
	Cleanup(func())
}) *Client {
	mock := &Client{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
