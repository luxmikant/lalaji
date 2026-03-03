package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jambotails/shipping-service/pkg/response"
)

type bucket struct {
	tokens    float64
	lastCheck time.Time
}

// RateLimiter provides a simple in-memory token-bucket rate limiter per client IP.
// maxTokens = bucket size, refillRate = tokens added per second.
func RateLimiter(maxTokens float64, refillRate float64) gin.HandlerFunc {
	var mu sync.Mutex
	buckets := make(map[string]*bucket)

	// Background goroutine to clean up stale entries every 5 minutes.
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			mu.Lock()
			cutoff := time.Now().Add(-10 * time.Minute)
			for ip, b := range buckets {
				if b.lastCheck.Before(cutoff) {
					delete(buckets, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return func(c *gin.Context) {
		ip := c.ClientIP()
		now := time.Now()

		mu.Lock()
		b, exists := buckets[ip]
		if !exists {
			b = &bucket{tokens: maxTokens, lastCheck: now}
			buckets[ip] = b
		}

		// Refill tokens based on elapsed time.
		elapsed := now.Sub(b.lastCheck).Seconds()
		b.tokens += elapsed * refillRate
		if b.tokens > maxTokens {
			b.tokens = maxTokens
		}
		b.lastCheck = now

		if b.tokens < 1 {
			mu.Unlock()
			c.AbortWithStatusJSON(http.StatusTooManyRequests,
				response.Error(c, http.StatusTooManyRequests, "rate limit exceeded, try again later"),
			)
			return
		}

		b.tokens--
		mu.Unlock()

		c.Next()
	}
}
