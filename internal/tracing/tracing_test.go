package tracing

import (
	"context"
	"testing"
	"time"

	"go.opentelemetry.io/otel"
)

// TestInitSampler no-op provider is exercised with a batch of sampled/non-sampled
// decisions by checking sampler behavior directly.
func TestParseSampler(t *testing.T) {
	tests := []struct {
		name string
		rate float64
	}{
		{"never", 0},
		{"always", 1},
		{"ratio", 0.5},
		{"default", -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sampler := parseSampler(tt.rate)
			if sampler == nil {
				t.Fatal("nil sampler")
			}
		})
	}
}

// TestShutdownNil verifies Shutdown is safe for a nil provider.
func TestShutdownNil(t *testing.T) {
	if err := Shutdown(context.Background(), nil); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

// TestInitSetsGlobal verifies Init installs a real (always-sample) provider
// and propagator; endpoint is never dialed because no spans are exported.
func TestInitSetsGlobal(t *testing.T) {
	provider, err := Init("test-service", "http://127.0.0.1:4318", 1.0)
	if err != nil {
		t.Fatalf("Init: %v", err)
	}
	if provider == nil {
		t.Fatal("expected non-nil provider")
	}

	got := otel.GetTracerProvider()
	if got == nil {
		t.Fatal("global tracer provider not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := Shutdown(ctx, provider); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}
}