package logging

import (
	"io"
	"log/slog"
	"os"

	"github.com/armandwipangestu/fiber-boilerplate/internal/config"
	"gopkg.in/natefinch/lumberjack.v2"
)

// NewLogger builds the application logger. Output is controlled by
// cfg.LogOutput ("stdout", "file", "both") and level by cfg.LogLevel.
// In non-development environments the log output is sanitized to
// prevent sensitive data (tokens, credentials, card numbers) leakage.
func NewLogger(cfg config.Config) *slog.Logger {
	writer := newOutputWriter(cfg)

	level := parseLevel(cfg.LogLevel)
	baseOpts := &slog.HandlerOptions{Level: level}

	var handler slog.Handler
	if cfg.AppEnv == "development" {
		handler = slog.NewTextHandler(writer, baseOpts)
	} else {
		handler = slog.NewJSONHandler(writer, baseOpts)
	}

	// Source handler adds file/function/line to every record.
	handler = NewSourceHandler(handler, 3)

	// Sanitize sensitive data in non-development environments.
	if cfg.AppEnv != "development" {
		handler = NewSanitizeHandler(handler)
	}

	return slog.New(handler).With("service", cfg.AppName)
}

// newOutputWriter chooses the writer(s) based on cfg.LogOutput.
func newOutputWriter(cfg config.Config) io.Writer {
	switch cfg.LogOutput {
	case "file":
		return newRotatingWriter(cfg)
	case "both":
		return io.MultiWriter(os.Stdout, newRotatingWriter(cfg))
	default:
		return os.Stdout
	}
}

// newRotatingWriter returns a lumberjack-based size-rotated file writer.
func newRotatingWriter(cfg config.Config) io.Writer {
	return &lumberjack.Logger{
		Filename:   cfg.LogFilePath,
		MaxSize:    cfg.LogMaxSizeMB, // MB
		MaxBackups: cfg.LogMaxBackups,
		MaxAge:     cfg.LogMaxAgeDays, // days
		Compress:   cfg.LogCompress,
	}
}

func parseLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
