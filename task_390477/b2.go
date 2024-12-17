package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

// Initialize a secret key for JWT signing
var secretKey = []byte("super-secret-key") // Replace with a secure environment variable

// In-memory cache for revoked tokens (use Redis, Memcached, etc., in production)
var revokedTokens sync.Map

// Define custom claims
type MyClaims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

// Method to check if a token is revoked
func isTokenRevoked(tokenString string) bool {
	_, exists := revokedTokens.Load(tokenString)
	return exists
}

// Method to revoke a token
func revokeToken(tokenString string) {
	revokedTokens.Store(tokenString, struct{}{})
}

// Middleware to validate and check for revoked tokens
func jwtMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authorization := c.Request.Header.Get("Authorization")
		if authorization == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing authorization header"})
			c.Abort()
			return
		}

		tokenString := authorization[len("Bearer "):]

		// Validate the token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHS256); !ok {
				return nil, fmt.Errorf("unexpected signing method")
			}
			return secretKey, nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		claims := token.Claims.(*MyClaims)

		// Check if the token is revoked
		if isTokenRevoked(tokenString) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token revoked"})
			c.Abort()
			return
		}

		// Add claims to context for later use
		ctx := context.WithValue(c.Request.Context(), "claims", claims)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

// Generate a new JWT
func generateJWT(username string) string {
	exp := time.Now().Add(time.Hour * 24).Unix()
	claims := &MyClaims{
		Username: username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: exp,
			Issuer:    "your-app",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return base64.StdEncoding.EncodeToString([]byte(token.SignedString(secretKey)))
}

// Login handler to generate a JWT
func loginHandler(c *gin.Context) {
	username := c.PostForm("username")
	token := generateJWT(username)
	c.JSON(http.StatusOK, gin.H{"token": token})
}

// Logout handler to revoke a JWT
func logoutHandler(c *gin.Context) {
	authorization := c.Request.Header.Get("Authorization")
	if authorization == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing authorization header"})
		c.Abort()
		return
	}

	tokenString := authorization[len("Bearer "):]
	revokeToken(tokenString)
	c.JSON(http.StatusOK, gin.H{"message": "Token revoked"})
}

// Protected route handler
func protectedHandler(c *gin.Context) {
	claims := c.Request.Context().Value("claims").(*MyClaims)
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Hello, %s!", claims.Username)})
}

func main() {
	r := gin.Default()
	r.POST("/login", loginHandler)
	r.POST("/logout", jwtMiddleware(), logoutHandler)
	r.GET("/protected", jwtMiddleware(), protectedHandler)

	log.Fatal(http.ListenAndServe(":8080", r))
}