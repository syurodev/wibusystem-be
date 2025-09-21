// Package migrator provides database migration abstractions for the wibusystem monorepo
package migrator

import (
	"context"

	"wibusystem/pkg/database/interfaces"
)

// Migrator defines the interface for database migrations
type Migrator interface {
	// Up runs all pending migrations
	Up(ctx context.Context) error

	// Down rolls back the last migration
	Down(ctx context.Context) error

	// Steps runs a specific number of migrations up or down
	Steps(ctx context.Context, n int) error

	// Version returns the current migration version and dirty state
	Version(ctx context.Context) (uint, bool, error)

	// Force sets the migration version without running migrations
	Force(ctx context.Context, version int) error

	// Close closes the migrator
	Close() error
}

// MigrationConfig defines configuration for migrations
type MigrationConfig struct {
	// Path to migration files (e.g., "file://migrations" or "embed://")
	SourceURL string

	// Database name for the migration driver
	DatabaseName string

	// Table name to store migration state (optional, defaults to "schema_migrations")
	MigrationsTable string

	// Enable verbose logging
	Verbose bool
}

// MigratorFactory creates migrators for different database types
type MigratorFactory interface {
	// CreateMigrator creates a migrator for the given database type
	CreateMigrator(dbType interfaces.DatabaseType, db interface{}, config *MigrationConfig) (Migrator, error)

	// SupportedDatabases returns list of supported database types
	SupportedDatabases() []interfaces.DatabaseType
}
