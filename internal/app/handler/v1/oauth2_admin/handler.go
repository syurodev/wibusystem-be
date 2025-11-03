package oauth2_admin

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid/v5"

	"system/internal/domain"
	"system/pkg/util/crypto"
	"system/pkg/util/random"
	"system/pkg/util/response"
)

type Handler struct {
	clientRepo domain.OAuth2ClientRepository
}

func NewHandler(clientRepo domain.OAuth2ClientRepository) *Handler {
	return &Handler{
		clientRepo: clientRepo,
	}
}

// CreateClient godoc
// @Summary Create a new OAuth2 client
// @Description Create a new OAuth2 client with the provided configuration
// @Tags OAuth2 Admin
// @Accept json
// @Produce json
// @Param request body CreateClientRequest true "Client creation request"
// @Success 201 {object} ClientResponse
// @Failure 400 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /admin/oauth2/clients [post]
func (h *Handler) CreateClient(c *gin.Context) {
	var req CreateClientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_FAILED", "validation.failed", err.Error())
		return
	}

	// Validate grant types, response types, and scopes
	if invalidType, err := validateGrantTypes(req.GrantTypes); err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_GRANT_TYPE", "validation.invalid_grant_type", map[string]any{"type": invalidType})
		return
	}

	if invalidType, err := validateResponseTypes(req.ResponseTypes); err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_RESPONSE_TYPE", "validation.invalid_response_type", map[string]any{"type": invalidType})
		return
	}

	// Generate client ID and secret
	clientID := uuid.Must(uuid.NewV7())
	var clientSecret string
	var secretHash string

	if !req.IsPublic {
		// Generate secure random secret
		secret, err := random.GenerateRandomString(32)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, "SECRET_GENERATION_FAILED", "client.secret_generation_failed", nil)
			return
		}
		clientSecret = secret

		// Hash the secret
		hash, err := crypto.HashPassword(clientSecret)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, "SECRET_HASH_FAILED", "client.secret_hash_failed", nil)
			return
		}
		secretHash = hash
	}

	// Parse tenant ID if provided
	var tenantID *uuid.UUID
	if req.TenantID != nil {
		tid, err := uuid.FromString(*req.TenantID)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "INVALID_TENANT_ID", "validation.invalid_tenant_id", nil)
			return
		}
		tenantID = &tid
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
		TenantID:          tenantID,
		ClientURI:         req.ClientURI,
		LogoURL:           req.LogoURL,
		Active:            true,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	// Save to database
	if err := h.clientRepo.Create(c.Request.Context(), client); err != nil {
		response.Error(c, http.StatusInternalServerError, "CREATE_FAILED", "client.create_failed", nil)
		return
	}

	// Build response
	clientResponse := h.buildClientResponse(client, &clientSecret)
	response.Success(c, http.StatusCreated, "client.created", clientResponse, nil)
}

// GetClient godoc
// @Summary Get an OAuth2 client by ID
// @Description Retrieve details of a specific OAuth2 client
// @Tags OAuth2 Admin
// @Produce json
// @Param id path string true "Client ID"
// @Success 200 {object} ClientResponse
// @Failure 404 {object} map[string]any
// @Router /admin/oauth2/clients/{id} [get]
func (h *Handler) GetClient(c *gin.Context) {
	clientID, err := uuid.FromString(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_CLIENT_ID", "validation.invalid_client_id", nil)
		return
	}

	client, err := h.clientRepo.GetByID(c.Request.Context(), clientID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "CLIENT_NOT_FOUND", "client.not_found", nil)
		return
	}

	clientResponse := h.buildClientResponse(client, nil)
	response.Success(c, http.StatusOK, "client.retrieved", clientResponse, nil)
}

// ListClients godoc
// @Summary List OAuth2 clients
// @Description Retrieve a list of OAuth2 clients with optional filtering
// @Tags OAuth2 Admin
// @Produce json
// @Param tenant_id query string false "Filter by tenant ID"
// @Param active query boolean false "Filter by active status"
// @Param limit query int false "Limit number of results" default(50)
// @Param offset query int false "Offset for pagination" default(0)
// @Success 200 {object} ClientListResponse
// @Failure 500 {object} map[string]any
// @Router /admin/oauth2/clients [get]
func (h *Handler) ListClients(c *gin.Context) {
	var tenantID *uuid.UUID
	if tenantIDStr := c.Query("tenant_id"); tenantIDStr != "" {
		tid, err := uuid.FromString(tenantIDStr)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "INVALID_TENANT_ID", "validation.invalid_tenant_id", nil)
			return
		}
		tenantID = &tid
	}

	var active *bool
	if activeStr := c.Query("active"); activeStr != "" {
		if activeStr == "true" {
			t := true
			active = &t
		} else if activeStr == "false" {
			f := false
			active = &f
		}
	}

	limit := 50
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	offset := 0
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	clients, total, err := h.clientRepo.List(c.Request.Context(), tenantID, active, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "LIST_FAILED", "client.list_failed", nil)
		return
	}

	responses := make([]ClientResponse, len(clients))
	for i, client := range clients {
		responses[i] = h.buildClientResponse(client, nil)
	}

	listResponse := ClientListResponse{
		Clients: responses,
		Total:   total,
		Limit:   limit,
		Offset:  offset,
	}
	response.Success(c, http.StatusOK, "client.listed", listResponse, nil)
}

