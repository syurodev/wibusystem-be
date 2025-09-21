// Package factory provides database factory and manager for multi-database support
package factory

import (
	"context"
	"fmt"
	"log"

	"wibusystem/pkg/database/config"
	"wibusystem/pkg/database/interfaces"
	"wibusystem/pkg/database/migrator"
	"wibusystem/pkg/database/providers/mongodb"
	"wibusystem/pkg/database/providers/postgres"
	"wibusystem/pkg/database/providers/redis"
	"wibusystem/pkg/database/providers/timescale"
)

// DatabaseManager manages multiple database connections and migrations
type DatabaseManager struct {
	primary          interfaces.RelationalDatabase
	cache            interfaces.CacheDatabase
	document         interfaces.DocumentDatabase
	timeseries       interfaces.TimeSeriesDatabase
	config           *config.DatabaseConfig
	migrationManager *migrator.Manager
}

// NewDatabaseManager creates a new database manager with the given configuration
func NewDatabaseManager(config *config.DatabaseConfig) (*DatabaseManager, error) {
	if config == nil {
		return nil, fmt.Errorf("database config cannot be nil")
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid database config: %w", err)
	}

	return &DatabaseManager{
		config:           config,
		migrationManager: migrator.NewManager(),
	}, nil
}

// Connect establishes connections to all configured databases
func (dm *DatabaseManager) Connect(ctx context.Context) error {
	var errors []error

	// Connect to primary database (required)
	if dm.config.Primary != nil {
		primary, err := dm.createRelationalDatabase(dm.config.Primary)
		if err != nil {
			errors = append(errors, fmt.Errorf("failed to create primary database: %w", err))
		} else {
			if err := primary.Connect(ctx); err != nil {
				errors = append(errors, fmt.Errorf("failed to connect to primary database: %w", err))
			} else {
				dm.primary = primary
				log.Printf("Primary database connected: %s", primary.GetType())
			}
		}
	}

	// Connect to cache database (optional)
	if dm.config.Cache != nil {
		cache, err := dm.createCacheDatabase(dm.config.Cache)
		if err != nil {
			log.Printf("Warning: failed to create cache database: %v", err)
		} else {
			if err := cache.Connect(ctx); err != nil {
				log.Printf("Warning: failed to connect to cache database: %v", err)
			} else {
				dm.cache = cache
				log.Printf("Cache database connected: %s", cache.GetType())
			}
		}
	}

	// Connect to document database (optional)
	if dm.config.Document != nil {
		document, err := dm.createDocumentDatabase(dm.config.Document)
		if err != nil {
			log.Printf("Warning: failed to create document database: %v", err)
		} else {
			if err := document.Connect(ctx); err != nil {
				log.Printf("Warning: failed to connect to document database: %v", err)
			} else {
				dm.document = document
				log.Printf("Document database connected: %s", document.GetType())
			}
		}
	}

	// Connect to time-series database (optional)
	if dm.config.TimeSeries != nil {
		timeseries, err := dm.createTimeSeriesDatabase(dm.config.TimeSeries)
		if err != nil {
			log.Printf("Warning: failed to create timeseries database: %v", err)
		} else {
			if err := timeseries.Connect(ctx); err != nil {
				log.Printf("Warning: failed to connect to timeseries database: %v", err)
			} else {
				dm.timeseries = timeseries
				log.Printf("TimeSeries database connected: %s", timeseries.GetType())
			}
		}
	}

	// If primary database failed, return error
	if len(errors) > 0 && dm.primary == nil {
		return fmt.Errorf("critical database connections failed: %v", errors)
	}

	return nil
}

// Close closes all database connections and migrators
func (dm *DatabaseManager) Close() error {
	var errors []error

	// Close migrators first
	if dm.migrationManager != nil {
		if err := dm.migrationManager.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close migration manager: %w", err))
		}
	}

	if dm.primary != nil {
		if err := dm.primary.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close primary database: %w", err))
		}
	}

	if dm.cache != nil {
		if err := dm.cache.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close cache database: %w", err))
		}
	}

	if dm.document != nil {
		if err := dm.document.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close document database: %w", err))
		}
	}

	if dm.timeseries != nil {
		if err := dm.timeseries.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close timeseries database: %w", err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors occurred while closing databases: %v", errors)
	}

	log.Println("All database connections and migrators closed")
	return nil
}

