package main

import (
    "errors"

    "./errors"
    "./service"
)

func main() {
    // Custom error handler that could include logging, notification, etc.
    customErrorHandler := func(err error) {
        if dbErr, ok := err.(*errors.DatabaseError); ok {
            log.Printf("Critical Database Error: %v\n", dbErr.Cause())
        } else {
            errors.DefaultErrorHandler(err)
        }
    }

    // Create service with custom error handler
    svc := service.NewConcreteService(customErrorHandler)

    if err := svc.DoSomething(); err != nil {
        return
    }

    log.Println("Operation successful.")
}