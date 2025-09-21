// Package factory provides database factory and manager for multi-database support.
// Config types have been moved to the config package to avoid circular imports.
package factory

// Re-export commonly used config types for backward compatibility
import "wibusystem/pkg/database/config"

type DatabaseConfig = config.DatabaseConfig
type RelationalConfig = config.RelationalConfig
type CacheConfig = config.CacheConfig
type DocumentConfig = config.DocumentConfig
type TimeSeriesConfig = config.TimeSeriesConfig
type RetentionPolicy = config.RetentionPolicy
type CompressionPolicy = config.CompressionPolicy

// DefaultDatabaseConfig returns a default configuration suitable for development
func DefaultDatabaseConfig() *DatabaseConfig {
	return config.DefaultDatabaseConfig()
}
