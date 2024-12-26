

package main

import (
	"errors"
	"fmt"
	"log"
	"myapp/errorreporter" // Replace with actual module path
	"net/http"
)

func main() {
	// Initialize the error reporting service (Sentry)
	reporter := &errorreporter.Reporter{}
	err := reporter.Init("https://examplePublicKey@o123456.ingest.sentry.io/987654") // Use your actual DSN here
	if err != nil {
		log.Fatalf("Failed to initialize Sentry: %v", err)
	}
	defer reporter.Close()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Simulate an application error
		err := errors.New("an example application error")
		context := map[string]interface{}{
			"user_id":   "user_12345",
			"request":   r.URL.Path,
			"client_ip": r.RemoteAddr,
		}

		// Capture the error through the unified error reporter
		reporter.Capture(err, context, errorreporter.CategoryLogic)

		// Respond to the client
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "An error occurred and has been reported.")
	})

	log.Println("Server is running on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		reporter.Capture(err, nil, errorreporter.CategoryNetwork)
		log.Fatalf("Server failed: %v", err)
	}
}




