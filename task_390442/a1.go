package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
)

var mySigningKey = []byte("secret") // Weak signing key for demonstration

func GenerateJWT(w http.ResponseWriter, r *http.Request) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": "user1",
		"exp":      time.Now().Add(time.Hour * 1).Unix(),
	})

	tokenString, err := token.SignedString(mySigningKey)
	if err != nil {
		http.Error(w, "Could not generate token", http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, tokenString)
}

func ValidateJWT(w http.ResponseWriter, r *http.Request) {
	tokenString := r.Header.Get("Authorization")

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Check the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return mySigningKey, nil
	})

	if err != nil || !token.Valid {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		fmt.Fprintf(w, "Welcome %v!", claims["username"])
	} else {
		http.Error(w, "Invalid claims", http.StatusUnauthorized)
	}
}

func main() {
	http.HandleFunc("/generate", GenerateJWT)
	http.HandleFunc("/validate", ValidateJWT)
	log.Fatal(http.ListenAndServe(":8080", nil))
}