// errorreporter/error_reporter.go
package errorreporter

import (
	"log"

	"github.com/getsentry/sentry-go"
)

// Initialize the Sentry SDK
func Init(dsn string) error {
	return sentry.Init(sentry.ClientOptions{
		Dsn: dsn,
	})
}

// CaptureError captures and sends the error to Sentry
func CaptureError(err error, context map[string]interface{}) {
	if err == nil {
		return
	}
	
	// Attach additional context
	sentry.WithScope(func(scope *sentry.Scope) {
		for key, value := range context {
			scope.SetExtra(key, value)
		}
		sentry.CaptureException(err)
	})
}