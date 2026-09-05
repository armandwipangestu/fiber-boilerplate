package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisCache is a Cache backed by Redis, suitable for multi-instance
// deployments that must share the same memoized lookups.
type RedisCache struct {
	client *redis.Client
}

// NewRedisCache wraps an existing Redis client.
func NewRedisCache(client *redis.Client) *RedisCache {
	return &RedisCache{client: client}
}

func (c *RedisCache) Get(ctx context.Context, key string) ([]byte, error) {
	val, err := c.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, ErrCacheMiss
	}
	return val, err
}

func (c *RedisCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	exp := ttl
	if ttl <= 0 {
		exp = 0 // keep forever
	}
	return c.client.Set(ctx, key, value, exp).Err()
}

func (c *RedisCache) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

// DeletePattern removes all keys matching the glob pattern using SCAN so it
// works on large key sets without blocking.
func (c *RedisCache) DeletePattern(ctx context.Context, pattern string) error {
	var cursor uint64
	var keys []string
	for {
		var err error
		keys, cursor, err = c.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return err
		}
		if len(keys) > 0 {
			if err := c.client.Del(ctx, keys...).Err(); err != nil {
				return err
			}
		}
		if cursor == 0 {
			return nil
		}
	}
}

func (c *RedisCache) Close() error {
	return c.client.Close()
}
