package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jambotails/shipping-service/pkg/response"
	"go.uber.org/zap"
)

// Recovery catches panics and returns a 500 JSON response.
func Recovery(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				requestID, _ := c.Get("requestId")

				logger.Error("panic recovered",
					zap.String("requestId", requestID.(string)),
					zap.Any("error", err),
					zap.String("path", c.Request.URL.Path),
				)

				c.AbortWithStatusJSON(http.StatusInternalServerError,
					response.Error(c, http.StatusInternalServerError, "internal server error"),
				)
			}
		}()
		c.Next()
	}
}
