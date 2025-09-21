package config

import (
	"fmt"
	"time"

	"wibusystem/pkg/database/interfaces"
)

// DatabaseConfig holds configuration for multiple database types
type DatabaseConfig struct {
	// Primary database (usually PostgreSQL)
	Primary *RelationalConfig `json:"primary,omitempty"`

	// Cache database (usually Redis)
	Cache *CacheConfig `json:"cache,omitempty"`

	// Document database (usually MongoDB)
	Document *DocumentConfig `json:"document,omitempty"`

	// Time-series database (usually TimescaleDB)
	TimeSeries *TimeSeriesConfig `json:"timeseries,omitempty"`
}

// RelationalConfig configures PostgreSQL/TimescaleDB connections
type RelationalConfig struct {
	Type            interfaces.DatabaseType `json:"type"`
	Host            string                  `json:"host"`
	Port            int                     `json:"port"`
	Database        string                  `json:"database"`
	Username        string                  `json:"username"`
	Password        string                  `json:"password"`
	SSLMode         string                  `json:"ssl_mode"`
	MaxConns        int32                   `json:"max_conns"`
	MinConns        int32                   `json:"min_conns"`
	MaxConnLifetime time.Duration           `json:"max_conn_lifetime"`
	MaxConnIdleTime time.Duration           `json:"max_conn_idle_time"`
	MigrationsPath  string                  `json:"migrations_path,omitempty"`
}

// CacheConfig configures Redis connections
type CacheConfig struct {
	Host            string        `json:"host"`
	Port            int           `json:"port"`
	Password        string        `json:"password"`
	Database        int           `json:"database"`
	MaxRetries      int           `json:"max_retries"`
	MinRetryBackoff time.Duration `json:"min_retry_backoff"`
	MaxRetryBackoff time.Duration `json:"max_retry_backoff"`
	DialTimeout     time.Duration `json:"dial_timeout"`
	ReadTimeout     time.Duration `json:"read_timeout"`
	WriteTimeout    time.Duration `json:"write_timeout"`
	PoolSize        int           `json:"pool_size"`
	MinIdleConns    int           `json:"min_idle_conns"`
	MaxConnAge      time.Duration `json:"max_conn_age"`
	PoolTimeout     time.Duration `json:"pool_timeout"`
	IdleTimeout     time.Duration `json:"idle_timeout"`
}

// DocumentConfig configures MongoDB connections
type DocumentConfig struct {
	URI                    string         `json:"uri"`
	Database               string         `json:"database"`
	MaxPoolSize            *uint64        `json:"max_pool_size,omitempty"`
	MinPoolSize            *uint64        `json:"min_pool_size,omitempty"`
	MaxConnIdleTime        *time.Duration `json:"max_conn_idle_time,omitempty"`
	ConnectTimeout         time.Duration  `json:"connect_timeout"`
	ServerSelectionTimeout time.Duration  `json:"server_selection_timeout"`
	SocketTimeout          time.Duration  `json:"socket_timeout"`
}

// TimeSeriesConfig configures TimescaleDB connections (extends RelationalConfig)
type TimeSeriesConfig struct {
	*RelationalConfig
	RetentionPolicy   *RetentionPolicy   `json:"retention_policy,omitempty"`
	CompressionPolicy *CompressionPolicy `json:"compression_policy,omitempty"`
}

// RetentionPolicy defines data retention for time-series data
type RetentionPolicy struct {
	Enabled     bool          `json:"enabled"`
	Interval    time.Duration `json:"interval"`     // How long to keep data
	JobSchedule time.Duration `json:"job_schedule"` // How often to run cleanup
}

// CompressionPolicy defines compression settings for time-series data
type CompressionPolicy struct {
	Enabled   bool          `json:"enabled"`
	AfterAge  time.Duration `json:"after_age"` // Compress data older than this
	SegmentBy []string      `json:"segment_by,omitempty"`
	OrderBy   []string      `json:"order_by,omitempty"`
}

