package customerror

import (
	"errors"
	"fmt"
)

// CustomError is a reusable custom error type with additional fields for structured logging.
type CustomError struct {
	Message string     `json:"message"`
	Code    int        `json:"code"`
	Err     error      `json:"cause,omitempty"`
	Context map[string]interface{} `json:"context,omitempty"`
}

// New creates a new CustomError.
func New(message string, code int, err error, context map[string]interface{}) *CustomError {
	return &CustomError{
		Message: message,
		Code:    code,
		Err:     err,
		Context: context,
	}
}

// Error implements the error interface for CustomError.
func (e *CustomError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("Code: %d, Message: %s, Cause: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("Code: %d, Message: %s", e.Code, e.Message)
}

// Unwrap retrieves the wrapped error.
func (e *CustomError) Unwrap() error {
	return e.Err
}