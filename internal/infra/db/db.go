package db

// Database connection setup for PostgreSQL.

import (
	"database/sql"

	_ "github.com/lib/pq"
)

func NewPostgresDB(dsn string) (*sql.DB, error) {
	return sql.Open("postgres", dsn)
}
