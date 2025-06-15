package db

// Database connection setup for PostgreSQL.

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	_ "github.com/lib/pq"
)

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	DBName          string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
	PingTimeout     time.Duration
}

// getEnvInt gets an integer from environment variable with default value
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvDuration gets a duration from environment variable with default value
func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

// getEnvString gets a string from environment variable with default value
func getEnvString(key string, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetDatabaseConfig returns database configuration from environment variables
func GetDatabaseConfig() DatabaseConfig {
	return DatabaseConfig{
		Host:            getEnvString("POSTGRES_HOST", "localhost"),
		Port:            getEnvString("POSTGRES_PORT", "5432"),
		User:            getEnvString("POSTGRES_USER", "user"),
		Password:        getEnvString("POSTGRES_PASSWORD", "password"),
		DBName:          getEnvString("POSTGRES_DBNAME", "orderdb"),
		SSLMode:         getEnvString("POSTGRES_SSLMODE", "disable"),
		MaxOpenConns:    getEnvInt("DB_MAX_OPEN_CONNS", 300),
		MaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNS", 150),
		ConnMaxLifetime: getEnvDuration("DB_CONN_MAX_LIFETIME", 45*time.Minute),
		ConnMaxIdleTime: getEnvDuration("DB_CONN_MAX_IDLE_TIME", 20*time.Minute),
		PingTimeout:     getEnvDuration("DB_PING_TIMEOUT", 15*time.Second),
	}
}

// buildDSN constructs the PostgreSQL DSN from individual components
func (config DatabaseConfig) buildDSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.DBName,
		config.SSLMode,
	)
}

// NewPostgresDB creates a new PostgreSQL database connection using environment configuration
func NewPostgresDB() (*sql.DB, error) {
	config := GetDatabaseConfig()
	return NewPostgresDBWithConfig(config)
}

// NewPostgresDBWithConfig creates a new PostgreSQL database connection with custom configuration
func NewPostgresDBWithConfig(config DatabaseConfig) (*sql.DB, error) {
	dsn := config.buildDSN()

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(config.ConnMaxLifetime)
	db.SetConnMaxIdleTime(config.ConnMaxIdleTime)

	// Test the connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Printf("✅ Database connection pool configured:")
	log.Printf("   Host: %s:%s", config.Host, config.Port)
	log.Printf("   Database: %s", config.DBName)
	log.Printf("   User: %s", config.User)
	log.Printf("   SSLMode: %s", config.SSLMode)
	log.Printf("   MaxOpenConns: %d", config.MaxOpenConns)
	log.Printf("   MaxIdleConns: %d", config.MaxIdleConns)
	log.Printf("   ConnMaxLifetime: %v", config.ConnMaxLifetime)
	log.Printf("   ConnMaxIdleTime: %v", config.ConnMaxIdleTime)

	return db, nil
}
