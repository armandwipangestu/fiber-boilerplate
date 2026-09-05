package logging

import (
	"bytes"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/armandwipangestu/fiber-boilerplate/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSourceHandler_AddsCallerInfo(t *testing.T) {
	var buf bytes.Buffer
	inner := slog.NewTextHandler(&buf, nil)
	handler := NewSourceHandler(inner, 3)

	logger := slog.New(handler)
	logger.Info("hello")

	out := buf.String()
	assert.Contains(t, out, "file=")
	assert.Contains(t, out, "function=")
	assert.Contains(t, out, "line=")
}

func TestSanitizeString_CreditCard(t *testing.T) {
	got := SanitizeString("4111111111111111")
	assert.Equal(t, "[REDACTED]", got)
}

func TestSanitizeString_CVV(t *testing.T) {
	got := SanitizeString(`{"cvv":"123"}`)
	assert.NotContains(t, got, "123")
	assert.Contains(t, got, "[REDACTED]")
}

func TestSanitizeString_SSN(t *testing.T) {
	got := SanitizeString("123-45-6789")
	assert.Equal(t, "[REDACTED]", got)
}

func TestSanitizeString_JWT(t *testing.T) {
	got := SanitizeString("eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxMjMifQ.sig")
	assert.Equal(t, "[REDACTED]", got)
}

func TestSanitizeString_PasswordField(t *testing.T) {
	got := SanitizeString(`password=supersecret123`)
	assert.NotContains(t, got, "supersecret123")
	assert.Contains(t, got, "[REDACTED]")
}

func TestSanitizeString_BearerToken(t *testing.T) {
	got := SanitizeString("Authorization: Bearer abc.def.ghi.jkl.mno")
	assert.Contains(t, got, "[REDACTED]")
	assert.NotContains(t, got, "Bearer abc.def")
}

func TestSanitizeString_DatabaseURL(t *testing.T) {
	got := SanitizeString("postgres://user:hunter2@localhost:5432/db")
	assert.NotContains(t, got, "hunter2")
	assert.Contains(t, got, "[REDACTED]")
}

func TestSanitizeHandler_MasksAttributes(t *testing.T) {
	var buf bytes.Buffer
	inner := slog.NewJSONHandler(&buf, nil)
	handler := NewSanitizeHandler(inner)

	logger := slog.New(handler)
	logger.Info("login", "token", "eyJhbGciOiJIUzI1NiJ9.xxx.yyy", "email", "test@example.com")

	out := buf.String()
	assert.NotContains(t, out, "eyJhbGciOiJIUzI1NiJ9.xxx.yyy")
	assert.Contains(t, out, "[REDACTED]")
}

func TestSensitiveString_HidesValue(t *testing.T) {
	s := NewSensitiveString("secret-token-value")

	assert.Equal(t, "[REDACTED]", s.String())
	require.NotEqual(t, "secret-token-value", s.String())

	data, err := s.MarshalJSON()
	require.NoError(t, err)
	assert.Equal(t, `"[REDACTED]"`, string(data))

	assert.Equal(t, "secret-token-value", s.Value())
}

func TestDailyRotator_WritesToDatedFile(t *testing.T) {
	cfg := configForTest(t)
	rotator := newDailyRotator(cfg)
	defer rotator.Close()

	n, err := rotator.Write([]byte("hello\n"))
	require.NoError(t, err)
	assert.Greater(t, n, 0)

	// First segment: ensure a file exists with today's date.
	files := listLogFiles(t, cfg)
	require.GreaterOrEqual(t, len(files), 1)
	assert.True(t, strings.Contains(files[0], logFilePrefix(cfg)))
}

func TestLogger_DevTextHandler(t *testing.T) {
	cfg := configForTest(t)
	cfg.AppEnv = "development"
	cfg.LogOutput = "stdout"

	logger := NewLogger(cfg)
	assert.NotNil(t, logger)
}
func configForTest(t *testing.T) config.Config {
	t.Helper()
	return config.Config{
		AppName:       "test-app",
		AppEnv:        "production",
		LogOutput:     "file",
		LogFilePath:   filepath.Join(t.TempDir(), "logs", "app.log"),
		LogMaxSizeMB:  1,
		LogMaxBackups: 1,
		LogMaxAgeDays: 1,
	}
}

func listLogFiles(t *testing.T, cfg config.Config) []string {
	t.Helper()
	dir := filepath.Dir(cfg.LogFilePath)
	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	files := make([]string, 0)
	for _, e := range entries {
		if !e.IsDir() {
			files = append(files, e.Name())
		}
	}
	return files
}

func logFilePrefix(cfg config.Config) string {
	base := filepath.Base(cfg.LogFilePath)
	ext := filepath.Ext(base)
	return base[:len(base)-len(ext)]
}
