package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/getsentry/sentry-go"
	"github.com/getsentry/sentry-go/integrations/stdlog"
)

// CustomError represents a unified error struct
type CustomError struct {
	message string
	code    int
	details map[string]interface{}
}

func (e CustomError) Error() string {
	return fmt.Sprintf("Error %d: %s", e.code, e.message)
}

// InitializeSentry initializes Sentry with DSN and standard log integration
func InitializeSentry(dsn string) {
	err := sentry.Init(sentry.ClientOptions{
		Dsn:             dsn,
		Integrations:    []sentry.Integration{stdlog.New()},
		Release:         "your-app@1.0.0", // set to your app version
		BeforeSend:      sanitizeBeforeSend,
	})
	if err != nil {
		log.Fatalf("Could not initialize Sentry: %v", err)
	}
}

// sanitizeBeforeSend removes sensitive information from error events
func sanitizeBeforeSend(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
	// Example: Remove sensitive fields from the "extra" section
	if extra, ok := event.Extra["sensitive_data"].(map[string]interface{}); ok {
		delete(extra, "password")
		delete(extra, "credit_card")
		event.Extra["sensitive_data"] = extra
	}
	return event
}

// HandleHTTPError handles errors within HTTP handlers
func HandleHTTPError(w http.ResponseWriter, r *http.Request, err error, code int) {
	// Report error to Sentry
	ctx := sentry.WithScope(sentry.NewScope().
		SetTag("http.method", r.Method).
		SetTag("http.path", r.URL.Path))

	sentry.CaptureException(err, ctx)

	// Send HTTP response
	http.Error(w, err.Error(), code)
}

// Application Logic Error Example
func someApplicationFunction(ctx context.Context) error {
	if someCondition {
		return CustomError{message: "Application logic error", code: 500, details: map[string]interface{}{"input": "invalid data"}}
	}
	return nil
}

// Network Error Example
func someNetworkFunction(ctx context.Context) error {
	return fmt.Errorf("network error: %w", http.ErrClientConnLost)
}

func main() {
	InitializeSentry("your-sentry-dsn")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Simulate an application logic error
		err1 := someApplicationFunction(ctx)
		if err1 != nil {
			HandleHTTPError(w, r, err1, http.StatusInternalServerError)
			return
		}

		// Simulate a network error
		err2 := someNetworkFunction(ctx)
		if err2 != nil {
			HandleHTTPError(w, r, err2, http.StatusServiceUnavailable)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "Success")
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}