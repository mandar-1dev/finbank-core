package db

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"corebank/backend/internal/config"
)

// Connect opens a pooled MySQL connection using the standard library
// database/sql package. We deliberately avoid an ORM so that every query
// used for money movement is explicit and easy to reason about.
func Connect(cfg config.Config) (*sql.DB, error) {
	conn, err := sql.Open("mysql", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("opening db: %w", err)
	}

	conn.SetMaxOpenConns(25)
	conn.SetMaxIdleConns(10)
	conn.SetConnMaxLifetime(5 * time.Minute)

	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("pinging db: %w", err)
	}

	return conn, nil
}
