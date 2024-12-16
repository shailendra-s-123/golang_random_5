package main

import (
	"context"
	"fmt"
	"net/http"
	"log"

	"github.com/gorilla/mux"

	"github.com/your-org/your-app/errors"
	"github.com/your-org/your-app/handlers"
	"github.com/your-org/your-app/middleware"
	"github.com/your-org/your-app/service"
)

func main() {
	r := mux.NewRouter()

	// Create service with injected error handler
	svc := service.NewConcreteService(handlers.DefaultErrorHandler)

	// HTTP handler for demonstration
	r.HandleFunc("/do", middleware.ErrorMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()
		if err := svc.DoSomething(ctx); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintln(w, "Operation successful.")
	}), handlers.DefaultErrorHandler)).Methods("GET")

	log.Fatal(http.ListenAndServe(":8080", r))
}