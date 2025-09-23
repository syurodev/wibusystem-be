package middleware

import (
	"log"

	"github.com/gin-gonic/gin"
	authmw "wibusystem/pkg/middleware/auth"
	"wibusystem/services/catalog/config"
)

// AuthMiddleware wraps the shared authentication middleware for the catalog service
type AuthMiddleware struct {
	*authmw.Middleware
}

// NewAuthMiddleware creates a new authentication middleware for the catalog service
func NewAuthMiddleware(cfg *config.Config) (*AuthMiddleware, error) {
	// Create auth config from catalog service config
	authConfig := &authmw.Config{
		IdentifyGRPCURL: cfg.Integrations.IdentifyGRPCURL,
		GRPCTimeout:     cfg.Server.ReadTimeout, // Use server read timeout for gRPC calls
		CacheEnabled:    true,                   // Enable caching for better performance
		CacheTTL:        cfg.Media.SignedURLExpiry, // Use media expiry as a reasonable cache TTL
		MaxCacheSize:    1000,
		TokenHeader:     "Authorization",
		TokenPrefix:     "Bearer",
		SkipPaths: []string{
			"/health",
			"/metrics",
			"/favicon.ico",
			"/api/v1/public",  // Public API endpoints
			"/.well-known",    // Well-known endpoints
		},
	}

	// Create the middleware
	middleware, err := authmw.NewMiddleware(authConfig)
	if err != nil {
		return nil, err
	}

	log.Printf("Authentication middleware initialized with identify service at %s", authConfig.IdentifyGRPCURL)

	return &AuthMiddleware{
		Middleware: middleware,
	}, nil
}

// SetupProtectedAPIMiddleware returns middleware for protected API routes
// This includes authentication but no specific scope requirements
func (a *AuthMiddleware) SetupProtectedAPIMiddleware() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		a.RequireAuth(),
	}
}

// SetupAdminAPIMiddleware returns middleware for admin API routes
// This includes authentication and admin scope requirement
func (a *AuthMiddleware) SetupAdminAPIMiddleware() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		a.RequireAuth(),
		a.RequireAdmin(),
	}
}

// SetupOptionalAuthMiddleware returns middleware for routes with optional authentication
func (a *AuthMiddleware) SetupOptionalAuthMiddleware() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		a.OptionalAuth(),
	}
}