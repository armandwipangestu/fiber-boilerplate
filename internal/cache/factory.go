package cache

import (
	"context"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

// NewCache selects the cache backend: Redis when redisURL is configured and
// reachable, otherwise an in-memory store. The active backend is logged.
func NewCache(ctx context.Context, redisURL string, logger *slog.Logger) (Cache, error) {
	if redisURL == "" {
		logger.Info("cache backend selected", "backend", "memory")
		return NewMemoryCache(), nil
	}

	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}
	client := redis.NewClient(opts)

	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	if err := client.Ping(pingCtx).Err(); err != nil {
		_ = client.Close()
		if logger != nil {
			logger.Warn("redis unreachable, falling back to in-memory cache", "error", err)
		}
		return NewMemoryCache(), nil
	}

	if logger != nil {
		logger.Info("cache backend selected", "backend", "redis", "addr", opts.Addr)
	}
	return NewRedisCache(client), nil
}
