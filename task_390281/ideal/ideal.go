// main.go


package main

import (
	"errors"
	"log"
	"net/http"
	"nain/errorreporter"
)

func main() {
	// Initialize Sentry with your DSN
	err := errorreporter.Init("https://examplePublicKey@o123456.ingest.sentry.io/987654")
	if err != nil {
		log.Fatalf("Failed to initialize Sentry: %v", err)
	}
	defer errorreporter.Flush(2 * time.Second)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Simulate an error
		err := errors.New("an example error occurred")
		context := map[string]interface{}{
			"user_id": "12345",
			"path":    r.URL.Path,
		}
		errorreporter.CaptureError(err, context)

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error captured and reported."))
	})

	log.Println("Server is running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
