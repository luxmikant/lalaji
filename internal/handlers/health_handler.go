package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jambotails/shipping-service/internal/cache"
)

// HealthHandler handles health check requests.
type HealthHandler struct {
	db    *sql.DB
	redis *cache.RedisClient
}

// NewHealthHandler creates a new HealthHandler.
func NewHealthHandler(db *sql.DB, redis *cache.RedisClient) *HealthHandler {
	return &HealthHandler{db: db, redis: redis}
}

// Check handles GET /health.
func (h *HealthHandler) Check(c *gin.Context) {
	status := "ok"
	httpCode := http.StatusOK
	details := gin.H{}

	// Check DB
	if err := h.db.PingContext(c.Request.Context()); err != nil {
		status = "degraded"
		httpCode = http.StatusServiceUnavailable
		details["database"] = "unreachable"
	} else {
		details["database"] = "ok"
	}

	// Check Redis
	if h.redis != nil && h.redis.IsAvailable() {
		if err := h.redis.Ping(c.Request.Context()); err != nil {
			status = "degraded"
			details["redis"] = "unreachable"
		} else {
			details["redis"] = "ok"
		}
	} else {
		details["redis"] = "disabled"
	}

	c.JSON(httpCode, gin.H{
		"status":  status,
		"details": details,
	})
}
