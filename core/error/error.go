package error

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
