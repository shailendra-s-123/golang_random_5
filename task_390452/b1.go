package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/handler"
	"github.com/99designs/gqlgen/model"
	"github.com/99designs/gqlgen/resolver"
)

// Query type for the GraphQL schema
type Query struct {
	Resolver *resolver.Resolver
}

// Mutation type for the GraphQL schema
type Mutation struct {
	Resolver *resolver.Resolver
}

// Resolver is the main resolver struct
type Resolver struct {
	Query     *Query
	Mutation *Mutation
}

func main() {
	// Create a new server
	s := handler.NewDefaultServer(NewExecutableSchema(Config{Resolvers: &Resolver{}}))

	// Serve HTTP requests
	log.Fatal(http.ListenAndServe(":8080", s))
}

// LongRunningQuery simulates a long-running operation
func (r *Resolver) LongRunningQuery(ctx context.Context) (string, error) {
	log.Println("Starting long-running query...")
	select {
	case <-ctx.Done():
		return "", ctx.Err() // Return error if cancelled
	case <-time.After(5 * time.Second):
		return "Query completed successfully", nil
	}
}

// Send a cancellation signal after 3 seconds
func cancelAfterThreeSeconds(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	go func() {
		select {
		case <-time.After(3 * time.Second):
			log.Println("Cancelling request after 3 seconds")
			cancel()
		}
	}()

	graphql.Do(ctx, w, r)
}