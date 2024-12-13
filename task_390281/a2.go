// main.go
package main

import (
	"errors"
	"log"
	"net/http"

	"yourmodule/errorreporting" // Adjust the import path to your project
)

func main() {
	// Initialize Sentry
	err := errorreporting.Init("YOUR_SENTRY_DSN")
	if err != nil {
		log.Fatalf("sentry.Init: %s", err)
	}
	defer sentry.Flush(2 * 1000)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if err := riskyOperation(); err != nil {
			context := map[string]interface{}{
				"user_id": r.Header.Get("X-User-ID"), // Assume we're getting the User ID from the header
				"request": r.URL.Path,
				"method":  r.Method,
			}
			errorreporting.CaptureError(err, context)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		w.Write([]byte("Success!"))
	})

	http.ListenAndServe(":8080", nil)
}

func riskyOperation() error {
	// Simulating a network error
	return &errorreporting.NetworkError{msg: "Failed to reach the endpoint"}
}