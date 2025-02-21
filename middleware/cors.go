package middleware

import (
	"easy-chat/config"
	"github.com/gin-gonic/gin"
	"net/http"
)

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg := config.Get()
		allowedOrigins := cfg.AllowedOrigin
		origin := c.GetHeader("Origin")
		if !isOriginAllowed(origin, allowedOrigins) {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		setHeaders(c)

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func isOriginAllowed(origin string, allowedOrigins []string) bool {
	for _, allowedOrigin := range allowedOrigins {
		if origin == allowedOrigin {
			return true
		}
	}
	return false
}

func setHeaders(c *gin.Context) {
	origin := c.GetHeader("Origin")
	c.Header("Access-Control-Allow-Origin", origin)
	c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
	c.Header("Access-Control-Allow-Credentials", "true")
}
