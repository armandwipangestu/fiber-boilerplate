package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_DefaultValues(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost:5432/test")
	t.Setenv("JWT_SECRET", "test-secret")

	cfg, err := Load()

	require.NoError(t, err)
	assert.Equal(t, "fiber-boilerplate", cfg.AppName)
	assert.Equal(t, "development", cfg.AppEnv)
	assert.Equal(t, 8080, cfg.AppPort)
	assert.Equal(t, "0.0.0.0", cfg.AppHost)
	assert.Equal(t, 30*time.Second, cfg.ShutdownTimeout)
	assert.Equal(t, "postgres", cfg.DatabaseDriver)
	assert.Equal(t, 25, cfg.DatabaseMaxOpenConns)
	assert.Equal(t, 10, cfg.DatabaseMaxIdleConns)
	assert.Equal(t, 5*time.Minute, cfg.DatabaseConnMaxLifetime)
	assert.Equal(t, 3*time.Minute, cfg.DatabaseConnMaxIdleTime)
	assert.Equal(t, true, cfg.RateLimitEnabled)
	assert.Equal(t, 100, cfg.RateLimitRequests)
	assert.Equal(t, 60*time.Second, cfg.RateLimitExpiration)
	assert.Equal(t, []string{"*"}, cfg.CORSAllowedOrigins)
	assert.Equal(t, "info", cfg.LogLevel)
	assert.Equal(t, "stdout", cfg.LogOutput)
	assert.Equal(t, 15*time.Minute, cfg.JWTAccessTokenExpiry)
	assert.Equal(t, 720*time.Hour, cfg.JWTRefreshTokenExpiry)
	assert.Equal(t, false, cfg.OTELEnabled)
	assert.Equal(t, true, cfg.SwaggerEnabled)
}

func TestLoad_EnvironmentOverride(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost:5432/test")
	t.Setenv("JWT_SECRET", "test-secret")
	t.Setenv("APP_NAME", "my-app")
	t.Setenv("APP_PORT", "9090")
	t.Setenv("APP_ENV", "production")

	cfg, err := Load()

	require.NoError(t, err)
	assert.Equal(t, "my-app", cfg.AppName)
	assert.Equal(t, 9090, cfg.AppPort)
	assert.Equal(t, "production", cfg.AppEnv)
}

func TestLoad_MissingDatabaseURL(t *testing.T) {
	os.Unsetenv("DATABASE_URL")
	t.Setenv("JWT_SECRET", "test-secret")

	_, err := Load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "DATABASE_URL is required")
}

func TestLoad_MissingJWTSecret(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost:5432/test")
	os.Unsetenv("JWT_SECRET")

	_, err := Load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "JWT_SECRET is required")
}

func TestLoad_DurationParsing(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost:5432/test")
	t.Setenv("JWT_SECRET", "test-secret")
	t.Setenv("SHUTDOWN_TIMEOUT", "15s")
	t.Setenv("DATABASE_CONN_MAX_LIFETIME", "10m")

	cfg, err := Load()

	require.NoError(t, err)
	assert.Equal(t, 15*time.Second, cfg.ShutdownTimeout)
	assert.Equal(t, 10*time.Minute, cfg.DatabaseConnMaxLifetime)
}

func TestLoad_CSVSlice(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost:5432/test")
	t.Setenv("JWT_SECRET", "test-secret")
	t.Setenv("CORS_ALLOWED_ORIGINS", "http://a.com,http://b.com")

	cfg, err := Load()

	require.NoError(t, err)
	assert.Equal(t, []string{"http://a.com", "http://b.com"}, cfg.CORSAllowedOrigins)
}
