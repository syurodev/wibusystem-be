// Package config defines strongly-typed runtime configuration for the
// Identity Service and helpers to load values from environment variables.
//
// All durations are parsed using time.ParseDuration syntax. Sensible defaults
// are provided for local development; override via environment in production.
package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"wibusystem/pkg/database/config"
	"wibusystem/pkg/database/interfaces"
	grpcconfig "wibusystem/pkg/grpc/config"
)

// Config aggregates all configuration sections for the service.
type Config struct {
	Server       ServerConfig       `json:"server"`
	GRPC         GRPCConfig         `json:"grpc"`
	Database     DatabaseConfig     `json:"database"`
	OAuth2       OAuth2Config       `json:"oauth2"`
	Security     SecurityConfig     `json:"security"`
	Localization LocalizationConfig `json:"localization"`
}

// ServerConfig controls HTTP server and runtime behavior.
type ServerConfig struct {
	Host         string        `json:"host"`
	Port         int           `json:"port"`
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
	IdleTimeout  time.Duration `json:"idle_timeout"`
	Environment  string        `json:"environment"`
}

// DatabaseConfig configures the database connections and migration path.
type DatabaseConfig struct {
	*config.DatabaseConfig
	MigrationsPath string `json:"migrations_path"`
}

// OAuth2Config configures OAuth2/OIDC lifespans and signing.
//
// Note: In development, AllowInsecureEndpoints may be true to simplify
// testing. Ensure secure settings in production deployments.
type OAuth2Config struct {
	Issuer                 string        `json:"issuer"`
	AccessTokenLifespan    time.Duration `json:"access_token_lifespan"`
	RefreshTokenLifespan   time.Duration `json:"refresh_token_lifespan"`
	AuthorizeCodeLifespan  time.Duration `json:"authorize_code_lifespan"`
	IDTokenLifespan        time.Duration `json:"id_token_lifespan"`
	JWTSigningKey          string        `json:"jwt_signing_key"`
	AllowInsecureEndpoints bool          `json:"allow_insecure_endpoints"`
}

// SecurityConfig groups security-related knobs such as password hashing cost,
// JWT secret (if used), session duration, CORS, and rate limiting.
type SecurityConfig struct {
	BCryptCost      int                `json:"bcrypt_cost"`
	JWTSecret       string             `json:"jwt_secret"`
	SessionDuration time.Duration      `json:"session_duration"`
	CORS            CORSConfig         `json:"cors"`
	RateLimit       RateLimitConfig    `json:"rate_limit"`
	Registration    RegistrationConfig `json:"registration"`
	LoginPageURL    string             `json:"login_page_url"`
}

// LocalizationConfig defines how translation files are loaded and which
// languages are supported when building localized responses for clients.
type LocalizationConfig struct {
	DefaultLanguage    string   `json:"default_language"`
	SupportedLanguages []string `json:"supported_languages"`
	BundlePath         string   `json:"bundle_path"`
	QueryParam         string   `json:"query_param"`
}

// CORSConfig defines cross-origin resource sharing policy.
type CORSConfig struct {
	AllowOrigins     []string `json:"allow_origins"`
	AllowMethods     []string `json:"allow_methods"`
	AllowHeaders     []string `json:"allow_headers"`
	ExposeHeaders    []string `json:"expose_headers"`
	AllowCredentials bool     `json:"allow_credentials"`
	MaxAge           int      `json:"max_age"`
}

// RateLimitConfig tunes request rate limiting behavior.
type RateLimitConfig struct {
	RequestsPerMinute int `json:"requests_per_minute"`
	BurstSize         int `json:"burst_size"`
}

type RegistrationConfig struct {
	InitialAccessToken            string `json:"initial_access_token"`
	RegistrationAccessTokenSecret string `json:"registration_access_token_secret"`
}

// GRPCConfig controls gRPC server behavior.
type GRPCConfig struct {
	*grpcconfig.ServerConfig
}

