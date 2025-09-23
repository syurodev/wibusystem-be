package auth

import (
	"time"
)

// Config holds configuration for the authentication middleware
type Config struct {
	// IdentifyGRPCURL is the address of the identify service gRPC server
	IdentifyGRPCURL string

	// GRPCTimeout is the timeout for gRPC calls to the identify service
	GRPCTimeout time.Duration

	// CacheEnabled enables token validation result caching
	CacheEnabled bool

	// CacheTTL is the time-to-live for cached validation results
	CacheTTL time.Duration

	// MaxCacheSize is the maximum number of cached validation results
	MaxCacheSize int

	// TokenHeader is the HTTP header name for the auth token (default: Authorization)
	TokenHeader string

	// TokenPrefix is the prefix for the auth token (default: Bearer)
	TokenPrefix string

	// SkipPaths are paths that should skip authentication
	SkipPaths []string
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		IdentifyGRPCURL: "localhost:9091",
		GRPCTimeout:     5 * time.Second,
		CacheEnabled:    true,
		CacheTTL:        5 * time.Minute,
		MaxCacheSize:    1000,
		TokenHeader:     "Authorization",
		TokenPrefix:     "Bearer",
		SkipPaths:       []string{"/health", "/metrics", "/favicon.ico"},
	}
}

// WithIdentifyGRPCURL sets the identify service gRPC URL
func (c *Config) WithIdentifyGRPCURL(url string) *Config {
	c.IdentifyGRPCURL = url
	return c
}

// WithGRPCTimeout sets the gRPC call timeout
func (c *Config) WithGRPCTimeout(timeout time.Duration) *Config {
	c.GRPCTimeout = timeout
	return c
}

// WithCaching enables/disables caching with specified TTL and max size
func (c *Config) WithCaching(enabled bool, ttl time.Duration, maxSize int) *Config {
	c.CacheEnabled = enabled
	c.CacheTTL = ttl
	c.MaxCacheSize = maxSize
	return c
}

// WithTokenFormat sets the token header and prefix
func (c *Config) WithTokenFormat(header, prefix string) *Config {
	c.TokenHeader = header
	c.TokenPrefix = prefix
	return c
}

// WithSkipPaths sets paths that should skip authentication
func (c *Config) WithSkipPaths(paths []string) *Config {
	c.SkipPaths = paths
	return c
}