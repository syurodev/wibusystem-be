// Package migrator provides migration factory implementation
package migrator

import (
	"fmt"

	"wibusystem/pkg/database/interfaces"
)

// DefaultMigratorFactory implements MigratorFactory
type DefaultMigratorFactory struct{}

// CreateMigrator creates a migrator for the given database type
func (f *DefaultMigratorFactory) CreateMigrator(dbType interfaces.DatabaseType, db interface{}, migrationConfig *MigrationConfig) (Migrator, error) {
	switch dbType {
	case interfaces.PostgreSQL:
		return f.createPostgresMigrator(db, migrationConfig)
	case interfaces.TimescaleDB:
		// TimescaleDB uses PostgreSQL migrations
		return f.createPostgresMigrator(db, migrationConfig)
	default:
		return nil, fmt.Errorf("unsupported database type for migrations: %s", dbType)
	}
}

// createPostgresMigrator creates a PostgreSQL migrator
func (f *DefaultMigratorFactory) createPostgresMigrator(db interface{}, migrationConfig *MigrationConfig) (Migrator, error) {
	// Extract relational config from the database interface
	relationalConfig, ok := db.(interfaces.RelationalConfigInterface)
	if !ok {
		return nil, fmt.Errorf("database config does not implement RelationalConfigInterface")
	}

	return NewPostgresMigratorFromInterface(relationalConfig, migrationConfig)
}

// SupportedDatabases returns list of supported database types
func (f *DefaultMigratorFactory) SupportedDatabases() []interfaces.DatabaseType {
	return []interfaces.DatabaseType{
		interfaces.PostgreSQL,
		interfaces.TimescaleDB,
	}
}

// NewMigratorFactory creates a new migrator factory
func NewMigratorFactory() MigratorFactory {
	return &DefaultMigratorFactory{}
}
