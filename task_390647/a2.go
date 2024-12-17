package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type TokenClaims struct {
	Role string `json:"role"`
	jwt.StandardClaims
}

// Environment variable for secret
const jwtSecretEnv = "JWT_SECRET_KEY"

// GenerateJWT generates a JWT based on the user role.
func GenerateJWT(role string) (string, error) {
	var accessTokenExpiry, refreshTokenExpiry time.Duration

	// Define token lifetime based on user role
	switch role {
	case "admin":
		accessTokenExpiry = 15 * time.Minute   // Admins may have a longer expiry
		refreshTokenExpiry = 7 * 24 * time.Hour // Longer refresh token for admins
	case "user":
		accessTokenExpiry = 15 * time.Minute   // Regular users
		refreshTokenExpiry = 3 * 24 * time.Hour // Shorter refresh token for regular users
	default:
		accessTokenExpiry = 5 * time.Minute    // Guests or others
		refreshTokenExpiry = 1 * 24 * time.Hour // Shortest refresh token for guests
	}

	// Generate access token
	accessTokenClaims := TokenClaims{
		Role: role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(accessTokenExpiry).Unix(),
			Issuer:    "my-service",
		},
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessTokenClaims)
	accessTokenString, err := accessToken.SignedString([]byte(os.Getenv(jwtSecretEnv)))
	if err != nil {
		return "", err
	}

	// Generate refresh token
	refreshTokenClaims := jwt.StandardClaims{
		ExpiresAt: time.Now().Add(refreshTokenExpiry).Unix(),
		Issuer:    "my-service",
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshTokenClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(os.Getenv(jwtSecretEnv)))
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Access Token: %s\nRefresh Token: %s", accessTokenString, refreshTokenString), nil
}

// AuthMiddleware checks the validity of the JWT token.
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenStr := r.Header.Get("Authorization")
		if tokenStr == "" {
			http.Error(w, "Missing token", http.StatusUnauthorized)
			return
		}

		claims := &TokenClaims{}
		_, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv(jwtSecretEnv)), nil
		})
		if err != nil || !claims.Valid() {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Pass the claims onward
		next.ServeHTTP(w, r)
	})
}

// Simple protected handler
func protectedHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello! You have accessed a protected route.")
}

func main() {
	// Example of generating JWT on user request
	role := "user" // This would normally come from user authentication
	tokens, err := GenerateJWT(role)
	if err != nil {
		fmt.Println("Error generating tokens:", err)
		return
	}
	fmt.Println(tokens)

	http.Handle("/protected", AuthMiddleware(http.HandlerFunc(protectedHandler)))
	http.ListenAndServe(":8080", nil)
}