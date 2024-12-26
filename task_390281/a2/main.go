package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"your_module_path/errorreporter" // replace with actual module path
	"time"
)

func main() {
	// Initialize the error reporting service (Sentry)
	reporter := &errorreporter.Reporter{}
	err := reporter.Init("https://examplePublicKey@o123456.ingest.sentry.io/987654")
	if err != nil {
		log.Fatalf("Failed to initialize Sentry: %v", err)
	}
	defer reporter.Close()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Simulate an application error
		err := errors.New("an example application error")
		context := map[string]interface{}{
			"user_id": "user_12345",
			"request":  r.URL.Path,
		}

		// Instead of propagating the error, we log it through the unified reporter
		reporter.Capture(err, context, errorreporter.CategoryLogic)

		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "An error occurred and has been reported.")
	})

	log.Println("Server is running on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		reporter.Capture(err, nil, errorreporter.CategoryNetwork)
		log.Fatalf("Server failed: %v", err)
	}
}