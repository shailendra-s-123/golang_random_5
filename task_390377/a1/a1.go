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
	testExpiredToken(t, expiredTokenString)

	// Test weak signing algorithm (HS224)
	weakToken := jwt.NewWithClaims(jwt.SigningMethodHS224, jwt.StandardClaims{
		Subject:  "1234567890",
		Name:     "John Doe",
		IssuedAt: time.Now().Unix(),
	})
	weakTokenString, err := weakToken.SignedString(secret)
	if err != nil {
		t.Fatal(err)
	}
	testWeakSigningAlgorithm(t, weakTokenString, secret)
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

func testExpiredToken(t *testing.T, tokenString string) {
	_, err := jwt.ParseWithClaims(tokenString, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		return nil, nil
	})

	if err == nil || !errors.Is(err, jwt.ErrSignatureExpired) {
		t.Errorf("Expected expired token error, got: %v", err)
	}
}

func testWeakSigningAlgorithm(t *testing.T, tokenString string, secret []byte) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok || token.Method.Algorithm() == jwt.HS224 {
			return nil, fmt.Errorf("Weak signing algorithm detected: %v", token.Header["alg"])
		}
		return secret, nil
	})

	if err == nil {
		t.Errorf("Expected weak signing algorithm error, got valid token: %+v", token)
	}
}

func main() {
	testing.Run(os.Args[0], func(m *testing.M) {
		os.Exit(m.Run())
	})
}