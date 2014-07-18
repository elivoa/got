package exception

import (
	"fmt"
)

// TODO: common error.

type CoreError struct {
	Message string
	Reason  string
}

func (e *CoreError) Error() string { return e.Message }

func NewCoreError(message string) *CoreError { return &CoreError{Message: message} }
func NewCoreErrorf(message string, v ...interface{}) *CoreError {
	return &CoreError{Message: fmt.Sprintf(message, v)}
}

// --------------------------------------------------------------------------------
type CoercionError struct {
	CoreError
}

func NewCoercionError(message string) *CoercionError {
	return &CoercionError{CoreError{Message: message}}
}

func NewCoercionErrorf(message string, v ...interface{}) *CoercionError {
	return &CoercionError{CoreError{Message: fmt.Sprintf(message, v)}}
}