// Health checks the health of all connected databases
func (dm *DatabaseManager) Health(ctx context.Context) error {
	var errors []error

	if dm.primary != nil {
		if err := dm.primary.Health(ctx); err != nil {
			errors = append(errors, fmt.Errorf("primary database unhealthy: %w", err))
		}
	}

	if dm.cache != nil {
		if err := dm.cache.Health(ctx); err != nil {
			errors = append(errors, fmt.Errorf("cache database unhealthy: %w", err))
		}
	}

	if dm.document != nil {
		if err := dm.document.Health(ctx); err != nil {
			errors = append(errors, fmt.Errorf("document database unhealthy: %w", err))
		}
	}

	if dm.timeseries != nil {
		if err := dm.timeseries.Health(ctx); err != nil {
			errors = append(errors, fmt.Errorf("timeseries database unhealthy: %w", err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("database health check failed: %v", errors)
	}

	return nil
}

// Getters for specific database types

// GetPrimary returns the primary relational database
func (dm *DatabaseManager) GetPrimary() interfaces.RelationalDatabase {
	return dm.primary
}

// GetCache returns the cache database
func (dm *DatabaseManager) GetCache() interfaces.CacheDatabase {
	return dm.cache
}

// GetDocument returns the document database
func (dm *DatabaseManager) GetDocument() interfaces.DocumentDatabase {
	return dm.document
}

// GetTimeSeries returns the time-series database
func (dm *DatabaseManager) GetTimeSeries() interfaces.TimeSeriesDatabase {
	return dm.timeseries
}

// GetConfig returns the database configuration
func (dm *DatabaseManager) GetConfig() *config.DatabaseConfig {
	return dm.config
}

// GetConfig returns the database configuration as interface (for migration manager)
func (dm *DatabaseManager) GetConfigInterface() interfaces.DatabaseConfigInterface {
	return dm.config
}

// Factory methods for creating database providers

func (dm *DatabaseManager) createRelationalDatabase(config *config.RelationalConfig) (interfaces.RelationalDatabase, error) {
	switch config.Type {
	case interfaces.PostgreSQL:
		return postgres.NewPostgresProvider(config)
	case interfaces.TimescaleDB:
		// TimescaleDB uses TimeSeriesConfig, so we need to wrap RelationalConfig
		tsConfig := &TimeSeriesConfig{RelationalConfig: config}
		return timescale.NewTimescaleProvider(tsConfig)
	default:
		return nil, fmt.Errorf("unsupported relational database type: %s", config.Type)
	}
}

func (dm *DatabaseManager) createCacheDatabase(config *config.CacheConfig) (interfaces.CacheDatabase, error) {
	// Currently only Redis is supported for caching
	return redis.NewRedisProvider(config)
}

func (dm *DatabaseManager) createDocumentDatabase(config *config.DocumentConfig) (interfaces.DocumentDatabase, error) {
	// Currently only MongoDB is supported for document storage
	return mongodb.NewMongoProvider(config)
}

func (dm *DatabaseManager) createTimeSeriesDatabase(config *config.TimeSeriesConfig) (interfaces.TimeSeriesDatabase, error) {
	// Currently only TimescaleDB is supported for time-series
	return timescale.NewTimescaleProvider(config)
}

// Utility methods

// GetDatabasesStatus returns the status of all databases
func (dm *DatabaseManager) GetDatabasesStatus(ctx context.Context) map[string]interface{} {
	status := make(map[string]interface{})

	if dm.primary != nil {
		err := dm.primary.Health(ctx)
		status["primary"] = map[string]interface{}{
			"type":      dm.primary.GetType(),
			"connected": err == nil,
			"error":     err,
		}
	}

	if dm.cache != nil {
		err := dm.cache.Health(ctx)
		status["cache"] = map[string]interface{}{
			"type":      dm.cache.GetType(),
			"connected": err == nil,
			"error":     err,
		}
	}

	if dm.document != nil {
		err := dm.document.Health(ctx)
		status["document"] = map[string]interface{}{
			"type":      dm.document.GetType(),
			"connected": err == nil,
			"error":     err,
		}
	}

	if dm.timeseries != nil {
		err := dm.timeseries.Health(ctx)
		status["timeseries"] = map[string]interface{}{
			"type":      dm.timeseries.GetType(),
			"connected": err == nil,
			"error":     err,
		}
	}

	return status
}

// HasDatabase checks if a specific database type is configured and connected
func (dm *DatabaseManager) HasDatabase(dbType interfaces.DatabaseType) bool {
	switch dbType {
	case interfaces.PostgreSQL:
		return dm.primary != nil && dm.primary.GetType() == interfaces.PostgreSQL
	case interfaces.TimescaleDB:
		return dm.timeseries != nil || (dm.primary != nil && dm.primary.GetType() == interfaces.TimescaleDB)
	case interfaces.MongoDB:
		return dm.document != nil
	case interfaces.Redis:
		return dm.cache != nil
	default:
		return false
	}
}

// SetupMigrations configures migrations for the database manager
func (dm *DatabaseManager) SetupMigrations(ctx context.Context, migrationsBasePath string) error {
	if dm.migrationManager == nil {
		return fmt.Errorf("migration manager not initialized")
	}

	return dm.migrationManager.SetupMigrations(ctx, dm, migrationsBasePath)
}

// RunMigrations runs migrations for all configured databases
func (dm *DatabaseManager) RunMigrations(ctx context.Context) error {
	if dm.migrationManager == nil {
		return fmt.Errorf("migration manager not initialized")
	}

	return dm.migrationManager.RunMigrations(ctx)
}

// RollbackMigrations rolls back the last migration for all configured databases
func (dm *DatabaseManager) RollbackMigrations(ctx context.Context) error {
	if dm.migrationManager == nil {
		return fmt.Errorf("migration manager not initialized")
	}

	return dm.migrationManager.RollbackMigrations(ctx)
}

// GetMigrationStatus returns the migration status for all configured databases
func (dm *DatabaseManager) GetMigrationStatus(ctx context.Context) (map[interfaces.DatabaseType]migrator.MigrationStatus, error) {
	if dm.migrationManager == nil {
		return nil, fmt.Errorf("migration manager not initialized")
	}

	return dm.migrationManager.Status(ctx)
}

// NewDefaultManager creates a database manager with default configuration
func NewDefaultManager() (*DatabaseManager, error) {
	config := config.DefaultDatabaseConfig()
	return NewDatabaseManager(config)
}
