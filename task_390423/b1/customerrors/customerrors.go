package customerror

import (
	"errors"
	"fmt"
)

// CustomError is a custom error type with an additional Message field.
type CustomError struct {
	error
	Message string `json:"message,omitempty"`
}

// New creates a new CustomError with a specified message.
func New(msg string, err error) *CustomError {
	return &CustomError{
		error:  err,
		Message: msg,
	}
}

// Error implements the error interface, returning a formatted string.
func (ce *CustomError) Error() string {
	return fmt.Sprintf("custom error: %s; underlying: %v", ce.Message, ce.error)
}

// Wrap creates a new CustomError that wraps an existing error with a message.
func Wrap(msg string, err error) error {
	return New(msg, err)
}

// Is checks if an error is of type CustomError and matches the given message.
func Is(err error, msg string) bool {
	if ce, ok := err.(*CustomError); ok {
		return ce.Message == msg
	}
	return false
}