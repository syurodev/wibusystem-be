// Package middleware contains HTTP middleware for the Identity module.
package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	AllowCredentials bool
	ExposeHeaders    []string
	MaxAge           int
}

// DefaultCORSConfig returns the default CORS configuration
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowOrigins: []string{
			"http://localhost:3000",
			"http://localhost:3001",
			"http://localhost:5173",
			"http://localhost:5174",
		},
		AllowMethods: []string{
			"GET",
			"POST",
			"PUT",
			"PATCH",
			"DELETE",
			"OPTIONS",
			"HEAD",
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
			"X-Requested-With",
			"X-Request-ID",
			"X-Tenant-ID",
		},
		AllowCredentials: true,
		ExposeHeaders: []string{
			"Content-Length",
			"Content-Type",
			"X-Request-ID",
		},
		MaxAge: 86400, // 24 hours
	}
}

// ProductionCORSConfig returns CORS configuration for production
func ProductionCORSConfig(allowedOrigins []string) CORSConfig {
	config := DefaultCORSConfig()
	config.AllowOrigins = allowedOrigins
	return config
}

// CORSMiddleware creates a CORS middleware with the given configuration
func CORSMiddleware(config CORSConfig) fiber.Handler {
	return cors.New(cors.Config{
		AllowOrigins:     strings.Join(config.AllowOrigins, ","),
		AllowMethods:     strings.Join(config.AllowMethods, ","),
		AllowHeaders:     strings.Join(config.AllowHeaders, ","),
		AllowCredentials: config.AllowCredentials,
		ExposeHeaders:    strings.Join(config.ExposeHeaders, ","),
		MaxAge:           config.MaxAge,
	})
}

// DevelopmentCORS returns a configurable CORS middleware for development environments.
// It enables credentials support so that cookies/session data work during local development.
func DevelopmentCORS(allowedOrigins []string) fiber.Handler {
	config := DefaultCORSConfig()

	if len(allowedOrigins) > 0 {
		config.AllowOrigins = allowedOrigins
	}

	return CORSMiddleware(config)
}

// ProductionCORS returns a strict CORS middleware for production
func ProductionCORS(allowedOrigins []string) fiber.Handler {
	config := ProductionCORSConfig(allowedOrigins)
	return CORSMiddleware(config)
}

// CustomOriginValidator creates a CORS middleware with custom origin validation
func CustomOriginValidator(validator func(origin string) bool) fiber.Handler {
	return cors.New(cors.Config{
		AllowOriginsFunc: func(origin string) bool {
			return validator(origin)
		},
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS,HEAD",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization,X-Requested-With,X-Request-ID,X-Tenant-ID",
		AllowCredentials: true,
		MaxAge:           86400,
	})
}

// IsDevelopmentOrigin checks if an origin is a development origin
func IsDevelopmentOrigin(origin string) bool {
	developmentOrigins := []string{
		"http://localhost",
		"http://127.0.0.1",
		"http://0.0.0.0",
	}

	for _, dev := range developmentOrigins {
		if strings.HasPrefix(origin, dev) {
			return true
		}
	}

	return false
}

// IsProductionOrigin checks if an origin matches production domains
func IsProductionOrigin(origin string, allowedDomains []string) bool {
	for _, domain := range allowedDomains {
		if strings.HasSuffix(origin, domain) {
			return true
		}
	}
	return false
}
