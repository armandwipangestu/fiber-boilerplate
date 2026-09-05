# Phase 16 — OpenTelemetry (OTel) Tracing

> Distributed tracing with W3C `traceparent` propagation, OTLP/HTTP export, and
> log/trace correlation (slog fields `trace_id`/`span_id`).

## Deliverables

### 16.1 Tracing setup (`internal/tracing`)
- `tracing.Init(serviceName, endpoint, sampleRate)` configures an OTLP HTTP
  exporter and a `sdktrace.TracerProvider`, installs it as the OpenTelemetry
  global along with `TraceContext` + `Baggage` propagators.
- Sampler: `>1` = always, `<=0` = never, else ratio-based.
- `tracing.Shutdown(ctx, provider)` flushes pending spans, drains on exit.

### 16.2 Tracing middleware (`internal/middleware/tracing.go`)
- Starts a server span per request (`METHOD /route`), extracts incoming trace
  context from `traceparent` headers.
- Records HTTP semantic attributes (`http.request.method`, `http.route`,
  `url.full`, `client.address`, `user_agent.original`,
  `http.response.status_code`).
- Marks the span as errored on 5xx and records handler errors.
- Stores `trace_id`/`span_id` into Fiber locals; the logging middleware emits
  them in every request log line → trace-to-log correlation in the collector.

### 16.3 Wiring
- `cmd/server/main.go` initializes tracing when `OTEL_ENABLED=true` (fail-soft:
  logs and continues without tracing on error); provider shutdown is deferred.
- Enabled off by default; `.env.example` carries `OTEL_ENABLED`,
  `OTEL_ENDPOINT` (default `http://localhost:4318`), `OTEL_SAMPLE_RATE`.

## Verify
- Point the app at an OTLP HTTP collector (`docker compose` Collector, Phase 22)
  with `OTEL_ENABLED=true`; every request appears as a span and logged entries
  carry the same `trace_id`.