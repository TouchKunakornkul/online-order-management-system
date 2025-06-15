package db

import (
	"database/sql"
	"errors"
	"fmt"

	"online-order-management-system/pkg/logger"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// MigrationManager handles database migrations
type MigrationManager struct {
	db     *sql.DB
	logger *logger.Logger
}

// NewMigrationManager creates a new migration manager
func NewMigrationManager(db *sql.DB) *MigrationManager {
	return &MigrationManager{
		db:     db,
		logger: logger.New("migration-manager", "1.0.0"),
	}
}

// RunMigrations runs all pending migrations
func (m *MigrationManager) RunMigrations(migrationsPath string) error {
	driver, err := postgres.WithInstance(m.db, &postgres.Config{})
	if err != nil {
		m.logger.WithError(err).Error("Failed to create postgres driver instance")
		return fmt.Errorf("failed to create postgres driver: %w", err)
	}

	migration, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsPath),
		"postgres",
		driver,
	)
	if err != nil {
		m.logger.WithError(err).Error("Failed to create migration instance")
		return fmt.Errorf("failed to create migration instance: %w", err)
	}
	defer migration.Close()

	// Run migrations
	if err := migration.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			m.logger.Info("No pending migrations to run")
			return nil
		}
		m.logger.WithError(err).Error("Failed to run migrations")
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	m.logger.Info("Successfully ran all pending migrations")
	return nil
}

// RollbackMigration rolls back one migration
func (m *MigrationManager) RollbackMigration(migrationsPath string) error {
	driver, err := postgres.WithInstance(m.db, &postgres.Config{})
	if err != nil {
		m.logger.WithError(err).Error("Failed to create postgres driver instance")
		return fmt.Errorf("failed to create postgres driver: %w", err)
	}

	migration, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsPath),
		"postgres",
		driver,
	)
	if err != nil {
		m.logger.WithError(err).Error("Failed to create migration instance")
		return fmt.Errorf("failed to create migration instance: %w", err)
	}
	defer migration.Close()

	// Rollback one step
	if err := migration.Steps(-1); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			m.logger.Info("No migrations to rollback")
			return nil
		}
		m.logger.WithError(err).Error("Failed to rollback migration")
		return fmt.Errorf("failed to rollback migration: %w", err)
	}

	m.logger.Info("Successfully rolled back one migration")
	return nil
}

// GetMigrationVersion returns the current migration version
func (m *MigrationManager) GetMigrationVersion(migrationsPath string) (uint, bool, error) {
	driver, err := postgres.WithInstance(m.db, &postgres.Config{})
	if err != nil {
		m.logger.WithError(err).Error("Failed to create postgres driver instance")
		return 0, false, fmt.Errorf("failed to create postgres driver: %w", err)
	}

	migration, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsPath),
		"postgres",
		driver,
	)
	if err != nil {
		m.logger.WithError(err).Error("Failed to create migration instance")
		return 0, false, fmt.Errorf("failed to create migration instance: %w", err)
	}
	defer migration.Close()

	version, dirty, err := migration.Version()
	if err != nil {
		if errors.Is(err, migrate.ErrNilVersion) {
			m.logger.Info("No migrations have been applied yet")
			return 0, false, nil
		}
		m.logger.WithError(err).Error("Failed to get migration version")
		return 0, false, fmt.Errorf("failed to get migration version: %w", err)
	}

	m.logger.WithFields(map[string]interface{}{
		"version": version,
		"dirty":   dirty,
	}).Info("Current migration version")

	return version, dirty, nil
}