// Load reads configuration from environment variables with defaults suitable
// for local development. Override via env for staging/production.
func Load() *Config {
	migrationsPath := getEnv("DB_MIGRATIONS_PATH", "../../pkg/database/migrations/postgres/identify")

	jwtSigningKey := getEnv("OAUTH2_JWT_SIGNING_KEY", "")
	if jwtSigningKey == "" {
		if keyFromFile, err := os.ReadFile("oidc-signing.pem"); err == nil {
			jwtSigningKey = string(keyFromFile)
		}
	}

	return &Config{
		Server: ServerConfig{
			Host:         getEnv("SERVER_HOST", "localhost"),
			Port:         getEnvAsInt("SERVER_PORT", 8080),
			ReadTimeout:  getEnvAsDuration("SERVER_READ_TIMEOUT", 30*time.Second),
			WriteTimeout: getEnvAsDuration("SERVER_WRITE_TIMEOUT", 30*time.Second),
			IdleTimeout:  getEnvAsDuration("SERVER_IDLE_TIMEOUT", 120*time.Second),
			Environment:  getEnv("ENVIRONMENT", "development"),
		},
		GRPC: GRPCConfig{
			ServerConfig: &grpcconfig.ServerConfig{
				Host:                  getEnv("GRPC_HOST", "localhost"),
				Port:                  getEnvAsInt("GRPC_PORT", 9090),
				MaxRecvMsgSize:        getEnvAsInt("GRPC_MAX_RECV_MSG_SIZE", 4*1024*1024),
				MaxSendMsgSize:        getEnvAsInt("GRPC_MAX_SEND_MSG_SIZE", 4*1024*1024),
				ConnectionTimeout:     getEnvAsDuration("GRPC_CONNECTION_TIMEOUT", 5*time.Second),
				MaxConnectionIdle:     getEnvAsDuration("GRPC_MAX_CONNECTION_IDLE", 15*time.Second),
				MaxConnectionAge:      getEnvAsDuration("GRPC_MAX_CONNECTION_AGE", 30*time.Second),
				MaxConnectionAgeGrace: getEnvAsDuration("GRPC_MAX_CONNECTION_AGE_GRACE", 5*time.Second),
				Time:                  getEnvAsDuration("GRPC_KEEPALIVE_TIME", 10*time.Second),
				Timeout:               getEnvAsDuration("GRPC_KEEPALIVE_TIMEOUT", 3*time.Second),
				EnableReflection:      getEnvAsBool("GRPC_ENABLE_REFLECTION", true),
			},
		},
		Database: DatabaseConfig{
			DatabaseConfig: &config.DatabaseConfig{
				Primary: &config.RelationalConfig{
					Type:            interfaces.PostgreSQL,
					Host:            getEnv("DB_HOST", "localhost"),
					Port:            getEnvAsInt("DB_PORT", 5432),
					Database:        getEnv("DB_NAME", "identify_service"),
					Username:        getEnv("DB_USER", "identify_service"),
					Password:        getEnv("DB_PASSWORD", "c2f960ed5f802b6acac2d4d928f21ada"),
					MaxConns:        int32(getEnvAsInt("DB_MAX_CONNS", 25)),
					MinConns:        int32(getEnvAsInt("DB_MIN_CONNS", 5)),
					MaxConnLifetime: getEnvAsDuration("DB_MAX_CONN_LIFETIME", time.Hour),
					MaxConnIdleTime: getEnvAsDuration("DB_MAX_CONN_IDLE_TIME", 30*time.Minute),
					SSLMode:         getEnv("DB_SSL_MODE", "disable"),
				},
				// Cache, Document, and TimeSeries can be added later as needed
			},
			MigrationsPath: migrationsPath,
		},
		OAuth2: OAuth2Config{
			Issuer:                 getEnv("OAUTH2_ISSUER", "http://localhost:8080"),
			AccessTokenLifespan:    getEnvAsDuration("OAUTH2_ACCESS_TOKEN_LIFESPAN", 1*time.Hour),
			RefreshTokenLifespan:   getEnvAsDuration("OAUTH2_REFRESH_TOKEN_LIFESPAN", 24*time.Hour),
			AuthorizeCodeLifespan:  getEnvAsDuration("OAUTH2_AUTHORIZE_CODE_LIFESPAN", 10*time.Minute),
			IDTokenLifespan:        getEnvAsDuration("OAUTH2_ID_TOKEN_LIFESPAN", 1*time.Hour),
			JWTSigningKey:          jwtSigningKey,
			AllowInsecureEndpoints: getEnvAsBool("OAUTH2_ALLOW_INSECURE_ENDPOINTS", true),
		},
		Security: SecurityConfig{
			BCryptCost:      getEnvAsInt("BCRYPT_COST", 12),
			JWTSecret:       getEnv("JWT_SECRET", "wibusystem-jwt-secret-change-in-production"),
			SessionDuration: getEnvAsDuration("SESSION_DURATION", 24*time.Hour),
			CORS: CORSConfig{
				AllowOrigins:     []string{"http://localhost:3000", "http://localhost:3001"},
				AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
				AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
				ExposeHeaders:    []string{"Content-Length"},
				AllowCredentials: true,
				MaxAge:           86400, // 24 hours
			},
			RateLimit: RateLimitConfig{
				RequestsPerMinute: getEnvAsInt("RATE_LIMIT_RPM", 60),
				BurstSize:         getEnvAsInt("RATE_LIMIT_BURST", 20),
			},
			Registration: RegistrationConfig{
				InitialAccessToken:            getEnv("DCR_INITIAL_ACCESS_TOKEN", ""),
				RegistrationAccessTokenSecret: getEnv("DCR_REG_ACCESS_TOKEN_SECRET", "change-me-in-prod"),
			},
			LoginPageURL: getEnv("LOGIN_PAGE_URL", ""),
		},
		Localization: LocalizationConfig{
			DefaultLanguage:    getEnv("LOCALIZATION_DEFAULT_LANGUAGE", "en"),
			SupportedLanguages: []string{"en", "vi"},
			BundlePath:         getEnv("LOCALIZATION_BUNDLE_PATH", "locales"),
			QueryParam:         getEnv("LOCALIZATION_QUERY_PARAM", "lang"),
		},
	}
}

// Helper functions to get environment variables with type-safe fallbacks.

// getEnv returns the string value of key or defaultValue if unset.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt returns the integer value for key or defaultValue if unset or invalid.
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
		log.Printf("Invalid integer value for %s: %s, using default: %d", key, value, defaultValue)
	}
	return defaultValue
}

// getEnvAsBool returns the boolean value for key or defaultValue if unset or invalid.
func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
		log.Printf("Invalid boolean value for %s: %s, using default: %t", key, value, defaultValue)
	}
	return defaultValue
}

// getEnvAsDuration returns a parsed duration for key or defaultValue if unset or invalid.
func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
		log.Printf("Invalid duration value for %s: %s, using default: %s", key, value, defaultValue)
	}
	return defaultValue
}
