package tracing

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.27.0"
)

// Init configures an OTLP HTTP tracer provider, exports spans to the given
// endpoint, and sets it (plus the text map propagator) as the OpenTelemetry
// global. Returns the provider so callers can Shutdown it during graceful
// shutdown.
func Init(serviceName, endpoint string, sampleRate float64) (*sdktrace.TracerProvider, error) {
	exporter, err := otlptracehttp.New(context.Background(), otlptracehttp.WithEndpointURL(endpoint))
	if err != nil {
		return nil, err
	}

	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithSampler(parseSampler(sampleRate)),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
		)),
	)

	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return provider, nil
}

// Shutdown flushes any pending spans and stops the exporter. Idempotent for
// nil providers.
func Shutdown(ctx context.Context, provider *sdktrace.TracerProvider) error {
	if provider == nil {
		return nil
	}
	return provider.Shutdown(ctx)
}

func parseSampler(rate float64) sdktrace.Sampler {
	switch {
	case rate <= 0:
		return sdktrace.NeverSample()
	case rate >= 1:
		return sdktrace.AlwaysSample()
	default:
		return sdktrace.TraceIDRatioBased(rate)
	}
}

// ShutdownTimeout is the default budget the caller (graceful shutdown) grants
// the tracer provider to drain pending spans.
const ShutdownTimeout = 5 * time.Second
