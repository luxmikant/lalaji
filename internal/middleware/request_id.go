package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestID injects a unique X-Request-ID header and context value on every request.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Use existing header if provided, else generate a new UUID
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		c.Set("requestId", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}
