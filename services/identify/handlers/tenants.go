package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	d "wibusystem/pkg/common/dto"
	r "wibusystem/pkg/common/response"
	"wibusystem/pkg/i18n"
	"wibusystem/services/identify/middleware"
	"wibusystem/services/identify/services/interfaces"
)

// TenantHandler handles tenant management endpoints
type TenantHandler struct {
	tenantService interfaces.TenantServiceInterface
	loc           *i18n.Translator
}

// NewTenantHandler creates a new tenant handler
func NewTenantHandler(tenantService interfaces.TenantServiceInterface, translator *i18n.Translator) *TenantHandler {
	return &TenantHandler{
		tenantService: tenantService,
		loc:           translator,
	}
}

// ListTenants handles GET /tenants
func (h *TenantHandler) ListTenants(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	// Check if user wants to see only their tenants
	userTenantsOnly := c.Query("user_only") == "true"

	if userTenantsOnly {
		userInfo, exists := middleware.GetUserFromContext(c)
		if !exists {
			message := i18n.Localize(c, "identify.errors.unauthorized.message", "Unauthorized")
			description := i18n.Localize(c, "identify.errors.unauthorized.description", "User authentication required")
			c.JSON(http.StatusUnauthorized, r.StandardResponse{
				Success: false,
				Message: message,
				Data:    nil,
				Error:   &r.ErrorDetail{Code: "unauthorized", Description: description},
				Meta:    map[string]interface{}{},
			})
			return
		}

		userID, _ := uuid.Parse(userInfo.Subject)
		tenants, err := h.tenantService.GetUserTenants(ctx, userID)
		if err != nil {
			status, code, message, description := mapServiceError(c, err, "list_user_tenants")
			c.JSON(status, r.StandardResponse{
				Success: false,
				Message: message,
				Data:    nil,
				Error:   &r.ErrorDetail{Code: code, Description: description},
				Meta:    map[string]interface{}{},
			})
			return
		}

		message := i18n.Localize(c, "identify.tenants.list_success", "Tenants fetched successfully")
		c.JSON(http.StatusOK, r.StandardResponse{
			Success: true,
			Message: message,
			Data:    tenants,
			Error:   nil,
			Meta:    map[string]interface{}{},
		})
		return
	}

	// Get all tenants with pagination (admin only)
	if !isAdmin(c) {
		message := i18n.Localize(c, "identify.errors.access_denied.message", "Access denied")
		description := i18n.Localize(c, "identify.errors.access_denied.description", "Admin privileges required to list all tenants")
		c.JSON(http.StatusForbidden, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "access_denied", Description: description},
			Meta:    map[string]interface{}{},
		})
		return
	}

	tenants, total, err := h.tenantService.ListTenants(ctx, page, pageSize)
	if err != nil {
		status, code, message, description := mapServiceError(c, err, "list")
		c.JSON(status, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: code, Description: description},
			Meta:    map[string]interface{}{},
		})
		return
	}

	// Calculate pagination metadata
	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	message := i18n.Localize(c, "identify.tenants.list_success", "Tenants fetched successfully")
	c.JSON(http.StatusOK, r.StandardResponse{
		Success: true,
		Message: message,
		Data:    tenants,
		Error:   nil,
		Meta: map[string]interface{}{
			"page":        page,
			"page_size":   pageSize,
			"total_pages": totalPages,
			"total_items": total,
		},
	})
}

// GetTenant handles GET /tenants/:id
func (h *TenantHandler) GetTenant(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse tenant ID
	tenantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		message := i18n.Localize(c, "identify.errors.invalid_tenant_id.message", "Invalid tenant id")
		description := i18n.LocalizeWithData(c, "identify.errors.invalid_tenant_id.description", "Tenant ID must be a valid UUID", map[string]interface{}{"id": c.Param("id")})
		c.JSON(http.StatusBadRequest, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "invalid_tenant_id", Description: description},
			Meta:    map[string]interface{}{},
		})
		return
	}

	// Check if user has access to this tenant
	if !h.canAccessTenant(c, tenantID) {
		message := i18n.Localize(c, "identify.errors.access_denied.message", "Access denied")
		description := i18n.Localize(c, "identify.errors.access_denied.description", "You don't have access to this tenant")
		c.JSON(http.StatusForbidden, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "access_denied", Description: description},
			Meta:    map[string]interface{}{},
		})
		return
	}

	// Get tenant through service
	tenant, err := h.tenantService.GetTenantByID(ctx, tenantID)
	if err != nil {
		status, code, message, description := mapServiceError(c, err, "get")
		c.JSON(status, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: code, Description: description},
			Meta:    map[string]interface{}{},
		})
		return
	}

	message := i18n.Localize(c, "identify.tenants.get_success", "Tenant fetched successfully")
	c.JSON(http.StatusOK, r.StandardResponse{
		Success: true,
		Message: message,
		Data:    tenant,
		Error:   nil,
		Meta:    map[string]interface{}{},
	})
}

