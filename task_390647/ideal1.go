package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// User represents the user details.
type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

// CustomClaims represents the JWT claims structure.
type CustomClaims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// Secret keys for signing tokens (use environment variables in production).
var (
	AccessTokenSecret  = []byte("your-access-token-secret")
	RefreshTokenSecret = []byte("your-refresh-token-secret")
)

// GetTokenLifetimes determines token lifetimes dynamically based on the user's role.
func GetTokenLifetimes(role string) (time.Duration, time.Duration) {
	switch role {
	case "admin":
		return 10 * time.Minute, 1 * time.Hour // Shorter for higher privilege roles
	default:
		return 15 * time.Minute, 12 * time.Hour
	}
}

// GenerateTokens generates access and refresh tokens with dynamic lifetimes.
func GenerateTokens(user User) (string, string, error) {
	accessLifetime, refreshLifetime := GetTokenLifetimes(user.Role)

	// Access Token
	accessClaims := &CustomClaims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(accessLifetime)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "your-app",
		},
	}
	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString(AccessTokenSecret)
	if err != nil {
		log.Printf("Error signing access token for user %s: %v", user.Username, err)
		return "", "", fmt.Errorf("failed to sign access token: %w", err)
	}

	// Refresh Token
	refreshClaims := &CustomClaims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(refreshLifetime)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "your-app",
		},
	}
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString(RefreshTokenSecret)
	if err != nil {
		log.Printf("Error signing refresh token for user %s: %v", user.Username, err)
		return "", "", fmt.Errorf("failed to sign refresh token: %w", err)
	}

	log.Printf("Generated tokens for user %s (Role: %s)", user.Username, user.Role)
	return accessToken, refreshToken, nil
}

// HandleLogin handles user login and token generation.
func HandleLogin(w http.ResponseWriter, r *http.Request) {
	log.Println("Received login request")
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		log.Printf("Error decoding request payload: %v", err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if user.Username == "" || user.Role == "" {
		log.Printf("Invalid login attempt: Missing username or role")
		http.Error(w, "Username and role are required", http.StatusBadRequest)
		return
	}

	log.Printf("Processing login for user %s (Role: %s)", user.Username, user.Role)
	accessToken, refreshToken, err := GenerateTokens(user)
	if err != nil {
		log.Printf("Error generating tokens for user %s: %v", user.Username, err)
		http.Error(w, fmt.Sprintf("Error generating tokens: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response for user %s: %v", user.Username, err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}

	log.Printf("Login successful for user %s", user.Username)
}

func main() {
	http.HandleFunc("/login", HandleLogin)
	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
