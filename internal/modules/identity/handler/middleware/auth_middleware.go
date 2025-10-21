// Package middleware contains HTTP middleware for the Identity module.
package middleware

import (
	"context"
	"strings"

	"wibusystem/internal/modules/identity/service"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// ContextKey is a custom type for context keys to avoid collisions
type ContextKey string

const (
	// UserIDKey is the context key for user ID
	UserIDKey ContextKey = "user_id"
	// SessionIDKey is the context key for session ID
	SessionIDKey ContextKey = "session_id"
	// TenantIDKey is the context key for tenant ID
	TenantIDKey ContextKey = "tenant_id"
)

// AuthMiddleware creates a middleware that validates session tokens
func AuthMiddleware(authService service.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Extract token from Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return NewAppError(
				ErrUnauthorized,
				"Missing authorization header",
				fiber.StatusUnauthorized,
			)
		}

		// Extract token (support both "Bearer <token>" and "<token>")
		token := extractToken(authHeader)
		if token == "" {
			return NewAppError(
				ErrUnauthorized,
				"Invalid authorization header format",
				fiber.StatusUnauthorized,
			)
		}

		// Validate session token
		ctx := c.Context()
		session, err := authService.ValidateSession(ctx, token)
		if err != nil {
			return NewAppError(
				ErrUnauthorized,
				"Invalid or expired session",
				fiber.StatusUnauthorized,
			).WithCode("INVALID_SESSION")
		}

		// Store user ID and session ID in context
		c.Locals(string(UserIDKey), session.UserID)
		c.Locals(string(SessionIDKey), session.ID)

		return c.Next()
	}
}

// OptionalAuthMiddleware is like AuthMiddleware but doesn't fail if no token is provided
func OptionalAuthMiddleware(authService service.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			// No auth header, continue without authentication
			return c.Next()
		}

		token := extractToken(authHeader)
		if token == "" {
			// Invalid format, but optional, so continue
			return c.Next()
		}

		// Try to validate session
		ctx := c.Context()
		session, err := authService.ValidateSession(ctx, token)
		if err != nil {
			// Invalid session, but optional, so continue
			return c.Next()
		}

		// Store user ID and session ID in context
		c.Locals(string(UserIDKey), session.UserID)
		c.Locals(string(SessionIDKey), session.ID)

		return c.Next()
	}
}

// RequireTenantMembership creates a middleware that checks if the user is a member of a specific tenant
func RequireTenantMembership(tenantService service.TenantService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get user ID from context
		userID, ok := GetUserID(c)
		if !ok {
			return NewAppError(
				ErrUnauthorized,
				"User not authenticated",
				fiber.StatusUnauthorized,
			)
		}

		// Get tenant ID from route params
		tenantIDStr := c.Params("tenantId")
		if tenantIDStr == "" {
			return NewAppError(
				ErrBadRequest,
				"Tenant ID is required",
				fiber.StatusBadRequest,
			)
		}

		tenantID, err := uuid.Parse(tenantIDStr)
		if err != nil {
			return NewAppError(
				ErrBadRequest,
				"Invalid tenant ID format",
				fiber.StatusBadRequest,
			)
		}

		// Check if user is a member of the tenant
		ctx := c.Context()
		isMember, err := tenantService.IsMember(ctx, tenantID, userID)
		if err != nil {
			return NewAppError(
				ErrInternalServer,
				"Failed to check tenant membership",
				fiber.StatusInternalServerError,
			)
		}

		if !isMember {
			return NewAppError(
				ErrForbidden,
				"You are not a member of this tenant",
				fiber.StatusForbidden,
			).WithCode("NOT_TENANT_MEMBER")
		}

		// Store tenant ID in context
		c.Locals(string(TenantIDKey), tenantID)

		return c.Next()
	}
}

// RequireTenantRole creates a middleware that checks if the user has a specific role in a tenant
func RequireTenantRole(tenantService service.TenantService, requiredRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get user ID from context
		userID, ok := GetUserID(c)
		if !ok {
			return NewAppError(
				ErrUnauthorized,
				"User not authenticated",
				fiber.StatusUnauthorized,
			)
		}

		// Get tenant ID from route params
		tenantIDStr := c.Params("tenantId")
		if tenantIDStr == "" {
			return NewAppError(
				ErrBadRequest,
				"Tenant ID is required",
				fiber.StatusBadRequest,
			)
		}

		tenantID, err := uuid.Parse(tenantIDStr)
		if err != nil {
			return NewAppError(
				ErrBadRequest,
				"Invalid tenant ID format",
				fiber.StatusBadRequest,
			)
		}

		// Get user's role in the tenant
		ctx := c.Context()
		member, err := tenantService.GetMember(ctx, tenantID, userID)
		if err != nil {
			return NewAppError(
				ErrForbidden,
				"You are not a member of this tenant",
				fiber.StatusForbidden,
			).WithCode("NOT_TENANT_MEMBER")
		}

		// Check if user has one of the required roles
		hasRole := false
		for _, role := range requiredRoles {
			if string(member.Role) == role {
				hasRole = true
				break
			}
		}

		if !hasRole {
			return NewAppError(
				ErrForbidden,
				"You don't have permission to perform this action",
				fiber.StatusForbidden,
			).WithCode("INSUFFICIENT_PERMISSIONS")
		}

		// Store tenant ID in context
		c.Locals(string(TenantIDKey), tenantID)

		return c.Next()
	}
}

// Helper functions

// extractToken extracts the token from the Authorization header
// Supports both "Bearer <token>" and "<token>" formats
func extractToken(authHeader string) string {
	parts := strings.Split(authHeader, " ")
	if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
		return parts[1]
	}
	// If no "Bearer" prefix, assume the entire header is the token
	return authHeader
}

// GetUserID retrieves the user ID from the fiber context
func GetUserID(c *fiber.Ctx) (uuid.UUID, bool) {
	userID := c.Locals(string(UserIDKey))
	if userID == nil {
		return uuid.Nil, false
	}
	id, ok := userID.(uuid.UUID)
	return id, ok
}

// GetSessionID retrieves the session ID from the fiber context
func GetSessionID(c *fiber.Ctx) (uuid.UUID, bool) {
	sessionID := c.Locals(string(SessionIDKey))
	if sessionID == nil {
		return uuid.Nil, false
	}
	id, ok := sessionID.(uuid.UUID)
	return id, ok
}

// GetTenantID retrieves the tenant ID from the fiber context
func GetTenantID(c *fiber.Ctx) (uuid.UUID, bool) {
	tenantID := c.Locals(string(TenantIDKey))
	if tenantID == nil {
		return uuid.Nil, false
	}
	id, ok := tenantID.(uuid.UUID)
	return id, ok
}

// MustGetUserID retrieves the user ID from context or panics
func MustGetUserID(c *fiber.Ctx) uuid.UUID {
	userID, ok := GetUserID(c)
	if !ok {
		panic("user ID not found in context")
	}
	return userID
}

// GetUserIDFromContext retrieves user ID from standard context
func GetUserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	userID := ctx.Value(UserIDKey)
	if userID == nil {
		return uuid.Nil, false
	}
	id, ok := userID.(uuid.UUID)
	return id, ok
}
