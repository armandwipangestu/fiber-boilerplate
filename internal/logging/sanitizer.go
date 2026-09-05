package logging

import (
	"context"
	"log/slog"
	"regexp"
	"strings"
)

var (
	// Sensitive data patterns (PRD Section 41.2).
	reCreditCard    = regexp.MustCompile(`\b(?:4[0-9]{12}(?:[0-9]{3})?|5[1-5][0-9]{14}|3[47][0-9]{13}|6(?:011|5[0-9]{2})[0-9]{12})\b`)
	reCVV           = regexp.MustCompile(`\b[0-9]{3,4}\b`)
	reSSN           = regexp.MustCompile(`\b[0-9]{3}-[0-9]{2}-[0-9]{4}\b`)
	reJWT           = regexp.MustCompile(`\beyJ[A-Za-z0-9-_]*\.[A-Za-z0-9-_]*\.[A-Za-z0-9-_]*\b`)
	reRefreshToken  = regexp.MustCompile(`\b[0-9a-f]{40,}\b`)
	reAPIKey        = regexp.MustCompile(`\b(?:sk|pk|rk|ak)_[A-Za-z0-9]{16,}\b`)
	reBearer        = regexp.MustCompile(`(?i)\bBearer\s+[A-Za-z0-9._~+\-/:=]{20,}`)
	rePrivateKey    = regexp.MustCompile(`-----BEGIN (?:RSA |EC |OPENSSH )?PRIVATE KEY-----.*?-----END (?:RSA |EC |OPENSSH )?PRIVATE KEY-----`)
	reDatabaseURL   = regexp.MustCompile(`(?i)([a-z0-9]+://)[^@/\s]+(:[^@/\s]+)?@`)
	rePasswordField = regexp.MustCompile(`(?i)(password|passwd|pwd|secret|token|api.?key|authorization|refresh.?token|access.?token)(["']?\s*[:=]\s*["']?)[^"'\s,}\]]+`)
	reEmailToken    = regexp.MustCompile(`(?i)[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}`)
)

// SanitizeString applies every masking pattern to a raw string, replacing
// matches with a redaction token.
func SanitizeString(s string) string {
	out := s
	// Order matters: run the specific high-precision patterns BEFORE the
	// broad 3-4 digit CVV pattern (which would otherwise mangle SSNs).
	out = reSSN.ReplaceAllString(out, "[REDACTED]")
	out = reCreditCard.ReplaceAllString(out, "[REDACTED]")
	out = reAPIKey.ReplaceAllString(out, "[REDACTED]")
	out = reJWT.ReplaceAllString(out, "[REDACTED]")
	out = rePrivateKey.ReplaceAllString(out, "[REDACTED]")
	out = reBearer.ReplaceAllString(out, "[REDACTED]")
	out = reDatabaseURL.ReplaceAllString(out, "${1}[REDACTED]@")
	out = rePasswordField.ReplaceAllString(out, "${1}${2}[REDACTED]")
	out = reRefreshToken.ReplaceAllString(out, "[REDACTED]")
	out = reCVV.ReplaceAllString(out, "[REDACTED]")
	return out
}

// NewSanitizeHandler wraps an inner slog.Handler and masks sensitive
// values found in the message and in string attributes.
func NewSanitizeHandler(inner slog.Handler) slog.Handler {
	return &sanitizeHandler{inner: inner}
}

type sanitizeHandler struct {
	inner slog.Handler
}

func (h *sanitizeHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.inner.Enabled(ctx, level)
}

func (h *sanitizeHandler) Handle(ctx context.Context, r slog.Record) error {
	r.Message = SanitizeString(r.Message)

	attrs := make([]slog.Attr, 0, r.NumAttrs())
	r.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, sanitizeAttr(a))
		return true
	})

	// slog.Record cannot be mutated after creation, so we rebuild it.
	rebuilt := slog.NewRecord(r.Time, r.Level, r.Message, r.PC)
	rebuilt.AddAttrs(attrs...)

	return h.inner.Handle(ctx, rebuilt)
}

func (h *sanitizeHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	sanitized := make([]slog.Attr, len(attrs))
	for i, a := range attrs {
		sanitized[i] = sanitizeAttr(a)
	}
	return &sanitizeHandler{inner: h.inner.WithAttrs(sanitized)}
}

func (h *sanitizeHandler) WithGroup(name string) slog.Handler {
	return &sanitizeHandler{inner: h.inner.WithGroup(name)}
}

func sanitizeAttr(a slog.Attr) slog.Attr {
	if a.Value.Kind() == slog.KindString {
		a.Value = slog.StringValue(SanitizeString(a.Value.String()))
	}
	if a.Value.Kind() == slog.KindGroup {
		group := a.Value.Group()
		for i, g := range group {
			group[i] = sanitizeAttr(g)
		}
		a.Value = slog.GroupValue(group...)
	}
	return a
}

// SanitizeQuery is a convenience for one-off redaction of a whole string
// (e.g. URL strings, stack traces).
func SanitizeQuery(s string) string {
	return strings.TrimSpace(SanitizeString(s))
}
