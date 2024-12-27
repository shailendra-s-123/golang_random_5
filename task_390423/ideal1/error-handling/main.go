package main

import (
    "errors"
    "fmt"
    "log"

    "error-handling/customerror"
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
    // Attempt to perform a task and handle errors
    err := performTask()
    if err != nil {
        // Log the custom error
        log.Printf("Error occurred: %v", err)

        // If needed, unwrap the original error
        if unwrappedErr := errors.Unwrap(err); unwrappedErr != nil {
            log.Printf("Original error: %v", unwrappedErr)
        }

        return
    }

    fmt.Println("Task completed successfully.")
}
