package main

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"net/http"
	"testing"
)

// Create a JWT with the 'none' signing method
func createNoneAlgorithmToken() (string, error) {
	noneToken := jwt.New(jwt.SigningMethodNone)
	claims := make(jwt.MapClaims)
	claims["user"] = "testuser"
	// Set the expiration to 1 hour
	claims["exp"] = jwt.TimeFunc().Add(1 * time.Hour).Unix()
	noneToken.Claims = claims

	tokenString, err := noneToken.SignedString(nil)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// Test if the application accepts the token
func TestWeakSigningAlgorithm(t *testing.T) {
	token, err := createNoneAlgorithmToken()
	if err != nil {
		t.Fatalf("Error creating token: %v", err)
	}

	resp, err := http.Get("http://localhost:8080/api/protected?token=" + token)
	if err != nil {
		t.Fatalf("Error making request: %v", err)
	}
	if resp.StatusCode == http.StatusOK {
		t.Errorf("Weak signing algorithm accepted a token: %s", token)
	}
}