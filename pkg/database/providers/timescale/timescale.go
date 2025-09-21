// Package timescale implements TimescaleDB provider for the wibusystem monorepo
package timescale

import (
	"context"
	"fmt"
	"log"
	"strings"

	"wibusystem/pkg/database/config"
	"wibusystem/pkg/database/interfaces"
	"wibusystem/pkg/database/providers/postgres"
)

// TimescaleProvider implements the TimeSeriesDatabase interface
// It embeds PostgresProvider since TimescaleDB is built on PostgreSQL
type TimescaleProvider struct {
	*postgres.PostgresProvider
	config *config.TimeSeriesConfig
}

// NewTimescaleProvider creates a new TimescaleDB provider
func NewTimescaleProvider(config *config.TimeSeriesConfig) (*TimescaleProvider, error) {
	if config == nil {
		return nil, fmt.Errorf("timescaledb config cannot be nil")
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid timescaledb config: %w", err)
	}

	// Create underlying PostgreSQL provider
	pgProvider, err := postgres.NewPostgresProvider(config.RelationalConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres provider: %w", err)
	}

	return &TimescaleProvider{
		PostgresProvider: pgProvider,
		config:           config,
	}, nil
}

// GetType returns the database type
func (t *TimescaleProvider) GetType() interfaces.DatabaseType {
	return interfaces.TimescaleDB
}

// Connect establishes connection to TimescaleDB and ensures TimescaleDB extension is enabled
func (t *TimescaleProvider) Connect(ctx context.Context) error {
	// First connect using the underlying PostgreSQL provider
	if err := t.PostgresProvider.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect to TimescaleDB: %w", err)
	}

	// Verify TimescaleDB extension is available
	if err := t.verifyTimescaleExtension(ctx); err != nil {
		return fmt.Errorf("TimescaleDB extension verification failed: %w", err)
	}

	log.Printf("Successfully connected to TimescaleDB database: %s", t.config.RelationalConfig.Database)
	return nil
}

// verifyTimescaleExtension checks if TimescaleDB extension is enabled
func (t *TimescaleProvider) verifyTimescaleExtension(ctx context.Context) error {
	pool := t.PostgresProvider.GetPool()
	if pool == nil {
		return fmt.Errorf("postgres connection pool not available")
	}

	var version string
	query := "SELECT extversion FROM pg_extension WHERE extname = 'timescaledb'"
	err := pool.QueryRow(ctx, query).Scan(&version)
	if err != nil {
		return fmt.Errorf("TimescaleDB extension not found - ensure CREATE EXTENSION timescaledb; has been run")
	}

	log.Printf("TimescaleDB extension version: %s", version)
	return nil
}

// Hypertable operations

// CreateHypertable creates a TimescaleDB hypertable
func (t *TimescaleProvider) CreateHypertable(ctx context.Context, table string, timeColumn string, options interfaces.HypertableOptions) error {
	pool := t.PostgresProvider.GetPool()
	if pool == nil {
		return fmt.Errorf("postgres connection pool not available")
	}

	// Build the CREATE HYPERTABLE query
	var queryParts []string
	var args []interface{}
	argIndex := 1

	queryParts = append(queryParts, "SELECT create_hypertable($"+fmt.Sprintf("%d", argIndex)+", $"+fmt.Sprintf("%d", argIndex+1))
	args = append(args, table, timeColumn)
	argIndex += 2

	// Add optional parameters
	if options.ChunkTimeInterval != nil {
		queryParts[0] += ", chunk_time_interval => $" + fmt.Sprintf("%d", argIndex)
		args = append(args, fmt.Sprintf("%d microseconds", options.ChunkTimeInterval.Microseconds()))
		argIndex++
	}

	if !options.CreateDefaultIndexes {
		queryParts[0] += ", create_default_indexes => $" + fmt.Sprintf("%d", argIndex)
		args = append(args, false)
		argIndex++
	}

	if options.IfNotExists {
		queryParts[0] += ", if_not_exists => $" + fmt.Sprintf("%d", argIndex)
		args = append(args, true)
		argIndex++
	}

	if options.PartitioningColumn != nil {
		queryParts[0] += ", partitioning_column => $" + fmt.Sprintf("%d", argIndex)
		args = append(args, *options.PartitioningColumn)
		argIndex++

		if options.NumberPartitions != nil {
			queryParts[0] += ", number_partitions => $" + fmt.Sprintf("%d", argIndex)
			args = append(args, *options.NumberPartitions)
			argIndex++
		}
	}

	query := strings.Join(queryParts, "")

	_, err := pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to create hypertable %s: %w", table, err)
	}

	log.Printf("Created hypertable: %s with time column: %s", table, timeColumn)
	return nil
}

// DropHypertable drops a TimescaleDB hypertable
func (t *TimescaleProvider) DropHypertable(ctx context.Context, table string) error {
	pool := t.PostgresProvider.GetPool()
	if pool == nil {
		return fmt.Errorf("postgres connection pool not available")
	}

	query := "SELECT drop_hypertable($1, if_exists => true)"
	_, err := pool.Exec(ctx, query, table)
	if err != nil {
		return fmt.Errorf("failed to drop hypertable %s: %w", table, err)
	}

	log.Printf("Dropped hypertable: %s", table)
	return nil
}

// Compression operations

