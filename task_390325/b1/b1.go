package main

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// Error interface
type Error interface {
	error
	Code() int
	IsClientError() bool
}

// ClientError represents a client-related error
type ClientError struct {
	error
	code int
}

func (ce ClientError) Code() int {
	return ce.code
}

func (ce ClientError) IsClientError() bool {
	return true
}

// ServerError represents a server-related error
type ServerError struct {
	error
	code int
}

func (se ServerError) Code() int {
	return se.code
}

func (se ServerError) IsClientError() bool {
	return false
}

// WrapError adds context to an error
func WrapError(err error, msg string) error {
	return errors.New(msg + ": " + err.Error())
}

// Logger interface for centralized logging
type Logger interface {
	Errorf(format string, args ...interface{})
}

type defaultLogger struct {
}

func (d *defaultLogger) Errorf(format string, args ...interface{}) {
	log.Printf("Error: "+format, args...)
}

// HTTPErrorHandler middleware for HTTP responses
func HTTPErrorHandler(next http.Handler, logger Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				panicError := err.(error)
				http.Error(w, panicError.Error(), http.StatusInternalServerError)
				logger.Errorf("Panic caught: %v", panicError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func main() {
	r := mux.NewRouter()
	logger := &defaultLogger{}

	r.HandleFunc("/api/data", GetDataHandler(logger)).Methods("GET")

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}
}

// GetDataHandler is a service layer handler
func GetDataHandler(logger Logger) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		data, err := service.GetData(ctx, "some-key")
		if err != nil {
			httpError := mapHttpError(err, logger)
			http.Error(w, httpError.Error(), httpError.Code())
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(data))
	})
}

func mapHttpError(err error, logger Logger) error {
	if ce, ok := err.(ClientError); ok {
		logger.Errorf("Client Error: %v", err)
		return ce
	} else if se, ok := err.(ServerError); ok {
		logger.Errorf("Server Error: %v", err)
		return se
	} else {
		logger.Errorf("Internal Error: %v", err)
		return ServerError{err, http.StatusInternalServerError}
	}
}

// Service interface
type Service interface {
	GetData(ctx context.Context, key string) (string, error)
}

type defaultService struct {
	repo Repository
}

func NewDefaultService(repo Repository) Service {
	return &defaultService{repo: repo}
}

func (s *defaultService) GetData(ctx context.Context, key string) (string, error) {
	data, err := s.repo.Get(ctx, key)
	if err != nil {
		return "", WrapError(err, "Failed to get data")
	}
	return data, nil
}

// Repository interface
type Repository interface {
	Get(ctx context.Context, key string) (string, error)
}

type defaultRepository struct {
}

func NewDefaultRepository() Repository {
	return &defaultRepository{}
}

func (r *defaultRepository) Get(ctx context.Context, key string) (string, error) {
	if key == "" {
		return "", ClientError{errors.New("Key is required"), http.StatusBadRequest}
	}
	// Simulate a database error
	return "", ServerError{errors.New("Database error"), http.StatusInternalServerError}
}