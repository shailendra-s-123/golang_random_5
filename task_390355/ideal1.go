package main

import (
	"log"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/websocket"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
	"golang.org/x/crypto/acme/autocert"
)

// JWT validation function
func validateToken(tokenString string) (string, error) {
	// Replace with your own JWT secret
	secret := []byte("your-secret-key")
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	if err != nil || !token.Valid {
		return "", err
	}

	// Extract user info from the token (e.g., userID)
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", err
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return "", err
	}

	return userID, nil
}

// WebSocket upgrader for connection
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // In production, implement proper CORS policy
	},
}

// WebSocket handler for managing subscriptions
func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Could not upgrade connection", http.StatusBadRequest)
		return
	}
	defer conn.Close()

	// Extract JWT token from URL query parameter
	token := r.URL.Query().Get("token")
	userID, err := validateToken(token)
	if err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte("Unauthorized"))
		return
	}

	// Example of using the userID in your application logic
	log.Printf("User %s connected", userID)

	// Handle GraphQL subscription here (you can use userID to customize subscriptions)
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			break // Exit on error
		}

		// Example of using the msg (subscription request)
		log.Printf("Received subscription message: %s", msg)

		// Process incoming subscription requests and send back responses
		conn.WriteMessage(websocket.TextMessage, []byte("Subscription data"))
	}
}

// GraphQL schema example
var rootQuery = graphql.Fields{
	"hello": &graphql.Field{
		Type: graphql.String,
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			return "Hello, world!", nil
		},
	},
}

var schema, _ = graphql.NewSchema(graphql.SchemaConfig{
	Query: graphql.NewObject(graphql.ObjectConfig{
		Name:   "RootQuery",
		Fields: rootQuery,
	}),
})

// Serve GraphQL queries via HTTP
func graphqlHandler(w http.ResponseWriter, r *http.Request) {
	handler := handler.New(&handler.Config{
		Schema: &schema,
		Pretty: true,
	})
	handler.ServeHTTP(w, r)
}

func main() {
	// Set up auto TLS (useful for production, self-signed certs for dev)
	m := autocert.Manager{
		Cache:      autocert.DirCache("./certs"),
		HostPolicy: autocert.HostWhitelist("localhost"),
	}

	// Set up HTTP server with routes
	http.HandleFunc("/graphql", graphqlHandler) // GraphQL queries
	http.HandleFunc("/ws", wsHandler)           // WebSocket for subscriptions

	// HTTPS server with automatic TLS management
	server := &http.Server{
		Addr:         ":8080",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		TLSConfig:    m.TLSConfig(),
	}

	log.Printf("Starting server on https://localhost:8080")
	if err := server.ListenAndServeTLS("", ""); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

