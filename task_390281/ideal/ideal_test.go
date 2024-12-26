// errorreporter/error_reporter.go

package errorreporter

import (
	"github.com/getsentry/sentry-go"
	"log"
	"time"
)

// Init initializes Sentry with the provided DSN
func Init(dsn string) error {
	err := sentry.Init(sentry.ClientOptions{
		Dsn: dsn,
	})
	if err != nil {
		return err
	}
	log.Println("Sentry initialized successfully.")
	return nil
}

// Flush flushes any buffered events to Sentry
func Flush(timeout time.Duration) {
	sentry.Flush(timeout)
}

// CaptureError sends an error to Sentry with additional context
func CaptureError(err error, context map[string]interface{}) {
	if err == nil {
		return
	}
	sentry.WithScope(func(scope *sentry.Scope) {
		for key, value := range context {
			scope.SetExtra(key, value)
		}
		sentry.CaptureException(err)
	})
	log.Printf("Error captured: %v\n", err)
}
