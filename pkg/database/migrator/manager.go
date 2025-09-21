// Package migrator provides migration manager for database factory integration
package migrator

import (
	"context"
	"fmt"
	"log"
	"path/filepath"

	"wibusystem/pkg/database/interfaces"
)

// Manager manages migrations for multiple databases
type Manager struct {
	factory   MigratorFactory
	migrators map[interfaces.DatabaseType]Migrator
}

// NewManager creates a new migration manager
func NewManager() *Manager {
	return &Manager{
		factory:   NewMigratorFactory(),
		migrators: make(map[interfaces.DatabaseType]Migrator),
	}
}

// SetupMigrations configures migrations for a database manager
func (m *Manager) SetupMigrations(ctx context.Context, dbManager interfaces.DatabaseManagerInterface, migrationsBasePath string) error {
	// Setup migrations for primary database if it exists
	if primary := dbManager.GetPrimary(); primary != nil {
		dbConfig := dbManager.GetConfigInterface()
		if primaryConfig := dbConfig.GetPrimary(); primaryConfig != nil {
			migrationConfig := &MigrationConfig{
				SourceURL:       fmt.Sprintf("file://%s", filepath.Join(migrationsBasePath, "postgres")),
				DatabaseName:    "postgres",
				MigrationsTable: "schema_migrations",
				Verbose:         true,
			}

			migrator, err := m.factory.CreateMigrator(primary.GetType(), primaryConfig, migrationConfig)
			if err != nil {
				return fmt.Errorf("failed to create migrator for primary database: %w", err)
			}

			m.migrators[primary.GetType()] = migrator
			log.Printf("Migration setup completed for primary database: %s", primary.GetType())
		}
	}

	// Setup migrations for TimescaleDB if it exists
	if timeseries := dbManager.GetTimeSeries(); timeseries != nil {
		dbConfig := dbManager.GetConfigInterface()
		if timeseriesConfig := dbConfig.GetTimeSeries(); timeseriesConfig != nil {
			migrationConfig := &MigrationConfig{
				SourceURL:       fmt.Sprintf("file://%s", filepath.Join(migrationsBasePath, "timescale")),
				DatabaseName:    "postgres", // TimescaleDB uses postgres driver
				MigrationsTable: "timescale_migrations",
				Verbose:         true,
			}

			relationalConfig := timeseriesConfig.GetRelationalConfig()
			migrator, err := m.factory.CreateMigrator(timeseries.GetType(), relationalConfig, migrationConfig)
			if err != nil {
				return fmt.Errorf("failed to create migrator for TimescaleDB: %w", err)
			}

			m.migrators[interfaces.TimescaleDB] = migrator
			log.Printf("Migration setup completed for TimescaleDB")
		}
	}

	return nil
}

// RunMigrations runs migrations for all configured databases
func (m *Manager) RunMigrations(ctx context.Context) error {
	for dbType, migrator := range m.migrators {
		log.Printf("Running migrations for %s...", dbType)

		if err := migrator.Up(ctx); err != nil {
			return fmt.Errorf("failed to run migrations for %s: %w", dbType, err)
		}

		version, dirty, err := migrator.Version(ctx)
		if err != nil {
			log.Printf("Warning: failed to get migration version for %s: %v", dbType, err)
		} else {
			log.Printf("Migration completed for %s. Version: %d, Dirty: %t", dbType, version, dirty)
		}
	}

	return nil
}

// RollbackMigrations rolls back the last migration for all configured databases
func (m *Manager) RollbackMigrations(ctx context.Context) error {
	for dbType, migrator := range m.migrators {
		log.Printf("Rolling back migrations for %s...", dbType)

		if err := migrator.Down(ctx); err != nil {
			return fmt.Errorf("failed to rollback migrations for %s: %w", dbType, err)
		}

		version, dirty, err := migrator.Version(ctx)
		if err != nil {
			log.Printf("Warning: failed to get migration version for %s: %v", dbType, err)
		} else {
			log.Printf("Rollback completed for %s. Version: %d, Dirty: %t", dbType, version, dirty)
		}
	}

	return nil
}

// GetMigrator returns the migrator for a specific database type
func (m *Manager) GetMigrator(dbType interfaces.DatabaseType) (Migrator, bool) {
	migrator, exists := m.migrators[dbType]
	return migrator, exists
}

// Close closes all migrators
func (m *Manager) Close() error {
	var errors []error

	for dbType, migrator := range m.migrators {
		if err := migrator.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close migrator for %s: %w", dbType, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors occurred while closing migrators: %v", errors)
	}

	log.Println("All migrators closed successfully")
	return nil
}

// Status returns the migration status for all configured databases
func (m *Manager) Status(ctx context.Context) (map[interfaces.DatabaseType]MigrationStatus, error) {
	status := make(map[interfaces.DatabaseType]MigrationStatus)

	for dbType, migrator := range m.migrators {
		version, dirty, err := migrator.Version(ctx)
		status[dbType] = MigrationStatus{
			Version: version,
			Dirty:   dirty,
			Error:   err,
		}
	}

	return status, nil
}

// MigrationStatus represents the migration status for a database
type MigrationStatus struct {
	Version uint
	Dirty   bool
	Error   error
}