// DefaultDatabaseConfig returns a default configuration suitable for development
func DefaultDatabaseConfig() *DatabaseConfig {
	return &DatabaseConfig{
		Primary: &RelationalConfig{
			Type:            interfaces.PostgreSQL,
			Host:            "localhost",
			Port:            5432,
			Database:        "wibusystem",
			Username:        "wibusystem",
			Password:        "wibusystem",
			SSLMode:         "disable",
			MaxConns:        25,
			MinConns:        5,
			MaxConnLifetime: time.Hour,
			MaxConnIdleTime: 30 * time.Minute,
			MigrationsPath:  "migrations/postgres",
		},
		Cache: &CacheConfig{
			Host:            "localhost",
			Port:            6379,
			Password:        "",
			Database:        0,
			MaxRetries:      3,
			MinRetryBackoff: 8 * time.Millisecond,
			MaxRetryBackoff: 512 * time.Millisecond,
			DialTimeout:     5 * time.Second,
			ReadTimeout:     3 * time.Second,
			WriteTimeout:    3 * time.Second,
			PoolSize:        10,
			MinIdleConns:    2,
			MaxConnAge:      30 * time.Minute,
			PoolTimeout:     4 * time.Second,
			IdleTimeout:     5 * time.Minute,
		},
		Document: &DocumentConfig{
			URI:                    "mongodb://localhost:27017",
			Database:               "wibusystem",
			ConnectTimeout:         10 * time.Second,
			ServerSelectionTimeout: 5 * time.Second,
			SocketTimeout:          3 * time.Second,
		},
		TimeSeries: &TimeSeriesConfig{
			RelationalConfig: &RelationalConfig{
				Type:            interfaces.TimescaleDB,
				Host:            "localhost",
				Port:            5433,
				Database:        "wibusystem_timeseries",
				Username:        "wibusystem",
				Password:        "wibusystem",
				SSLMode:         "disable",
				MaxConns:        25,
				MinConns:        5,
				MaxConnLifetime: time.Hour,
				MaxConnIdleTime: 30 * time.Minute,
				MigrationsPath:  "migrations/timescale",
			},
			RetentionPolicy: &RetentionPolicy{
				Enabled:     true,
				Interval:    90 * 24 * time.Hour, // 90 days
				JobSchedule: 24 * time.Hour,      // Daily cleanup
			},
			CompressionPolicy: &CompressionPolicy{
				Enabled:   true,
				AfterAge:  7 * 24 * time.Hour, // Compress after 7 days
				SegmentBy: []string{"device_id"},
				OrderBy:   []string{"timestamp DESC"},
			},
		},
	}
}

// Validate validates the database configuration
func (c *DatabaseConfig) Validate() error {
	if c.Primary == nil {
		return fmt.Errorf("primary database configuration is required")
	}

	if err := c.Primary.Validate(); err != nil {
		return fmt.Errorf("primary database config validation failed: %w", err)
	}

	if c.Cache != nil {
		if err := c.Cache.Validate(); err != nil {
			return fmt.Errorf("cache database config validation failed: %w", err)
		}
	}

	if c.Document != nil {
		if err := c.Document.Validate(); err != nil {
			return fmt.Errorf("document database config validation failed: %w", err)
		}
	}

	if c.TimeSeries != nil {
		if err := c.TimeSeries.Validate(); err != nil {
			return fmt.Errorf("timeseries database config validation failed: %w", err)
		}
	}

	return nil
}

// Validate validates relational database configuration
func (c *RelationalConfig) Validate() error {
	if c.Host == "" {
		return fmt.Errorf("host is required")
	}
	if c.Port <= 0 || c.Port > 65535 {
		return fmt.Errorf("invalid port: %d", c.Port)
	}
	if c.Database == "" {
		return fmt.Errorf("database name is required")
	}
	if c.Username == "" {
		return fmt.Errorf("username is required")
	}
	return nil
}

// Validate validates cache database configuration
func (c *CacheConfig) Validate() error {
	if c.Host == "" {
		return fmt.Errorf("host is required")
	}
	if c.Port <= 0 || c.Port > 65535 {
		return fmt.Errorf("invalid port: %d", c.Port)
	}
	return nil
}

// Validate validates document database configuration
func (c *DocumentConfig) Validate() error {
	if c.URI == "" {
		return fmt.Errorf("URI is required")
	}
	if c.Database == "" {
		return fmt.Errorf("database name is required")
	}
	return nil
}

// Validate validates time-series database configuration
func (c *TimeSeriesConfig) Validate() error {
	if c.RelationalConfig == nil {
		return fmt.Errorf("relational config is required for TimescaleDB")
	}
	return c.RelationalConfig.Validate()
}

// ConnectionString returns the connection string for relational databases
func (c *RelationalConfig) ConnectionString() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.Username, c.Password, c.Host, c.Port, c.Database, c.SSLMode)
}

// GetType returns the database type
func (c *RelationalConfig) GetType() interfaces.DatabaseType {
	return c.Type
}

// GetMigrationsPath returns the migrations path
func (c *RelationalConfig) GetMigrationsPath() string {
	return c.MigrationsPath
}

// GetRelationalConfig returns the underlying relational config for TimescaleDB
func (c *TimeSeriesConfig) GetRelationalConfig() interfaces.RelationalConfigInterface {
	return c.RelationalConfig
}

// GetPrimary returns the primary database config
func (c *DatabaseConfig) GetPrimary() interfaces.RelationalConfigInterface {
	return c.Primary
}

// GetTimeSeries returns the timeseries database config
func (c *DatabaseConfig) GetTimeSeries() interfaces.TimeSeriesConfigInterface {
	return c.TimeSeries
}
