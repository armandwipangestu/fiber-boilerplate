package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/armandwipangestu/fiber-boilerplate/internal/config"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/mattn/go-sqlite3"
)

// NewDatabase opens a connection to the configured driver, applies pool
// settings, and pings the database to verify connectivity.
func NewDatabase(cfg config.Config) (*sql.DB, error) {
	db, err := sql.Open(cfg.DatabaseDriver, cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("database open: %w", err)
	}

	db.SetMaxOpenConns(cfg.DatabaseMaxOpenConns)
	db.SetMaxIdleConns(cfg.DatabaseMaxIdleConns)
	db.SetConnMaxLifetime(cfg.DatabaseConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.DatabaseConnMaxIdleTime)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("database ping: %w", err)
	}

	return db, nil
}
