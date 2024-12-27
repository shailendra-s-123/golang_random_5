// errors/errors.go
package errors

import (
	"fmt"
)

// CustomError is a type that includes error details.
type CustomError struct {
	Message string
	Code    int
	Err     error
}

// New creates a new CustomError.
func New(message string, code int, err error) *CustomError {
	return &CustomError{
		Message: message,
		Code:    code,
		Err:     err,
	}
}

// Error implements the error interface.
func (e *CustomError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("Code: %d, Message: %s, Original Error: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("Code: %d, Message: %s", e.Code, e.Message)
}

// Unwrap retrieves the original error.
func (e *CustomError) Unwrap() error {
	return e.Err
}