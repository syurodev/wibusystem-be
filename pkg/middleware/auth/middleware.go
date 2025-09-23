package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	r "wibusystem/pkg/common/response"
)

// Middleware provides authentication middleware functionality
type Middleware struct {
	grpcClient *GRPCClient
	cache      *Cache
	config     *Config
}

// NewMiddleware creates a new authentication middleware
func NewMiddleware(config *Config) (*Middleware, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// Create gRPC client
	grpcClient, err := NewGRPCClient(config)
	if err != nil {
		return nil, err
	}

	// Create cache
	cache := NewCache(config.CacheEnabled, config.CacheTTL, config.MaxCacheSize)

	return &Middleware{
		grpcClient: grpcClient,
		cache:      cache,
		config:     config,
	}, nil
}

// RequireAuth returns a middleware that requires valid authentication
func (m *Middleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip authentication for configured paths
		if m.shouldSkipPath(c.Request.URL.Path) {
			c.Next()
			return
		}

		// Extract token from request
		token, err := m.extractToken(c)
		if err != nil {
			m.respondUnauthorized(c, "Missing or invalid authorization token")
			return
		}

		// Validate token
		result, err := m.validateToken(c.Request.Context(), token)
		if err != nil {
			m.respondError(c, "Token validation failed", http.StatusInternalServerError)
			return
		}

		if !result.Valid {
			m.respondUnauthorized(c, result.Error)
			return
		}

		// Set authentication context
		m.setAuthContext(c, token, result.UserContext)

		c.Next()
	}
}

// RequireScope returns a middleware that requires specific scopes
func (m *Middleware) RequireScope(scopes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// First check if user is authenticated
		user, exists := GetUserFromContext(c)
		if !exists {
			m.respondUnauthorized(c, "Authentication required")
			return
		}

		// Check if user has required scopes
		if !user.HasAllScopes(scopes...) {
			m.respondForbidden(c, "Insufficient permissions")
			return
		}

		c.Next()
	}
}

// RequireAnyScope returns a middleware that requires at least one of the specified scopes
func (m *Middleware) RequireAnyScope(scopes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// First check if user is authenticated
		user, exists := GetUserFromContext(c)
		if !exists {
			m.respondUnauthorized(c, "Authentication required")
			return
		}

		// Check if user has any of the required scopes
		if !user.HasAnyScope(scopes...) {
			m.respondForbidden(c, "Insufficient permissions")
			return
		}

		c.Next()
	}
}

// RequireAdmin returns a middleware that requires admin privileges
func (m *Middleware) RequireAdmin() gin.HandlerFunc {
	return m.RequireScope("admin")
}

// OptionalAuth returns a middleware that optionally authenticates but doesn't require it
func (m *Middleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip authentication for configured paths
		if m.shouldSkipPath(c.Request.URL.Path) {
			c.Next()
			return
		}

		// Try to extract token from request
		token, err := m.extractToken(c)
		if err != nil {
			// No token found, continue without authentication
			c.Next()
			return
		}

		// Try to validate token
		result, err := m.validateToken(c.Request.Context(), token)
		if err != nil || !result.Valid {
			// Invalid token, continue without authentication
			c.Next()
			return
		}

		// Set authentication context if token is valid
		m.setAuthContext(c, token, result.UserContext)

		c.Next()
	}
}

// extractToken extracts the authentication token from the request
func (m *Middleware) extractToken(c *gin.Context) (string, error) {
	// Get authorization header
	authHeader := c.GetHeader(m.config.TokenHeader)
	if authHeader == "" {
		return "", ErrMissingToken
	}

	// Check for Bearer prefix
	prefix := m.config.TokenPrefix + " "
	if !strings.HasPrefix(authHeader, prefix) {
		return "", ErrInvalidTokenFormat
	}

	// Extract token
	token := strings.TrimSpace(authHeader[len(prefix):])
	if token == "" {
		return "", ErrMissingToken
	}

	return token, nil
}

// validateToken validates a token using cache and/or gRPC call
func (m *Middleware) validateToken(ctx context.Context, token string) (*ValidationResult, error) {
	// Try cache first if enabled
	if m.config.CacheEnabled {
		if cachedResult, found := m.cache.Get(token); found {
			return cachedResult, nil
		}
	}

	// Validate via gRPC
	result, err := m.grpcClient.ValidateToken(ctx, token)
	if err != nil {
		return nil, err
	}

	// Cache the result if caching is enabled and result is valid
	if m.config.CacheEnabled && result != nil {
		m.cache.Set(token, result)
	}

	return result, nil
}

// setAuthContext sets authentication information in the Gin context
func (m *Middleware) setAuthContext(c *gin.Context, token string, user *UserContext) {
	c.Set(UserContextKey, user)
	c.Set(TokenKey, token)
	c.Set(AuthenticatedKey, true)
}

// shouldSkipPath checks if a path should skip authentication
func (m *Middleware) shouldSkipPath(path string) bool {
	for _, skipPath := range m.config.SkipPaths {
		if path == skipPath || strings.HasPrefix(path, skipPath) {
			return true
		}
	}
	return false
}

// respondUnauthorized sends a 401 Unauthorized response
func (m *Middleware) respondUnauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, r.StandardResponse{
		Success: false,
		Message: "Unauthorized",
		Data:    nil,
		Error: &r.ErrorDetail{
			Code:        "unauthorized",
			Description: message,
		},
		Meta: map[string]interface{}{},
	})
	c.Abort()
}

// respondForbidden sends a 403 Forbidden response
func (m *Middleware) respondForbidden(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, r.StandardResponse{
		Success: false,
		Message: "Forbidden",
		Data:    nil,
		Error: &r.ErrorDetail{
			Code:        "forbidden",
			Description: message,
		},
		Meta: map[string]interface{}{},
	})
	c.Abort()
}

// respondError sends an error response
func (m *Middleware) respondError(c *gin.Context, message string, statusCode int) {
	c.JSON(statusCode, r.StandardResponse{
		Success: false,
		Message: "Authentication Error",
		Data:    nil,
		Error: &r.ErrorDetail{
			Code:        "auth_error",
			Description: message,
		},
		Meta: map[string]interface{}{},
	})
	c.Abort()
}

// Close cleans up the middleware resources
func (m *Middleware) Close() error {
	if m.cache != nil {
		m.cache.Stop()
	}
	if m.grpcClient != nil {
		return m.grpcClient.Close()
	}
	return nil
}

// HealthCheck performs a health check of the authentication system
func (m *Middleware) HealthCheck(ctx context.Context) error {
	return m.grpcClient.HealthCheck(ctx)
}