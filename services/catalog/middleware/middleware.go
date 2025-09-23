package middleware

import (
	"time"

	"github.com/gin-gonic/gin"

	"wibusystem/pkg/i18n"
	"wibusystem/services/catalog/config"
)

// Manager wires middleware commonly used by the Catalog service.
type Manager struct {
	Locale    *LocaleMiddleware
	RateLimit *RateLimiter
	Config    *config.Config
}

// NewManager constructs a middleware manager aligned with Identify service setup.
func NewManager(cfg *config.Config, translator *i18n.Translator) *Manager {
	var locale *LocaleMiddleware
	if translator != nil {
		locale = NewLocaleMiddleware(translator)
	}

	limit := 120
	if cfg.Server.Environment == "production" {
		limit = 60
	}

	return &Manager{
		Locale:    locale,
		RateLimit: NewRateLimiter(limit, time.Minute),
		Config:    cfg,
	}
}

// SetupCommonMiddleware applies common middleware chain for all HTTP routes.
func (m *Manager) SetupCommonMiddleware(router *gin.Engine) {
	if m.Locale != nil {
		router.Use(m.Locale.Handler())
	}

	router.Use(SecurityHeaders())
	router.Use(RequestLogger())
	router.Use(RequestID())
	router.Use(ErrorHandler())
	router.Use(APIVersionMiddleware())

	if m.RateLimit != nil {
		router.Use(m.RateLimit.RateLimit())
	}
}

// SetupPublicAPIMiddleware returns middleware for public API routes.
func (m *Manager) SetupPublicAPIMiddleware() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		ValidateContentType(),
	}
}

// SetupInternalAPIMiddleware returns middleware chain for internal routes constrained by stricter limits.
func (m *Manager) SetupInternalAPIMiddleware() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		ValidateContentType(),
		NewRateLimiter(60, time.Minute).RateLimit(),
	}
}
