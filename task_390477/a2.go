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

// Configuration for JWT
var jwtKey = []byte("my_secret_key")
var tokenRevokedMutex sync.RWMutex
var revokedTokens = make(map[string]struct{})

// Claims struct to hold JWT claims
type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

// Middleware to check for revoked tokens
func tokenRevocationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is missing"})
			c.Abort()
			return
		}

		// Validate the token
		claims, err := validateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		// Check if the token is revoked
		tokenRevokedMutex.RLock()
		_, revoked := revokedTokens[tokenString]
		tokenRevokedMutex.RUnlock()

		if revoked {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token revoked"})
			c.Abort()
			return
		}

		// Set claims in the context
		ctx := context.WithValue(c.Request.Context(), "claims", claims)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// Function to validate the token
func validateToken(tokenString string) (*Claims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHS256); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return jwtKey, nil
	})

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return &Claims{
			Username: claims["username"].(string),
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: int64(claims["exp"].(float64)),
			},
		}, nil
	} else {
		return nil, err
	}
}

// Function to generate a new token
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

// Handler to log in and generate a token
func loginHandler(c *gin.Context) {
	username := c.PostForm("username")
	token, err := generateToken(username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token})
}

// Handler to revoke a token
func logoutHandler(c *gin.Context) {
	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Authorization header is missing"})
		return
	}

	// Lock for writing before revoking
	tokenRevokedMutex.Lock()
	defer tokenRevokedMutex.Unlock()
	revokedTokens[tokenString] = struct{}{}
	c.JSON(http.StatusOK, gin.H{"message": "Token revoked"})
}

// Handler for a protected resource
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