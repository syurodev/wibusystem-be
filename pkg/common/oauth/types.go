// Package oauth provides simple types for OAuth2/OIDC token validation
// that can be shared across services without any external dependencies.
package oauth

import "time"

// UserInfo represents authenticated user information from token validation.
type UserInfo struct {
	Subject   string            `json:"sub"`
	Username  string            `json:"username,omitempty"`
	Email     string            `json:"email,omitempty"`
	Name      string            `json:"name,omitempty"`
	Verified  bool              `json:"email_verified,omitempty"`
	TenantID  string            `json:"tenant_id,omitempty"`  // Current tenant ID for the user
	Extra     map[string]string `json:"extra,omitempty"`
	UpdatedAt *time.Time        `json:"updated_at,omitempty"`
}

// TokenInfo contains metadata about the validated token.
type TokenInfo struct {
	Active    bool      `json:"active"`
	TokenType string    `json:"token_type,omitempty"`
	Scope     []string  `json:"scope,omitempty"`
	ClientID  string    `json:"client_id,omitempty"`
	Audience  []string  `json:"aud,omitempty"`
	Issuer    string    `json:"iss,omitempty"`
	Subject   string    `json:"sub,omitempty"`
	ExpiresAt time.Time `json:"exp,omitempty"`
	IssuedAt  time.Time `json:"iat,omitempty"`
}

// ValidationResult contains the complete result of token validation.
type ValidationResult struct {
	Valid     bool       `json:"valid"`
	TokenInfo *TokenInfo `json:"token_info,omitempty"`
	UserInfo  *UserInfo  `json:"user_info,omitempty"`
	Error     string     `json:"error,omitempty"`
}

// ValidationRequest represents a token validation request.
type ValidationRequest struct {
	Token     string   `json:"token"`
	TokenType string   `json:"token_type,omitempty"` // defaults to "access_token"
	Scopes    []string `json:"required_scopes,omitempty"`
}

// HasScope checks if the token has all required scopes.
func (v *ValidationResult) HasScope(requiredScopes ...string) bool {
	if !v.Valid || v.TokenInfo == nil {
		return false
	}

	grantedScopes := make(map[string]bool)
	for _, scope := range v.TokenInfo.Scope {
		grantedScopes[scope] = true
	}

	for _, required := range requiredScopes {
		if !grantedScopes[required] {
			return false
		}
	}

	return true
}

// HasAnyScope checks if the token has at least one of the required scopes.
func (v *ValidationResult) HasAnyScope(requiredScopes ...string) bool {
	if !v.Valid || v.TokenInfo == nil {
		return false
	}

	grantedScopes := make(map[string]bool)
	for _, scope := range v.TokenInfo.Scope {
		grantedScopes[scope] = true
	}

	for _, required := range requiredScopes {
		if grantedScopes[required] {
			return true
		}
	}

	return false
}

// IsAdmin checks if the token has admin privileges.
func (v *ValidationResult) IsAdmin() bool {
	return v.HasScope("admin")
}