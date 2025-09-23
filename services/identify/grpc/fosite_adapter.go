// Package grpc provides gRPC server integration for the Identity Service
package grpc

import (
	"context"

	"github.com/ory/fosite"
	"github.com/ory/fosite/handler/openid"

	"wibusystem/pkg/common/oauth"
	"wibusystem/services/identify/oauth2"
)

// FositeTokenValidator implements TokenValidator interface using fosite OAuth2Provider
type FositeTokenValidator struct {
	provider *oauth2.Provider
}

// NewFositeTokenValidator creates a new fosite-based token validator
func NewFositeTokenValidator(provider *oauth2.Provider) *FositeTokenValidator {
	return &FositeTokenValidator{
		provider: provider,
	}
}

// ValidateToken validates a token using fosite OAuth2Provider
func (v *FositeTokenValidator) ValidateToken(ctx context.Context, req *oauth.ValidationRequest) (*oauth.ValidationResult, error) {
	// Set default token type
	tokenType := req.TokenType
	if tokenType == "" {
		tokenType = "access_token"
	}

	// Use fosite to introspect the token
	_, ar, err := v.provider.OAuth2Provider.IntrospectToken(
		ctx,
		req.Token,
		fosite.AccessToken,
		&openid.DefaultSession{},
	)

	if err != nil {
		// Token is invalid
		return &oauth.ValidationResult{
			Valid: false,
			Error: "invalid or expired token",
		}, nil
	}

	// Token is valid, extract information
	result := &oauth.ValidationResult{
		Valid: true,
	}

	// Extract token information
	result.TokenInfo = &oauth.TokenInfo{
		Active:    true,
		TokenType: "Bearer",
		Scope:     argumentsToStringSlice(ar.GetGrantedScopes()),
		ClientID:  ar.GetClient().GetID(),
		Audience:  argumentsToStringSlice(ar.GetGrantedAudience()),
		Subject:   ar.GetSession().GetSubject(),
	}

	// Extract user information from session
	if session, ok := ar.GetSession().(*openid.DefaultSession); ok {
		result.UserInfo = &oauth.UserInfo{
			Subject:  session.GetSubject(),
			Username: session.GetUsername(),
		}

		// Extract additional claims if available
		if session.Claims != nil && session.Claims.Extra != nil {
			if email, ok := session.Claims.Extra["email"].(string); ok {
				result.UserInfo.Email = email
			}
			if name, ok := session.Claims.Extra["name"].(string); ok {
				result.UserInfo.Name = name
			}
			if emailVerified, ok := session.Claims.Extra["email_verified"].(bool); ok {
				result.UserInfo.Verified = emailVerified
			}
			if preferredUsername, ok := session.Claims.Extra["preferred_username"].(string); ok && result.UserInfo.Username == "" {
				result.UserInfo.Username = preferredUsername
			}
		}

		// Set issued at and expires at times
		if session.Claims != nil {
			if !session.Claims.IssuedAt.IsZero() {
				result.TokenInfo.IssuedAt = session.Claims.IssuedAt
			}
			if !session.Claims.ExpiresAt.IsZero() {
				result.TokenInfo.ExpiresAt = session.Claims.ExpiresAt
			}
		}

		// Set issuer from session or config
		if session.Claims != nil && session.Claims.Issuer != "" {
			result.TokenInfo.Issuer = session.Claims.Issuer
		} else {
			result.TokenInfo.Issuer = v.provider.Config.AccessTokenIssuer
		}
	}

	// Check required scopes if specified
	if len(req.Scopes) > 0 {
		grantedScopes := make(map[string]bool)
		for _, scope := range result.TokenInfo.Scope {
			grantedScopes[scope] = true
		}

		for _, requiredScope := range req.Scopes {
			if !grantedScopes[requiredScope] {
				result.Valid = false
				result.Error = "insufficient scope: missing " + requiredScope
				break
			}
		}
	}

	return result, nil
}

// argumentsToStringSlice converts fosite.Arguments to []string
// argumentsToStringSlice converts fosite.Arguments to []string
func argumentsToStringSlice(args fosite.Arguments) []string {
	result := make([]string, len(args))
	copy(result, args)
	return result
}
