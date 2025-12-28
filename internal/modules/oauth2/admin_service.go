package oauth2

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gofrs/uuid/v5"

	"system/internal/domain"
	"system/pkg/util/crypto"
	"system/pkg/util/stringutil"
)

// OAuth2AdminService chứa business logic cho OAuth2 client management.
type oauth2AdminServiceImpl struct {
	clientRepo domain.OAuth2ClientRepository
}

// NewOAuth2AdminService tạo instance mới của OAuth2AdminService.
func NewAdminService(clientRepo domain.OAuth2ClientRepository) OAuth2AdminService {
	return &oauth2AdminServiceImpl{
		clientRepo: clientRepo,
	}
}

// ValidateGrantTypes validates grant types.
// Returns the invalid grant type and error if validation fails.
func (s *oauth2AdminServiceImpl) ValidateGrantTypes(grantTypes []string) (string, error) {
	validGrantTypes := map[string]bool{
		"authorization_code": true,
		"refresh_token":      true,
		"client_credentials": true,
		"password":           true,
		"implicit":           true,
	}

	for _, gt := range grantTypes {
		if !validGrantTypes[gt] {
			return gt, errors.New("invalid grant type")
		}
	}
	return "", nil
}

// ValidateResponseTypes validates response types.
// Returns the invalid response type and error if validation fails.
func (s *oauth2AdminServiceImpl) ValidateResponseTypes(responseTypes []string) (string, error) {
	validResponseTypes := map[string]bool{
		"code":     true,
		"token":    true,
		"id_token": true,
	}

	for _, rt := range responseTypes {
		if !validResponseTypes[rt] {
			return rt, errors.New("invalid response type")
		}
	}
	return "", nil
}

// GenerateClientSecret generates and hashes a client secret.
// Returns plaintext secret, hashed secret, and error.
func (s *oauth2AdminServiceImpl) GenerateClientSecret() (plainSecret string, hashedSecret string, err error) {
	secret, err := stringutil.GenerateRandomString(32)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate secret: %w", err)
	}

	hash, err := crypto.HashPassword(secret)
	if err != nil {
		return "", "", fmt.Errorf("failed to hash secret: %w", err)
	}

	return secret, hash, nil
}

// CreateClientRequest contains the data needed to create an OAuth2 client.
type AdminCreateClientRequest struct {
	ClientName        string
	RedirectURIs      []string
	GrantTypes        []string
	ResponseTypes     []string
	Scopes            []string
	IsPublic          bool
	IsInternal        bool
	TokenEndpointAuth string
	OrganizationID    *uuid.UUID
	ClientURI         *string
	LogoURL           *string
}

// CreateClient creates a new OAuth2 client.
// Returns the created client and the plaintext secret (if not public).
func (s *oauth2AdminServiceImpl) CreateClient(ctx context.Context, req AdminCreateClientRequest) (*domain.OAuth2Client, *string, error) {
	// Generate client ID
	clientID := uuid.Must(uuid.NewV7())

	// Generate secret for confidential clients
	var clientSecret *string
	var secretHash string

	if !req.IsPublic {
		plain, hash, err := s.GenerateClientSecret()
		if err != nil {
			return nil, nil, fmt.Errorf("failed to generate client secret: %w", err)
		}
		clientSecret = &plain
		secretHash = hash
	}

	// Create client entity
	client := &domain.OAuth2Client{
		ID:                clientID,
		ClientName:        req.ClientName,
		SecretHash:        secretHash,
		RedirectURIs:      req.RedirectURIs,
		GrantTypes:        req.GrantTypes,
		ResponseTypes:     req.ResponseTypes,
		Scopes:            req.Scopes,
		IsPublic:          req.IsPublic,
		IsInternal:        req.IsInternal,
		TokenEndpointAuth: req.TokenEndpointAuth,
		OrganizationID:    req.OrganizationID,
		ClientURI:         req.ClientURI,
		LogoURL:           req.LogoURL,
		Active:            true,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	// Save to database
	if err := s.clientRepo.Create(ctx, client); err != nil {
		return nil, nil, fmt.Errorf("failed to create client: %w", err)
	}

	return client, clientSecret, nil
}

// GetClientByID retrieves a client by ID.
func (s *oauth2AdminServiceImpl) GetClientByID(ctx context.Context, clientID uuid.UUID) (*domain.OAuth2Client, error) {
	client, err := s.clientRepo.GetByID(ctx, clientID)
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}
	return client, nil
}

