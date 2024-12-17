package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

// User struct represents a user
type User struct {
	ID      int    `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

// Claims struct represents the JWT claims
type Claims struct {
	User    User    `json:"user"`
	Expires time.Time `json:"exp"`
	jwt.StandardClaims
}

// GetTokenLifetime dynamically determines token lifetime based on user role
func GetTokenLifetime(role string) (accessTime, refreshTime time.Duration) {
	switch role {
	case "admin":
		accessTime = 10 * time.Minute
		refreshTime = 1 * time.Hour
	default:
		accessTime = 15 * time.Minute
		refreshTime = 12 * time.Hour
	}
	return
}

// GenerateJWT generates a JWT with a dynamically configured lifetime
func GenerateJWT(user User) (string, string, error) {
	accessTime, refreshTime := GetTokenLifetime(user.Role)

	// Access Token
	accessClaims := &Claims{
		User:      user,
		Expires:   time.Now().Add(accessTime),
		IssuedAt:  time.Now(),
		Subject:   user.Username,
		Issuer:    "your-app",
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessString, err := accessToken.SignedString([]byte("your-secret-key"))
	if err != nil {
		return "", "", err
	}

	// Refresh Token
	refreshClaims := &Claims{
		User:      user,
		Expires:   time.Now().Add(refreshTime),
		IssuedAt:  time.Now(),
		Subject:   user.Username,
		Issuer:    "your-app",
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshString, err := refreshToken.SignedString([]byte("your-secret-key"))
	if err != nil {
		return "", "", err
	}

	return accessString, refreshString, nil
}

// HandleLogin handles user login and returns tokens
func HandleLogin(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	accessToken, refreshToken, err := GenerateJWT(user)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

func main() {
	http.HandleFunc("/login", HandleLogin)

	// Assuming we run in a container where service mesh manages traffic
	fmt.Println("Listening on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}