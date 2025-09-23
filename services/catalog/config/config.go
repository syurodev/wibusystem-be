// Package config defines runtime configuration for the Catalog Service,
// mirroring the structure used in the Identify Service to keep a consistent
// developer experience across services.
package config

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	dbconfig "wibusystem/pkg/database/config"
	"wibusystem/pkg/database/interfaces"
)

// Config aggregates configuration sections for the Catalog service.
type Config struct {
	Server       ServerConfig       `json:"server"`
	Database     DatabaseConfig     `json:"database"`
	Localization LocalizationConfig `json:"localization"`
	Content      ContentConfig      `json:"content"`
	Media        MediaConfig        `json:"media"`
	Security     SecurityConfig     `json:"security"`
	Integrations IntegrationsConfig `json:"integrations"`
}

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	Host         string        `json:"host"`
	Port         int           `json:"port"`
	GRPCPort     int           `json:"grpc_port"`
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
	IdleTimeout  time.Duration `json:"idle_timeout"`
	Environment  string        `json:"environment"`
}

// DatabaseConfig wraps the shared database configuration and a service-specific
// migration path so each service can manage its own schema.
type DatabaseConfig struct {
	*dbconfig.DatabaseConfig
	MigrationsPath string `json:"migrations_path"`
}

// LocalizationConfig controls language support for responses and metadata.
type LocalizationConfig struct {
	DefaultLanguage    string   `json:"default_language"`
	SupportedLanguages []string `json:"supported_languages"`
	BundlePath         string   `json:"bundle_path"`
	QueryParam         string   `json:"query_param"`
	HeaderName         string   `json:"header_name"`
	CookieName         string   `json:"cookie_name"`
}

// ContentConfig toggles catalog verticals on/off without redeploying.
type ContentConfig struct {
	EnableAnime bool `json:"enable_anime"`
	EnableManga bool `json:"enable_manga"`
	EnableNovel bool `json:"enable_novel"`
}

// MediaConfig captures knobs for media delivery integration (CDN, image proxy).
type MediaConfig struct {
	CDNBaseURL      string        `json:"cdn_base_url"`
	ImageProxyURL   string        `json:"image_proxy_url"`
	SignedURLExpiry time.Duration `json:"signed_url_expiry"`
}

// SecurityConfig currently holds CORS settings for HTTP responses.
type SecurityConfig struct {
	CORS CORSConfig `json:"cors"`
}

// CORSConfig defines allowed origins, methods, and headers for browser clients.
type CORSConfig struct {
	AllowOrigins     []string `json:"allow_origins"`
	AllowMethods     []string `json:"allow_methods"`
	AllowHeaders     []string `json:"allow_headers"`
	ExposeHeaders    []string `json:"expose_headers"`
	AllowCredentials bool     `json:"allow_credentials"`
	MaxAge           int      `json:"max_age"`
}

// IntegrationsConfig holds downstream service endpoints and clients.
type IntegrationsConfig struct {
	IdentifyGRPCURL string `json:"identify_grpc_url"`
}

