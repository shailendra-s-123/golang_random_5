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

// Logger interface for dependency injection
type Logger interface {
	LogError(err error, context string)
}

type SimpleLogger struct{}

func (l *SimpleLogger) LogError(err error, context string) {
	log.Printf("Error: %s | Context: %s", err, context)
}

// Retry function with exponential backoff
func retry(ctx context.Context, attempts int, backoff time.Duration, f func(context.Context) error) error {
	for attempt := 1; attempt <= attempts; attempt++ {
		err := f(ctx)
		if err == nil {
			return nil
		}
		log.Printf("Attempt %d failed: %v, retrying...", attempt, err)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoff * time.ExponentialBackoff(1)):
		}
	}
	return fmt.Errorf("max attempts reached: %d", attempts)
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
	return "", nil
}

// Service layer
type UserService struct {
	repo   *UserRepository
	logger Logger
	maxAttempts int
	backoff    time.Duration
}

func NewUserService(repo *UserRepository, logger Logger, maxAttempts int, backoff time.Duration) *UserService {
	return &UserService{repo: repo, logger: logger, maxAttempts: maxAttempts, backoff: backoff}
}

func (s *UserService) GetUserByID(ctx context.Context, id string) (string, error) {
	user, err := retry(
		ctx,
		s.maxAttempts,
		s.backoff,
		func(ctx context.Context) error {
			return s.repo.FindUserByID(ctx, id)
		},
	)
	if err != nil {
		s.logger.LogError(err, "Service.GetUserByID")
		return "", err
	}
	return user, nil
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