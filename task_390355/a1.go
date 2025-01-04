package main

import (
    "github.com/gorilla/websocket"
    "github.com/graphql-go/graphql"
    "github.com/graphql-go/handler"
    "net/http"
    "time"
)

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true // Set proper CORS policy here
    },
}

// JWT validation function
func validateToken(token string) (string, error) {
    // Implement your JWT validation logic here
    // Return user ID or any user information if valid, else return error
}

// WebSocket handler
func wsHandler(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        http.Error(w, "Could not upgrade connection", http.StatusBadRequest)
        return
    }
    defer conn.Close()

    // Extract token from query parameters
    token := r.URL.Query().Get("token")
    userID, err := validateToken(token)
    if err != nil {
        conn.WriteMessage(websocket.TextMessage, []byte("Unauthorized"))
        return
    }

    // Handle subscription logic here
    // Use userID to determine what data the user can subscribe to

    for {
        _, msg, err := conn.ReadMessage()
        if err != nil {
            break
        }
        // Handle incoming message (GraphQL subscription)
        // Respond to client with updated data
    }
}

func main() {
    http.Handle("/", handler.NewGraphQLHandler(/* your GraphQL schema */))

    http.HandleFunc("/ws", wsHandler)

    // Start the server
    srv := &http.Server{
        Addr:         ":8080",
        ReadTimeout:  10 * time.Second,
        WriteTimeout: 10 * time.Second,
    }
    srv.ListenAndServe()
}