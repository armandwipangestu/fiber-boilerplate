package rbac

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/armandwipangestu/fiber-boilerplate/internal/cache"
)

const (
	cacheTTL           = 5 * time.Minute
	permissionsPrefix  = "rbac:permissions:"
	rolesPrefix        = "rbac:roles:"
	permissionsPattern = permissionsPrefix + "*"
	rolesPattern       = rolesPrefix + "*"
)

// Cache stores per-user permission and role lookups.
type Cache interface {
	GetPermissions(userID string) []string
	SetPermissions(userID string, perms []string)
	GetRoles(userID string) []string
	SetRoles(userID string, roles []string)
	Invalidate(userID string)
	InvalidateAll()
}

// NewInMemoryCache builds a standalone in-process cache. Kept for tests and
// callers that do not need a shared store.
func NewInMemoryCache() Cache {
	return newStoreCache(cache.NewMemoryCache())
}

// NewCache builds an RBAC cache on top of a generic cache store (in-memory or
// Redis). Multi-instance deployments share the same memoized lookups.
func NewCache(store cache.Cache) Cache {
	return newStoreCache(store)
}

// storeCache adapts the generic cache.Cache to the RBAC query interface.
type storeCache struct {
	store cache.Cache
	ctx   context.Context
	log   *slog.Logger
}

func newStoreCache(store cache.Cache) *storeCache {
	return &storeCache{store: store, ctx: context.Background(), log: slog.Default()}
}

func (c *storeCache) GetPermissions(userID string) []string {
	return c.get(permissionsPrefix+userID, "permissions")
}

func (c *storeCache) SetPermissions(userID string, perms []string) {
	c.set(permissionsPrefix+userID, perms)
}

func (c *storeCache) GetRoles(userID string) []string {
	return c.get(rolesPrefix+userID, "roles")
}

func (c *storeCache) SetRoles(userID string, roles []string) {
	c.set(rolesPrefix+userID, roles)
}

func (c *storeCache) Invalidate(userID string) {
	_ = c.store.Delete(c.ctx, permissionsPrefix+userID)
	_ = c.store.Delete(c.ctx, rolesPrefix+userID)
}

func (c *storeCache) InvalidateAll() {
	_ = c.store.DeletePattern(c.ctx, permissionsPattern)
	_ = c.store.DeletePattern(c.ctx, rolesPattern)
}

func (c *storeCache) get(key, what string) []string {
	raw, err := c.store.Get(c.ctx, key)
	if err != nil {
		return nil
	}
	var out []string
	if err := json.Unmarshal(raw, &out); err != nil {
		c.log.Warn("corrupt cache entry", "what", what, "key", key, "error", err)
		_ = c.store.Delete(c.ctx, key)
		return nil
	}
	return out
}

func (c *storeCache) set(key string, value []string) {
	if value == nil {
		return
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return
	}
	_ = c.store.Set(c.ctx, key, raw, cacheTTL)
}
