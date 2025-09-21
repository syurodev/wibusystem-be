// Package migrator provides PostgreSQL migration implementation
package migrator

import (
	"context"
	"database/sql"
	"fmt"
	"wibusystem/pkg/database/interfaces"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// PostgresMigrator implements Migrator for PostgreSQL databases
type PostgresMigrator struct {
	db      *sql.DB
	migrate *migrate.Migrate
	config  *MigrationConfig
}

// NewPostgresMigratorFromInterface creates a new PostgreSQL migrator using interface-based config
func NewPostgresMigratorFromInterface(relationalConfig interfaces.RelationalConfigInterface, migrationConfig *MigrationConfig) (*PostgresMigrator, error) {
	// Create SQL connection for migrations (golang-migrate requires database/sql)
	connectionString := relationalConfig.ConnectionString()

	db, err := sql.Open("pgx", connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection for migrations: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Create postgres driver instance
	driverConfig := &postgres.Config{}
	if migrationConfig.MigrationsTable != "" {
		driverConfig.MigrationsTable = migrationConfig.MigrationsTable
	}

	driver, err := postgres.WithInstance(db, driverConfig)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create postgres migration driver: %w", err)
	}

	// Create migrate instance
	m, err := migrate.NewWithDatabaseInstance(
		migrationConfig.SourceURL,
		migrationConfig.DatabaseName,
		driver,
	)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create migrator: %w", err)
	}

	// Enable verbose logging if requested
	if migrationConfig.Verbose {
		m.Log = &MigrationLogger{}
	}

	return &PostgresMigrator{
		db:      db,
		migrate: m,
		config:  migrationConfig,
	}, nil
}

// Up runs all pending migrations
func (m *PostgresMigrator) Up(ctx context.Context) error {
	if err := m.migrate.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run up migrations: %w", err)
	}
	return nil
}

// Down rolls back the last migration
func (m *PostgresMigrator) Down(ctx context.Context) error {
	if err := m.migrate.Down(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run down migrations: %w", err)
	}
	return nil
}

// Steps runs a specific number of migrations up or down
func (m *PostgresMigrator) Steps(ctx context.Context, n int) error {
	if err := m.migrate.Steps(n); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run %d migration steps: %w", n, err)
	}
	return nil
}

// Version returns the current migration version and dirty state
func (m *PostgresMigrator) Version(ctx context.Context) (uint, bool, error) {
	version, dirty, err := m.migrate.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return 0, false, fmt.Errorf("failed to get migration version: %w", err)
	}
	return version, dirty, nil
}

// Force sets the migration version without running migrations
func (m *PostgresMigrator) Force(ctx context.Context, version int) error {
	if err := m.migrate.Force(version); err != nil {
		return fmt.Errorf("failed to force migration version to %d: %w", version, err)
	}
	return nil
}

// Close closes the migrator
func (m *PostgresMigrator) Close() error {
	sourceErr, databaseErr := m.migrate.Close()
	if sourceErr != nil {
		return sourceErr
	}
	if databaseErr != nil {
		return databaseErr
	}
	return m.db.Close()
}

// MigrationLogger implements migrate.Logger for verbose output
type MigrationLogger struct{}

func (l *MigrationLogger) Printf(format string, v ...interface{}) {
	fmt.Printf("[MIGRATION] "+format, v...)
}

func (l *MigrationLogger) Verbose() bool {
	return true
}
