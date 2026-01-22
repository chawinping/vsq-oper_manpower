package middleware

import (
	"vsq-oper-manpower/backend/internal/config"

	"github.com/gin-gonic/gin"
)

func CORS(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		
		// Check if the origin is in the allowed list
		allowedOrigin := ""
		for _, allowed := range cfg.CORS.AllowedOrigins {
			if origin == allowed {
				allowedOrigin = origin
				break
			}
		}
		
		// If origin matches an allowed origin, use it; otherwise, don't set the header
		// (browsers will reject if credentials are used and origin doesn't match)
		if allowedOrigin != "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		}
		
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}


