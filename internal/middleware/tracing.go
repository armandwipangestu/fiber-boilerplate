package middleware

import (
	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"

	"github.com/armandwipangestu/fiber-boilerplate/internal/pkg"
)

const traceIDCtxKey = "trace_id"
const spanIDCtxKey = "span_id"

// NewTracingMiddleware starts a server span per request, propagates the
// incoming trace context from headers, and records the handler result. The
// trace/span ids are stored into Fiber locals so the logging middleware can
// correlate log entries with traces.
func NewTracingMiddleware() fiber.Handler {
	tracer := otel.Tracer("fiber-boilerplate")

	return func(c *fiber.Ctx) error {
		ctx := otel.GetTextMapPropagator().Extract(
			c.Context(),
			propagation.HeaderCarrier(c.GetReqHeaders()),
		)

		route := c.Route().Path

		ctx, span := tracer.Start(ctx, c.Method()+" "+route,
			trace.WithAttributes(
				attribute.String("http.request.method", c.Method()),
				attribute.String("http.route", route),
				attribute.String("url.full", c.BaseURL()+string(c.Request().RequestURI())),
				attribute.String("user_agent.original", c.Get("User-Agent")),
				attribute.String("client.address", c.IP()),
			),
		)
		ctx = trace.ContextWithSpan(ctx, span)
		spanCtx := span.SpanContext()
		if spanCtx.IsValid() {
			pkg.SetTraceID(c, spanCtx.TraceID().String())
			pkg.SetSpanID(c, spanCtx.SpanID().String())
		}

		c.SetUserContext(ctx)

		err := c.Next()

		span.SetAttributes(attribute.Int("http.response.status_code", c.Response().StatusCode()))
		if c.Response().StatusCode() >= 500 {
			span.SetStatus(codes.Error, "request failed")
		}
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		span.End()

		return err
	}
}