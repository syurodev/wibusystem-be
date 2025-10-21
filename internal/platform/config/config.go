// Package config provides centralized configuration management for the modular monolith.
// It consolidates all module-specific configurations and platform-level settings.
package config

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config is the root configuration for the entire application.
type Config struct {
	Server       ServerConfig       `json:"server"`
	Database     DatabaseConfig     `json:"database"`
	Redis        RedisConfig        `json:"redis"`
	Security     SecurityConfig     `json:"security"`
	Localization LocalizationConfig `json:"localization"`

	// Module-specific configurations
	Identity IdentityModuleConfig `json:"identity"`
	Catalog  CatalogModuleConfig  `json:"catalog"`
}

// ServerConfig controls HTTP and gRPC server behavior.
type ServerConfig struct {
	Host         string        `json:"host"`
	Port         int           `json:"port"`
	GRPCPort     int           `json:"grpc_port"`
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
	IdleTimeout  time.Duration `json:"idle_timeout"`
	Environment  string        `json:"environment"` // development, staging, production
}

// DatabaseConfig configures PostgreSQL connection with multiple schemas.
type DatabaseConfig struct {
	Host            string        `json:"host"`
	Port            int           `json:"port"`
	Database        string        `json:"database"`
	User            string        `json:"user"`
	Password        string        `json:"password"`
	SSLMode         string        `json:"ssl_mode"`
	MaxOpenConns    int           `json:"max_open_conns"`
	MaxIdleConns    int           `json:"max_idle_conns"`
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime"`
	MigrationsPath  string        `json:"migrations_path"`

	// Schema names for each module
	IdentitySchema  string `json:"identity_schema"`
	CatalogSchema   string `json:"catalog_schema"`
	CommunitySchema string `json:"community_schema"`
	PaymentSchema   string `json:"payment_schema"`
}

// RedisConfig configures Redis for caching and pub/sub.
type RedisConfig struct {
	Host        string `json:"host"`
	Port        int    `json:"port"`
	Password    string `json:"password"`
	DB          int    `json:"db"`
	MaxRetries  int    `json:"max_retries"`
	PoolSize    int    `json:"pool_size"`
	MinIdleConn int    `json:"min_idle_conn"`
}

// SecurityConfig groups security-related settings.
type SecurityConfig struct {
	BCryptCost      int             `json:"bcrypt_cost"`
	JWTSecret       string          `json:"jwt_secret"`
	SessionDuration time.Duration   `json:"session_duration"`
	CORS            CORSConfig      `json:"cors"`
	RateLimit       RateLimitConfig `json:"rate_limit"`
}

// CORSConfig defines CORS policy.
type CORSConfig struct {
	AllowOrigins     []string `json:"allow_origins"`
	AllowMethods     []string `json:"allow_methods"`
	AllowHeaders     []string `json:"allow_headers"`
	ExposeHeaders    []string `json:"expose_headers"`
	AllowCredentials bool     `json:"allow_credentials"`
	MaxAge           int      `json:"max_age"`
}

// RateLimitConfig controls rate limiting.
type RateLimitConfig struct {
	Enabled           bool `json:"enabled"`
	RequestsPerMinute int  `json:"requests_per_minute"`
	BurstSize         int  `json:"burst_size"`
}

// LocalizationConfig manages i18n settings.
type LocalizationConfig struct {
	DefaultLanguage    string   `json:"default_language"`
	SupportedLanguages []string `json:"supported_languages"`
	BundlePath         string   `json:"bundle_path"`
	QueryParam         string   `json:"query_param"`
	HeaderName         string   `json:"header_name"`
	CookieName         string   `json:"cookie_name"`
}

// IdentityModuleConfig holds Identity module specific settings.
type IdentityModuleConfig struct {
	OAuth2                        OAuth2Config `json:"oauth2"`
	RegistrationAccessTokenSecret string       `json:"registration_access_token_secret"`
	InitialAccessToken            string       `json:"initial_access_token"`
	LoginPageURL                  string       `json:"login_page_url"`
}

