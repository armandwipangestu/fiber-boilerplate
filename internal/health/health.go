package health

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

// CheckResult is the outcome of pinging a single dependency.
type CheckResult struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

// Handler reports liveness and readiness of the service and its dependencies.
type Handler struct {
	db    *sql.DB
	redis *redis.Client
}

// NewHandler wires a Handler to the database. Pass a nil redis client when
// Redis is not configured.
func NewHandler(db *sql.DB, redis *redis.Client) *Handler {
	return &Handler{db: db, redis: redis}
}

// Live always reports healthy (process is up).
func (h *Handler) Live(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"status": "ok"})
}

// Ready pings every dependency and returns 200 only when all are healthy.
func (h *Handler) Ready(c *fiber.Ctx) error {
	results := h.checkAll(c.Context())
	if allHealthy(results) {
		return c.JSON(healthResponse("ready", results))
	}
	return c.Status(fiber.StatusServiceUnavailable).JSON(healthResponse("not ready", results))
}

// Health is an alias for Ready used by orchestrators expecting /health.
func (h *Handler) Health(c *fiber.Ctx) error {
	return h.Ready(c)
}

func (h *Handler) checkAll(ctx context.Context) []CheckResult {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	dbErr := errors.New("database not configured")
	if h.db != nil {
		dbErr = h.db.PingContext(ctx)
	}
	results := []CheckResult{{
		Name:   "database",
		Status: statusOf(dbErr),
		Error:  errText(dbErr),
	}}

	if h.redis != nil {
		redisErr := h.redis.Ping(ctx).Err()
		results = append(results, CheckResult{
			Name:   "redis",
			Status: statusOf(redisErr),
			Error:  errText(redisErr),
		})
	}

	return results
}

func statusOf(err error) string {
	if err != nil {
		return "unhealthy"
	}
	return "ok"
}

func errText(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func allHealthy(results []CheckResult) bool {
	for _, r := range results {
		if r.Status != "ok" {
			return false
		}
	}
	return true
}

func healthResponse(status string, results []CheckResult) fiber.Map {
	return fiber.Map{"status": status, "checks": results}
}
