package console

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/armandwipangestu/fiber-boilerplate/internal/config"
	"github.com/armandwipangestu/fiber-boilerplate/internal/database"
	"github.com/redis/go-redis/v9"
)

// ConfigCheck loads the full configuration (validating DATABASE_URL and
// JWT_SECRET), prints every resolved setting with secrets masked, and probes
// the configured database and cache. It fails when the configuration is
// invalid or a configured dependency is unreachable.
func ConfigCheck() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("configuration invalid: %w", err)
	}

	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "SETTING\tVALUE")
	rows := []struct{ key, value string }{
		{"APP_NAME", cfg.AppName},
		{"APP_ENV", cfg.AppEnv},
		{"APP_HOST", cfg.AppHost},
		{"APP_PORT", fmt.Sprintf("%d", cfg.AppPort)},
		{"SHUTDOWN_TIMEOUT", cfg.ShutdownTimeout.String()},
		{"DATABASE_DRIVER", cfg.DatabaseDriver},
		{"DATABASE_URL", maskURL(cfg.DatabaseURL)},
		{"DATABASE_MAX_OPEN_CONNS", fmt.Sprintf("%d", cfg.DatabaseMaxOpenConns)},
		{"REDIS_URL", maskURL(cfg.RedisURL)},
		{"STORAGE_PATH", cfg.StoragePath},
		{"PUBLIC_URL", cfg.PublicURL},
		{"S3_REGION", cfg.S3Region},
		{"S3_BUCKET", cfg.S3Bucket},
		{"S3_ENDPOINT", cfg.S3Endpoint},
		{"S3_ACCESS_KEY_ID", maskSecret(cfg.S3AccessKeyID)},
		{"S3_SECRET_ACCESS_KEY", maskSecret(cfg.S3SecretAccessKey)},
		{"JWT_SECRET", maskSecret(cfg.JWTSecret)},
		{"JWT_ACCESS_TOKEN_EXPIRY", cfg.JWTAccessTokenExpiry.String()},
		{"JWT_REFRESH_TOKEN_EXPIRY", cfg.JWTRefreshTokenExpiry.String()},
		{"RATE_LIMIT_ENABLED", fmt.Sprintf("%t", cfg.RateLimitEnabled)},
		{"CORS_ALLOWED_ORIGINS", strings.Join(cfg.CORSAllowedOrigins, ",")},
		{"LOG_LEVEL", cfg.LogLevel},
		{"LOG_OUTPUT", cfg.LogOutput},
		{"OTEL_ENABLED", fmt.Sprintf("%t", cfg.OTELEnabled)},
		{"SWAGGER_ENABLED", fmt.Sprintf("%t", cfg.SwaggerEnabled)},
		{"DEFAULT_ADMIN_EMAIL", cfg.DefaultAdminEmail},
		{"DEFAULT_ADMIN_PASSWORD", maskSecret(cfg.DefaultAdminPassword)},
	}
	for _, r := range rows {
		fmt.Fprintf(w, "%s\t%s\n", r.key, r.value)
	}
	if err := w.Flush(); err != nil {
		return err
	}

	var problems []string
	if err := probeDatabase(*cfg); err != nil {
		problems = append(problems, err.Error())
	}
	if err := probeRedis(*cfg); err != nil {
		problems = append(problems, err.Error())
	}

	for _, p := range problems {
		printf("FAIL: %s\n", p)
	}
	printf("check: %s\n", ok(len(problems) == 0))
	if len(problems) > 0 {
		return fmt.Errorf("config:check found %d problem(s)", len(problems))
	}
	return nil
}

func probeDatabase(cfg config.Config) error {
	db, err := database.NewDatabase(cfg)
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("database unreachable: %w", err)
	}
	printf("database: ok\n")
	return nil
}

func probeRedis(cfg config.Config) error {
	if cfg.RedisURL == "" {
		printf("redis: not configured\n")
		return nil
	}
	opts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		return fmt.Errorf("redis url invalid: %w", err)
	}
	client := redis.NewClient(opts)
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis unreachable: %w", err)
	}
	printf("redis: ok\n")
	return nil
}

func ok(good bool) string {
	if good {
		return "ok"
	}
	return "failed"
}

func maskSecret(s string) string {
	if s == "" {
		return "—"
	}
	return "******"
}

// maskURL redacts credentials embedded in a URL while keeping the layout.
func maskURL(raw string) string {
	if raw == "" {
		return "—"
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return maskSecret(raw)
	}
	if parsed.User != nil {
		parsed.User = url.User("****")
	}
	return parsed.String()
}
