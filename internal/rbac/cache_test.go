package rbac

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCache_GetSetPermissions(t *testing.T) {
	c := NewInMemoryCache()
	assert.Nil(t, c.GetPermissions("u1"))

	c.SetPermissions("u1", []string{"users.view"})
	assert.Equal(t, []string{"users.view"}, c.GetPermissions("u1"))
}

func TestCache_Invalidate(t *testing.T) {
	c := NewInMemoryCache()
	c.SetPermissions("u1", []string{"users.view"})
	c.Invalidate("u1")
	assert.Nil(t, c.GetPermissions("u1"))
}

func TestCache_InvalidateAll(t *testing.T) {
	c := NewInMemoryCache()
	c.SetPermissions("u1", []string{"users.view"})
	c.SetRoles("u2", []string{"admin"})
	c.InvalidateAll()
	assert.Nil(t, c.GetPermissions("u1"))
	assert.Nil(t, c.GetRoles("u2"))
}
