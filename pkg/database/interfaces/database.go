// Package interfaces defines the core database abstractions for the wibusystem monorepo.
// It provides a unified interface for different database types while allowing provider-specific optimizations.
package interfaces

import (
	"context"
	"time"
)

// DatabaseType represents the type of database provider
type DatabaseType string

const (
	PostgreSQL  DatabaseType = "postgresql"
	MongoDB     DatabaseType = "mongodb"
	Redis       DatabaseType = "redis"
	TimescaleDB DatabaseType = "timescaledb"
)

// Database defines the core database operations that all providers must implement
type Database interface {
	// Core operations
	Connect(ctx context.Context) error
	Close() error
	Health(ctx context.Context) error
	GetType() DatabaseType

	// Transaction support (where applicable)
	BeginTx(ctx context.Context) (Transaction, error)
}

// Transaction defines transaction operations for providers that support them
type Transaction interface {
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

// Rows represents a query result set
type Rows interface {
	Next() bool
	Scan(dest ...interface{}) error
	Close() error
	Err() error
}

// Row represents a single query result row
type Row interface {
	Scan(dest ...interface{}) error
}

// Result represents the result of a command execution
type Result interface {
	LastInsertId() (int64, error)
	RowsAffected() (int64, error)
}

// RelationalDatabase extends Database for SQL-based databases (PostgreSQL, TimescaleDB)
type RelationalDatabase interface {
	Database

	// Query execution
	Query(ctx context.Context, sql string, args ...interface{}) (Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) Row
	Exec(ctx context.Context, sql string, args ...interface{}) (Result, error)

	// Migration support
	Migrate(ctx context.Context, direction MigrationDirection) error
	GetMigrationVersion(ctx context.Context) (uint, bool, error)
}

// DocumentDatabase extends Database for document-based databases (MongoDB)
type DocumentDatabase interface {
	Database

	// Collection operations
	CreateCollection(ctx context.Context, name string) error
	DropCollection(ctx context.Context, name string) error
	ListCollections(ctx context.Context) ([]string, error)

	// Index operations
	CreateIndex(ctx context.Context, collection string, index IndexDefinition) error
	DropIndex(ctx context.Context, collection, name string) error
}

// CacheDatabase extends Database for cache/key-value stores (Redis)
type CacheDatabase interface {
	Database

	// Key-value operations
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Delete(ctx context.Context, keys ...string) error
	Exists(ctx context.Context, keys ...string) (int64, error)

	// Hash operations
	HGet(ctx context.Context, key, field string) (string, error)
	HSet(ctx context.Context, key string, values ...interface{}) error
	HDelete(ctx context.Context, key string, fields ...string) error

	// List operations
	LPush(ctx context.Context, key string, values ...interface{}) error
	RPop(ctx context.Context, key string) (string, error)

	// Set operations
	SAdd(ctx context.Context, key string, members ...interface{}) error
	SMembers(ctx context.Context, key string) ([]string, error)
}

// TimeSeriesDatabase extends Database for time-series databases (TimescaleDB)
type TimeSeriesDatabase interface {
	RelationalDatabase

	// Hypertable operations
	CreateHypertable(ctx context.Context, table string, timeColumn string, options HypertableOptions) error
	DropHypertable(ctx context.Context, table string) error

	// Compression operations
	EnableCompression(ctx context.Context, table string, options CompressionOptions) error
	DisableCompression(ctx context.Context, table string) error

	// Continuous aggregate operations
	CreateContinuousAggregate(ctx context.Context, name string, query string, options ContinuousAggregateOptions) error
	DropContinuousAggregate(ctx context.Context, name string) error
}

// MigrationDirection defines the direction of database migrations
type MigrationDirection string

const (
	MigrationUp   MigrationDirection = "up"
	MigrationDown MigrationDirection = "down"
)

// IndexDefinition defines database index configuration
type IndexDefinition struct {
	Name    string                 `json:"name"`
	Fields  map[string]int         `json:"fields"` // field -> sort order (1 for asc, -1 for desc)
	Unique  bool                   `json:"unique"`
	Sparse  bool                   `json:"sparse"`
	TTL     *time.Duration         `json:"ttl,omitempty"`
	Options map[string]interface{} `json:"options,omitempty"`
}

// HypertableOptions defines TimescaleDB hypertable configuration
type HypertableOptions struct {
	ChunkTimeInterval    *time.Duration `json:"chunk_time_interval,omitempty"`
	CreateDefaultIndexes bool           `json:"create_default_indexes"`
	IfNotExists          bool           `json:"if_not_exists"`
	PartitioningColumn   *string        `json:"partitioning_column,omitempty"`
	NumberPartitions     *int           `json:"number_partitions,omitempty"`
}

// CompressionOptions defines TimescaleDB compression configuration
type CompressionOptions struct {
	SegmentBy         []string       `json:"segment_by,omitempty"`
	OrderBy           []string       `json:"order_by,omitempty"`
	ChunkTimeInterval *time.Duration `json:"chunk_time_interval,omitempty"`
}

// ContinuousAggregateOptions defines TimescaleDB continuous aggregate configuration
type ContinuousAggregateOptions struct {
	WithData      bool           `json:"with_data"`
	RefreshPolicy *RefreshPolicy `json:"refresh_policy,omitempty"`
}

// RefreshPolicy defines continuous aggregate refresh policy
type RefreshPolicy struct {
	StartOffset      time.Duration `json:"start_offset"`
	EndOffset        time.Duration `json:"end_offset"`
	ScheduleInterval time.Duration `json:"schedule_interval"`
}

// DatabaseManagerInterface defines the interface that migration manager needs from database manager
type DatabaseManagerInterface interface {
	GetPrimary() RelationalDatabase
	GetTimeSeries() TimeSeriesDatabase
	GetConfigInterface() DatabaseConfigInterface
}

// DatabaseConfigInterface defines the interface for database configuration that migration manager needs
type DatabaseConfigInterface interface {
	GetPrimary() RelationalConfigInterface
	GetTimeSeries() TimeSeriesConfigInterface
}

// RelationalConfigInterface defines the interface for relational database configuration
type RelationalConfigInterface interface {
	ConnectionString() string
	GetType() DatabaseType
	GetMigrationsPath() string
}

// TimeSeriesConfigInterface defines the interface for time-series database configuration
type TimeSeriesConfigInterface interface {
	GetRelationalConfig() RelationalConfigInterface
}