// CreateTenant handles POST /tenants
func (h *TenantHandler) CreateTenant(c *gin.Context) {
	var req d.CreateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		message := i18n.Localize(c, "identify.errors.invalid_request.message", "Invalid request")
		description := i18n.LocalizeWithData(c, "identify.errors.invalid_request.description", fmt.Sprintf("Invalid request body: %s", err.Error()), map[string]interface{}{"error": err.Error()})
		c.JSON(http.StatusBadRequest, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "invalid_request", Description: description},
			Meta:    map[string]interface{}{},
		})
		return
	}

	ctx := c.Request.Context()

	// Create tenant through service
	tenant, err := h.tenantService.CreateTenant(ctx, req)
	if err != nil {
		status, code, message, description := mapServiceError(c, err, "create")
		c.JSON(status, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: code, Description: description},
			Meta:    map[string]interface{}{},
		})
		return
	}

	// Get current user ID and add them as owner
	userInfo, exists := middleware.GetUserFromContext(c)
	if exists {
		userID, _ := uuid.Parse(userInfo.Subject)
		// Add user as owner (ignore error for now as this is optional)
		_ = h.tenantService.AddUserToTenant(ctx, tenant.ID, userID, "owner")
	}

	message := i18n.Localize(c, "identify.tenants.create_success", "Tenant created successfully")
	c.JSON(http.StatusCreated, r.StandardResponse{
		Success: true,
		Message: message,
		Data:    tenant,
		Error:   nil,
		Meta:    map[string]interface{}{},
	})
}

// UpdateTenant handles PUT /tenants/:id
func (h *TenantHandler) UpdateTenant(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse tenant ID
	tenantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		message := i18n.Localize(c, "identify.errors.invalid_tenant_id.message", "Invalid tenant id")
		description := i18n.LocalizeWithData(c, "identify.errors.invalid_tenant_id.description", "Tenant ID must be a valid UUID", map[string]interface{}{"id": c.Param("id")})
		c.JSON(http.StatusBadRequest, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "invalid_tenant_id", Description: description},
			Meta:    map[string]interface{}{},
		})
		return
	}

	// Check if user can modify this tenant
	if !h.canModifyTenant(c, tenantID) {
		message := i18n.Localize(c, "identify.errors.access_denied.message", "Access denied")
		description := i18n.Localize(c, "identify.errors.access_denied.description", "You don't have permission to modify this tenant")
		c.JSON(http.StatusForbidden, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "access_denied", Description: description},
			Meta:    map[string]interface{}{},
		})
		return
	}

	var req d.UpdateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		message := i18n.Localize(c, "identify.errors.invalid_request.message", "Invalid request")
		description := i18n.LocalizeWithData(c, "identify.errors.invalid_request.description", fmt.Sprintf("Invalid request body: %s", err.Error()), map[string]interface{}{"error": err.Error()})
		c.JSON(http.StatusBadRequest, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "invalid_request", Description: description},
			Meta:    map[string]interface{}{},
		})
		return
	}

	// Update tenant through service
	tenant, err := h.tenantService.UpdateTenant(ctx, tenantID, req)
	if err != nil {
		status, code, message, description := mapServiceError(c, err, "update")
		c.JSON(status, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: code, Description: description},
			Meta:    map[string]interface{}{},
		})
		return
	}

	message := i18n.Localize(c, "identify.tenants.update_success", "Tenant updated successfully")
	c.JSON(http.StatusOK, r.StandardResponse{
		Success: true,
		Message: message,
		Data:    tenant,
		Error:   nil,
		Meta:    map[string]interface{}{},
	})
}

