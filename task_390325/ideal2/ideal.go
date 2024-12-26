package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"
)

// Define the AppError interface with additional methods
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

// TransientServerError for server-side transient issues
type TransientServerError struct {
	msg  string
	code int
}

func (e *TransientServerError) Error() string {
	return e.msg
}

func (e *TransientServerError) Code() int {
	return e.code
}

func (e *TransientServerError) IsClientError() bool {
	return false
}

func (e *TransientServerError) IsTransient() bool {
	return true
}

// Logger interface for centralized logging
type Logger interface {
	LogError(err error, context string)
}

type SimpleLogger struct{}

func (l *SimpleLogger) LogError(err error, context string) {
	log.Printf("Error: %s | Context: %s", err, context)
}

// Retry function with exponential backoff
func retry(ctx context.Context, attempts int, backoff time.Duration, f func(context.Context) (string, error)) (string, error) {
	var result string
	var err error
	for attempt := 1; attempt <= attempts; attempt++ {
		result, err = f(ctx)
		if err == nil {
			return result, nil
		}

		// Check if the error is transient
		if appErr, ok := err.(AppError); ok && appErr.IsTransient() {
			log.Printf("Attempt %d failed with transient error: %v. Retrying...", attempt, err)
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(backoff * time.Duration(attempt)): // Exponential backoff
			}
			continue
		}
		// Non-transient errors are returned immediately
		return "", err
	}
	return "", fmt.Errorf("max retry attempts reached: %d, last error: %w", attempts, err)
}

// Repository layer
type UserRepository struct {
	logger Logger
}

func (r *UserRepository) FindUserByID(ctx context.Context, id string) (string, error) {
	if id == "" {
		return "", &ClientError{"user ID is required", http.StatusBadRequest}
	}
	// Simulate a transient database error
	if rand.Intn(2) == 0 {
		return "", &TransientServerError{"database connection failed transiently", http.StatusInternalServerError}
	}
	// Simulate successful response
	return "User Name", nil
}

// Service layer
type UserService struct {
	repo        *UserRepository
	logger      Logger
	maxAttempts int
	backoff     time.Duration
}

func NewUserService(repo *UserRepository, logger Logger, maxAttempts int, backoff time.Duration) *UserService {
	return &UserService{repo: repo, logger: logger, maxAttempts: maxAttempts, backoff: backoff}
}

func (s *UserService) GetUserByID(ctx context.Context, id string) (string, error) {
	return retry(ctx, s.maxAttempts, s.backoff, func(ctx context.Context) (string, error) {
		return s.repo.FindUserByID(ctx, id)
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

// Middleware for error handling
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
	// Seed random number generator
	rand.Seed(time.Now().UnixNano())

	// Initialize dependencies
	logger := &SimpleLogger{}
	repo := &UserRepository{logger: logger}
	service := NewUserService(repo, logger, 3, 100*time.Millisecond)

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