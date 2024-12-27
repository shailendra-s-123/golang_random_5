package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/yourusername/customerrors" // Replace with your module path
)

func someFunction(ctx context.Context) error {
	// Simulate an error
	return errors.New("some internal error")
}

func anotherFunction(ctx context.Context) error {
	// Call a function that might return an error
	err := someFunction(ctx)
	if err != nil {
		return customerrors.Wrap("failed in anotherFunction", err)
	}
	return nil
}

func main() {
	ctx := context.Background()

	// Attempt to perform an operation
	err := anotherFunction(ctx)
	if err != nil {
		// Handle the error
		if customerrors.Is(err, "failed in anotherFunction") {
			log.Fatalf("Specific error occurred: %v", err)
		} else {
			log.Fatalf("An unexpected error occurred: %v", err)
		}
	}

	fmt.Println("Operation completed successfully.")
}