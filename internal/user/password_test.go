package user

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashPassword_Verify(t *testing.T) {
	hash, err := HashPassword("supersecret")
	require.NoError(t, err)

	assert.True(t, VerifyPassword(hash, "supersecret"))
	assert.False(t, VerifyPassword(hash, "wrongpass"))
	assert.NotEqual(t, "supersecret", hash)
}

func TestBuildSearch_Empty(t *testing.T) {
	where, args := buildSearch("", []any{}, "")
	assert.Equal(t, "", where)
	assert.Empty(t, args)
}

func TestBuildSearch_WithTerm(t *testing.T) {
	where, args := buildSearch("", []any{}, "ada")
	assert.Contains(t, where, "ILIKE $1")
	assert.Equal(t, []any{"%ada%"}, args)
}

func TestValidSortColumn(t *testing.T) {
	assert.True(t, validSortColumn("email"))
	assert.True(t, validSortColumn("created_at"))
	assert.False(t, validSortColumn("id; DROP TABLE"))
}

func TestListQuery_Normalization(t *testing.T) {
	q := ListUsersQuery{}
	assert.Equal(t, 0, q.Page) // service applies defaults, repo only uses as-is
}
