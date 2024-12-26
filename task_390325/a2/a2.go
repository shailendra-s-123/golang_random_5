package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"
)

// Define the Error interface with additional methods
type AppError interface {
	error
	Code() int
	IsClientError() bool
	IsTransient() bool
}

// ClientError for client-side issues
type ClientError struct {
	msg  string
	code int
}

func (e *ClientError) Error() string {
	return e.msg
}

func (e *ClientError) Code() int {
	return e.code
}

func (e *ClientError) IsClientError() bool {
	return true
}

func (e *ClientError) IsTransient() bool {
	return false
}

// ServerError for server-side issues
type ServerError struct {
	msg  string
	code int
}

func (e *ServerError) Error() string {
	return e.msg
}

func (e *ServerError) Code() int {
	return e.code
}

func (e *ServerError) IsClientError() bool {
	return false
}

func (e *ServerError) IsTransient() bool {
	return false
}

// TransientError for temporary issues
type TransientError struct {
	msg  string
	code int
}

func (e *TransientError) Error() string {
	return e.msg
}

func (e *TransientError) Code() int {
	return e.code
}

func (e *TransientError) IsClientError() bool {
	return false
}

func (e *TransientError) IsTransient() bool {
	return true
}

// Logger interface for dependency injection
type Logger interface {
	LogError(err error, context string)
}

type SimpleLogger struct{}

func (l *SimpleLogger) LogError(err error, context string) {
	log.Printf("Error: %s | Context: %s", err, context)
}

// Repository layer
type UserRepository struct {
	logger Logger
}

func (r *UserRepository) FindUserByID(ctx context.Context, id string) (string, error) {
	if id == "" {
		return "", &ClientError{"user ID is required", http.StatusBadRequest}
	}
	// Simulate a database error with a probabilistic approach (e.g., downtime)
	if rand.Float32() < 0.5 {
		// Simulate transient error
		return "", &TransientError{"temporary database connection failed", http.StatusInternalServerError}
	}
	// Simulate a successful database response
	return "User Name", nil
}

// backoffRetries function with exponential backoff
func backoffRetries(maxRetries int, fn func() (string, error)) (string, error) {
	var result string
	var err error
	for i := 0; i < maxRetries; i++ {
		result, err = fn()
		if err == nil {
			return result, nil
		}

		// Check if the error is transient, if so, retry it
		if transientErr, ok := err.(AppError); ok && transientErr.IsTransient() {
			time.Sleep(time.Duration(1<<i) * time.Millisecond * 100) // Exponential backoff
			continue
		}
		return "", err // Non-transient or non-app error, return it immediately
	}
	return "", err // Return last error if retries exhausted
}

// Service layer
type UserService struct {
	repo   *UserRepository
	logger Logger
}

func NewUserService(repo *UserRepository, logger Logger) *UserService {
	return &UserService{repo: repo, logger: logger}
}

func (s *UserService) GetUserByID(ctx context.Context, id string) (string, error) {
	// Wrap the repo call to include retries
	return backoffRetries(3, func() (string, error) {
		return s.repo.FindUserByID(ctx, id) // Attempt to find user
	})
}

// HTTP handler
func userHandler(service *UserService, logger Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		user, err := service.GetUserByID(r.Context(), id)
		if err != nil {
			logger.LogError(err, "Handler.userHandler")
			if appErr, ok := err.(AppError); ok {
				http.Error(w, appErr.Error(), appErr.Code())
				return
			}
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("User: %s", user)))
	}
}

// Middleware for error handling at HTTP layer
func errorHandlingMiddleware(next http.Handler, logger Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				logger.LogError(fmt.Errorf("panic recovered: %v", rec), "Middleware")
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func main() {
	rand.Seed(time.Now().UnixNano()) // Seed random number generator

	// Initialize dependencies
	logger := &SimpleLogger{}
	repo := &UserRepository{logger: logger}
	service := NewUserService(repo, logger)

	// Create HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("/user", userHandler(service, logger))

	// Wrap with middleware
	handler := errorHandlingMiddleware(mux, logger)

	log.Println("Server running on :8080")
	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatal(err)
	}
}