// Load builds the config using environment variables with sensible defaults.
func Load() *Config {
	migrationsPath := getEnv("CONFIG_DB_MIGRATIONS_PATH", "../../pkg/database/migrations/postgres/catalog")

	return &Config{
		Server: ServerConfig{
			Host:         getEnv("CONFIG_SERVICE_HOST", "localhost"),
			Port:         getEnvAsInt("CONFIG_SERVICE_PORT", 8082),
			GRPCPort:     getEnvAsInt("CONFIG_SERVICE_GRPC_PORT", 40002),
			ReadTimeout:  getEnvAsDuration("CONFIG_SERVICE_READ_TIMEOUT", 30*time.Second),
			WriteTimeout: getEnvAsDuration("CONFIG_SERVICE_WRITE_TIMEOUT", 30*time.Second),
			IdleTimeout:  getEnvAsDuration("CONFIG_SERVICE_IDLE_TIMEOUT", 120*time.Second),
			Environment:  getEnv("CONFIG_ENVIRONMENT", "development"),
		},
		Database: DatabaseConfig{
			DatabaseConfig: &dbconfig.DatabaseConfig{
				Primary: &dbconfig.RelationalConfig{
					Type:            interfaces.PostgreSQL,
					Host:            getEnv("CONFIG_DB_HOST", "localhost"),
					Port:            getEnvAsInt("CONFIG_DB_PORT", 5432),
					Database:        getEnv("CONFIG_DB_NAME", "catalog_service"),
					Username:        getEnv("CONFIG_DB_USER", "catalog_service"),
					Password:        getEnv("CONFIG_DB_PASSWORD", "catalog_service"),
					SSLMode:         getEnv("CONFIG_DB_SSL_MODE", "disable"),
					MaxConns:        int32(getEnvAsInt("CONFIG_DB_MAX_CONNS", 30)),
					MinConns:        int32(getEnvAsInt("CONFIG_DB_MIN_CONNS", 5)),
					MaxConnLifetime: getEnvAsDuration("CONFIG_DB_MAX_CONN_LIFETIME", time.Hour),
					MaxConnIdleTime: getEnvAsDuration("CONFIG_DB_MAX_CONN_IDLE_TIME", 30*time.Minute),
				},
			},
			MigrationsPath: migrationsPath,
		},
		Localization: LocalizationConfig{
			DefaultLanguage:    getEnv("CONFIG_LOCALE_DEFAULT", "en"),
			SupportedLanguages: splitAndTrim(getEnv("CONFIG_LOCALE_SUPPORTED", "en,vi")),
			BundlePath:         getEnv("CONFIG_LOCALE_BUNDLE_PATH", "../../pkg/i18n/locales"),
			QueryParam:         getEnv("CONFIG_LOCALE_QUERY_PARAM", "lang"),
			HeaderName:         getEnv("CONFIG_LOCALE_HEADER", "Accept-Language"),
			CookieName:         getEnv("CONFIG_LOCALE_COOKIE", "locale"),
		},
		Content: ContentConfig{
			EnableAnime: getEnvAsBool("CONFIG_FEATURE_ANIME", true),
			EnableManga: getEnvAsBool("CONFIG_FEATURE_MANGA", true),
			EnableNovel: getEnvAsBool("CONFIG_FEATURE_NOVEL", true),
		},
		Media: MediaConfig{
			CDNBaseURL:      getEnv("CONFIG_MEDIA_CDN_BASE_URL", ""),
			ImageProxyURL:   getEnv("CONFIG_MEDIA_IMAGE_PROXY_URL", ""),
			SignedURLExpiry: getEnvAsDuration("CONFIG_MEDIA_SIGNED_URL_EXPIRY", 15*time.Minute),
		},
		Security: SecurityConfig{
			CORS: CORSConfig{
				AllowOrigins:     splitAndTrim(getEnv("CONFIG_CORS_ALLOW_ORIGINS", "*")),
				AllowMethods:     splitAndTrim(getEnv("CONFIG_CORS_ALLOW_METHODS", "GET,POST,PUT,PATCH,DELETE,OPTIONS")),
				AllowHeaders:     splitAndTrim(getEnv("CONFIG_CORS_ALLOW_HEADERS", "Authorization,Content-Type,Accept-Language")),
				ExposeHeaders:    splitAndTrim(getEnv("CONFIG_CORS_EXPOSE_HEADERS", "Content-Length,Content-Type")),
				AllowCredentials: getEnvAsBool("CONFIG_CORS_ALLOW_CREDENTIALS", true),
				MaxAge:           getEnvAsInt("CONFIG_CORS_MAX_AGE", 3600),
			},
		},
		Integrations: IntegrationsConfig{
			IdentifyGRPCURL: getEnv("CONFIG_IDENTIFY_GRPC_URL", "localhost:9090"),
		},
	}
}

// Helper functions ---------------------------------------------------------

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
		log.Printf("Invalid integer for %s: %s, defaulting to %d", key, value, defaultValue)
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
		log.Printf("Invalid boolean for %s: %s, defaulting to %t", key, value, defaultValue)
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
		log.Printf("Invalid duration for %s: %s, defaulting to %s", key, value, defaultValue)
	}
	return defaultValue
}

func splitAndTrim(value string) []string {
	if value == "" {
		return []string{}
	}
	parts := []string{}
	for _, item := range strings.Split(value, ",") {
		trimmed := strings.TrimSpace(item)
		if trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return parts
}
