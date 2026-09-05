package cache

import (
	"context"
	"path"
	"sync"
	"time"
)

type memoryEntry struct {
	value     []byte
	expiresAt time.Time
}

// MemoryCache is an in-process TTL store with periodic cleanup of expired
// entries. Suitable for single-instance deployments.
type MemoryCache struct {
	mu        sync.RWMutex
	entries   map[string]memoryEntry
	done      chan struct{}
	closeOnce sync.Once
}

// NewMemoryCache builds a MemoryCache and starts its background cleanup
// goroutine, which sweeps expired entries every minute.
func NewMemoryCache() *MemoryCache {
	c := &MemoryCache{
		entries: make(map[string]memoryEntry),
		done:    make(chan struct{}),
	}
	go c.cleanup()
	return c
}

func (c *MemoryCache) Get(_ context.Context, key string) ([]byte, error) {
	c.mu.RLock()
	e, ok := c.entries[key]
	c.mu.RUnlock()
	if !ok {
		return nil, ErrCacheMiss
	}
	if !e.expiresAt.IsZero() && time.Now().After(e.expiresAt) {
		c.mu.Lock()
		delete(c.entries, key)
		c.mu.Unlock()
		return nil, ErrCacheMiss
	}
	return e.value, nil
}

func (c *MemoryCache) Set(_ context.Context, key string, value []byte, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	e := memoryEntry{value: value}
	if ttl > 0 {
		e.expiresAt = time.Now().Add(ttl)
	}
	c.entries[key] = e
	return nil
}

func (c *MemoryCache) Delete(_ context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.entries, key)
	return nil
}

func (c *MemoryCache) DeletePattern(_ context.Context, pattern string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	for key := range c.entries {
		if ok, _ := path.Match(pattern, key); ok {
			delete(c.entries, key)
		}
	}
	return nil
}

func (c *MemoryCache) Close() error {
	c.closeOnce.Do(func() { close(c.done) })
	return nil
}

// cleanup periodically drops expired entries.
func (c *MemoryCache) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-c.done:
			return
		case now := <-ticker.C:
			c.mu.Lock()
			for key, e := range c.entries {
				if !e.expiresAt.IsZero() && now.After(e.expiresAt) {
					delete(c.entries, key)
				}
			}
			c.mu.Unlock()
		}
	}
}