// OAuth2Config for OAuth2/OIDC provider.
type OAuth2Config struct {
	Issuer                 string        `json:"issuer"`
	AccessTokenLifespan    time.Duration `json:"access_token_lifespan"`
	RefreshTokenLifespan   time.Duration `json:"refresh_token_lifespan"`
	AuthorizeCodeLifespan  time.Duration `json:"authorize_code_lifespan"`
	IDTokenLifespan        time.Duration `json:"id_token_lifespan"`
	JWTSigningKey          string        `json:"jwt_signing_key"`
	AllowInsecureEndpoints bool          `json:"allow_insecure_endpoints"`
}

// CatalogModuleConfig holds Catalog module specific settings.
type CatalogModuleConfig struct {
	EnableAnime     bool          `json:"enable_anime"`
	EnableManga     bool          `json:"enable_manga"`
	EnableNovel     bool          `json:"enable_novel"`
	CDNBaseURL      string        `json:"cdn_base_url"`
	ImageProxyURL   string        `json:"image_proxy_url"`
	SignedURLExpiry time.Duration `json:"signed_url_expiry"`
}

// Load builds the configuration from environment variables with sensible defaults.
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Host:         getEnv("SERVER_HOST", "localhost"),
			Port:         getEnvAsInt("SERVER_PORT", 8080),
			GRPCPort:     getEnvAsInt("SERVER_GRPC_PORT", 50051),
			ReadTimeout:  getEnvAsDuration("SERVER_READ_TIMEOUT", 30*time.Second),
			WriteTimeout: getEnvAsDuration("SERVER_WRITE_TIMEOUT", 30*time.Second),
			IdleTimeout:  getEnvAsDuration("SERVER_IDLE_TIMEOUT", 120*time.Second),
			Environment:  getEnv("ENVIRONMENT", "development"),
		},
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnvAsInt("DB_PORT", 5432),
			Database:        getEnv("DB_NAME", "system_dev"),
			User:            getEnv("DB_USER", "system_dev"),
			Password:        getEnv("DB_PASSWORD", "system_dev"),
			SSLMode:         getEnv("DB_SSL_MODE", "disable"),
			MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getEnvAsDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
			MigrationsPath:  getEnv("DB_MIGRATIONS_PATH", "./migrations"),
			IdentitySchema:  getEnv("DB_IDENTITY_SCHEMA", "identity"),
			CatalogSchema:   getEnv("DB_CATALOG_SCHEMA", "catalog"),
			CommunitySchema: getEnv("DB_COMMUNITY_SCHEMA", "community"),
			PaymentSchema:   getEnv("DB_PAYMENT_SCHEMA", "payment"),
		},
		Redis: RedisConfig{
			Host:        getEnv("REDIS_HOST", "localhost"),
			Port:        getEnvAsInt("REDIS_PORT", 6379),
			Password:    getEnv("REDIS_PASSWORD", "8767375c91873d04626c73c8aed8ce01"),
			DB:          getEnvAsInt("REDIS_DB", 0),
			MaxRetries:  getEnvAsInt("REDIS_MAX_RETRIES", 3),
			PoolSize:    getEnvAsInt("REDIS_POOL_SIZE", 10),
			MinIdleConn: getEnvAsInt("REDIS_MIN_IDLE_CONN", 2),
		},
		Security: SecurityConfig{
			BCryptCost:      getEnvAsInt("BCRYPT_COST", 12),
			JWTSecret:       getEnv("JWT_SECRET", "development-secret-change-in-production"),
			SessionDuration: getEnvAsDuration("SESSION_DURATION", 24*time.Hour),
			CORS: CORSConfig{
				AllowOrigins:     getEnvAsSlice("CORS_ALLOW_ORIGINS", []string{"http://localhost:3000"}),
				AllowMethods:     getEnvAsSlice("CORS_ALLOW_METHODS", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"}),
				AllowHeaders:     getEnvAsSlice("CORS_ALLOW_HEADERS", []string{"Origin", "Content-Type", "Authorization", "Accept-Language"}),
				ExposeHeaders:    getEnvAsSlice("CORS_EXPOSE_HEADERS", []string{"Content-Length"}),
				AllowCredentials: getEnvAsBool("CORS_ALLOW_CREDENTIALS", true),
				MaxAge:           getEnvAsInt("CORS_MAX_AGE", 43200),
			},
			RateLimit: RateLimitConfig{
				Enabled:           getEnvAsBool("RATE_LIMIT_ENABLED", false),
				RequestsPerMinute: getEnvAsInt("RATE_LIMIT_RPM", 60),
				BurstSize:         getEnvAsInt("RATE_LIMIT_BURST", 10),
			},
		},
		Localization: LocalizationConfig{
			DefaultLanguage:    getEnv("I18N_DEFAULT_LANGUAGE", "en"),
			SupportedLanguages: getEnvAsSlice("I18N_SUPPORTED_LANGUAGES", []string{"en", "vi", "ja"}),
			BundlePath:         getEnv("I18N_BUNDLE_PATH", "./locales"),
			QueryParam:         getEnv("I18N_QUERY_PARAM", "lang"),
			HeaderName:         getEnv("I18N_HEADER_NAME", "Accept-Language"),
			CookieName:         getEnv("I18N_COOKIE_NAME", "locale"),
		},
		Identity: IdentityModuleConfig{
			OAuth2: OAuth2Config{
				Issuer:                 getEnv("OAUTH2_ISSUER", "http://localhost:8080"),
				AccessTokenLifespan:    getEnvAsDuration("OAUTH2_ACCESS_TOKEN_LIFESPAN", 1*time.Hour),
				RefreshTokenLifespan:   getEnvAsDuration("OAUTH2_REFRESH_TOKEN_LIFESPAN", 720*time.Hour),
				AuthorizeCodeLifespan:  getEnvAsDuration("OAUTH2_AUTHORIZE_CODE_LIFESPAN", 10*time.Minute),
				IDTokenLifespan:        getEnvAsDuration("OAUTH2_ID_TOKEN_LIFESPAN", 1*time.Hour),
				JWTSigningKey:          getEnv("OAUTH2_JWT_SIGNING_KEY", ""),
				AllowInsecureEndpoints: getEnvAsBool("OAUTH2_ALLOW_INSECURE_ENDPOINTS", true),
			},
			RegistrationAccessTokenSecret: getEnv("IDENTITY_REG_ACCESS_TOKEN_SECRET", "change-this-secret-in-production"),
			InitialAccessToken:            getEnv("IDENTITY_INITIAL_ACCESS_TOKEN", ""),
			LoginPageURL:                  getEnv("IDENTITY_LOGIN_PAGE_URL", "http://localhost:3000/login"),
		},
		Catalog: CatalogModuleConfig{
			EnableAnime:     getEnvAsBool("CATALOG_ENABLE_ANIME", true),
			EnableManga:     getEnvAsBool("CATALOG_ENABLE_MANGA", true),
			EnableNovel:     getEnvAsBool("CATALOG_ENABLE_NOVEL", true),
			CDNBaseURL:      getEnv("CATALOG_CDN_BASE_URL", ""),
			ImageProxyURL:   getEnv("CATALOG_IMAGE_PROXY_URL", ""),
			SignedURLExpiry: getEnvAsDuration("CATALOG_SIGNED_URL_EXPIRY", 1*time.Hour),
		},
	}
}

