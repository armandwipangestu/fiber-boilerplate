package rbac

import (
	"sync"
	"time"
)

const (
	cacheTTL          = 5 * time.Minute
	permCacheKey      = "rbac:permissions"
	roleCacheKey      = "rbac:roles"
	permissionsPrefix = "rbac:permissions:"
	rolesPrefix       = "rbac:roles:"
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

type cacheEntry struct {
	value []string
	at    time.Time
}

type inMemoryCache struct {
	mu    sync.RWMutex
	perms map[string]cacheEntry
	roles map[string]cacheEntry
}

// NewInMemoryCache builds the default TTL cache. Phase 18 will layer Redis on
// top while keeping this interface.
func NewInMemoryCache() Cache {
	return &inMemoryCache{
		perms: make(map[string]cacheEntry),
		roles: make(map[string]cacheEntry),
	}
}

func (c *inMemoryCache) GetPermissions(userID string) []string {
	return c.get(c.perms, permissionsPrefix+userID)
}

func (c *inMemoryCache) SetPermissions(userID string, perms []string) {
	c.set(c.perms, permissionsPrefix+userID, perms)
}

func (c *inMemoryCache) GetRoles(userID string) []string {
	return c.get(c.roles, rolesPrefix+userID)
}

func (c *inMemoryCache) SetRoles(userID string, roles []string) {
	c.set(c.roles, rolesPrefix+userID, roles)
}

func (c *inMemoryCache) get(m map[string]cacheEntry, key string) []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	e, ok := m[key]
	if !ok || time.Since(e.at) > cacheTTL {
		return nil
	}
	return e.value
}

func (c *inMemoryCache) set(m map[string]cacheEntry, key string, value []string) {
	if value == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	m[key] = cacheEntry{value: value, at: time.Now()}
}

func (c *inMemoryCache) Invalidate(userID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.perms, permissionsPrefix+userID)
	delete(c.roles, rolesPrefix+userID)
}

func (c *inMemoryCache) InvalidateAll() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.perms = make(map[string]cacheEntry)
	c.roles = make(map[string]cacheEntry)
}
