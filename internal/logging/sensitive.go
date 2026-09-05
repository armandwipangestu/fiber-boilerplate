package logging

// SensitiveString is a string wrapper that NEVER exposes its value through
// String() or JSON marshaling — it renders as "[REDACTED]". Use it when you
// must keep a secret (API key, refresh token, DB password) in a struct that
// may be logged.
//
// Retrieve the original value with Value(); to avoid leaks, never pass a
// SensitiveString into a logger directly.
type SensitiveString struct {
	value string
}

func NewSensitiveString(value string) SensitiveString {
	return SensitiveString{value: value}
}

// String always returns "[REDACTED]".
func (s SensitiveString) String() string {
	return "[REDACTED]"
}

// GoString always returns "[REDACTED]" (used by %#v).
func (s SensitiveString) GoString() string {
	return "[REDACTED]"
}

// MarshalJSON always returns the JSON string "[REDACTED]".
func (s SensitiveString) MarshalJSON() ([]byte, error) {
	return []byte(`"[REDACTED]"`), nil
}

// Value returns the underlying secret. It is intentionally not exposed via
// formatting/marshaling; only call this when you genuinely need the raw value.
func (s SensitiveString) Value() string {
	return s.value
}