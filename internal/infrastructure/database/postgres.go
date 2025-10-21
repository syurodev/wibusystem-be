// Package database provides database infrastructure for the modular monolith.
// It manages PostgreSQL connections with multiple schema support for different modules.
package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"

	"wibusystem/internal/platform/config"
)

// DB wraps the database connection pool and provides schema-aware queries.
type DB struct {
	pool   *pgxpool.Pool
	config *config.DatabaseConfig
}

// New creates a new database connection with the given configuration.
func New(ctx context.Context, cfg *config.DatabaseConfig) (*DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.Database,
		cfg.SSLMode,
	)

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	// Configure connection pool
	poolConfig.MaxConns = int32(cfg.MaxOpenConns)
	poolConfig.MinConns = int32(cfg.MaxIdleConns)
	poolConfig.MaxConnLifetime = cfg.ConnMaxLifetime
	poolConfig.MaxConnIdleTime = 10 * time.Minute
	poolConfig.HealthCheckPeriod = 1 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Printf("Connected to database: %s@%s:%d/%s", cfg.User, cfg.Host, cfg.Port, cfg.Database)

	return &DB{
		pool:   pool,
		config: cfg,
	}, nil
}

// Pool returns the underlying connection pool.
func (db *DB) Pool() *pgxpool.Pool {
	return db.pool
}

// Close closes the database connection pool.
func (db *DB) Close() {
	if db.pool != nil {
		db.pool.Close()
		log.Println("Database connection closed")
	}
}

// Ping checks if the database is accessible.
func (db *DB) Ping(ctx context.Context) error {
	return db.pool.Ping(ctx)
}

// Stats returns connection pool statistics.
func (db *DB) Stats() *pgxpool.Stat {
	return db.pool.Stat()
}

// CreateSchemas creates all module schemas if they don't exist.
func (db *DB) CreateSchemas(ctx context.Context) error {
	schemas := []string{
		db.config.IdentitySchema,
		db.config.CatalogSchema,
		db.config.CommunitySchema,
		db.config.PaymentSchema,
	}

	for _, schema := range schemas {
		query := fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", schema)
		if _, err := db.pool.Exec(ctx, query); err != nil {
			return fmt.Errorf("failed to create schema %s: %w", schema, err)
		}
		log.Printf("Schema ensured: %s", schema)
	}

	return nil
}

// RunMigrations runs database migrations for all schemas.
func (db *DB) RunMigrations(migrationsPath string) error {
	if migrationsPath == "" {
		log.Println("No migrations path provided, skipping migrations")
		return nil
	}

	// Get standard database connection for migrate library
	stdDB := stdlib.OpenDBFromPool(db.pool)
	defer stdDB.Close()

	driver, err := postgres.WithInstance(stdDB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://"+migrationsPath,
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migration instance: %w", err)
	}

	// Run migrations
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	version, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return fmt.Errorf("failed to get migration version: %w", err)
	}

	if err == migrate.ErrNilVersion {
		log.Println("No migrations applied yet")
	} else {
		log.Printf("Migration completed: version=%d, dirty=%v", version, dirty)
	}

	return nil
}

// WithTransaction executes a function within a database transaction.
func (db *DB) WithTransaction(ctx context.Context, fn func(pgx.Tx) error) error {
	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("failed to rollback transaction: %w (original error: %v)", rbErr, err)
		}
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// HealthCheck performs a health check on the database.
func (db *DB) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	var result int
	err := db.pool.QueryRow(ctx, "SELECT 1").Scan(&result)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}

	if result != 1 {
		return fmt.Errorf("health check returned unexpected result: %d", result)
	}

	return nil
}

// SchemaExists checks if a schema exists in the database.
func (db *DB) SchemaExists(ctx context.Context, schemaName string) (bool, error) {
	var exists bool
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM information_schema.schemata
			WHERE schema_name = $1
		)
	`
	err := db.pool.QueryRow(ctx, query, schemaName).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check schema existence: %w", err)
	}
	return exists, nil
}

// DropSchema drops a schema and all its objects (USE WITH CAUTION).
func (db *DB) DropSchema(ctx context.Context, schemaName string) error {
	query := fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", schemaName)
	if _, err := db.pool.Exec(ctx, query); err != nil {
		return fmt.Errorf("failed to drop schema %s: %w", schemaName, err)
	}
	log.Printf("Schema dropped: %s", schemaName)
	return nil
}

// SetSearchPath sets the schema search path for a connection.
// This is useful when you want queries to automatically use a specific schema.
func (db *DB) SetSearchPath(ctx context.Context, schemas ...string) error {
	if len(schemas) == 0 {
		return fmt.Errorf("at least one schema is required")
	}

	searchPath := ""
	for i, schema := range schemas {
		if i > 0 {
			searchPath += ", "
		}
		searchPath += schema
	}

	query := fmt.Sprintf("SET search_path TO %s", searchPath)
	if _, err := db.pool.Exec(ctx, query); err != nil {
		return fmt.Errorf("failed to set search path: %w", err)
	}

	return nil
}

// TableExists checks if a table exists in a specific schema.
func (db *DB) TableExists(ctx context.Context, schemaName, tableName string) (bool, error) {
	var exists bool
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM information_schema.tables
			WHERE table_schema = $1
			AND table_name = $2
		)
	`
	err := db.pool.QueryRow(ctx, query, schemaName, tableName).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check table existence: %w", err)
	}
	return exists, nil
}

// GetSchemaVersion returns the current schema version from the schema_migrations table.
func (db *DB) GetSchemaVersion(ctx context.Context) (uint, bool, error) {
	var version uint
	var dirty bool

	query := `SELECT version, dirty FROM schema_migrations LIMIT 1`
	err := db.pool.QueryRow(ctx, query).Scan(&version, &dirty)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, false, nil
		}
		return 0, false, fmt.Errorf("failed to get schema version: %w", err)
	}

	return version, dirty, nil
}

// LogPoolStats logs current connection pool statistics.
func (db *DB) LogPoolStats() {
	stats := db.pool.Stat()
	log.Printf("DB Pool Stats - Acquired: %d, Idle: %d, Total: %d, Max: %d",
		stats.AcquiredConns(),
		stats.IdleConns(),
		stats.TotalConns(),
		stats.MaxConns(),
	)
}

// WaitForConnection waits for database to become available with retries.
func WaitForConnection(ctx context.Context, cfg *config.DatabaseConfig, maxRetries int, retryDelay time.Duration) (*DB, error) {
	var db *DB
	var err error

	for i := 0; i < maxRetries; i++ {
		db, err = New(ctx, cfg)
		if err == nil {
			return db, nil
		}

		log.Printf("Failed to connect to database (attempt %d/%d): %v", i+1, maxRetries, err)

		if i < maxRetries-1 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(retryDelay):
				continue
			}
		}
	}

	return nil, fmt.Errorf("failed to connect to database after %d attempts: %w", maxRetries, err)
}