// EnableCompression enables compression on a hypertable
func (t *TimescaleProvider) EnableCompression(ctx context.Context, table string, options interfaces.CompressionOptions) error {
	pool := t.PostgresProvider.GetPool()
	if pool == nil {
		return fmt.Errorf("postgres connection pool not available")
	}

	// First, enable compression on the hypertable
	var queryParts []string
	queryParts = append(queryParts, "ALTER TABLE "+table+" SET (timescaledb.compress = true")

	if len(options.SegmentBy) > 0 {
		queryParts = append(queryParts, "timescaledb.compress_segmentby = '"+strings.Join(options.SegmentBy, ",")+"'")
	}

	if len(options.OrderBy) > 0 {
		queryParts = append(queryParts, "timescaledb.compress_orderby = '"+strings.Join(options.OrderBy, ",")+"'")
	}

	query := strings.Join(queryParts, ", ") + ")"

	_, err := pool.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to enable compression on %s: %w", table, err)
	}

	// Add compression policy if chunk time interval is specified
	if options.ChunkTimeInterval != nil {
		policyQuery := "SELECT add_compression_policy($1, INTERVAL '%d seconds')"
		_, err = pool.Exec(ctx, fmt.Sprintf(policyQuery, int64(options.ChunkTimeInterval.Seconds())), table)
		if err != nil {
			return fmt.Errorf("failed to add compression policy for %s: %w", table, err)
		}
	}

	log.Printf("Enabled compression for hypertable: %s", table)
	return nil
}

// DisableCompression disables compression on a hypertable
func (t *TimescaleProvider) DisableCompression(ctx context.Context, table string) error {
	pool := t.PostgresProvider.GetPool()
	if pool == nil {
		return fmt.Errorf("postgres connection pool not available")
	}

	// Remove compression policy first (if exists)
	_, _ = pool.Exec(ctx, "SELECT remove_compression_policy($1, if_exists => true)", table)

	// Disable compression
	query := "ALTER TABLE " + table + " SET (timescaledb.compress = false)"
	_, err := pool.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to disable compression on %s: %w", table, err)
	}

	log.Printf("Disabled compression for hypertable: %s", table)
	return nil
}

// Continuous aggregate operations

// CreateContinuousAggregate creates a continuous aggregate
func (t *TimescaleProvider) CreateContinuousAggregate(ctx context.Context, name string, query string, options interfaces.ContinuousAggregateOptions) error {
	pool := t.PostgresProvider.GetPool()
	if pool == nil {
		return fmt.Errorf("postgres connection pool not available")
	}

	// Build continuous aggregate query
	caQuery := fmt.Sprintf("CREATE MATERIALIZED VIEW %s WITH (timescaledb.continuous) AS %s", name, query)

	if options.WithData {
		caQuery += " WITH DATA"
	} else {
		caQuery += " WITH NO DATA"
	}

	_, err := pool.Exec(ctx, caQuery)
	if err != nil {
		return fmt.Errorf("failed to create continuous aggregate %s: %w", name, err)
	}

	// Add refresh policy if specified
	if options.RefreshPolicy != nil {
		policy := options.RefreshPolicy
		policyQuery := `SELECT add_continuous_aggregate_policy($1,
			start_offset => INTERVAL '%d seconds',
			end_offset => INTERVAL '%d seconds',
			schedule_interval => INTERVAL '%d seconds')`

		_, err = pool.Exec(ctx,
			fmt.Sprintf(policyQuery,
				int64(policy.StartOffset.Seconds()),
				int64(policy.EndOffset.Seconds()),
				int64(policy.ScheduleInterval.Seconds())),
			name)
		if err != nil {
			return fmt.Errorf("failed to add refresh policy for continuous aggregate %s: %w", name, err)
		}
	}

	log.Printf("Created continuous aggregate: %s", name)
	return nil
}

// DropContinuousAggregate drops a continuous aggregate
func (t *TimescaleProvider) DropContinuousAggregate(ctx context.Context, name string) error {
	pool := t.PostgresProvider.GetPool()
	if pool == nil {
		return fmt.Errorf("postgres connection pool not available")
	}

	// Remove refresh policy first (if exists)
	_, _ = pool.Exec(ctx, "SELECT remove_continuous_aggregate_policy($1, if_exists => true)", name)

	// Drop the continuous aggregate
	query := "DROP MATERIALIZED VIEW IF EXISTS " + name
	_, err := pool.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to drop continuous aggregate %s: %w", name, err)
	}

	log.Printf("Dropped continuous aggregate: %s", name)
	return nil
}

// SetupRetentionPolicy sets up automatic data retention
func (t *TimescaleProvider) SetupRetentionPolicy(ctx context.Context, table string) error {
	if t.config.RetentionPolicy == nil || !t.config.RetentionPolicy.Enabled {
		return nil
	}

	pool := t.PostgresProvider.GetPool()
	if pool == nil {
		return fmt.Errorf("postgres connection pool not available")
	}

	policy := t.config.RetentionPolicy
	query := `SELECT add_retention_policy($1, INTERVAL '%d seconds',
		schedule_interval => INTERVAL '%d seconds')`

	_, err := pool.Exec(ctx,
		fmt.Sprintf(query,
			int64(policy.Interval.Seconds()),
			int64(policy.JobSchedule.Seconds())),
		table)
	if err != nil {
		return fmt.Errorf("failed to add retention policy for %s: %w", table, err)
	}

	log.Printf("Added retention policy for hypertable: %s (retain for %s)", table, policy.Interval)
	return nil
}

// SetupCompressionPolicy sets up automatic compression
func (t *TimescaleProvider) SetupCompressionPolicy(ctx context.Context, table string) error {
	if t.config.CompressionPolicy == nil || !t.config.CompressionPolicy.Enabled {
		return nil
	}

	policy := t.config.CompressionPolicy

	// Enable compression first
	compressionOptions := interfaces.CompressionOptions{
		SegmentBy:         policy.SegmentBy,
		OrderBy:           policy.OrderBy,
		ChunkTimeInterval: &policy.AfterAge,
	}

	return t.EnableCompression(ctx, table, compressionOptions)
}
