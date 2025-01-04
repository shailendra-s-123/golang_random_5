package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/handler"
	"github.com/gorilla/websocket"
)

// Define your schema and resolvers
// ... (omitted for brevity)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Replace this with your origin check logic
	},
}

func authenticateSubscription(ctx context.Context, conn *websocket.Conn) (context.Context, error) {
	// Read the JWT from the WebSocket connection headers
	authHeader := conn.Request().Header.Get("Authorization")
	if authHeader == "" {
		return nil, fmt.Errorf("missing authorization header")
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return nil, fmt.Errorf("invalid authorization header format")
	}

	token := parts[1]

	// Verify the token (implement your token verification logic here)
	claims, err := verifyToken(token)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	// Add the verified claims to the context
	return context.WithValue(ctx, "claims", claims), nil
}

func main() {
	// Define your resolvers
	// ... (omitted for brevity)

	srv := handler.NewServer(graphql.NewExecutableSchema(graphql.Config{Resolvers: resolvers}))

	http.HandleFunc("/graphql", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			srv.ServeHTTP(w, r)
		} else if r.Method == http.MethodUpgrade {
			conn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				log.Printf("upgrade error: %v", err)
				return
			}
			defer conn.Close()

			ctx, err := authenticateSubscription(context.Background(), conn)
			if err != nil {
				log.Printf("authentication error: %v", err)
				conn.WriteMessage(websocket.TextMessage, []byte(`{"error": "Authentication failed"}`))
				return
			}

			srv.HandleWebSocketConnection(ctx, conn)
		} else {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})

	log.Println("starting server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

// Implement your token verification logic
func verifyToken(token string) (map[string]interface{}, error) {
	// ... (implement token verification logic here)
	return nil, nil
}