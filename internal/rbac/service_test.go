package rbac

import "testing"

func TestServiceInterface_CompileTime(t *testing.T) {
	// Ensures the in-memory cache satisfies the Cache contract.
	var _ Cache = NewInMemoryCache()
}
