// Package oauth2 integrates github.com/ory/fosite to provide OAuth2/OIDC
// functionality (authorize, token, introspect, revoke) and related helpers.
package oauth2

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ory/fosite"
	"github.com/ory/fosite/compose"
	"github.com/ory/fosite/handler/openid"
	_ "github.com/ory/fosite/storage"
	"github.com/ory/fosite/token/jwt"

	"wibusystem/services/identify/config"
)

// Provider wraps fosite configured with storage, signing strategies, and
// helpers for OIDC session creation and JWKS exposure.
type Provider struct {
	OAuth2Provider fosite.OAuth2Provider
	Store          *Store
	Config         *fosite.Config
	JWTSigner      interface{}
	PrivateKey     *rsa.PrivateKey
	Kid            string
}

// NewProvider constructs a fosite OAuth2 provider backed by Postgres store
// and configured lifespans. It also generates or loads an RSA private key used
// to expose a JWKS for OIDC consumers.
//
// Note: Access/refresh tokens use the configured fosite strategy (HMAC here).
// JWKS is provided primarily for OIDC ID Tokens or future JWT usage.
func NewProvider(cfg *config.Config, pool *pgxpool.Pool) (*Provider, error) {
	// Generate or load RSA key pair
	privateKey, err := generateOrLoadRSAKey(cfg.OAuth2.JWTSigningKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create RSA key: %w", err)
	}

	// Create fosite config
	fositeConfig := &fosite.Config{
		AccessTokenLifespan:   cfg.OAuth2.AccessTokenLifespan,
		RefreshTokenLifespan:  cfg.OAuth2.RefreshTokenLifespan,
		AuthorizeCodeLifespan: cfg.OAuth2.AuthorizeCodeLifespan,
		IDTokenLifespan:       cfg.OAuth2.IDTokenLifespan,

		// Scopes
		ScopeStrategy:            fosite.HierarchicScopeStrategy,
		AudienceMatchingStrategy: fosite.DefaultAudienceMatchingStrategy,

		// Security
		DisableRefreshTokenValidation: false,
		SendDebugMessagesToClients:    cfg.Server.Environment != "production",

		// PKCE
		EnforcePKCE:                    true,
		EnforcePKCEForPublicClients:    true,
		EnablePKCEPlainChallengeMethod: false,

		// Token format
		AccessTokenIssuer: cfg.OAuth2.Issuer,

		// OIDC
		IDTokenIssuer: cfg.OAuth2.Issuer,
	}

	// Set HMAC global secret for access/refresh token signing.
	// Must be at least 32 bytes for SHA512/256 HMAC strategy.
	// Use Security.JWTSecret for dev; in production, provide a strong secret.
	if cfg.Security.JWTSecret != "" {
		fositeConfig.GlobalSecret = []byte(cfg.Security.JWTSecret)
	}

	// Build common strategies: HMAC for OAuth2 tokens, RSA for OIDC ID tokens.
	// OpenIDConnect strategy requires a key getter function and configurator.
	oidcStrategy := compose.NewOpenIDConnectStrategy(
		func(_ context.Context) (interface{}, error) { return privateKey, nil },
		fositeConfig,
	)
	strategy := compose.CommonStrategy{
		CoreStrategy:               compose.NewOAuth2HMACStrategy(fositeConfig),
		OpenIDConnectTokenStrategy: oidcStrategy,
	}

	// Create storage (requires access token lifespan)
	store := NewStore(pool, privateKey, fositeConfig.AccessTokenLifespan)

	// Compose OAuth2 provider with required handlers
	oauth2Provider := compose.Compose(
		fositeConfig,
		store,
		strategy,

		// OAuth2 core handlers (no password grant in production; TODO consider enabling for dev/testing only)
		compose.OAuth2AuthorizeExplicitFactory,
		compose.OAuth2AuthorizeImplicitFactory,
		compose.OAuth2ClientCredentialsGrantFactory,
		compose.OAuth2RefreshTokenGrantFactory,

		// PKCE
		compose.OAuth2PKCEFactory,

		// OIDC handlers
		compose.OpenIDConnectExplicitFactory,
		compose.OpenIDConnectImplicitFactory,
		compose.OpenIDConnectHybridFactory,
		compose.OpenIDConnectRefreshFactory,

		// Token introspection and revocation
		compose.OAuth2TokenIntrospectionFactory,
		compose.OAuth2TokenRevocationFactory,
	)

	// Precompute KID for JWKS and ID Token headers
	kid, err := computeKID(&privateKey.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to compute KID: %w", err)
	}

	return &Provider{
		OAuth2Provider: oauth2Provider,
		Store:          store,
		Config:         fositeConfig,
		JWTSigner:      strategy.CoreStrategy,
		PrivateKey:     privateKey,
		Kid:            kid,
	}, nil
}

