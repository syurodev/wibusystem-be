package auth

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// UserContext represents authenticated user information in the request context
type UserContext struct {
	// User ID extracted from token
	UserID uuid.UUID `json:"user_id"`

	// Username of the authenticated user
	Username string `json:"username"`

	// Email of the authenticated user
	Email string `json:"email"`

	// Name/display name of the user
	Name string `json:"name"`

	// EmailVerified indicates if the user's email is verified
	EmailVerified bool `json:"email_verified"`

	// Scopes granted to this token
	Scopes []string `json:"scopes"`

	// ClientID that issued the token
	ClientID string `json:"client_id"`

	// TokenType (usually "access_token")
	TokenType string `json:"token_type"`

	// ExpiresAt is when the token expires
	ExpiresAt time.Time `json:"expires_at"`

	// IssuedAt is when the token was issued
	IssuedAt time.Time `json:"issued_at"`

	// Extra claims from the token
	Extra map[string]string `json:"extra,omitempty"`
}

// HasScope checks if the user has a specific scope
func (u *UserContext) HasScope(scope string) bool {
	for _, s := range u.Scopes {
		if s == scope {
			return true
		}
	}
	return false
}

// HasAnyScope checks if the user has at least one of the specified scopes
func (u *UserContext) HasAnyScope(scopes ...string) bool {
	userScopes := make(map[string]bool)
	for _, scope := range u.Scopes {
		userScopes[scope] = true
	}

	for _, scope := range scopes {
		if userScopes[scope] {
			return true
		}
	}
	return false
}

// HasAllScopes checks if the user has all of the specified scopes
func (u *UserContext) HasAllScopes(scopes ...string) bool {
	userScopes := make(map[string]bool)
	for _, scope := range u.Scopes {
		userScopes[scope] = true
	}

	for _, scope := range scopes {
		if !userScopes[scope] {
			return false
		}
	}
	return true
}

// IsAdmin checks if the user has admin privileges
func (u *UserContext) IsAdmin() bool {
	return u.HasScope("admin")
}

// ValidationResult represents the result of token validation
type ValidationResult struct {
	Valid       bool         `json:"valid"`
	Error       string       `json:"error,omitempty"`
	UserContext *UserContext `json:"user_context,omitempty"`
}

// ContextKeys for storing authentication data in Gin context
const (
	// UserContextKey is the key for storing UserContext in Gin context
	UserContextKey = "auth_user"

	// TokenKey is the key for storing the raw token in Gin context
	TokenKey = "auth_token"

	// AuthenticatedKey is the key for storing authentication status
	AuthenticatedKey = "auth_authenticated"
)

// GetUserFromContext retrieves the authenticated user from Gin context
func GetUserFromContext(c *gin.Context) (*UserContext, bool) {
	value, exists := c.Get(UserContextKey)
	if !exists {
		return nil, false
	}

	user, ok := value.(*UserContext)
	return user, ok
}

// GetTokenFromContext retrieves the raw token from Gin context
func GetTokenFromContext(c *gin.Context) (string, bool) {
	value, exists := c.Get(TokenKey)
	if !exists {
		return "", false
	}

	token, ok := value.(string)
	return token, ok
}

// IsAuthenticated checks if the request is authenticated
func IsAuthenticated(c *gin.Context) bool {
	value, exists := c.Get(AuthenticatedKey)
	if !exists {
		return false
	}

	authenticated, ok := value.(bool)
	return ok && authenticated
}