// UpdateClient godoc
// @Summary Update an OAuth2 client
// @Description Update an existing OAuth2 client's configuration
// @Tags OAuth2 Admin
// @Accept json
// @Produce json
// @Param id path string true "Client ID"
// @Param request body UpdateClientRequest true "Client update request"
// @Success 200 {object} ClientResponse
// @Failure 400 {object} map[string]any
// @Failure 404 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /admin/oauth2/clients/{id} [put]
func (h *Handler) UpdateClient(c *gin.Context) {
	clientID, err := uuid.FromString(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_CLIENT_ID", "validation.invalid_client_id", nil)
		return
	}

	var req UpdateClientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_FAILED", "validation.failed", err.Error())
		return
	}

	// Get existing client
	client, err := h.clientRepo.GetByID(c.Request.Context(), clientID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "CLIENT_NOT_FOUND", "client.not_found", nil)
		return
	}

	// Update fields if provided
	if req.ClientName != nil {
		client.ClientName = *req.ClientName
	}
	if req.RedirectURIs != nil {
		client.RedirectURIs = *req.RedirectURIs
	}
	if req.GrantTypes != nil {
		if invalidType, err := validateGrantTypes(*req.GrantTypes); err != nil {
			response.Error(c, http.StatusBadRequest, "INVALID_GRANT_TYPE", "validation.invalid_grant_type", map[string]any{"type": invalidType})
			return
		}
		client.GrantTypes = *req.GrantTypes
	}
	if req.ResponseTypes != nil {
		if invalidType, err := validateResponseTypes(*req.ResponseTypes); err != nil {
			response.Error(c, http.StatusBadRequest, "INVALID_RESPONSE_TYPE", "validation.invalid_response_type", map[string]any{"type": invalidType})
			return
		}
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

	// Save updates
	if err := h.clientRepo.Update(c.Request.Context(), client); err != nil {
		response.Error(c, http.StatusInternalServerError, "UPDATE_FAILED", "client.update_failed", nil)
		return
	}

	clientResponse := h.buildClientResponse(client, nil)
	response.Success(c, http.StatusOK, "client.updated", clientResponse, nil)
}

// DeleteClient godoc
// @Summary Delete an OAuth2 client
// @Description Permanently delete an OAuth2 client
// @Tags OAuth2 Admin
// @Param id path string true "Client ID"
// @Success 204
// @Failure 404 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /admin/oauth2/clients/{id} [delete]
func (h *Handler) DeleteClient(c *gin.Context) {
	clientID, err := uuid.FromString(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_CLIENT_ID", "validation.invalid_client_id", nil)
		return
	}

	if err := h.clientRepo.Delete(c.Request.Context(), clientID); err != nil {
		response.Error(c, http.StatusInternalServerError, "DELETE_FAILED", "client.delete_failed", nil)
		return
	}

	response.Success(c, http.StatusOK, "client.deleted", nil, nil)
}

// RegenerateSecret godoc
// @Summary Regenerate client secret
// @Description Generate a new secret for an OAuth2 client
// @Tags OAuth2 Admin
// @Accept json
// @Produce json
// @Param id path string true "Client ID"
// @Success 200 {object} ClientResponse
// @Failure 400 {object} map[string]any
// @Failure 404 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /admin/oauth2/clients/{id}/regenerate-secret [post]
func (h *Handler) RegenerateSecret(c *gin.Context) {
	clientID, err := uuid.FromString(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_CLIENT_ID", "validation.invalid_client_id", nil)
		return
	}

	// Get existing client
	client, err := h.clientRepo.GetByID(c.Request.Context(), clientID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "CLIENT_NOT_FOUND", "client.not_found", nil)
		return
	}

	// Cannot regenerate secret for public clients
	if client.IsPublic {
		response.Error(c, http.StatusBadRequest, "PUBLIC_NO_SECRET", "client.public_no_secret", nil)
		return
	}

	// Generate new secret
	newSecret, err := random.GenerateRandomString(32)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "SECRET_GENERATION_FAILED", "client.secret_generation_failed", nil)
		return
	}

	// Hash the new secret
	secretHash, err := crypto.HashPassword(newSecret)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "SECRET_HASH_FAILED", "client.secret_hash_failed", nil)
		return
	}

	client.SecretHash = secretHash
	client.UpdatedAt = time.Now()

	// Save updates
	if err := h.clientRepo.Update(c.Request.Context(), client); err != nil {
		response.Error(c, http.StatusInternalServerError, "UPDATE_FAILED", "client.update_failed", nil)
		return
	}

	clientResponse := h.buildClientResponse(client, &newSecret)
	response.Success(c, http.StatusOK, "client.secret_regenerated", clientResponse, nil)
}

// Helper methods

func (h *Handler) buildClientResponse(client *domain.OAuth2Client, clientSecret *string) ClientResponse {
	var tenantID *string
	if client.TenantID != nil {
		tid := client.TenantID.String()
		tenantID = &tid
	}

	return ClientResponse{
		ID:                client.ID,
		ClientName:        client.ClientName,
		ClientSecret:      clientSecret,
		RedirectURIs:      client.RedirectURIs,
		GrantTypes:        client.GrantTypes,
		ResponseTypes:     client.ResponseTypes,
		Scopes:            client.Scopes,
		IsPublic:          client.IsPublic,
		IsInternal:        client.IsInternal,
		TokenEndpointAuth: client.TokenEndpointAuth,
		TenantID:          tenantID,
		ClientURI:         client.ClientURI,
		LogoURL:           client.LogoURL,
		Active:            client.Active,
		CreatedAt:         client.CreatedAt,
		UpdatedAt:         client.UpdatedAt,
	}
}

func validateGrantTypes(grantTypes []string) (string, error) {
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

func validateResponseTypes(responseTypes []string) (string, error) {
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