// generateOrLoadRSAKey generates a new RSA key pair or loads from a PEM/PKCS8
// key string. Used for OIDC JWKS exposure.
func generateOrLoadRSAKey(keyString string) (*rsa.PrivateKey, error) {
	// Try to parse as PEM first
	if block, _ := pem.Decode([]byte(keyString)); block != nil {
		privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err == nil {
			return privateKey, nil
		}

		// Try PKCS8 format
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err == nil {
			if rsaKey, ok := key.(*rsa.PrivateKey); ok {
				return rsaKey, nil
			}
		}
	}

	// Generate new key if parsing failed or no key provided
	return generateRSAKey()
}

// generateRSAKey creates a 2048-bit RSA key for development.
// Consider managing keys externally in production.
func generateRSAKey() (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, 2048)
}

// GetJWKS returns the JSON Web Key Set for this provider. The key ID (kid)
// is derived from a hash of the public key.
func (p *Provider) GetJWKS() (map[string]interface{}, error) {
	publicKey := &p.PrivateKey.PublicKey

	nB64, eB64, err := rsaKeyToJWKComponents(publicKey)
	if err != nil {
		return nil, err
	}

	jwk := map[string]interface{}{
		"kty": "RSA",
		"use": "sig",
		"alg": "RS256",
		"kid": p.Kid,
		"n":   nB64,
		"e":   eB64,
	}

	jwks := map[string]interface{}{
		"keys": []interface{}{jwk},
	}

	return jwks, nil
}

// ValidateToken validates an access token and returns the session info.
// TODO: Implement JWT validation if switching to JWT access tokens.
func (p *Provider) ValidateToken(_ string) (*openid.DefaultSession, error) {
	// This would validate the JWT token and return session info
	// Implementation would depend on your specific token validation needs
	return nil, fmt.Errorf("token validation not implemented")
}

// CreateCustomSession creates a session embedding basic user info as OIDC
// claims; add/adjust claims to fit your profile needs.
func (p *Provider) CreateCustomSession(userID string, username string, email string) *openid.DefaultSession {
	session := &openid.DefaultSession{
		Claims: &jwt.IDTokenClaims{
			Issuer:      p.Config.IDTokenIssuer,
			Subject:     userID,
			Audience:    []string{},
			ExpiresAt:   time.Now().Add(p.Config.IDTokenLifespan),
			IssuedAt:    time.Now(),
			RequestedAt: time.Now(),
			AuthTime:    time.Now(),
		},
		Headers: &jwt.Headers{
			Extra: map[string]interface{}{
				"kid": p.Kid,
			},
		},
		Subject:  userID,
		Username: username,
	}

	// Add custom claims
	session.Claims.Extra = map[string]interface{}{
		"email":              email,
		"email_verified":     true,
		"preferred_username": username,
		"name":               username,
	}

	return session
}

// computeKID creates a deterministic KID from the RSA public key
func computeKID(pub *rsa.PublicKey) (string, error) {
	keyBytes, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(keyBytes)
	return fmt.Sprintf("%x", hash[:8]), nil
}

// rsaKeyToJWKComponents encodes RSA public key to base64url n and e
func rsaKeyToJWKComponents(pub *rsa.PublicKey) (string, string, error) {
	// Modulus N
	nBytes := pub.N.Bytes()
	nB64 := base64.RawURLEncoding.EncodeToString(nBytes)

	// Exponent E
	// Convert int to big-endian bytes with minimal length
	eBytes := intToBytes(pub.E)
	eB64 := base64.RawURLEncoding.EncodeToString(eBytes)

	return nB64, eB64, nil
}

func intToBytes(i int) []byte {
	if i == 0 {
		return []byte{0}
	}
	b := []byte{}
	for i > 0 {
		b = append([]byte{byte(i & 0xff)}, b...)
		i >>= 8
	}
	return b
}

// GetUserInfo returns user information for the /userinfo endpoint.
// TODO: Extract from validated token/session rather than return a placeholder.
func (p *Provider) GetUserInfo(_ string) (map[string]interface{}, error) {
	// This would extract user info from the token
	// For now, return a placeholder
	return map[string]interface{}{
		"sub":   "user123",
		"email": "user@example.com",
		"name":  "Example User",
	}, nil
}
