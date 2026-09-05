package cache

import (
	"context"
	"errors"
	"time"
)

// ErrCacheMiss is returned by Get when the key does not exist or has expired.
var ErrCacheMiss = errors.New("cache miss")

// Cache is a TTL key/value store used for memoized lookups (e.g. RBAC
// permissions). Values are caller-encoded bytes (typically JSON).
type Cache interface {
	// Get returns the stored value, or ErrCacheMiss.
	Get(ctx context.Context, key string) ([]byte, error)
	// Set stores value until ttl elapses. A non-positive ttl means no
	// expiration.
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	// Delete removes a single key; deleting a missing key is a no-op.
	Delete(ctx context.Context, key string) error
	// DeletePattern removes every key matching the glob pattern.
	DeletePattern(ctx context.Context, pattern string) error
	// Close releases any resources. Safe to call multiple times.
	Close() error
}
