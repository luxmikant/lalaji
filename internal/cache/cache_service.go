package cache

import (
	"context"
	"encoding/json"
	"time"

	"go.uber.org/zap"
)

// CacheService provides a simple Get/Set/Delete abstraction over Redis
// with automatic JSON serialization and graceful fallback when Redis is unavailable.
type CacheService struct {
	redis  *RedisClient
	logger *zap.Logger
}

// NewCacheService creates a new cache service.
func NewCacheService(redis *RedisClient, logger *zap.Logger) *CacheService {
	return &CacheService{redis: redis, logger: logger}
}

// Get retrieves a value from cache and unmarshals into dest.
// Returns true if cache hit, false on miss or error (graceful fallback).
func (c *CacheService) Get(ctx context.Context, key string, dest interface{}) bool {
	if !c.redis.IsAvailable() {
		return false
	}

	val, err := c.redis.Client.Get(ctx, key).Result()
	if err != nil {
		// Cache miss or Redis error — not logged as error (expected flow)
		return false
	}

	if err := json.Unmarshal([]byte(val), dest); err != nil {
		c.logger.Warn("cache unmarshal failed", zap.String("key", key), zap.Error(err))
		return false
	}

	c.logger.Debug("cache hit", zap.String("key", key))
	return true
}

// Set stores a value in cache with the given TTL.
// Silently fails if Redis is unavailable (graceful degradation).
func (c *CacheService) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) {
	if !c.redis.IsAvailable() {
		return
	}

	data, err := json.Marshal(value)
	if err != nil {
		c.logger.Warn("cache marshal failed", zap.String("key", key), zap.Error(err))
		return
	}

	if err := c.redis.Client.Set(ctx, key, data, ttl).Err(); err != nil {
		c.logger.Warn("cache set failed", zap.String("key", key), zap.Error(err))
		return
	}

	c.logger.Debug("cache set", zap.String("key", key), zap.Duration("ttl", ttl))
}

// Delete removes a key from cache.
func (c *CacheService) Delete(ctx context.Context, key string) {
	if !c.redis.IsAvailable() {
		return
	}

	if err := c.redis.Client.Del(ctx, key).Err(); err != nil {
		c.logger.Warn("cache delete failed", zap.String("key", key), zap.Error(err))
	}
}
