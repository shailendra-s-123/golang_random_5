package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// TokenClaims defines the structure of our JWT claims
type TokenClaims struct {
	Role string `json:"role"`
	jwt.StandardClaims
}

// GenerateJWT generates a token based on user role, which determines its lifetime.
func GenerateJWT(role string) (string, error) {
	var expireDuration time.Duration
	switch role {
	case "admin":
		expireDuration = 24 * time.Hour // 24 hours for admins
	case "user":
		expireDuration = 1 * time.Hour // 1 hour for regular users
	default:
		expireDuration = 15 * time.Minute // 15 minutes for guests or others
	}
	
	// Create the claims
	claims := TokenClaims{
		Role: role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(expireDuration).Unix(),
			Issuer:    "my-service",
		},
	}
	
	// Generate token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signingKey := []byte(os.Getenv("JWT_SECRET_KEY")) // Store the secret key in an environment variable
	return token.SignedString(signingKey)
}

// Middleware to protect routes
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenStr := r.Header.Get("Authorization")
		if tokenStr == "" {
			http.Error(w, "Missing token", http.StatusUnauthorized)
			return
		}
		claims := &TokenClaims{}
		_, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET_KEY")), nil
		})
		if err != nil || !claims.Valid() {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Simple handler for testing
func protectedHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello! You have accessed a protected route.")
}

func main() {
	// Example of generating JWT on user request
	role := "user" // This would normally come from user authentication
	token, err := GenerateJWT(role)
	if err != nil {
		fmt.Println("Error generating token:", err)
		return
	}
	fmt.Println("Generated token:", token)
	
	http.Handle("/protected", AuthMiddleware(http.HandlerFunc(protectedHandler)))
	http.ListenAndServe(":8080", nil)
}