// ListClients retrieves a list of OAuth2 clients with filtering.
func (s *oauth2AdminServiceImpl) ListClients(ctx context.Context, organizationID *uuid.UUID, active *bool, limit, offset int) ([]*domain.OAuth2Client, int, error) {
	clients, total, err := s.clientRepo.List(ctx, organizationID, active, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list clients: %w", err)
	}

	return clients, total, nil
}

// UpdateClientRequest contains the data needed to update an OAuth2 client.
type AdminUpdateClientRequest struct {
	ClientName        *string
	RedirectURIs      *[]string
	GrantTypes        *[]string
	ResponseTypes     *[]string
	Scopes            *[]string
	IsPublic          *bool
	IsInternal        *bool
	TokenEndpointAuth *string
	ClientURI         *string
	LogoURL           *string
	Active            *bool
}

// UpdateClient updates an existing OAuth2 client.
func (s *oauth2AdminServiceImpl) UpdateClient(ctx context.Context, clientID uuid.UUID, req AdminUpdateClientRequest) (*domain.OAuth2Client, error) {
	// Get existing client
	client, err := s.clientRepo.GetByID(ctx, clientID)
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	// Update fields
	if req.ClientName != nil {
		client.ClientName = *req.ClientName
	}
	if req.RedirectURIs != nil {
		client.RedirectURIs = *req.RedirectURIs
	}
	if req.GrantTypes != nil {
		client.GrantTypes = *req.GrantTypes
	}
	if req.ResponseTypes != nil {
		client.ResponseTypes = *req.ResponseTypes
	}
	if req.Scopes != nil {
		client.Scopes = *req.Scopes
	}
	if req.IsPublic != nil {
		client.IsPublic = *req.IsPublic
	}
	if req.IsInternal != nil {
		client.IsInternal = *req.IsInternal
	}
	if req.TokenEndpointAuth != nil {
		client.TokenEndpointAuth = *req.TokenEndpointAuth
	}
	if req.ClientURI != nil {
		client.ClientURI = req.ClientURI
	}
	if req.LogoURL != nil {
		client.LogoURL = req.LogoURL
	}
	if req.Active != nil {
		client.Active = *req.Active
	}

	client.UpdatedAt = time.Now()

	// Save changes
	if err := s.clientRepo.Update(ctx, client); err != nil {
		return nil, fmt.Errorf("failed to update client: %w", err)
	}

	return client, nil
}

// DeleteClient deletes an OAuth2 client.
func (s *oauth2AdminServiceImpl) DeleteClient(ctx context.Context, clientID uuid.UUID) error {
	if err := s.clientRepo.Delete(ctx, clientID); err != nil {
		return fmt.Errorf("failed to delete client: %w", err)
	}
	return nil
}

// RegenerateClientSecret generates a new secret for an existing client.
// Returns the new plaintext secret.
func (s *oauth2AdminServiceImpl) RegenerateClientSecret(ctx context.Context, clientID uuid.UUID) (string, error) {
	// Get existing client
	client, err := s.clientRepo.GetByID(ctx, clientID)
	if err != nil {
		return "", fmt.Errorf("failed to get client: %w", err)
	}

	// Check if client is public
	if client.IsPublic {
		return "", errors.New("cannot regenerate secret for public client")
	}

	// Generate new secret
	plain, hash, err := s.GenerateClientSecret()
	if err != nil {
		return "", fmt.Errorf("failed to generate secret: %w", err)
	}

	// Update client
	client.SecretHash = hash
	client.UpdatedAt = time.Now()

	if err := s.clientRepo.Update(ctx, client); err != nil {
		return "", fmt.Errorf("failed to update client: %w", err)
	}

	return plain, nil
}
