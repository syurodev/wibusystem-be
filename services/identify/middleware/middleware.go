package middleware

import (
	"time"

	"github.com/gin-gonic/gin"

	"wibusystem/services/identify/config"
	"wibusystem/pkg/i18n"
	"wibusystem/services/identify/oauth2"
	"wibusystem/services/identify/repositories"
)

// Manager aggregates all middleware and provides helper combinations
// tailored to specific route groups (API, tenant, admin, OAuth2).
type Manager struct {
	Auth      *AuthMiddleware
	Tenant    *TenantMiddleware
	RateLimit *RateLimiter
	Config    *config.Config
	Locale    *LocaleMiddleware
}

// NewManager creates a new middleware manager.
func NewManager(cfg *config.Config, provider *oauth2.Provider, repos *repositories.Repositories, translator *i18n.Translator) *Manager {
	queryParam := cfg.Localization.QueryParam
	if queryParam == "" {
		queryParam = "lang"
	}

	var locale *LocaleMiddleware
	if translator != nil {
		locale = NewLocaleMiddleware(translator, queryParam)
	}

	return &Manager{
		Auth:      NewAuthMiddleware(provider, repos),
		Tenant:    NewTenantMiddleware(repos),
		RateLimit: NewRateLimiter(cfg.Security.RateLimit.RequestsPerMinute, time.Minute),
		Config:    cfg,
		Locale:    locale,
	}
}

// SetupCommonMiddleware sets up common middleware for all routes including
// security headers, logging, request ID, error handling, API versioning,
// and global rate limiting.
func (m *Manager) SetupCommonMiddleware(router *gin.Engine) {
	// Localization should run first so downstream middleware can rely on the
	// request-scoped localizer.
	if m.Locale != nil {
		router.Use(m.Locale.Handler())
	}
	// Security headers
	router.Use(SecurityHeaders())

	// Request logging
	router.Use(RequestLogger())

	// Request ID
	router.Use(RequestID())

	// Error handling
	router.Use(ErrorHandler())

	// API versioning
	router.Use(APIVersionMiddleware())

	// Rate limiting (global)
	router.Use(m.RateLimit.RateLimit())
}

// SetupAPIMiddleware returns middleware for general API routes with
// optional authentication and tenant resolution.
func (m *Manager) SetupAPIMiddleware() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		// Content type validation
		ValidateContentType(),

		// Optional authentication (for public endpoints)
		m.Auth.OptionalAuth(),

		// Tenant resolution (tries to resolve tenant from various sources)
		m.Tenant.ResolveTenant(),
	}
}

// SetupProtectedAPIMiddleware returns middleware for protected API routes
// that require authentication.
func (m *Manager) SetupProtectedAPIMiddleware() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		// Content type validation
		ValidateContentType(),

		// Required authentication
		m.Auth.RequireAuth(),

		// Tenant resolution
		m.Tenant.ResolveTenant(),
	}
}

// SetupTenantAPIMiddleware returns middleware for tenant-scoped routes
// that require authentication, resolved tenant, and active membership.
func (m *Manager) SetupTenantAPIMiddleware() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		// Content type validation
		ValidateContentType(),

		// Required authentication
		m.Auth.RequireAuth(),

		// Tenant resolution
		m.Tenant.ResolveTenant(),

		// Require tenant context
		m.Tenant.RequireTenant(),

		// Require tenant membership
		m.Tenant.RequireTenantMembership(),
	}
}

// SetupAdminMiddleware returns middleware for admin routes. In production,
// consider adding IP allowlists or additional safeguards.
func (m *Manager) SetupAdminMiddleware() []gin.HandlerFunc {
	middleware := []gin.HandlerFunc{
		// Content type validation
		ValidateContentType(),

		// Required authentication
		m.Auth.RequireAuth(),

		// Require admin scope
		m.Auth.RequireScope("admin"),

		// Require admin privileges
		m.Auth.RequireAdmin(),
	}

	// Add IP whitelist in production
	if m.Config.Server.Environment == "production" {
		// In production, you might want to restrict admin access to specific IPs
		// adminIPs := []string{"127.0.0.1", "10.0.0.0/8"}
		// middleware = append(middleware, IPWhitelist(adminIPs))
	}

	return middleware
}

// SetupOAuth2Middleware returns middleware for OAuth2 endpoints.
// Uses a stricter rate limit as these endpoints are sensitive.
func (m *Manager) SetupOAuth2Middleware() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		// More restrictive rate limiting for OAuth2 endpoints
		NewRateLimiter(30, time.Minute).RateLimit(), // 30 requests per minute

		// Content type validation
		ValidateContentType(),
	}
}

// Convenience methods for common middleware combinations

// RequireAuthAndTenant combines auth and tenant membership requirements.
func (m *Manager) RequireAuthAndTenant() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		m.Auth.RequireAuth(),
		m.Tenant.ResolveTenant(),
		m.Tenant.RequireTenant(),
		m.Tenant.RequireTenantMembership(),
	}
}

// RequireAuthAndScope requires authentication and specific scopes.
func (m *Manager) RequireAuthAndScope(scopes ...string) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		m.Auth.RequireAuth(),
		m.Auth.RequireScope(scopes...),
	}
}

// RequireTenantRole requires tenant membership and specific role.
func (m *Manager) RequireTenantRole(roles ...string) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		m.Auth.RequireAuth(),
		m.Tenant.ResolveTenant(),
		m.Tenant.RequireTenant(),
		m.Tenant.RequireTenantMembership(),
		m.Tenant.RequireTenantRole(roles...),
	}
}

// RequireTenantPermission requires tenant membership and specific permission.
// Not all permission checks are implemented yet; see TenantMiddleware.
func (m *Manager) RequireTenantPermission(permissions ...string) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		m.Auth.RequireAuth(),
		m.Tenant.ResolveTenant(),
		m.Tenant.RequireTenant(),
		m.Tenant.RequireTenantMembership(),
		m.Tenant.RequireTenantPermission(permissions...),
	}
}