// Helper functions to read environment variables with defaults

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		log.Printf("Warning: invalid int value for %s, using default %d", key, defaultValue)
		return defaultValue
	}
	return value
}

func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		log.Printf("Warning: invalid bool value for %s, using default %t", key, defaultValue)
		return defaultValue
	}
	return value
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := time.ParseDuration(valueStr)
	if err != nil {
		log.Printf("Warning: invalid duration value for %s, using default %s", key, defaultValue)
		return defaultValue
	}
	return value
}

func getEnvAsSlice(key string, defaultValue []string) []string {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	return strings.Split(valueStr, ",")
}

// IsDevelopment returns true if running in development mode.
func (c *Config) IsDevelopment() bool {
	return c.Server.Environment == "development"
}

// IsProduction returns true if running in production mode.
func (c *Config) IsProduction() bool {
	return c.Server.Environment == "production"
}

// DatabaseDSN returns PostgreSQL connection string.
func (c *Config) DatabaseDSN() string {
	return "host=" + c.Database.Host +
		" port=" + strconv.Itoa(c.Database.Port) +
		" user=" + c.Database.User +
		" password=" + c.Database.Password +
		" dbname=" + c.Database.Database +
		" sslmode=" + c.Database.SSLMode
}

// RedisDSN returns Redis connection string.
func (c *Config) RedisDSN() string {
	return c.Redis.Host + ":" + strconv.Itoa(c.Redis.Port)
}
