package main

import (
    "context"
    "fmt"
    "net/http"
    "github.com/gorilla/websocket"
    "github.com/dgrijalva/jwt-go"
)

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true // Allow all origins for simplicity; adjust in production
    },
}

// Secret key used for signing JWTs
var secretKey = []byte("your_secret_key")

// Function to validate the JWT token
func validateToken(tokenString string) (*jwt.Token, error) {
    return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        // Ensure that the token's signing method is HMAC
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return secretKey, nil
    })
}

// WebSocket connection handler
func wsHandler(w http.ResponseWriter, r *http.Request) {
    tokenString := r.URL.Query().Get("token") // Get token from query params

    // Validate the token
    token, err := validateToken(tokenString)
    if err != nil || !token.Valid {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    // Upgrade connection to WebSocket
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
        return
    }

    // Handle the connection (implement your subscription logic here)
    defer conn.Close()
    for {
        // Example of reading messages from the WebSocket
        _, msg, err := conn.ReadMessage()
        if err != nil {
            break
        }
        fmt.Printf("Received: %s\n", msg)
        
        // Echo the message back for demonstration
        err = conn.WriteMessage(websocket.TextMessage, msg)
        if err != nil {
            break
        }
    }
}

func main() {
    http.HandleFunc("/ws", wsHandler)
    fmt.Println("Server started on :8080")
    if err := http.ListenAndServeTLS(":8080", "server.crt", "server.key", nil); err != nil {
        panic("Failed to start server: " + err.Error())
    }
}