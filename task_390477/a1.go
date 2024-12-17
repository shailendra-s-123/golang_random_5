package main

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

// Secret key used to sign tokens
var jwtKey = []byte("my_secret_key")

// TokenRevocationList holds revoked tokens
var revokedTokens = make(map[string]struct{})
var mu sync.Mutex

type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

// Function to revoke a token
func revokeToken(tokenString string) {
	mu.Lock()
	defer mu.Unlock()
	revokedTokens[tokenString] = struct{}{}
}

// Middleware to check if the token is revoked
func tokenRevocationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if _, err := validateToken(tokenString); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		mu.Lock()
		_, exists := revokedTokens[tokenString]
		mu.Unlock()

		if exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token revoked"})
			c.Abort()
			return
		}

		// Set context with claims
		ctx := context.WithValue(c.Request.Context(), "claims", extractClaims(tokenString))
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// Validate the token
func validateToken(tokenString string) (*Claims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHS256); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return jwtKey, nil
	})

	if claims, ok := token.Claims.(jwt.Claims); ok && token.Valid {
		return claims.(*Claims), nil
	} else {
		return nil, err
	}
}

// Extract claims from the token
func extractClaims(tokenString string) *Claims {
	claims, _ := validateToken(tokenString)
	return claims
}

// Generate a new token
func generateToken(username string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		Username: username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

// Login Handler
func loginHandler(c *gin.Context) {
	username := c.PostForm("username")
	token, err := generateToken(username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token})
}

// Logout Handler (to revoke the token)
func logoutHandler(c *gin.Context) {
	tokenString := c.GetHeader("Authorization")
	revokeToken(tokenString) // Revoke token
	c.JSON(http.StatusOK, gin.H{"message": "Token revoked"})
}

// Protected resource
func protectedHandler(c *gin.Context) {
	claims := c.Request.Context().Value("claims").(*Claims)
	c.JSON(http.StatusOK, gin.H{"message": "success", "username": claims.Username})
}

func main() {
	router := gin.Default()

	router.POST("/login", loginHandler)
	router.POST("/logout", logoutHandler)
	router.GET("/protected", tokenRevocationMiddleware(), protectedHandler)

	router.Run(":8080")
}