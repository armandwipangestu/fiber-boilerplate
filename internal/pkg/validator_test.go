package pkg

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type validationFixture struct {
	Email    string `json:"email" validate:"required,email"`
	Name     string `json:"name" validate:"required,min=2,max=100"`
	Password string `json:"password" validate:"required,min=8"`
	Role     string `json:"role" validate:"oneof=admin user"`
	ID       string `json:"id" validate:"uuid"`
}

func TestValidateStruct_Valid(t *testing.T) {
	v := NewValidator()
	fields, err := v.ValidateStruct(validationFixture{
		Email:    "ada@example.com",
		Name:     "Ada",
		Password: "supersecret",
		Role:     "admin",
		ID:       "4d4fdb0f-6adc-4822-8176-dbde5999bb72",
	})
	require.NoError(t, err)
	assert.Empty(t, fields)
}

func TestValidateStruct_FieldErrors(t *testing.T) {
	v := NewValidator()
	fields, err := v.ValidateStruct(validationFixture{
		Email:    "not-an-email",
		Name:     "A",
		Password: "short",
		Role:     "superuser",
		ID:       "nope",
	})
	require.Error(t, err)
	assert.NotEmpty(t, fields["email"])
	assert.Equal(t, "must be a valid email address", fields["email"][0])
	assert.Contains(t, fields["name"][0], "at least 2")
	assert.Contains(t, fields["password"][0], "at least 8")
	assert.Contains(t, fields["role"][0], "admin")
	assert.Contains(t, fields["id"][0], "UUID")
}

func TestValidateStruct_JSONFieldNames(t *testing.T) {
	v := NewValidator()
	fields, err := v.ValidateStruct(struct {
		PerPage int `json:"per_page" validate:"min=1,max=100"`
	}{PerPage: 0})
	require.Error(t, err)
	_, ok := fields["per_page"]
	assert.True(t, ok, "field should be keyed by json tag name")
}
