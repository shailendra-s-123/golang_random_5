package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
)

// Custom error types
type ClientError struct {
	msg string
}

func (e *ClientError) Error() string {
	return e.msg
}

type ServerError struct {
	msg string
}

func (e *ServerError) Error() string {
	return e.msg
}

// Logger interface
type Logger interface {
	LogError(err error, context string)
}

// Simple Logger implementation
type SimpleLogger struct{}

func (l *SimpleLogger) LogError(err error, context string) {
	log.Printf("Error: %s, Context: %s", err, context)
}

// Error wrapping function
func wrapError(err error, context string) error {
	return fmt.Errorf("%s: %w", context, err)
}

// Repository layer
type UserRepository struct {
	logger Logger
}

func (r *UserRepository) FindUserByID(ctx context.Context, id string) (string, error) {
	// Simulating a database error
	return "", wrapError(&ServerError{"database connection failed"}, "UserRepository.FindUserByID")
}

// Service layer
type UserService struct {
	repo   *UserRepository
	logger Logger
}

func (s *UserService) GetUserByID(ctx context.Context, id string) (string, error) {
	user, err := s.repo.FindUserByID(ctx, id)
	if err != nil {
		s.logger.LogError(err, "UserService.GetUserByID")
		if errors.As(err, &ServerError{}) {
			return "", &ServerError{"internal server error"}
		}
		return "", &ClientError{"user not found"}
	}
	return user, nil
}

// HTTP Handler
func userHandler(s *UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		user, err := s.GetUserByID(r.Context(), id)
		if err != nil {
			switch err.(type) {
			case *ClientError:
				http.Error(w, err.Error(), http.StatusBadRequest)
			case *ServerError:
				http.Error(w, err.Error(), http.StatusInternalServerError)
			default:
				http.Error(w, "unexpected error", http.StatusInternalServerError)
			}
			return
		}
		w.Write([]byte(user))
	}
}

func main() {
	logger := &SimpleLogger{}
	repo := &UserRepository{logger: logger}
	service := &UserService{repo: repo, logger: logger}

	http.HandleFunc("/user", userHandler(service))
	log.Fatal(http.ListenAndServe(":8080", nil))
}