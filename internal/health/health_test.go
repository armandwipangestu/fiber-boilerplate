package health_test

import (
	"database/sql"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/armandwipangestu/fiber-boilerplate/internal/health"
)

func TestLive_Always200(t *testing.T) {
	h := health.NewHandler(nil, nil)
	app := fiber.New()
	app.Get("/health/live", h.Live)

	resp, err := app.Test(closedDBRequest("/health/live"))
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()
}

func TestReady_UnhealthyWithoutDB(t *testing.T) {
	h := health.NewHandler(nil, nil)
	app := fiber.New()
	app.Get("/health/ready", h.Ready)

	resp, err := app.Test(closedDBRequest("/health/ready"))
	require.NoError(t, err)
	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
	resp.Body.Close()
}

func TestReady_DBUnreachable(t *testing.T) {
	db, err := sql.Open("pgx", "postgres://postgres:postgres@localhost:1/nope?sslmode=disable")
	require.NoError(t, err)
	defer db.Close()

	h := health.NewHandler(db, nil)
	app := fiber.New()
	app.Get("/health/ready", h.Ready)

	resp, err := app.Test(closedDBRequest("/health/ready"))
	require.NoError(t, err)
	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
	resp.Body.Close()
}

func TestReady_WithBrokenRedis(t *testing.T) {
	db, err := sql.Open("pgx", "postgres://postgres:postgres@localhost:1/nope?sslmode=disable")
	require.NoError(t, err)
	defer db.Close()

	rdb := redis.NewClient(&redis.Options{Addr: "localhost:1"})
	defer rdb.Close()

	h := health.NewHandler(db, rdb)
	app := fiber.New()
	app.Get("/health/ready", h.Ready)

	resp, err := app.Test(closedDBRequest("/health/ready"))
	require.NoError(t, err)
	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
	resp.Body.Close()
}

func closedDBRequest(path string) *http.Request {
	req, _ := http.NewRequest(http.MethodGet, path, nil)
	return req
}
