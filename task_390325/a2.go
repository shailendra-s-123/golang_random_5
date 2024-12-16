// main.go
package main

import (
	"log"
	"net/http"
	"businesslogic"
	"customerrors"
	"errorhandler"
	"middleware"
)

func main() {
	// Set up a default error handler.
	defaultHandler := &errorhandler.DefaultHandler{}

	// Create business logic service with error handler.
	service := businesslogic.NewService(defaultHandler)

	// Set up the HTTP server with middleware.
	http.Handle("/do", &middleware.ErrorMiddleware{Handler: defaultHandler})

	http.HandleFunc("/do", func(w http.ResponseWriter, r *http.Request) {
		if err := service.DoAction(); err != nil {
			defaultHandler.Handle(err)
			http.Error(w, "An error occurred during processing.", http.StatusInternalServerError)
		} else {
			w.Write([]byte("Success!"))
		}
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}