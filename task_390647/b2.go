package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// TokenClaims holds JWT claims.
type TokenClaims struct {
	Role   string `json:"role"`
	IssuedAt int64 `json:"iat"`
	jwt.StandardClaims
}

// AccessTokenManager struct handles access and refresh token generation.
type AccessTokenManager struct {
	accessTokenDuration  time.Duration
	refreshTokenDuration time.Duration
	secretKey            []byte
}

// NewAccessTokenManager creates a new token manager.
func NewAccessTokenManager(
	accessTokenDuration, refreshTokenDuration time.Duration,
	secretKey string,
) (*AccessTokenManager, error) {
	if secretKey == "" {
		return nil, fmt.Errorf("secret key cannot be empty")
	}
	return &AccessTokenManager{
		accessTokenDuration:  accessTokenDuration,
		refreshTokenDuration: refreshTokenDuration,
		secretKey:            []byte(secretKey),
	}, nil
}

// GenerateTokens generates an access and refresh token based on the user's role.
func (atm *AccessTokenManager) GenerateTokens(role string) (accessToken, refreshToken string, err error) {
	accessClaims := TokenClaims{
		Role:           role,
		ExpiresAt:      time.Now().Add(atm.accessTokenDuration).Unix(),
		IssuedAt:       time.Now().Unix(),
		Issuer:         "your-service",
		Subject:        "user-subject",
	}

	refreshClaims := TokenClaims{
		Role:           role,
		ExpiresAt:      time.Now().Add(atm.refreshTokenDuration).Unix(),
		IssuedAt:       time.Now().Unix(),
		Issuer:         "your-service",
		Subject:        "user-subject",
	}

	accessToken = atm.signToken(accessClaims)
	refreshToken = atm.signToken(refreshClaims)
	return
}

func (atm *AccessTokenManager) signToken(claims *TokenClaims) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(atm.secretKey)
}

// VerifyToken verifies a JWT token.
func (atm *AccessTokenManager) VerifyToken(tokenString string) (*TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return atm.secretKey, nil
	})
	if err != nil || !token.Valid {
		return nil, err
	}
	return token.Claims.(*TokenClaims), nil
}

// HandleLogin handles the login endpoint, providing access and refresh tokens.
func HandleLogin(w http.ResponseWriter, r *http.Request, atm *AccessTokenManager) {
	// Normally you would validate user credentials here
	role := "user" // Assume this role is determined post-authentication

	accessToken, refreshToken, err := atm.GenerateTokens(role)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

// HandleProtected handles a protected route, verifying the access token.
func HandleProtected(w http.ResponseWriter, r *http.Request, atm *AccessTokenManager) {
	accessToken := r.Header.Get("Authorization")
	if accessToken == "" {
		http.Error(w, "Unauthorized: Access token required", http.StatusUnauthorized)
		return
	}

	claims, err := atm.VerifyToken(accessToken)
	if err != nil {
		http.Error(w, "Unauthorized: Invalid token", http.StatusUnauthorized)
		return
	}

	fmt.Fprintf(w, "Hello, %s!\n", claims.Role)
}

func main() {
	atm, err := NewAccessTokenManager(time.Hour, time.Hour*24, "your-super-secret-key")
	if err != nil {
		panic(err)
	}

	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		HandleLogin(w, r, atm)
	})

	http.HandleFunc("/protected", func(w http.ResponseWriter, r *http.Request) {
		HandleProtected(w, r, atm)
	})

	fmt.Println("Listening on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}