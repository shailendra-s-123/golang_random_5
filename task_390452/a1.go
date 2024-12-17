package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
)

// Query structure with an example property
var queryType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Query",
	Fields: graphql.Fields{
		"getData": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				ctx := p.Context // Get the context from the resolver parameters

				// Simulate an HTTP request which respects the context for cancellation
				req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://httpbin.org/delay/5", nil)
				if err != nil {
					return nil, err
				}

				client := &http.Client{}
				resp, err := client.Do(req)
				if err != nil {
					return nil, err
				}
				defer resp.Body.Close()

				// You can read the body and return the result
				return "Data fetched successfully", nil
			},
		},
	},
})

var schema, _ = graphql.NewSchema(graphql.SchemaConfig{
	Query: queryType,
})

func main() {
	// Create a new GraphQL HTTP handler
	graphQLHandler := handler.New(&handler.Config{
		Schema:   &schema,
		Pretty:   true,
		GraphiQL: true, // Enable GraphiQL interface
	})

	http.Handle("/graphql", graphQLHandler)
	fmt.Println("Server is running at :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}