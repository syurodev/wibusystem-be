package oauth2_admin

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

// CreateClientRequest là DTO cho việc tạo OAuth2 client mới
type CreateClientRequest struct {
	ClientName        string   `json:"client_name" binding:"required,min=3,max=255"`
	RedirectURIs      []string `json:"redirect_uris" binding:"required,min=1"`
	GrantTypes        []string `json:"grant_types" binding:"required,min=1"`
	ResponseTypes     []string `json:"response_types" binding:"required,min=1"`
	Scopes            []string `json:"scopes" binding:"required,min=1"`
	IsPublic          bool     `json:"is_public"`
	IsInternal        bool     `json:"is_internal"` // true = internal client (full access), false = external client (limited)
	TokenEndpointAuth string   `json:"token_endpoint_auth_method" binding:"required,oneof=client_secret_basic client_secret_post client_secret_jwt private_key_jwt none"`
	OrganizationID          *string  `json:"organization_id,omitempty"` // Optional, null = global client
	ClientURI         *string  `json:"client_uri,omitempty"`
	LogoURL           *string  `json:"logo_url,omitempty"`
}

// UpdateClientRequest là DTO cho việc update OAuth2 client
type UpdateClientRequest struct {
	ClientName        *string   `json:"client_name,omitempty" binding:"omitempty,min=3,max=255"`
	RedirectURIs      *[]string `json:"redirect_uris,omitempty" binding:"omitempty,min=1"`
	GrantTypes        *[]string `json:"grant_types,omitempty" binding:"omitempty,min=1"`
	ResponseTypes     *[]string `json:"response_types,omitempty" binding:"omitempty,min=1"`
	Scopes            *[]string `json:"scopes,omitempty" binding:"omitempty,min=1"`
	IsPublic          *bool     `json:"is_public,omitempty"`
	IsInternal        *bool     `json:"is_internal,omitempty"`
	TokenEndpointAuth *string   `json:"token_endpoint_auth_method,omitempty" binding:"omitempty,oneof=client_secret_basic client_secret_post client_secret_jwt private_key_jwt none"`
	ClientURI         *string   `json:"client_uri,omitempty"`
	LogoURL           *string   `json:"logo_url,omitempty"`
	Active            *bool     `json:"active,omitempty"`
}

// RegenerateSecretRequest là DTO cho việc regenerate client secret
type RegenerateSecretRequest struct {
	// Empty - just a POST request to trigger regeneration
}

// ClientResponse là DTO cho response
type ClientResponse struct {
	ID                uuid.UUID `json:"id"`
	ClientName        string    `json:"client_name"`
	ClientSecret      *string   `json:"client_secret,omitempty"` // Only returned on creation/regeneration
	RedirectURIs      []string  `json:"redirect_uris"`
	GrantTypes        []string  `json:"grant_types"`
	ResponseTypes     []string  `json:"response_types"`
	Scopes            []string  `json:"scopes"`
	IsPublic          bool      `json:"is_public"`
	IsInternal        bool      `json:"is_internal"`
	TokenEndpointAuth string    `json:"token_endpoint_auth_method"`
	OrganizationID          *string   `json:"organization_id,omitempty"`
	ClientURI         *string   `json:"client_uri,omitempty"`
	LogoURL           *string   `json:"logo_url,omitempty"`
	Active            bool      `json:"active"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// ClientListResponse là DTO cho list response
type ClientListResponse struct {
	Clients []ClientResponse `json:"clients"`
	Total   int              `json:"total"`
	Limit   int              `json:"limit"`
	Offset  int              `json:"offset"`
}
