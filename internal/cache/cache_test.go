package cache

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"
)

func TestMemoryCacheSetGet(t *testing.T) {
	c := NewMemoryCache()
	defer c.Close()

	ctx := context.Background()
	if _, err := c.Get(ctx, "k"); !errors.Is(err, ErrCacheMiss) {
		t.Fatalf("expected ErrCacheMiss, got %v", err)
	}

	if err := c.Set(ctx, "k", []byte("v"), time.Minute); err != nil {
		t.Fatalf("Set: %v", err)
	}
	got, err := c.Get(ctx, "k")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if string(got) != "v" {
		t.Fatalf("got %q, want %q", got, "v")
	}
}

func TestMemoryCacheTTL(t *testing.T) {
	c := NewMemoryCache()
	defer c.Close()

	ctx := context.Background()
	if err := c.Set(ctx, "short", []byte("x"), 50*time.Millisecond); err != nil {
		t.Fatal(err)
	}
	if err := c.Set(ctx, "forever", []byte("y"), 0); err != nil {
		t.Fatal(err)
	}
	time.Sleep(70 * time.Millisecond)

	if _, err := c.Get(ctx, "short"); !errors.Is(err, ErrCacheMiss) {
		t.Fatalf("expected expired key to miss, got %v", err)
	}
	if _, err := c.Get(ctx, "forever"); err != nil {
		t.Fatalf("non-expiring key should survive: %v", err)
	}
}

func TestMemoryCacheDeleteAndPattern(t *testing.T) {
	c := NewMemoryCache()
	defer c.Close()

	ctx := context.Background()
	_ = c.Set(ctx, "rbac:permissions:u1", []byte("a"), 0)
	_ = c.Set(ctx, "rbac:permissions:u2", []byte("b"), 0)
	_ = c.Set(ctx, "other:key", []byte("c"), 0)

	if err := c.Delete(ctx, "rbac:permissions:u1"); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Get(ctx, "rbac:permissions:u1"); !errors.Is(err, ErrCacheMiss) {
		t.Fatalf("delete should remove key, got %v", err)
	}

	if err := c.DeletePattern(ctx, "rbac:permissions:*"); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Get(ctx, "rbac:permissions:u2"); !errors.Is(err, ErrCacheMiss) {
		t.Fatalf("pattern delete should remove key, got %v", err)
	}
	if _, err := c.Get(ctx, "other:key"); err != nil {
		t.Fatalf("unrelated key should remain: %v", err)
	}
}

func TestMemoryCacheCloseIsIdempotent(t *testing.T) {
	c := NewMemoryCache()
	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestNewCacheNoRedis(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	c, err := NewCache(context.Background(), "", logger)
	if err != nil {
		t.Fatalf("NewCache: %v", err)
	}
	defer c.Close()
	if _, ok := c.(*MemoryCache); !ok {
		t.Fatalf("expected MemoryCache, got %T", c)
	}
}

// TestNewCacheRedisFallsBack exercises the fallback branch without a running
// Redis: an unreachable URL must yield an in-memory store instead of erroring.
func TestNewCacheRedisFallsBack(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	c, err := NewCache(context.Background(), "redis://127.0.0.1:1", logger)
	if err != nil {
		t.Fatalf("NewCache: %v", err)
	}
	defer c.Close()
	if _, ok := c.(*MemoryCache); !ok {
		t.Fatalf("expected fallback MemoryCache, got %T", c)
	}
}
