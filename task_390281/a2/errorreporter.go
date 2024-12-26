package errorreporter

import (
	"github.com/getsentry/sentry-go"
	"log"
	"time"
)

// Reporter is the structure for error reporting
type Reporter struct {
	dsn string
}

// Init initializes the error reporter with a given DSN.
func (r *Reporter) Init(dsn string) error {
	r.dsn = dsn
	if err := sentry.Init(sentry.ClientOptions{
		Dsn: r.dsn,
	}); err != nil {
		return err
	}
	log.Println("Sentry initialized successfully.")
	return nil
}

// Flush sends any buffered events to the monitoring service.
func (r *Reporter) Flush(timeout time.Duration) {
	sentry.Flush(timeout)
}

// Capture categorizes and captures an error with metadata.
func (r *Reporter) Capture(err error, context map[string]interface{}, category string) {
	if err == nil {
		return
	}

	sentry.WithScope(func(scope *sentry.Scope) {
		for key, value := range context {
			scope.SetExtra(key, value)
		}

		// Add the error category as a tag
		scope.SetTag("category", category)

		sentry.CaptureException(err)
	})
	log.Printf("Error captured (%s): %v\n", category, err)
}

// Close cleans up the reporter on application exit.
func (r *Reporter) Close() {
	r.Flush(2 * time.Second)
}