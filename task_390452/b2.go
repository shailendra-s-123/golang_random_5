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

// Query type definition
var queryType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Query",
	Fields: graphql.Fields{
		"fetchData": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				ctx := p.Context

				// Create a new context with a deadline of 2 seconds for demonstration purposes
				ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
				defer cancel()

				// Perform an HTTP request with the context
				req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://httpbin.org/delay/5", nil)
				if err != nil {
					return nil, fmt.Errorf("failed to create request: %w", err)
				}

				client := &http.Client{}
				resp, err := client.Do(req)
				if err != nil {
					// If context is done, it means the request was canceled
					if ctx.Err() == context.DeadlineExceeded || ctx.Err() == context.Canceled {
						return nil, fmt.Errorf("HTTP request canceled: %w", ctx.Err())
					}
					return nil, fmt.Errorf("failed to perform HTTP request: %w", err)
				}
				defer resp.Body.Close()

				body, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					return nil, fmt.Errorf("failed to read response body: %w", err)
				}

				return string(body), nil
			},
		},
	},
})

// Schema definition
var schema, _ = graphql.NewSchema(graphql.SchemaConfig{
	Query: queryType,
})

func main() {
	// Create a new GraphQL HTTP handler
	graphQLHandler := handler.New(&handler.Config{
		Schema: &schema,
	})

	http.Handle("/graphql", graphQLHandler)
	fmt.Println("Server is running at http://localhost:8080/graphql")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}