// DeleteTenant handles DELETE /tenants/:id
func (h *TenantHandler) DeleteTenant(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse tenant ID
	tenantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		message := i18n.Localize(c, "identify.errors.invalid_tenant_id.message", "Invalid tenant id")
		description := i18n.LocalizeWithData(c, "identify.errors.invalid_tenant_id.description", "Tenant ID must be a valid UUID", map[string]interface{}{"id": c.Param("id")})
		c.JSON(http.StatusBadRequest, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "invalid_tenant_id", Description: description},
			Meta:    map[string]interface{}{},
		})
		return
	}

	// Check if user can modify this tenant
	if !h.canModifyTenant(c, tenantID) {
		message := i18n.Localize(c, "identify.errors.access_denied.message", "Access denied")
		description := i18n.Localize(c, "identify.errors.access_denied.description", "You don't have permission to delete this tenant")
		c.JSON(http.StatusForbidden, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "access_denied", Description: description},
			Meta:    map[string]interface{}{},
		})
		return
	}

	// Delete tenant through service
	if err := h.tenantService.DeleteTenant(ctx, tenantID); err != nil {
		status, code, message, description := mapServiceError(c, err, "delete")
		c.JSON(status, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: code, Description: description},
			Meta:    map[string]interface{}{},
		})
		return
	}

	message := i18n.Localize(c, "identify.tenants.delete_success", "Tenant deleted successfully")
	c.JSON(http.StatusOK, r.StandardResponse{
		Success: true,
		Message: message,
		Data:    map[string]interface{}{},
		Error:   nil,
		Meta:    map[string]interface{}{},
	})
}

// GetTenantMembers handles GET /tenants/:id/members
func (h *TenantHandler) GetTenantMembers(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse tenant ID
	tenantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		message := i18n.Localize(c, "identify.errors.invalid_tenant_id.message", "Invalid tenant id")
		description := i18n.LocalizeWithData(c, "identify.errors.invalid_tenant_id.description", "Tenant ID must be a valid UUID", map[string]interface{}{"id": c.Param("id")})
		c.JSON(http.StatusBadRequest, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "invalid_tenant_id", Description: description},
			Meta:    map[string]interface{}{},
		})
		return
	}

	// Check if user has access to this tenant
	if !h.canAccessTenant(c, tenantID) {
		message := i18n.Localize(c, "identify.errors.access_denied.message", "Access denied")
		description := i18n.Localize(c, "identify.errors.access_denied.description", "You don't have access to this tenant")
		c.JSON(http.StatusForbidden, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "access_denied", Description: description},
			Meta:    map[string]interface{}{},
		})
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	// Get tenant members through service
	members, total, err := h.tenantService.GetTenantMembers(ctx, tenantID, page, pageSize)
	if err != nil {
		status, code, message, description := mapServiceError(c, err, "get_members")
		c.JSON(status, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: code, Description: description},
			Meta:    map[string]interface{}{},
		})
		return
	}

	// Calculate pagination metadata
	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	message := i18n.Localize(c, "identify.tenants.members_success", "Tenant members fetched successfully")
	c.JSON(http.StatusOK, r.StandardResponse{
		Success: true,
		Message: message,
		Data:    members,
		Error:   nil,
		Meta: map[string]interface{}{
			"page":        page,
			"page_size":   pageSize,
			"total_pages": totalPages,
			"total_items": total,
		},
	})
}

// Helper functions

// canAccessTenant checks if the current user can access a tenant
func (h *TenantHandler) canAccessTenant(c *gin.Context, tenantID uuid.UUID) bool {
	// Admin can access any tenant
	if isAdmin(c) {
		return true
	}

	userInfo, exists := middleware.GetUserFromContext(c)
	if !exists {
		return false
	}

	userID, _ := uuid.Parse(userInfo.Subject)

	// Check if user has access through service
	hasAccess, _ := h.tenantService.CheckUserTenantAccess(c.Request.Context(), userID, tenantID)
	return hasAccess
}

// canModifyTenant checks if the current user can modify a tenant
func (h *TenantHandler) canModifyTenant(c *gin.Context, tenantID uuid.UUID) bool {
	// Admin can modify any tenant
	if isAdmin(c) {
		return true
	}

	_, exists := middleware.GetUserFromContext(c)
	if !exists {
		return false
	}

	// For now, we'll use the same logic as canAccessTenant
	// In a real implementation, you might want to check for specific roles like "owner" or "admin"
	return h.canAccessTenant(c, tenantID)
}
