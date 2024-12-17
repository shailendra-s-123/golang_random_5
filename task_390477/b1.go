package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// RevocationList represents the store of revoked tokens
var RevocationList = make(map[string]bool)

// Create a new token
func createToken(user string) (string, error) {
	claims := &jwt.StandardClaims{
		Subject: user,
		ExpiresAt: time.Now().Add(time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	secret := []byte("your-secret-key")
	signedToken, err := token.SignedString(secret)
	return signedToken, err
}

// Revokes a token by adding it to the revocation list
func revokeToken(tokenString string) {
	RevocationList[tokenString] = true
}

// TokenCheckMiddleware checks if a token is valid and not revoked
func TokenCheckMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		if tokenString == "" || !strings.HasPrefix(tokenString, "Bearer ") {
			http.Error(w, "Token missing or invalid format", http.StatusUnauthorized)
			return
		}
		tokenString = tokenString[len("Bearer "):]

		token, err := jwt.ParseWithClaims(tokenString, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte("your-secret-key"), nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Token is invalid or expired", http.StatusUnauthorized)
			return
		}

		// Check if the token is revoked
		revoked := RevocationList[tokenString]
		ctx := context.WithValue(r.Context(), "tokenRevoked", revoked)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Example handler that checks the context for token revocation
func exampleHandler(w http.ResponseWriter, r *http.Request) {
	revoked := r.Context().Value("tokenRevoked").(bool)
	if revoked {
		http.Error(w, "Token has been revoked", http.StatusForbidden)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Hello, %s! Your token is valid.\n", r.Context().Value("tokenSubject").(string))
}

func main() {
	http.HandleFunc("/auth", func(w http.ResponseWriter, r *http.Request) {
		user := r.FormValue("user")
		token, err := createToken(user)
		if err != nil {
			http.Error(w, "Error creating token", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Token: %s\n", token)
	})

	http.HandleFunc("/revoke", func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.FormValue("token")
		revokeToken(tokenString)
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Token %s has been revoked.\n", tokenString)
	})

	http.HandleFunc("/protected", TokenCheckMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := r.Context().Value("jwtClaims").(*jwt.StandardClaims)
		ctx := context.WithValue(r.Context(), "tokenSubject", claims.Subject)
		exampleHandler(w, r.WithContext(ctx))
	})))

	fmt.Println("Server listening on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}