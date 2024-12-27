package customerror

import (
    "fmt"
)

// CustomError is a reusable custom error type with additional fields.
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
