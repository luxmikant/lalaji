package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// RedisClient wraps the go-redis client with health check and logging.
type RedisClient struct {
	Client *redis.Client
	logger *zap.Logger
}

// NewRedisClient creates and pings a new Redis connection.
func NewRedisClient(addr, password string, db int, logger *zap.Logger) (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           db,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     20,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		logger.Warn("Redis connection failed — cache will be disabled", zap.Error(err))
		return &RedisClient{Client: nil, logger: logger}, nil
	}

	logger.Info("Redis connected successfully", zap.String("addr", addr))
	return &RedisClient{Client: client, logger: logger}, nil
}

// IsAvailable returns true if Redis connection is established.
func (r *RedisClient) IsAvailable() bool {
	return r.Client != nil
}

// Ping checks Redis health.
func (r *RedisClient) Ping(ctx context.Context) error {
	if !r.IsAvailable() {
		return fmt.Errorf("redis not connected")
	}
	return r.Client.Ping(ctx).Err()
}

// Close gracefully closes the Redis connection.
func (r *RedisClient) Close() error {
	if r.Client != nil {
		return r.Client.Close()
	}
	return nil
}
