// errorreporting/error_reporting.go
package errorreporting

import (
	"fmt"
	"log"

	"github.com/getsentry/sentry-go"
)

// Custom error types
type LogicError struct {
	msg string
}

func (e *LogicError) Error() string { return e.msg }

type NetworkError struct {
	msg string
}

func (e *NetworkError) Error() string { return e.msg }

// Initialize Sentry
func Init(dsn string) error {
	return sentry.Init(sentry.ClientOptions{
		Dsn: dsn,
	})
}

// CaptureError captures and sends an error to Sentry with context
func CaptureError(err error, context map[string]interface{}) {
	if err == nil {
		return
	}

	// Redact sensitive data from context if necessary
	if userID, ok := context["user_id"]; ok {
		if IsSensitive(userID) { // Implement IsSensitive to check and redact
			context["user_id"] = "[REDACTED]"
		}
	}

	// Attach additional context
	sentry.WithScope(func(scope *sentry.Scope) {
		for key, value := range context {
			scope.SetExtra(key, value)
		}
		sentry.CaptureException(err)
	})
}

// Example function to check if a value is sensitive
func IsSensitive(value interface{}) bool {
	// Implement logic to determine if the value is sensitive
	// For example:
	if _, ok := value.(string); ok && value.(string) != "" {
		return true // Implement more sophisticated checks as needed
	}
	return false
}