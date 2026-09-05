package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppName          string
	AppEnv           string
	AppPort          int
	AppHost          string
	ShutdownTimeout  time.Duration

	DatabaseDriver          string
	DatabaseURL             string
	DatabaseMaxOpenConns    int
	DatabaseMaxIdleConns    int
	DatabaseConnMaxLifetime time.Duration
	DatabaseConnMaxIdleTime time.Duration

	RedisURL string

	JWTSecret             string
	JWTAccessTokenExpiry  time.Duration
	JWTRefreshTokenExpiry time.Duration

	RateLimitEnabled     bool
	RateLimitRequests    int
	RateLimitExpiration  time.Duration
	RateLimitStrategy    string

	CORSAllowedOrigins []string

	LogLevel       string
	LogOutput      string
	LogFilePath    string
	LogMaxSizeMB   int
	LogMaxBackups  int
	LogMaxAgeDays  int
	LogCompress    bool

	OTELEnabled    bool
	OTELEndpoint   string
	OTELSampleRate float64

	SwaggerEnabled bool
}

func Load() (*Config, error) {
	loadDotEnv()

	cfg := &Config{
		AppName:         getEnv("APP_NAME", "fiber-boilerplate"),
		AppEnv:          getEnv("APP_ENV", "development"),
		AppPort:         getEnvInt("APP_PORT", 8080),
		AppHost:         getEnv("APP_HOST", "0.0.0.0"),
		ShutdownTimeout: getEnvDuration("SHUTDOWN_TIMEOUT", 30*time.Second),

		DatabaseDriver:          getEnv("DATABASE_DRIVER", "postgres"),
		DatabaseURL:             getEnv("DATABASE_URL", ""),
		DatabaseMaxOpenConns:    getEnvInt("DATABASE_MAX_OPEN_CONNS", 25),
		DatabaseMaxIdleConns:    getEnvInt("DATABASE_MAX_IDLE_CONNS", 10),
		DatabaseConnMaxLifetime: getEnvDuration("DATABASE_CONN_MAX_LIFETIME", 5*time.Minute),
		DatabaseConnMaxIdleTime: getEnvDuration("DATABASE_CONN_MAX_IDLE_TIME", 3*time.Minute),

		RedisURL: getEnv("REDIS_URL", ""),

		JWTSecret:             getEnv("JWT_SECRET", ""),
		JWTAccessTokenExpiry:  getEnvDuration("JWT_ACCESS_TOKEN_EXPIRY", 15*time.Minute),
		JWTRefreshTokenExpiry: getEnvDuration("JWT_REFRESH_TOKEN_EXPIRY", 720*time.Hour),

		RateLimitEnabled:    getEnvBool("RATE_LIMIT_ENABLED", true),
		RateLimitRequests:   getEnvInt("RATE_LIMIT_REQUESTS", 100),
		RateLimitExpiration: getEnvDuration("RATE_LIMIT_EXPIRATION", 60*time.Second),
		RateLimitStrategy:   getEnv("RATE_LIMIT_STRATEGY", "ip"),

		CORSAllowedOrigins: getEnvSlice("CORS_ALLOWED_ORIGINS", "*"),

		LogLevel:      getEnv("LOG_LEVEL", "info"),
		LogOutput:     getEnv("LOG_OUTPUT", "stdout"),
		LogFilePath:   getEnv("LOG_FILE_PATH", "logs/app.log"),
		LogMaxSizeMB:  getEnvInt("LOG_MAX_SIZE_MB", 100),
		LogMaxBackups: getEnvInt("LOG_MAX_BACKUPS", 7),
		LogMaxAgeDays: getEnvInt("LOG_MAX_AGE_DAYS", 30),
		LogCompress:   getEnvBool("LOG_COMPRESS", true),

		OTELEnabled:    getEnvBool("OTEL_ENABLED", false),
		OTELEndpoint:   getEnv("OTEL_ENDPOINT", "http://localhost:4318"),
		OTELSampleRate: getEnvFloat("OTEL_SAMPLE_RATE", 1.0),

		SwaggerEnabled: getEnvBool("SWAGGER_ENABLED", true),
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) validate() error {
	if c.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	if c.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}
	return nil
}

func loadDotEnv() {
	data, err := os.ReadFile(".env")
	if err != nil {
		return
	}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if os.Getenv(key) == "" {
			os.Setenv(key, value)
		}
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return defaultVal
}

func getEnvFloat(key string, defaultVal float64) float64 {
	if val := os.Getenv(key); val != "" {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f
		}
	}
	return defaultVal
}

func getEnvBool(key string, defaultVal bool) bool {
	if val := os.Getenv(key); val != "" {
		if b, err := strconv.ParseBool(val); err == nil {
			return b
		}
	}
	return defaultVal
}

func getEnvDuration(key string, defaultVal time.Duration) time.Duration {
	if val := os.Getenv(key); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			return d
		}
	}
	return defaultVal
}

func getEnvSlice(key, defaultVal string) []string {
	val := getEnv(key, defaultVal)
	return strings.Split(val, ",")
}
