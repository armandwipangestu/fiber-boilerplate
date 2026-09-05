package cache

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

// TestRedisCacheIntegration exercises the Redis backend against a live server.
// Set TEST_REDIS_URL (default redis://localhost:6379/15) to run; skipped
// otherwise so the suite passes without a Redis instance.
func TestRedisCacheIntegration(t *testing.T) {
	url := os.Getenv("TEST_REDIS_URL")
	if url == "" {
		url = "redis://localhost:6379/15"
	}
	opts, err := redis.ParseURL(url)
	if err != nil {
		t.Fatal(err)
	}
	client := redis.NewClient(opts)
	ctx := context.Background()

	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	if err := client.Ping(pingCtx).Err(); err != nil {
		t.Skip("redis not reachable, skipping integration test")
	}
	defer client.FlushDB(ctx).Err()
	defer client.Close()

	c := NewRedisCache(client)

	// miss path
	if _, err := c.Get(ctx, "missing"); !errors.Is(err, ErrCacheMiss) {
		t.Fatalf("expected ErrCacheMiss, got %v", err)
	}

	// roundtrip
	if err := c.Set(ctx, "k", []byte("v"), 0); err != nil {
		t.Fatal(err)
	}
	got, err := c.Get(ctx, "k")
	if err != nil || string(got) != "v" {
		t.Fatalf("got %q, %v; want v", got, err)
	}

	// ttl expiration
	if err := c.Set(ctx, "exp", []byte("x"), 50*time.Millisecond); err != nil {
		t.Fatal(err)
	}
	time.Sleep(70 * time.Millisecond)
	if _, err := c.Get(ctx, "exp"); !errors.Is(err, ErrCacheMiss) {
		t.Fatalf("expected expired key to miss, got %v", err)
	}

	// delete + pattern delete
	_ = c.Set(ctx, "p:1", []byte("a"), 0)
	_ = c.Set(ctx, "p:2", []byte("b"), 0)
	_ = c.Set(ctx, "q:3", []byte("c"), 0)
	if err := c.DeletePattern(ctx, "p:*"); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Get(ctx, "p:1"); !errors.Is(err, ErrCacheMiss) {
		t.Fatalf("pattern delete failed for p:1, got %v", err)
	}
	if _, err := c.Get(ctx, "q:3"); err != nil {
		t.Fatalf("unrelated key should remain: %v", err)
	}
}
