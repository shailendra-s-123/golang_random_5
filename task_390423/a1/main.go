// main.go
package main

import (
	"fmt"
	"log"
	"error-handling-example/errors"
)

// Simulated function that returns an error
func riskyOperation() error {
	return fmt.Errorf("an unexpected error occurred")
}

func main() {
	if err := doWork(); err != nil {
		log.Fatalf("Application encountered an error: %v", err)
	}
}

func doWork() error {
	// Simulate a risky operation
	err := riskyOperation()
	if err != nil {
		// Wrap the underlying error with custom information
		return errors.New("failed to complete a risky operation", 1001, err)
	}

	fmt.Println("Work completed successfully!")
	return nil
}