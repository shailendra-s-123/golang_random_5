package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
)

// Define the GraphQL query type
var queryType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Query",
	Fields: graphql.Fields{
		"fetchData": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				// Get the context from the resolver parameters
				ctx := p.Context

				// Create an HTTP request using the context
				req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://httpbin.org/delay/5", nil)
				if err != nil {
					return nil, err
				}

				client := &http.Client{}
				resp, err := client.Do(req)
				if err != nil {
					// This error will occur if the request is canceled
					return nil, err
				}
				defer resp.Body.Close()

				// Read the response body
				body, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					return nil, err
				}

				// Return the response body as a string
				return string(body), nil
			},
		},
	},
})

// Set up the GraphQL schema
var schema, _ = graphql.NewSchema(graphql.SchemaConfig{
	Query: queryType,
})

func main() {
	// Create a new GraphQL HTTP handler
	graphQLHandler := handler.New(&handler.Config{
		Schema:   &schema,
		Pretty:   true,
		GraphiQL: true, // Enable GraphiQL interface for testing
	})

	// Start the HTTP server
	http.Handle("/graphql", graphQLHandler)
	fmt.Println("Server is running at :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}