package configs

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config là struct tổng hợp chứa tất cả các cấu hình của ứng dụng.
type Config struct {
	Server   ServerConfig
	DB       DatabaseConfig
	Redis    RedisConfig
	CORS     CORSConfig
	Log      LogConfig
	OAuth2   OAuthConfig
	Email    EmailConfig
	WebAuthn WebAuthnConfig
}

// OAuthConfig chứa cấu hình cho OAuth 2.0 Server.
type OAuthConfig struct {
	Issuer     string
	PrivateKey *rsa.PrivateKey
	KeyID      string
	HMACSecret string
}

// ServerConfig chứa cấu hình cho Web Server và gRPC.
type ServerConfig struct {
	Port     string
	GRPCPort string
	Env      string // development, staging, production
	IsProd   bool
}

// DatabaseConfig chứa cấu hình kết nối PostgreSQL.
type DatabaseConfig struct {
	Host            string
	Port            string
	Name            string
	User            string
	Password        string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration

	// Schemas
	IdentifySchema  string
	CatalogSchema   string
	CommunitySchema string
	PaymentSchema   string
}

// RedisConfig chứa cấu hình kết nối Redis.
type RedisConfig struct {
	Host         string
	Port         string
	Password     string
	DB           int
	MaxRetries   int
	PoolSize     int
	MinIdleConns int
}

// CORSConfig chứa cấu hình CORS cho Gin.
type CORSConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	AllowCredentials bool
	MaxAge           int // tính bằng giây
}

// LogConfig chứa cấu hình cho Zap Logger.
type LogConfig struct {
	Level        string // info, debug, error
	DBLogQueries bool
}

// EmailConfig chứa cấu hình cho Email Service (Resend).
type EmailConfig struct {
	Provider         string // resend, sendgrid, ses, etc.
	APIKey           string
	FromEmail        string
	FromName         string
	BaseURL          string // Base URL for email links (e.g., https://yourdomain.com)
	VerificationURL  string // URL template for email verification
	PasswordResetURL string // URL template for password reset
}

// WebAuthnConfig chứa cấu hình cho WebAuthn/Passkey authentication.
type WebAuthnConfig struct {
	RPID      string   // Relying Party ID (e.g., "example.com" or "localhost")
	RPName    string   // Relying Party Name (e.g., "Example Corp")
	RPOrigins []string // Allowed origins (e.g., ["https://example.com", "http://localhost:8080"])
	Timeout   int      // Timeout in milliseconds (default: 60000)
}

