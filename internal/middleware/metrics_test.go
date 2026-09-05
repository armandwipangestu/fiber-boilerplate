package middleware

import (
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus"
	io_prometheus_client "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricsMiddleware_RecordsAndNormalizes(t *testing.T) {
	app := fiber.New()
	app.Use(NewMetricsMiddleware())
	app.Get("/ping", func(c *fiber.Ctx) error { return c.SendStatus(http.StatusOK) })
	app.Get("/users/:id", func(c *fiber.Ctx) error { return c.SendStatus(http.StatusOK) })

	for i := 0; i < 2; i++ {
		resp, err := app.Test(mustRequest(t, "/ping"))
		require.NoError(t, err)
		resp.Body.Close()
	}
	resp, err := app.Test(mustRequest(t, "/users/123e4567-e89b-12d3-a456-426614174000"))
	require.NoError(t, err)
	resp.Body.Close()

	families, err := prometheus.DefaultGatherer.Gather()
	require.NoError(t, err)
	counter := counterForLabels(t, families, map[string]string{"method": "GET", "path": "/users/:id", "status": "200"})
	require.NotNil(t, counter, "expected normalized /users/:id metric")
	assert.Equal(t, float64(1), counter.GetValue())

	ping := counterForLabels(t, families, map[string]string{"method": "GET", "path": "/ping", "status": "200"})
	require.NotNil(t, ping)
	assert.Equal(t, float64(2), ping.GetValue())
}

func counterForLabels(t *testing.T, families []*io_prometheus_client.MetricFamily, want map[string]string) *io_prometheus_client.Counter {
	t.Helper()
	for _, mf := range families {
		if mf.GetName() != "http_requests_total" {
			continue
		}
		for _, m := range mf.GetMetric() {
			labels := map[string]string{}
			for _, lp := range m.GetLabel() {
				labels[lp.GetName()] = lp.GetValue()
			}
			match := true
			for k, v := range want {
				if labels[k] != v {
					match = false
					break
				}
			}
			if match {
				return m.GetCounter()
			}
		}
	}
	return nil
}
