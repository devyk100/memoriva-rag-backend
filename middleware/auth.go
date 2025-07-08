package middleware

import (
	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates the user session
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the user ID from the request header (passed from frontend)
		userID := c.GetHeader("X-User-ID")
		userEmail := c.GetHeader("X-User-Email")

		// For development, allow requests without authentication
		// In production, you'd validate the NextAuth session token
		if userID == "" && userEmail == "" {
			// Allow demo user for development
			c.Set("userID", "demo-user-id")
			c.Set("userEmail", "devyk100@gmail.com")
		} else {
			c.Set("userID", userID)
			c.Set("userEmail", userEmail)
		}

		c.Next()
	}
}

// CORSMiddleware handles CORS
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-User-ID, X-User-Email")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
