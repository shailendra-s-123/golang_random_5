package main

import (
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
)

func TestWeakSigningAlgorithms(t *testing.T) {
	// Validate with HS256
	validToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		ExpiresAt: time.Now().Add(time.Hour).Unix(),
	})
	secret := []byte("secret")
	validTokenString, _ := validToken.SignedString(secret)

	_, err := jwt.ParseWithClaims(validTokenString, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	if err != nil {
		t.Fatalf("Failed to parse valid token: %v", err)
	}

	// Test with 'none' algorithm
	noneToken := jwt.New(jwt.SigningMethodNone)
	noneTokenString, _ := noneToken.SignedString(nil)

	_, err = jwt.ParseWithClaims(noneTokenString, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Method)
		}
		return secret, nil
	})
	if err == nil {
		t.Error("Expected error for 'none' signing method")
	}
}