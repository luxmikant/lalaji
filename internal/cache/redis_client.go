package cache

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// RedisClient wraps the go-redis client with health check and logging.
type RedisClient struct {
	Client *redis.Client
	logger *zap.Logger
}

// NewRedisClient creates a Redis connection.
// When rawURL is a redis:// or rediss:// URL (e.g. Upstash), it is parsed directly
// so TLS is configured automatically for rediss:// scheme.
// When rawURL is empty, addr + password + db are used instead.
func NewRedisClient(rawURL, addr, password string, db int, logger *zap.Logger) (*RedisClient, error) {
	var client *redis.Client

	if strings.HasPrefix(rawURL, "redis://") || strings.HasPrefix(rawURL, "rediss://") {
		opt, err := redis.ParseURL(rawURL)
		if err != nil {
			logger.Warn("failed to parse REDIS_URL — cache disabled", zap.Error(err))
			return &RedisClient{Client: nil, logger: logger}, nil
		}
		opt.DialTimeout = 5 * time.Second
		opt.ReadTimeout = 3 * time.Second
		opt.WriteTimeout = 3 * time.Second
		opt.PoolSize = 10 // keep low; Upstash free tier limits concurrent connections
		client = redis.NewClient(opt)
		logger.Info("using Redis URL (TLS auto-detected)", zap.String("url", maskURL(rawURL)))
	} else {
		client = redis.NewClient(&redis.Options{
			Addr:         addr,
			Password:     password,
			DB:           db,
			DialTimeout:  5 * time.Second,
			ReadTimeout:  3 * time.Second,
			WriteTimeout: 3 * time.Second,
			PoolSize:     20,
		})
		logger.Info("using Redis addr", zap.String("addr", addr))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		logger.Warn("Redis ping failed — cache will be disabled", zap.Error(err))
		return &RedisClient{Client: nil, logger: logger}, nil
	}

	logger.Info("Redis connected successfully")
	return &RedisClient{Client: client, logger: logger}, nil
}

// maskURL hides credentials in log output.
func maskURL(rawURL string) string {
	// Replace everything between :// and @ with ***
	if at := strings.Index(rawURL, "@"); at != -1 {
		scheme := rawURL[:strings.Index(rawURL, "://")+3]
		return scheme + "***@" + rawURL[at+1:]
	}
	return rawURL
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
