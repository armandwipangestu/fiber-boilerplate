package logging

import (
	"context"
	"log/slog"
	"runtime"
)

// sourceKey is an internal context key used to skip a configurable number
// of goroutine call frames when resolving the caller location.
type sourceKey struct{}

// NewSourceHandler wraps an inner slog.Handler and attaches the caller's
// file, function, and line to every log record. skipFrames tells the
// handler how many stack frames to skip to reach the user code that
// actually called the logger.
func NewSourceHandler(inner slog.Handler, skipFrames int) slog.Handler {
	return &sourceHandler{
		inner:      inner,
		skipFrames: skipFrames,
	}
}

type sourceHandler struct {
	inner      slog.Handler
	skipFrames int
}

func (h *sourceHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.inner.Enabled(ctx, level)
}

func (h *sourceHandler) Handle(ctx context.Context, r slog.Record) error {
	if pc, file, line, ok := runtime.Caller(h.skipFrames); ok {
		fn := runtime.FuncForPC(pc)
		f := runtime.FuncForPC(pc)
		if fn != nil {
			r.AddAttrs(
				slog.String("file", file),
				slog.String("function", f.Name()),
				slog.Int("line", line),
			)
		}
	}
	return h.inner.Handle(ctx, r)
}

func (h *sourceHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &sourceHandler{
		inner:      h.inner.WithAttrs(attrs),
		skipFrames: h.skipFrames,
	}
}

func (h *sourceHandler) WithGroup(name string) slog.Handler {
	return &sourceHandler{
		inner:      h.inner.WithGroup(name),
		skipFrames: h.skipFrames,
	}
}

// contextSourceCalls stores the number of frames to skip in the context,
// letting tests inject an accurate skip value.
func contextSourceCalls(ctx context.Context, skip int) context.Context {
	return context.WithValue(ctx, sourceKey{}, skip)
}

func sourceSkipFromContext(ctx context.Context) (int, bool) {
	if v := ctx.Value(sourceKey{}); v != nil {
		return v.(int), true
	}
	return 0, false
}