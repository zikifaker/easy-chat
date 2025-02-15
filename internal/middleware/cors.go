package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func CORSMiddleware() gin.HandlerFunc {
	allowedOrigins := []string{"http://localhost:7002"}

	return func(c *gin.Context) {
		handleCORS(c, allowedOrigins)
		c.Next()
	}
}

func handleCORS(c *gin.Context, allowedOrigins []string) {
	origin := c.GetHeader("Origin")
	allowed := false
	for _, allowedOrigin := range allowedOrigins {
		if origin == allowedOrigin {
			c.Header("Access-Control-Allow-Origin", origin)
			allowed = true
			break
		}
	}

	if !allowed {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
	c.Header("Access-Control-Allow-Credentials", "true")

	if c.Request.Method == "OPTIONS" {
		c.AbortWithStatus(http.StatusNoContent)
		return
	}
}
