package main

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"error-handling/customerror" // Replace with your module path
)

// riskyOperation simulates an operation that may fail.
func riskyOperation() error {
	return errors.New("unexpected failure in database connection")
}

// performTask wraps errors and uses the custom error type.
func performTask() error {
	err := riskyOperation()
	if err != nil {
		// Wrap the error using the custom error package
		return customerror.New("failed to execute task", 1001, err)
	}
	return nil
}

func main() {
	// Initialize logrus logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter())
	logger.SetLevel(logrus.InfoLevel)

	// Attempt to perform a task and handle errors
	err := performTask()
	if err != nil {
		// Log the custom error using structured logging
		if customErr, ok := err.(*customerror.CustomError); ok {
			customErr.LogError(logger)
		} else {
			logger.WithError(err).Error("An unrecognized error occurred")
		}
		return
	}

	fmt.Println("Task completed successfully.")
}