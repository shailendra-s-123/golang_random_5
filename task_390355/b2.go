package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/dgrijalva/jwt-go"
)

// Define a secret key for JWT signing
var secretKey = []byte("your_secret_key")

// ClaimSet contains the data stored in the JWT
type ClaimSet struct {
	UserID string `json:"user_id"`
	jwt.StandardClaims
}

// IssueJWT generates a signed JWT for a given user ID
func IssueJWT(userID string) string {
	claims := &ClaimSet{
		UserID: userID,
		ExpiresAt: time.Now().Add(time.Hour * 1).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(secretKey)
	if err != nil {
		log.Fatal("Error signing JWT:", err)
	}
	return signedToken
}

// ValidateJWT extracts and validates the JWT from the WebSocket request
func ValidateJWT(r *http.Request) (*ClaimSet, error) {
	tokenString := r.URL.Query().Get("token")
	if tokenString == "" {
		return nil, fmt.Errorf("no token provided")
	}

	token, err := jwt.ParseWithClaims(tokenString, &ClaimSet{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secretKey, nil
	})

	if err != nil || !token.Valid {
		return nil, err
	}

	return token.Claims.(*ClaimSet), nil
}

// WebSocket handler with JWT authentication
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for simplicity; adjust in production
	},
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	claims, err := ValidateJWT(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	log.Printf("User %s connected\n", claims.UserID)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error("Upgrade failed:", err)
		return
	}
	defer conn.Close()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Error("ReadMessage failed:", err)
			break
		}

		log.Printf("Received message: %s\n", message)

		err = conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Error("WriteMessage failed:", err)
			break
		}
	}
}

func main() {
	// Secure server setup with TLS
	fs := http.FileServer(http.Dir("."))
	http.HandleFunc("/ws", wsHandler)
	http.Handle("/", fs)

	srv := &http.Server{
		Addr:      ":443",
		Handler:   nil, // Set the actual handler using the certificate handler
		TLSConfig: &tls.Config{
			RootCAs:      x509.NewCertPool(),
			Certificates: []tls.Certificate{{Certificate: []byte(serverCert), PrivateKey: []byte(serverKey)}},
		},
	}

	// Make sure to load your server.crt and server.key
	serverCert := []byte(`-----BEGIN CERTIFICATE-----
...your certificate here...
-----END CERTIFICATE-----`)
	serverKey := []byte(`-----BEGIN PRIVATE KEY-----
...your private key here...
-----END PRIVATE KEY-----`)

	err := srv.ListenAndServeTLS("", "")
	if err != nil {
		log.Fatal("ListenAndServeTLS error:", err)
	}
}