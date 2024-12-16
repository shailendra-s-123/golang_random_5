package main

import (
    "fmt"
    "log"
    "net/http"
)

// Custom error types
var (
    ErrNotFound       = fmt.Errorf("resource not found")
    ErrInvalidInput   = fmt.Errorf("invalid input")
    ErrInternalServer = fmt.Errorf("internal server error")
)

// AppError is a type for application-specific errors.
type AppError struct {
    Err       error
    Message   string
    StatusCode int
}

// Error implements the error interface.
func (e *AppError) Error() string {
    return e.Err.Error()
}

// Handler function
type Handler interface {
    ServeHTTP(w http.ResponseWriter, r *http.Request)
}

// Dependency Injection for logging
type Logger interface {
    LogError(err error)
}

// Struct for application
type App struct {
    Logger Logger
}

func (a *App) HandleRequest(w http.ResponseWriter, r *http.Request) {
    // Simulate some business logic
    if r.URL.Path != "/item" {
        a.Logger.LogError(ErrNotFound)
        http.Error(w, "Not Found", http.StatusNotFound)
        return
    }
    // Simulating potential error
    if r.Method != http.MethodGet {
        appErr := &AppError{
            Err: ErrInvalidInput,
            Message: "Only GET method is allowed",
            StatusCode: http.StatusBadRequest,
        }
        a.Logger.LogError(appErr)
        http.Error(w, appErr.Message, appErr.StatusCode)
        return
    }
    fmt.Fprintln(w, "Item found")
}

// SimpleLogger implementation for logging
type SimpleLogger struct {
    *log.Logger
}

func (l *SimpleLogger) LogError(err error) {
    l.Printf("Error occurred: %v\n", err)
}

func main() {
    logger := &SimpleLogger{log.New(log.Writer(), "LOG: ", log.Lshortfile)}
    app := &App{
        Logger: logger,
    }
    
    // Setting up the HTTP server with our application
    http.HandleFunc("/", app.HandleRequest)
    log.Fatal(http.ListenAndServe(":8080", nil))
}