// LoadConfig tải file .env và ánh xạ các biến môi trường vào struct Config.
func LoadConfig(envPath string) (*Config, error) {
	// 1. Tải file .env
	if err := godotenv.Load(envPath); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("error loading .env file: %w", err)
	}

	cfg := &Config{}

	// SERVER CONFIG
	cfg.Server.Port = getEnv("SERVER_PORT", "8080")
	cfg.Server.GRPCPort = getEnv("SERVER_GRPC_PORT", "50051")
	cfg.Server.Env = getEnv("ENVIRONMENT", "development")
	cfg.Server.IsProd = cfg.Server.Env == "production"

	// DATABASE CONFIG
	cfg.DB.Host = getEnv("DB_HOST", "localhost")
	cfg.DB.Port = getEnv("DB_PORT", "5432")
	cfg.DB.Name = getEnv("DB_NAME", "system_dev")
	cfg.DB.User = getEnv("DB_USER", "system_dev")
	cfg.DB.Password = getEnv("DB_PASSWORD", "system_dev")
	cfg.DB.SSLMode = getEnv("DB_SSL_MODE", "disable")
	cfg.DB.MaxOpenConns = getEnvAsInt("DB_MAX_OPEN_CONNS", 25)
	cfg.DB.MaxIdleConns = getEnvAsInt("DB_MAX_IDLE_CONNS", 5)
	connLifetime, err := time.ParseDuration(getEnv("DB_CONN_MAX_LIFETIME", "5m"))
	if err != nil {
		return nil, fmt.Errorf("invalid DB_CONN_MAX_LIFETIME format: %w", err)
	}
	cfg.DB.ConnMaxLifetime = connLifetime
	cfg.DB.IdentifySchema = getEnv("DB_IDENTIFY_SCHEMA", "identify")
	cfg.DB.CatalogSchema = getEnv("DB_CATALOG_SCHEMA", "catalog")
	cfg.DB.CommunitySchema = getEnv("DB_COMMUNITY_SCHEMA", "community")
	cfg.DB.PaymentSchema = getEnv("DB_PAYMENT_SCHEMA", "payment")

	// REDIS CONFIG
	cfg.Redis.Host = getEnv("REDIS_HOST", "localhost")
	cfg.Redis.Port = getEnv("REDIS_PORT", "6379")
	cfg.Redis.Password = getEnv("REDIS_PASSWORD", "")
	cfg.Redis.DB = getEnvAsInt("REDIS_DB", 0)
	cfg.Redis.MaxRetries = getEnvAsInt("REDIS_MAX_RETRIES", 3)
	cfg.Redis.PoolSize = getEnvAsInt("REDIS_POOL_SIZE", 10)
	cfg.Redis.MinIdleConns = getEnvAsInt("REDIS_MIN_IDLE_CONNS", 5)

	// CORS CONFIG
	cfg.CORS.AllowOrigins = getEnvAsSlice("CORS_ALLOW_ORIGINS", nil)
	cfg.CORS.AllowMethods = getEnvAsSlice("CORS_ALLOW_METHODS", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"})
	cfg.CORS.AllowHeaders = getEnvAsSlice("CORS_ALLOW_HEADERS", []string{"Origin", "Content-Type", "Authorization", "Accept-Language"})
	cfg.CORS.ExposeHeaders = getEnvAsSlice("CORS_EXPOSE_HEADERS", []string{"Content-Length"})
	cfg.CORS.AllowCredentials = getEnvAsBool("CORS_ALLOW_CREDENTIALS", true)
	cfg.CORS.MaxAge = getEnvAsInt("CORS_MAX_AGE", 43200)

	// LOG CONFIG
	cfg.Log.Level = getEnv("LOG_LEVEL", "info")
	cfg.Log.DBLogQueries = getEnvAsBool("DB_LOG_QUERIES", false)

	// OAUTH2 CONFIG
	cfg.OAuth2.Issuer = getEnv("OAUTH2_ISSUER", "http://localhost:8080")
	cfg.OAuth2.KeyID = getEnv("OAUTH2_KEY_ID", "default-key-id")
	privateKeyPath := getEnv("OAUTH2_PRIVATE_KEY_PATH", "")
	if privateKeyPath != "" {
		keyData, err := os.ReadFile(privateKeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read private key file %s: %w", privateKeyPath, err)
		}
		block, _ := pem.Decode(keyData)
		if block == nil {
			return nil, fmt.Errorf("failed to decode PEM block from %s", privateKeyPath)
		}
		privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
		cfg.OAuth2.PrivateKey = privateKey
	}

	cfg.OAuth2.HMACSecret = getEnv("OAUTH2_HMAC_SECRET", "")
	if len(cfg.OAuth2.HMACSecret) < 32 {
		return nil, fmt.Errorf("OAUTH2_HMAC_SECRET must be at least 32 bytes long")
	}

	// EMAIL CONFIG
	cfg.Email.Provider = getEnv("EMAIL_PROVIDER", "resend")
	cfg.Email.APIKey = getEnv("EMAIL_API_KEY", "")
	cfg.Email.FromEmail = getEnv("EMAIL_FROM_EMAIL", "noreply@localhost")
	cfg.Email.FromName = getEnv("EMAIL_FROM_NAME", "System")
	cfg.Email.BaseURL = getEnv("EMAIL_BASE_URL", "http://localhost:8080")
	cfg.Email.VerificationURL = getEnv("EMAIL_VERIFICATION_URL", cfg.Email.BaseURL+"/oauth2/verify-email?token={{.Token}}")
	cfg.Email.PasswordResetURL = getEnv("EMAIL_PASSWORD_RESET_URL", cfg.Email.BaseURL+"/oauth2/reset-password?token={{.Token}}")

	// WEBAUTHN CONFIG
	cfg.WebAuthn.RPID = getEnv("WEBAUTHN_RP_ID", "localhost")
	cfg.WebAuthn.RPName = getEnv("WEBAUTHN_RP_NAME", "System")
	cfg.WebAuthn.RPOrigins = getEnvAsSlice("WEBAUTHN_RP_ORIGINS", []string{"http://localhost:8080", "http://localhost:3000"})
	cfg.WebAuthn.Timeout = getEnvAsInt("WEBAUTHN_TIMEOUT", 60000)

	return cfg, nil
}

// --- Helper Functions ---

func getEnv(key string, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := getEnv(key, "")
	if valueStr == "true" {
		return true
	}
	if valueStr == "false" {
		return false
	}
	return defaultValue
}

func getEnvAsSlice(key string, defaultValue []string) []string {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}
	return strings.Split(strings.ReplaceAll(valueStr, " ", ""), ",")
}
