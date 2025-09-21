// Package middleware contains Gin middleware for authentication, tenant
// resolution, rate limiting, and common API concerns.
package middleware

import (
	_ "context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ory/fosite"
	"github.com/ory/fosite/handler/openid"

	r "wibusystem/pkg/common/response"
	"wibusystem/services/identify/oauth2"
	"wibusystem/services/identify/repositories"
)

// AuthMiddleware validates OAuth2 access tokens (via fosite) and
// enriches Gin context with authenticated user information and scopes.
type AuthMiddleware struct {
	provider *oauth2.Provider
	repos    *repositories.Repositories
}

// NewAuthMiddleware creates new authentication middleware.
func NewAuthMiddleware(provider *oauth2.Provider, repos *repositories.Repositories) *AuthMiddleware {
	return &AuthMiddleware{
		provider: provider,
		repos:    repos,
	}
}

// RequireAuth returns middleware that requires a valid OAuth2 Access Token.
// On success, user info, token, and scopes are injected into Gin context.
func (am *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c.Request)
		if token == "" {
			c.JSON(http.StatusUnauthorized, r.StandardResponse{
				Success: false,
				Message: "Unauthorized",
				Data:    nil,
				Error:   &r.ErrorDetail{Code: "unauthorized", Description: "Access token required"},
				Meta:    map[string]interface{}{},
			})
			c.Abort()
			return
		}

		// Validate token with Fosite
		tokenType, ar, err := am.provider.OAuth2Provider.IntrospectToken(
			c.Request.Context(),
			token,
			fosite.AccessToken,
			&openid.DefaultSession{},
		)
		if err != nil || tokenType != fosite.AccessToken {
			c.JSON(http.StatusUnauthorized, r.StandardResponse{
				Success: false,
				Message: "Invalid token",
				Data:    nil,
				Error:   &r.ErrorDetail{Code: "invalid_token", Description: "Invalid or expired access token"},
				Meta:    map[string]interface{}{},
			})
			c.Abort()
			return
		}

		// Extract user information from the session returned by fosite.
		session := ar.GetSession().(*openid.DefaultSession)
		userInfo := &UserInfo{
			Subject:  session.GetSubject(),
			Username: session.GetUsername(),
			Scopes:   ar.GetGrantedScopes(),
			ClientID: ar.GetClient().GetID(),
		}

		// Store user info in context
		c.Set("user", userInfo)
		c.Set("token", token)
		c.Set("scopes", ar.GetGrantedScopes())

		c.Next()
	}
}

// RequireScope enforces that the current token has all of the specified scopes.
func (am *AuthMiddleware) RequireScope(requiredScopes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		scopes, exists := c.Get("scopes")
		if !exists {
			c.JSON(http.StatusForbidden, r.StandardResponse{
				Success: false,
				Message: "Insufficient scope",
				Data:    nil,
				Error:   &r.ErrorDetail{Code: "insufficient_scope", Description: "Token does not have required scopes"},
				Meta:    map[string]interface{}{},
			})
			c.Abort()
			return
		}

		grantedScopes, ok := scopes.(fosite.Arguments)
		if !ok {
			c.JSON(http.StatusForbidden, r.StandardResponse{
				Success: false,
				Message: "Insufficient scope",
				Data:    nil,
				Error:   &r.ErrorDetail{Code: "insufficient_scope", Description: "Invalid scope information"},
				Meta:    map[string]interface{}{},
			})
			c.Abort()
			return
		}

		// Check if all required scopes are present.
		for _, required := range requiredScopes {
			if !grantedScopes.Has(required) {
				c.JSON(http.StatusForbidden, r.StandardResponse{
					Success: false,
					Message: "Insufficient scope",
					Data:    nil,
					Error:   &r.ErrorDetail{Code: "insufficient_scope", Description: "Token missing required scope: " + required},
					Meta:    map[string]interface{}{},
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// OptionalAuth extracts user info if a token is present, but does not require
// authentication. It continues as anonymous when no/invalid token is provided.
func (am *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c.Request)
		if token == "" {
			c.Next()
			return
		}

		// Try to validate token; ignore failures to allow anonymous access.
		tokenType, ar, err := am.provider.OAuth2Provider.IntrospectToken(
			c.Request.Context(),
			token,
			fosite.AccessToken,
			&openid.DefaultSession{},
		)
		if err == nil && tokenType == fosite.AccessToken {
			// Extract user information from session
			session := ar.GetSession().(*openid.DefaultSession)
			userInfo := &UserInfo{
				Subject:  session.GetSubject(),
				Username: session.GetUsername(),
				Scopes:   ar.GetGrantedScopes(),
				ClientID: ar.GetClient().GetID(),
			}

			// Store user info in context
			c.Set("user", userInfo)
			c.Set("token", token)
			c.Set("scopes", ar.GetGrantedScopes())
		}

		c.Next()
	}
}

// RequireAdmin enforces that the current authenticated principal has admin
// privileges (currently checked via the "admin" scope).
func (am *AuthMiddleware) RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		userInfo, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, r.StandardResponse{
				Success: false,
				Message: "Unauthorized",
				Data:    nil,
				Error:   &r.ErrorDetail{Code: "unauthorized", Description: "Authentication required"},
				Meta:    map[string]interface{}{},
			})
			c.Abort()
			return
		}

		user, ok := userInfo.(*UserInfo)
		if !ok {
			c.JSON(http.StatusInternalServerError, r.StandardResponse{
				Success: false,
				Message: "Server error",
				Data:    nil,
				Error:   &r.ErrorDetail{Code: "server_error", Description: "Invalid user information"},
				Meta:    map[string]interface{}{},
			})
			c.Abort()
			return
		}

		// Check if user has admin scope or is system admin
		if !user.Scopes.Has("admin") {
			c.JSON(http.StatusForbidden, r.StandardResponse{
				Success: false,
				Message: "Insufficient privileges",
				Data:    nil,
				Error:   &r.ErrorDetail{Code: "insufficient_privileges", Description: "Admin privileges required"},
				Meta:    map[string]interface{}{},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// UserInfo represents authenticated user information derived from an access
// token/session. Subject typically contains the user ID string.
type UserInfo struct {
	Subject  string           `json:"sub"`
	Username string           `json:"username"`
	Scopes   fosite.Arguments `json:"scopes"`
	ClientID string           `json:"client_id"`
}

// extractToken returns the bearer token from the Authorization header.
// Note: Falling back to the "access_token" query param is for debugging only.
func extractToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		// Try to get token from query parameter (for debugging only)
		return r.URL.Query().Get("access_token")
	}

	// Check for Bearer token
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}

	return ""
}

// GetUserFromContext extracts previously stored user information from Gin context.
func GetUserFromContext(c *gin.Context) (*UserInfo, bool) {
	userInfo, exists := c.Get("user")
	if !exists {
		return nil, false
	}

	user, ok := userInfo.(*UserInfo)
	return user, ok
}

// GetUserIDFromContext extracts user ID from context and converts to UUID.
//
// TODO: This currently returns a placeholder and should parse and validate
// user.Subject to a proper UUID user ID.
func GetUserIDFromContext(c *gin.Context) (uuid.UUID, error) {
	_, exists := GetUserFromContext(c)
	if !exists {
		return uuid.Nil, nil
	}

	// Parse subject as user ID (in production, you might want to validate this)
	// For now, assuming subject is the user ID as string
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001") // This should be parsed from user.Subject
	return userID, nil
}
