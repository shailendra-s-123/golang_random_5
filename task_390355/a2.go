package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/websocket"
)

var (
	// Secret key for signing JWT
	secretKey = []byte("your_secret_key")
	upgrader   = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Change this for production to securely manage origins
		},
	}
)

// User structure should match your needs
type User struct {
	Username string
	Password string
}

// For demonstration, we're using a hardcoded user.
var validUser = User{Username: "test", Password: "password"}

// Create JWT token
func generateJWT(username string) (string, error) {
	claims := jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(time.Hour * 1).Unix(), // Token is valid for 1 hour
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secretKey)
}

// Middleware for getting JWT token from header
func getTokenFromHeader(r *http.Request) string {
	authorization := r.Header.Get("Authorization")
	if len(authorization) > 7 && authorization[:7] == "Bearer " {
		return authorization[7:]
	}
	return ""
}

// Handler to simulate user authentication and JWT generation
func loginHandler(w http.ResponseWriter, r *http.Request) {
	// In a production app, you would handle request parsing and password comparisons.
	username, password := r.URL.Query().Get("username"), r.URL.Query().Get("password")
	if username == validUser.Username && password == validUser.Password {
		token, err := generateJWT(username)
		if err == nil {
			w.Write([]byte(token))
		} else {
			http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		}
	} else {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
	}
}

// Validate JWT token
func validateToken(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secretKey, nil
	})
}

// WebSocket handler
func wsHandler(w http.ResponseWriter, r *http.Request) {
	token := getTokenFromHeader(r)

	// Validate the token
	tkn, err := validateToken(token)
	if err != nil || !tkn.Valid {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Upgrade connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
		return
	}
	defer conn.Close()

	// Handle messages
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}
		fmt.Printf("Received from %s: %s\n", tkn.Claims.(jwt.MapClaims)["username"], msg)
		err = conn.WriteMessage(websocket.TextMessage, msg) 
		if err != nil {
			break
		}
	}
}

func main() {
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/ws", wsHandler)

	fmt.Println("Server is running on https://localhost:8443")
	err := http.ListenAndServeTLS(":8443", "server.crt", "server.key", nil) // Use correct .crt and .key files
	if err != nil {
		panic("Server failed to start: " + err.Error())
	}
}