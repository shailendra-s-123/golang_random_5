// errorreporter/error_reporter.go

package errorreporter

import (
	"context"
	"errors"
	"fmt"
	"github.com/getsentry/sentry-go"
	"log"
	"net/http"
	"time"
)

var (
	// ErrNetwork is used to categorize network-related errors
	ErrNetwork = errors.New("network error")
	// ErrLogic is used to categorize application logic errors
	ErrLogic = errors.New("logic error")
)

// Init initializes the Sentry client with the provided DSN
func Init(dsn string) error {
	return sentry.Init(sentry.ClientOptions{
		Dsn:         dsn,
		Release:     "your-app@1.0.0", // Replace with your application version
		Environment: "production",     // Replace with your environment
	})
}

// CaptureError sends an error to Sentry with additional context
func CaptureError(err error, ctx context.Context, metadata map[string]interface{}) {
	if err == nil {
		return
	}

	sentry.WithScope(func(scope *sentry.Scope) {
		// Add metadata
		for key, value := range metadata {
			scope.SetExtra(key, value)
		}

		// Set error type for categorization
		switch {
		case errors.Is(err, ErrNetwork):
			scope.SetTag("error.classification", "network")
		case errors.Is(err, ErrLogic):
			scope.SetTag("error.classification", "logic")
		default:
			scope.SetTag("error.classification", "unknown")
		}

		// Protect sensitive data
		if sensitive, ok := value.(string); ok && sensitive == "REDACTED" {
			scope.SetFingerprint([]string{"$message"})
		}

		// Capture the error
		sentry.CaptureException(err)
	})

	log.Printf("Error captured: %v\n", err)
}

// Flush flushes any buffered events to Sentry
func Flush() {
	sentry.Flush(2 * time.Second)
}