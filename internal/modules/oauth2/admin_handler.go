package oauth2

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid/v5"

	"system/internal/domain"

	"system/pkg/util/response"
)

// AdminHandler xử lý OAuth2 client management endpoints
type AdminHandler struct {
	adminService OAuth2AdminService
}

// NewAdminHandler tạo instance mới của AdminHandler
func NewAdminHandler(adminService OAuth2AdminService) *AdminHandler {
	return &AdminHandler{
		adminService: adminService,
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
func (h *AdminHandler) CreateClient(c *gin.Context) {
	var req CreateClientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_FAILED", I18nValidationFailed, err.Error())
		return
	}

	// Validate grant types and response types
	if invalidType, err := h.adminService.ValidateGrantTypes(req.GrantTypes); err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_GRANT_TYPE", I18nInvalidGrantType, map[string]any{"type": invalidType})
		return
	}

	if invalidType, err := h.adminService.ValidateResponseTypes(req.ResponseTypes); err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_RESPONSE_TYPE", I18nInvalidResponseType, map[string]any{"type": invalidType})
		return
	}

	// Parse tenant ID if provided
	var organizationID *uuid.UUID
	if req.OrganizationID != nil {
		tid, err := uuid.FromString(*req.OrganizationID)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "INVALID_TENANT_ID", I18nInvalidTenantID, nil)
			return
		}
		organizationID = &tid
	}

	// Create client via service
	serviceReq := AdminCreateClientRequest{
		ClientName:        req.ClientName,
		RedirectURIs:      req.RedirectURIs,
		GrantTypes:        req.GrantTypes,
		ResponseTypes:     req.ResponseTypes,
		Scopes:            req.Scopes,
		IsPublic:          req.IsPublic,
		IsInternal:        req.IsInternal,
		TokenEndpointAuth: req.TokenEndpointAuth,
		OrganizationID:          organizationID,
		ClientURI:         req.ClientURI,
		LogoURL:           req.LogoURL,
	}

	client, clientSecret, err := h.adminService.CreateClient(c.Request.Context(), serviceReq)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "CREATE_FAILED", I18nClientCreateFailed, nil)
		return
	}

	// Build response
	clientResponse := h.buildClientResponse(client, clientSecret)
	response.Success(c, http.StatusCreated, I18nClientCreated, clientResponse, nil)
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
func (h *AdminHandler) GetClient(c *gin.Context) {
	clientID, err := uuid.FromString(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_CLIENT_ID", I18nInvalidClientID, nil)
		return
	}

	client, err := h.adminService.GetClientByID(c.Request.Context(), clientID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "CLIENT_NOT_FOUND", I18nClientNotFound, nil)
		return
	}

	clientResponse := h.buildClientResponse(client, nil)
	response.Success(c, http.StatusOK, I18nClientRetrieved, clientResponse, nil)
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
func (h *AdminHandler) ListClients(c *gin.Context) {
	var organizationID *uuid.UUID
	if organizationIDStr := c.Query("tenant_id"); organizationIDStr != "" {
		tid, err := uuid.FromString(organizationIDStr)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "INVALID_TENANT_ID", I18nInvalidTenantID, nil)
			return
		}
		organizationID = &tid
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

	clients, total, err := h.adminService.ListClients(c.Request.Context(), organizationID, active, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "LIST_FAILED", I18nClientListFailed, nil)
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
	response.Success(c, http.StatusOK, I18nClientListed, listResponse, nil)
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
func (h *AdminHandler) UpdateClient(c *gin.Context) {
	clientID, err := uuid.FromString(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_CLIENT_ID", I18nInvalidClientID, nil)
		return
	}

	var req UpdateClientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_FAILED", I18nValidationFailed, err.Error())
		return
	}

	// Validate grant types and response types if provided
	if req.GrantTypes != nil {
		if invalidType, err := h.adminService.ValidateGrantTypes(*req.GrantTypes); err != nil {
			response.Error(c, http.StatusBadRequest, "INVALID_GRANT_TYPE", I18nInvalidGrantType, map[string]any{"type": invalidType})
			return
		}
	}
	if req.ResponseTypes != nil {
		if invalidType, err := h.adminService.ValidateResponseTypes(*req.ResponseTypes); err != nil {
			response.Error(c, http.StatusBadRequest, "INVALID_RESPONSE_TYPE", I18nInvalidResponseType, map[string]any{"type": invalidType})
			return
		}
	}

	// Update client via service
	serviceReq := AdminUpdateClientRequest{
		ClientName:        req.ClientName,
		RedirectURIs:      req.RedirectURIs,
		GrantTypes:        req.GrantTypes,
		ResponseTypes:     req.ResponseTypes,
		Scopes:            req.Scopes,
		IsPublic:          req.IsPublic,
		IsInternal:        req.IsInternal,
		TokenEndpointAuth: req.TokenEndpointAuth,
		ClientURI:         req.ClientURI,
		LogoURL:           req.LogoURL,
		Active:            req.Active,
	}

	client, err := h.adminService.UpdateClient(c.Request.Context(), clientID, serviceReq)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "UPDATE_FAILED", I18nClientUpdateFailed, nil)
		return
	}

	clientResponse := h.buildClientResponse(client, nil)
	response.Success(c, http.StatusOK, I18nClientUpdated, clientResponse, nil)
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
func (h *AdminHandler) DeleteClient(c *gin.Context) {
	clientID, err := uuid.FromString(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_CLIENT_ID", I18nInvalidClientID, nil)
		return
	}

	if err := h.adminService.DeleteClient(c.Request.Context(), clientID); err != nil {
		response.Error(c, http.StatusInternalServerError, "DELETE_FAILED", I18nClientDeleteFailed, nil)
		return
	}

	response.Success(c, http.StatusOK, I18nClientDeleted, nil, nil)
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
func (h *AdminHandler) RegenerateSecret(c *gin.Context) {
	clientID, err := uuid.FromString(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_CLIENT_ID", I18nInvalidClientID, nil)
		return
	}

	// Regenerate secret via service
	newSecret, err := h.adminService.RegenerateClientSecret(c.Request.Context(), clientID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "SECRET_REGENERATION_FAILED", I18nClientSecretGenFailed, nil)
		return
	}

	// Get updated client for response
	client, err := h.adminService.GetClientByID(c.Request.Context(), clientID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "CLIENT_RETRIEVAL_FAILED", I18nClientNotFound, nil)
		return
	}

	clientResponse := h.buildClientResponse(client, &newSecret)
	response.Success(c, http.StatusOK, I18nClientSecretRegenerated, clientResponse, nil)
}

// Helper methods

func (h *AdminHandler) buildClientResponse(client *domain.OAuth2Client, clientSecret *string) ClientResponse {
	var organizationID *string
	if client.OrganizationID != nil {
		tid := client.OrganizationID.String()
		organizationID = &tid
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
		OrganizationID:          organizationID,
		ClientURI:         client.ClientURI,
		LogoURL:           client.LogoURL,
		Active:            client.Active,
		CreatedAt:         client.CreatedAt,
		UpdatedAt:         client.UpdatedAt,
	}
}
