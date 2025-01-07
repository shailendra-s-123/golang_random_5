package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

type CustomClaims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// Configuration for token secret keys
var (
	AccessTokenSecret  = []byte("your-access-token-secret")
	RefreshTokenSecret = []byte("your-refresh-token-secret")
)

// GetTokenLifetimes determines token lifetimes based on user role
func GetTokenLifetimes(role string) (time.Duration, time.Duration) {
	switch role {
	case "admin":
		return 5 * time.Minute, 1 * time.Hour // Shorter for admin
	default:
		return 15 * time.Minute, 1 * time.Hour
	}
}

// GenerateTokens creates the JWT tokens for a user
func GenerateTokens(user User) (string, string, error) {
	accessLifetime, refreshLifetime := GetTokenLifetimes(user.Role)

	// Access token
	accessClaims := &CustomClaims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(accessLifetime)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "myapp",
		},
	}
	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString(AccessTokenSecret)
	if err != nil {
		return "", "", fmt.Errorf("failed to sign access token: %w", err)
	}

	// Refresh token
	refreshClaims := &CustomClaims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(refreshLifetime)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "myapp",
		},
	}
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString(RefreshTokenSecret)
	if err != nil {
		return "", "", fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return accessToken, refreshToken, nil
}

// HandleLogin handles user login and token generation
func HandleLogin(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Usually username/password validation must go here...

	accessToken, refreshToken, err := GenerateTokens(user)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error generating tokens: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func main() {
	http.HandleFunc("/login", HandleLogin)
	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}