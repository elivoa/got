package exception

import (
	"fmt"
	"github.com/elivoa/got/config"
	"github.com/elivoa/got/debug"
)

// TODO: common error.

type CoreError struct {
	Message  string
	Reason   string
	innererr error  // embed error
	stack    []byte // embed error's stack trace
}

func (e *CoreError) Error() string     { return e.Message }
func (e *CoreError) InnerError() error { return e.innererr }
func (e *CoreError) Stack() []byte     { return e.stack }

// Create Core Error, Usages:
//   NewCoreError(nil, "some error")   // message: "some error"
//   NewCoreError(nil, "ERR: %s", "A") // message: "ERR: A"; formated message.
//   NewCoreError(innerErr, nil)       // use the same message as innerErr.
//   NewCoreError(innerErr, "msg")     // the same as no innerErr
//
func NewCoreError(innerErr error, message string, v ...interface{}) *CoreError {
	var msg string
	if nil != v && len(v) > 0 {
		msg = fmt.Sprintf(msg, v...)
	} else if message == "" {
		msg = innerErr.Error()
	} else {
		msg = message
	}

	e := &CoreError{
		Message:  msg,
		innererr: innerErr,
	}
	if config.ProductionMode == false {
		e.stack = debug.Stack()
	}
	return e
}

// --------------------------------------------------------------------------------
type CoercionError struct {
	CoreError
}

func NewCoercionError(message string) *CoercionError {
	return &CoercionError{CoreError{Message: message}}
}

func NewCoercionErrorf(message string, v ...interface{}) *CoercionError {
	return &CoercionError{CoreError{Message: fmt.Sprintf(message, v...)}}
}

// --------------------------------------------------------------------------------

// login error
type PageNotFoundError struct{ CoreError }

func NewPageNotFoundError(message string) *PageNotFoundError {
	return &PageNotFoundError{CoreError{Message: message}}
}

func NewPageNotFoundErrorf(message string, v ...interface{}) *PageNotFoundError {
	return &PageNotFoundError{CoreError{Message: fmt.Sprintf(message, v...)}}
}

// --------------------------------------------------------------------------------
// login error
type AccessDeniedError struct{ CoreError }

func NewAccessDeniedError(message string) *AccessDeniedError {
	return &AccessDeniedError{CoreError{Message: message}}
}

func NewAccessDeniedErrorf(message string, v ...interface{}) *AccessDeniedError {
	return &AccessDeniedError{CoreError{Message: fmt.Sprintf(message, v...)}}
}
