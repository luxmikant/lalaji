package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jambotails/shipping-service/pkg/response"
)

// Auth validates a Bearer JWT and puts claims into the context.
func Auth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized,
				response.Error(c, http.StatusUnauthorized, "missing authorization header"),
			)
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized,
				response.Error(c, http.StatusUnauthorized, "invalid authorization format"),
			)
			return
		}

		token, err := jwt.Parse(parts[1], func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized,
				response.Error(c, http.StatusUnauthorized, "invalid or expired token"),
			)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized,
				response.Error(c, http.StatusUnauthorized, "invalid token claims"),
			)
			return
		}

		c.Set("claims", claims)
		if sub, ok := claims["sub"].(string); ok {
			c.Set("userId", sub)
		}

		c.Next()
	}
}
