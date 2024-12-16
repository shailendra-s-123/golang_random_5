package main

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"testing"
)

func TestJWTVulnerabilities(t *testing.T) {
	// Sample secret for HMAC signing
	secret := []byte("supersecret")

	// Test token signing with 'none' algorithm
	noneToken := jwt.New(jwt.SigningMethodNone)
	noneToken.Claims = map[string]interface{}{
		"sub": "1234567890",
		"name": "John Doe",
		"iat": time.Now().Unix(),
	}
	noneTokenString, err := noneToken.SignedString(nil)
	if err != nil {
		t.Fatal(err)
	}
	testInvalidToken(t, noneTokenString, secret)

	// Test expired token
	expiredToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		Subject:  "1234567890",
		Name:     "John Doe",
		IssuedAt: time.Now().Unix(),
		ExpiresAt: time.Now().Add(-1 * time.Minute).Unix(),
	})
	expiredTokenString, err := expiredToken.SignedString(secret)
	if err != nil {
		t.Fatal(err)
	}
	testInvalidToken(t, expiredTokenString, secret)

	// Test token with incorrect issuer
	incorrectIssuerToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		Subject:  "1234567890",
		Name:     "John Doe",
		IssuedAt: time.Now().Unix(),
		Issuer:   "invalid_issuer",
	})
	incorrectIssuerTokenString, err := incorrectIssuerToken.SignedString(secret)
	if err != nil {
		t.Fatal(err)
	}
	testInvalidToken(t, incorrectIssuerTokenString, secret)
}

func testInvalidToken(t *testing.T, tokenString string, secret []byte) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return secret, nil
	})

	if err == nil {
		t.Errorf("Expected error, got valid token: %+v", token)
	}
}

func main() {
	testing.Run(os.Args[0], func(m *testing.M) {
		os.Exit(m.Run())
	})
}