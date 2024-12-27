package main

import (
	"errors"
	"log"

	"github.com/sirupsen/logrus"
	"error-handling/customerror" // Replace with your module path
)

func init() {
	// Configure logrus for JSON output
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetLevel(logrus.DebugLevel)
}

// riskyOperation simulates an operation that may fail.
func riskyOperation() error {
	return errors.New("unexpected failure in database connection")
}

// performTask wraps errors and uses the custom error type.
func performTask() error {
	context := map[string]interface{}{
		"function": "performTask",
		"input":    "some data",
	}

	err := riskyOperation()
	if err != nil {
		// Wrap the error using the custom error package with context
		return customerror.New("failed to execute task", 1001, err, context)
	}
	return nil
}

func main() {
	// Attempt to perform a task and handle errors with structured logging
	err := performTask()
	if err != nil {
		// Log the custom error using logrus
		logrus.Error(err)

		// If needed, unwrap the original error
		if unwrappedErr := errors.Unwrap(err); unwrappedErr != nil {
			logrus.Errorf("Original error: %v", unwrappedErr)
		}

		return
	}

	logrus.Info("Task completed successfully.")
}