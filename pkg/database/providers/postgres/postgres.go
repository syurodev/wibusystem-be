// Package postgres implements PostgreSQL database provider for the wibusystem monorepo
package postgres

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"

	"wibusystem/pkg/database/config"
	"wibusystem/pkg/database/interfaces"
)

// PostgresProvider implements the RelationalDatabase interface
type PostgresProvider struct {
	pool   *pgxpool.Pool
	config *config.RelationalConfig
}

// NewPostgresProvider creates a new PostgreSQL database provider
func NewPostgresProvider(config *config.RelationalConfig) (*PostgresProvider, error) {
	if config == nil {
		return nil, fmt.Errorf("postgres config cannot be nil")
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid postgres config: %w", err)
	}

	return &PostgresProvider{
		config: config,
	}, nil
}

// Connect establishes connection to PostgreSQL database
func (p *PostgresProvider) Connect(ctx context.Context) error {
	poolConfig, err := pgxpool.ParseConfig(p.config.ConnectionString())
	if err != nil {
		return fmt.Errorf("failed to parse connection string: %w", err)
	}

	// Set connection pool configuration
	poolConfig.MaxConns = p.config.MaxConns
	poolConfig.MinConns = p.config.MinConns
	poolConfig.MaxConnLifetime = p.config.MaxConnLifetime
	poolConfig.MaxConnIdleTime = p.config.MaxConnIdleTime

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test the connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return fmt.Errorf("failed to ping database: %w", err)
	}

	p.pool = pool
	log.Printf("Successfully connected to PostgreSQL database: %s", p.config.Database)
	return nil
}

// Close closes the database connection pool
func (p *PostgresProvider) Close() error {
	if p.pool != nil {
		p.pool.Close()
		log.Printf("PostgreSQL connection pool closed for database: %s", p.config.Database)
	}
	return nil
}

// Health checks if the database connection is healthy
func (p *PostgresProvider) Health(ctx context.Context) error {
	if p.pool == nil {
		return fmt.Errorf("postgres provider not connected")
	}
	return p.pool.Ping(ctx)
}

// GetType returns the database type
func (p *PostgresProvider) GetType() interfaces.DatabaseType {
	return interfaces.PostgreSQL
}

// BeginTx starts a new transaction
func (p *PostgresProvider) BeginTx(ctx context.Context) (interfaces.Transaction, error) {
	if p.pool == nil {
		return nil, fmt.Errorf("postgres provider not connected")
	}

	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	return &PostgresTransaction{tx: tx}, nil
}

// Query executes a query that returns rows
func (p *PostgresProvider) Query(ctx context.Context, sql string, args ...interface{}) (interfaces.Rows, error) {
	if p.pool == nil {
		return nil, fmt.Errorf("postgres provider not connected")
	}

	rows, err := p.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	return &PostgresRows{rows: rows}, nil
}

// QueryRow executes a query that returns at most one row
func (p *PostgresProvider) QueryRow(ctx context.Context, sql string, args ...interface{}) interfaces.Row {
	if p.pool == nil {
		return &PostgresRow{err: fmt.Errorf("postgres provider not connected")}
	}

	row := p.pool.QueryRow(ctx, sql, args...)
	return &PostgresRow{row: row}
}

// Exec executes a query that doesn't return rows
func (p *PostgresProvider) Exec(ctx context.Context, sql string, args ...interface{}) (interfaces.Result, error) {
	if p.pool == nil {
		return nil, fmt.Errorf("postgres provider not connected")
	}

	result, err := p.pool.Exec(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute command: %w", err)
	}

	return &PostgresResult{result: result}, nil
}

// Migrate runs database migrations (placeholder - will be implemented with migration system)
func (p *PostgresProvider) Migrate(ctx context.Context, direction interfaces.MigrationDirection) error {
	// This will be implemented when we create the migration system
	return fmt.Errorf("migration not implemented yet")
}

// GetMigrationVersion returns the current migration version (placeholder)
func (p *PostgresProvider) GetMigrationVersion(ctx context.Context) (uint, bool, error) {
	// This will be implemented when we create the migration system
	return 0, false, fmt.Errorf("migration not implemented yet")
}

// GetPool returns the underlying connection pool for advanced operations
func (p *PostgresProvider) GetPool() *pgxpool.Pool {
	return p.pool
}
