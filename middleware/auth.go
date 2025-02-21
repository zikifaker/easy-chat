package middleware

import (
	"easy-chat/config"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

var (
	ErrMissedToken             = errors.New("missed token")
	ErrInvalidTokenFormat      = errors.New("invalid token format")
	ErrUnexpectedSigningMethod = errors.New("unexpected signing method")
	ErrInvalidToken            = errors.New("invalid token")
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := authenticateRequest(c); err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.Next()
	}
}

func authenticateRequest(c *gin.Context) error {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return ErrMissedToken
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {
		return ErrInvalidTokenFormat
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	return validateToken(tokenString)
}

func validateToken(tokenString string) error {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("%w: %v", ErrUnexpectedSigningMethod, token.Header["alg"])
		}
		return []byte(config.Get().SecretKey.JWT), nil
	})
	if err != nil || !token.Valid {
		return ErrInvalidToken
	}
	return nil
}
