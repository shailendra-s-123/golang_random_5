// main.go
package main

import (
	"errors"
	"log"
	"net/http"

	"github.com/getsentry/sentry-go"
	"yourmodule/errorreporter" // import your error reporting package
)

func main() {
	err := errorreporter.Init("YOUR_SENTRY_DSN_HERE")
	if err != nil {
		log.Fatalf("sentry.Init: %s", err)
	}
	defer sentry.Flush(2 * 1000)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if err := riskyOperation(); err != nil {
			context := map[string]interface{}{
				"user_id": r.Header.Get("X-User-ID"),
				"request": r.URL.Path,
			}
			errorreporter.CaptureError(err, context)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		w.Write([]byte("Success!"))
	})

	http.ListenAndServe(":8080", nil)
}

func riskyOperation() error {
	// Simulate an error
	return errors.New("an unexpected error occurred")
}