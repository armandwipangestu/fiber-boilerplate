package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const namespace = "http"

var (
	// RequestsTotal counts every handled request by method, normalized path,
	// and response status.
	RequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "requests_total",
		Help:      "Total HTTP requests processed.",
	}, []string{"method", "path", "status"})

	// RequestDuration is a histogram of request latency in seconds.
	RequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: namespace,
		Name:      "request_duration_seconds",
		Help:      "HTTP request latency distribution in seconds.",
		Buckets:   prometheus.DefBuckets,
	}, []string{"method", "path"})

	// RequestsInFlight is the current number of in-flight requests.
	RequestsInFlight = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "requests_in_flight",
		Help:      "Current number of HTTP requests being processed.",
	})
)

// Register is a stability hook: metrics are auto-registered through
// promauto's default registerer. Kept explicit for tests.
func Register() {
	_ = prometheus.Register(RequestsTotal)
	_ = prometheus.Register(RequestDuration)
	_ = prometheus.Register(RequestsInFlight)
}

// Handler returns the Prometheus scrape handler.
func Handler() http.Handler {
	return promhttp.Handler()
}
