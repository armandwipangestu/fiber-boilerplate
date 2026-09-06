# Phase 18 — Cache Layer

> Generic TTL cache with interchangeable backends. Redis when configured and
> reachable, in-memory otherwise — enabling shared memoized lookups across
> multiple instances.

## Deliverables

### 18.1 Cache interface (`internal/cache/cache.go`)
- `Cache`: `Get(ctx, key) ([]byte, error)`, `Set(ctx, key, value, ttl)`,
  `Delete(ctx, key)`, `DeletePattern(ctx, pattern)`, `Close()`.
- `ErrCacheMiss` sentinel for missing/expired keys.

### 18.2 In-memory cache (`internal/cache/memory.go`)
- `MemoryCache`: `map` guarded by `sync.RWMutex`, per-entry `expiresAt`.
- Periodic cleanup goroutine (every 1 minute) sweeps expired entries; `Close`
  stops it (idempotent).

### 18.3 Redis cache (`internal/cache/redis.go`)
- `RedisCache` wraps `*redis.Client`; `Get` maps `redis.Nil` → `ErrCacheMiss`;
  `Set` uses native TTL; `DeletePattern` uses `SCAN`/`DEL` (non-blocking).

### 18.4 Factory (`internal/cache/factory.go`)
- `NewCache(ctx, redisURL, logger)`: empty URL → memory; reachable URL → Redis;
  unreachable Redis → memory with a warning. Backend choice is logged.

### 18.5 Integration
- RBAC cache now sits on the generic store (`rbac.NewCache(store)`) — keys
  `rbac:permissions:*` / `rbac:roles:*`, JSON-encoded slices, 5-min TTL,
  `InvalidateAll` uses `DeletePattern`. `NewInMemoryCache` is retained for
  tests.
- `cmd/app/main.go` builds the store from `REDIS_URL` and closes it during
  graceful shutdown.

## Verify
- Memory unit tests: round-trip, TTL expiry, delete, pattern delete, idempotent
  close, factory fallback.
- Live Redis integration test (runs when a Redis is reachable, skips
  otherwise): round-trip, TTL, pattern delete.