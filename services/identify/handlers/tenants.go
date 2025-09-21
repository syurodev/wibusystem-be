package handlers

import (
	"net/http"
	"strconv"

	d "wibusystem/pkg/common/dto"
	m "wibusystem/pkg/common/model"
	r "wibusystem/pkg/common/response"
	"wibusystem/pkg/i18n"
	"wibusystem/services/identify/middleware"
	"wibusystem/services/identify/repositories"

	"github.com/google/uuid"

	"github.com/gin-gonic/gin"
)

// TenantHandler handles tenant management endpoints
type TenantHandler struct {
	repos *repositories.Repositories
	loc   *i18n.Translator
}

// NewTenantHandler creates a new tenant handler
func NewTenantHandler(repos *repositories.Repositories, translator *i18n.Translator) *TenantHandler {
	return &TenantHandler{
		repos: repos,
		loc:   translator,
	}
}

// ListTenants handles GET /tenants
func (h *TenantHandler) ListTenants(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	// Check if user wants to see only their tenants
	userTenantsOnly := c.Query("user_only") == "true"

	if userTenantsOnly {
		userInfo, exists := middleware.GetUserFromContext(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, r.StandardResponse{Success: false, Message: "Unauthorized", Data: nil, Error: &r.ErrorDetail{Code: "unauthorized", Description: "User authentication required"}, Meta: map[string]interface{}{}})
			return
		}

		userID, _ := uuid.Parse(userInfo.Subject)
		tenants, err := h.repos.Tenant.GetByUserID(ctx, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, r.StandardResponse{Success: false, Message: "Database error", Data: nil, Error: &r.ErrorDetail{Code: "database_error", Description: "Failed to fetch user tenants: " + err.Error()}, Meta: map[string]interface{}{}})
			return
		}

		c.JSON(http.StatusOK, r.StandardResponse{Success: true, Message: "Tenants fetched successfully", Data: tenants, Error: nil, Meta: map[string]interface{}{}})
		return
	}

	// Get all tenants with pagination (admin only)
	if !isAdmin(c) {
		c.JSON(http.StatusForbidden, r.StandardResponse{Success: false, Message: "Access denied", Data: nil, Error: &r.ErrorDetail{Code: "access_denied", Description: "Admin privileges required to list all tenants"}, Meta: map[string]interface{}{}})
		return
	}

	tenants, total, err := h.repos.Tenant.List(ctx, pageSize, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, r.StandardResponse{Success: false, Message: "Database error", Data: nil, Error: &r.ErrorDetail{Code: "database_error", Description: "Failed to fetch tenants: " + err.Error()}, Meta: map[string]interface{}{}})
		return
	}

	// Calculate pagination metadata
	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, r.StandardResponse{Success: true, Message: "Tenants fetched successfully", Data: tenants, Error: nil, Meta: map[string]interface{}{"page": page, "page_size": pageSize, "total_pages": totalPages, "total_items": total}})
}

// GetTenant handles GET /tenants/:id
func (h *TenantHandler) GetTenant(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse tenant ID
	tenantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, r.StandardResponse{Success: false, Message: "Invalid tenant id", Data: nil, Error: &r.ErrorDetail{Code: "invalid_tenant_id", Description: "Tenant ID must be a valid integer"}, Meta: map[string]interface{}{}})
		return
	}

	// Get tenant
	tenant, err := h.repos.Tenant.GetByID(ctx, tenantID)
	if err != nil {
		c.JSON(http.StatusNotFound, r.StandardResponse{Success: false, Message: "Tenant not found", Data: nil, Error: &r.ErrorDetail{Code: "tenant_not_found", Description: "Tenant not found"}, Meta: map[string]interface{}{}})
		return
	}

	// Check if user has access to this tenant
	if !canAccessTenant(c, tenantID) {
		c.JSON(http.StatusForbidden, r.StandardResponse{Success: false, Message: "Access denied", Data: nil, Error: &r.ErrorDetail{Code: "access_denied", Description: "You don't have access to this tenant"}, Meta: map[string]interface{}{}})
		return
	}

	// Get tenant members if requested
	includeMembers := c.Query("include_members") == "true"
	var memberships []*m.Membership
	if includeMembers {
		memberships, _ = h.repos.Membership.ListByTenantID(ctx, tenantID)
	}

	responseData := map[string]interface{}{
		"tenant": tenant,
	}

	if includeMembers {
		responseData["memberships"] = memberships
	}

	c.JSON(http.StatusOK, r.StandardResponse{Success: true, Message: "Tenant fetched successfully", Data: responseData, Error: nil, Meta: map[string]interface{}{}})
}

// CreateTenant handles POST /tenants
func (h *TenantHandler) CreateTenant(c *gin.Context) {
	var req d.CreateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, r.StandardResponse{Success: false, Message: "Invalid request", Data: nil, Error: &r.ErrorDetail{Code: "invalid_request", Description: "Invalid request body: " + err.Error()}, Meta: map[string]interface{}{}})
		return
	}

	ctx := c.Request.Context()

	// Check if tenant name already exists (optional - depends on business rules)
	existingTenant, _ := h.repos.Tenant.GetByName(ctx, req.Name)
	if existingTenant != nil {
		c.JSON(http.StatusConflict, r.StandardResponse{Success: false, Message: "Tenant exists", Data: nil, Error: &r.ErrorDetail{Code: "tenant_exists", Description: "Tenant with this name already exists"}, Meta: map[string]interface{}{}})
		return
	}

	// Create tenant
	tenant := &m.Tenant{
		Name: req.Name,
	}

	if err := h.repos.Tenant.Create(ctx, tenant); err != nil {
		c.JSON(http.StatusInternalServerError, r.StandardResponse{Success: false, Message: "Tenant creation failed", Data: nil, Error: &r.ErrorDetail{Code: "tenant_creation_failed", Description: "Failed to create tenant: " + err.Error()}, Meta: map[string]interface{}{}})
		return
	}

	// Get current user ID
	userInfo, exists := middleware.GetUserFromContext(c)
	if exists {
		userID, _ := uuid.Parse(userInfo.Subject)

		// Create membership for the creator as owner
		membership := &m.Membership{
			UserID:   userID,
			TenantID: tenant.ID,
			Status:   "active",
		}

		// Create membership (ignore error for now)
		_ = h.repos.Membership.Create(ctx, membership)
	}

	c.JSON(http.StatusCreated, r.StandardResponse{Success: true, Message: "Tenant created successfully", Data: tenant, Error: nil, Meta: map[string]interface{}{}})
}

// UpdateTenant handles PUT /tenants/:id
func (h *TenantHandler) UpdateTenant(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse tenant ID
	tenantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, r.StandardResponse{Success: false, Message: "Invalid tenant id", Data: nil, Error: &r.ErrorDetail{Code: "invalid_tenant_id", Description: "Tenant ID must be a valid integer"}, Meta: map[string]interface{}{}})
		return
	}

	// Check if tenant exists
	tenant, err := h.repos.Tenant.GetByID(ctx, tenantID)
	if err != nil {
		c.JSON(http.StatusNotFound, r.StandardResponse{Success: false, Message: "Tenant not found", Data: nil, Error: &r.ErrorDetail{Code: "tenant_not_found", Description: "Tenant not found"}, Meta: map[string]interface{}{}})
		return
	}

	// Check if user can modify this tenant
	if !canModifyTenant(c, tenantID) {
		c.JSON(http.StatusForbidden, r.StandardResponse{Success: false, Message: "Access denied", Data: nil, Error: &r.ErrorDetail{Code: "access_denied", Description: "You don't have permission to modify this tenant"}, Meta: map[string]interface{}{}})
		return
	}

	var req struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, r.StandardResponse{Success: false, Message: "Invalid request", Data: nil, Error: &r.ErrorDetail{Code: "invalid_request", Description: "Invalid request body: " + err.Error()}, Meta: map[string]interface{}{}})
		return
	}

	// Check if name is already taken by another tenant
	if req.Name != tenant.Name {
		existingTenant, _ := h.repos.Tenant.GetByName(ctx, req.Name)
		if existingTenant != nil && existingTenant.ID != tenantID {
			c.JSON(http.StatusConflict, r.StandardResponse{Success: false, Message: "Tenant name taken", Data: nil, Error: &r.ErrorDetail{Code: "tenant_name_taken", Description: "Tenant name is already taken"}, Meta: map[string]interface{}{}})
			return
		}
	}

	// Update tenant
	tenant.Name = req.Name
	if err := h.repos.Tenant.Update(ctx, tenant); err != nil {
		c.JSON(http.StatusInternalServerError, r.StandardResponse{Success: false, Message: "Tenant update failed", Data: nil, Error: &r.ErrorDetail{Code: "tenant_update_failed", Description: "Failed to update tenant: " + err.Error()}, Meta: map[string]interface{}{}})
		return
	}

	c.JSON(http.StatusOK, r.StandardResponse{Success: true, Message: "Tenant updated successfully", Data: tenant, Error: nil, Meta: map[string]interface{}{}})
}

// DeleteTenant handles DELETE /tenants/:id
func (h *TenantHandler) DeleteTenant(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse tenant ID
	tenantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, r.StandardResponse{Success: false, Message: "Invalid tenant id", Data: nil, Error: &r.ErrorDetail{Code: "invalid_tenant_id", Description: "Tenant ID must be a valid integer"}, Meta: map[string]interface{}{}})
		return
	}

	// Check if tenant exists
	_, err = h.repos.Tenant.GetByID(ctx, tenantID)
	if err != nil {
		c.JSON(http.StatusNotFound, r.StandardResponse{Success: false, Message: "Tenant not found", Data: nil, Error: &r.ErrorDetail{Code: "tenant_not_found", Description: "Tenant not found"}, Meta: map[string]interface{}{}})
		return
	}

	// Check if user can delete this tenant (admin only)
	if !isAdmin(c) {
		c.JSON(http.StatusForbidden, r.StandardResponse{Success: false, Message: "Access denied", Data: nil, Error: &r.ErrorDetail{Code: "access_denied", Description: "Admin privileges required to delete tenants"}, Meta: map[string]interface{}{}})
		return
	}

	// Delete tenant (this will cascade to memberships, roles, etc.)
	if err := h.repos.Tenant.Delete(ctx, tenantID); err != nil {
		c.JSON(http.StatusInternalServerError, r.StandardResponse{Success: false, Message: "Tenant deletion failed", Data: nil, Error: &r.ErrorDetail{Code: "tenant_deletion_failed", Description: "Failed to delete tenant: " + err.Error()}, Meta: map[string]interface{}{}})
		return
	}

	c.JSON(http.StatusOK, r.StandardResponse{Success: true, Message: "Tenant deleted successfully", Data: map[string]interface{}{}, Error: nil, Meta: map[string]interface{}{}})
}

// GetTenantMembers handles GET /tenants/:id/members
func (h *TenantHandler) GetTenantMembers(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse tenant ID
	tenantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, r.StandardResponse{Success: false, Message: "Invalid tenant id", Data: nil, Error: &r.ErrorDetail{Code: "invalid_tenant_id", Description: "Tenant ID must be a valid integer"}, Meta: map[string]interface{}{}})
		return
	}

	// Check if user has access to this tenant
	if !canAccessTenant(c, tenantID) {
		c.JSON(http.StatusForbidden, r.StandardResponse{Success: false, Message: "Access denied", Data: nil, Error: &r.ErrorDetail{Code: "access_denied", Description: "You don't have access to this tenant"}, Meta: map[string]interface{}{}})
		return
	}

	// Get tenant memberships
	memberships, err := h.repos.Membership.ListByTenantID(ctx, tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, r.StandardResponse{Success: false, Message: "Database error", Data: nil, Error: &r.ErrorDetail{Code: "database_error", Description: "Failed to fetch tenant members: " + err.Error()}, Meta: map[string]interface{}{}})
		return
	}

	c.JSON(http.StatusOK, r.StandardResponse{Success: true, Message: "Tenant members fetched successfully", Data: memberships, Error: nil, Meta: map[string]interface{}{}})
}

// AddTenantMember handles POST /tenants/:id/members
func (h *TenantHandler) AddTenantMember(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse tenant ID
	tenantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, r.StandardResponse{Success: false, Message: "Invalid tenant id", Data: nil, Error: &r.ErrorDetail{Code: "invalid_tenant_id", Description: "Tenant ID must be a valid integer"}, Meta: map[string]interface{}{}})
		return
	}

	// Check if user can modify this tenant
	if !canModifyTenant(c, tenantID) {
		c.JSON(http.StatusForbidden, r.StandardResponse{Success: false, Message: "Access denied", Data: nil, Error: &r.ErrorDetail{Code: "access_denied", Description: "You don't have permission to modify this tenant"}, Meta: map[string]interface{}{}})
		return
	}

	var req struct {
		UserID uuid.UUID `json:"user_id" binding:"required"`
		Status string    `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, r.StandardResponse{Success: false, Message: "Invalid request", Data: nil, Error: &r.ErrorDetail{Code: "invalid_request", Description: "Invalid request body: " + err.Error()}, Meta: map[string]interface{}{}})
		return
	}

	if req.Status == "" {
		req.Status = "active"
	}

	// Check if user exists
	_, err = h.repos.User.GetByID(ctx, req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, r.StandardResponse{Success: false, Message: "User not found", Data: nil, Error: &r.ErrorDetail{Code: "user_not_found", Description: "User not found"}, Meta: map[string]interface{}{}})
		return
	}

	// Check if membership already exists
	existingMembership, _ := h.repos.Membership.GetByUserAndTenant(ctx, req.UserID, tenantID)
	if existingMembership != nil {
		c.JSON(http.StatusConflict, r.StandardResponse{Success: false, Message: "Membership exists", Data: nil, Error: &r.ErrorDetail{Code: "membership_exists", Description: "User is already a member of this tenant"}, Meta: map[string]interface{}{}})
		return
	}

	// Create membership
	membership := &m.Membership{
		UserID:   req.UserID,
		TenantID: tenantID,
		Status:   req.Status,
	}

	if err := h.repos.Membership.Create(ctx, membership); err != nil {
		c.JSON(http.StatusInternalServerError, r.StandardResponse{Success: false, Message: "Membership creation failed", Data: nil, Error: &r.ErrorDetail{Code: "membership_creation_failed", Description: "Failed to add user to tenant: " + err.Error()}, Meta: map[string]interface{}{}})
		return
	}

	c.JSON(http.StatusCreated, r.StandardResponse{Success: true, Message: "User added to tenant successfully", Data: membership, Error: nil, Meta: map[string]interface{}{}})
}

// RemoveTenantMember handles DELETE /tenants/:id/members/:user_id
func (h *TenantHandler) RemoveTenantMember(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse tenant ID
	tenantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, r.StandardResponse{Success: false, Message: "Invalid tenant id", Data: nil, Error: &r.ErrorDetail{Code: "invalid_tenant_id", Description: "Tenant ID must be a valid integer"}, Meta: map[string]interface{}{}})
		return
	}

	// Parse user ID
	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, r.StandardResponse{Success: false, Message: "Invalid user id", Data: nil, Error: &r.ErrorDetail{Code: "invalid_user_id", Description: "User ID must be a valid integer"}, Meta: map[string]interface{}{}})
		return
	}

	// Check if user can modify this tenant
	if !canModifyTenant(c, tenantID) {
		c.JSON(http.StatusForbidden, r.StandardResponse{Success: false, Message: "Access denied", Data: nil, Error: &r.ErrorDetail{Code: "access_denied", Description: "You don't have permission to modify this tenant"}, Meta: map[string]interface{}{}})
		return
	}

	// Get membership
	membership, err := h.repos.Membership.GetByUserAndTenant(ctx, userID, tenantID)
	if err != nil {
		c.JSON(http.StatusNotFound, r.StandardResponse{Success: false, Message: "Membership not found", Data: nil, Error: &r.ErrorDetail{Code: "membership_not_found", Description: "User is not a member of this tenant"}, Meta: map[string]interface{}{}})
		return
	}

	// Delete membership
	if err := h.repos.Membership.Delete(ctx, membership.ID); err != nil {
		c.JSON(http.StatusInternalServerError, r.StandardResponse{Success: false, Message: "Membership deletion failed", Data: nil, Error: &r.ErrorDetail{Code: "membership_deletion_failed", Description: "Failed to remove user from tenant: " + err.Error()}, Meta: map[string]interface{}{}})
		return
	}

	c.JSON(http.StatusOK, r.StandardResponse{Success: true, Message: "User removed from tenant successfully", Data: map[string]interface{}{}, Error: nil, Meta: map[string]interface{}{}})
}

// Helper functions

// canAccessTenant checks if the current user can access the specified tenant
func canAccessTenant(c *gin.Context, tenantID uuid.UUID) bool {
	// Admins can access any tenant
	if isAdmin(c) {
		return true
	}

	// Check if user is a member of the tenant
	_, exists := middleware.GetUserFromContext(c)
	if !exists {
		return false
	}

	// This is simplified - in a full implementation, you'd check membership in the database
	// For now, we'll allow access if tenant context is resolved
	contextTenantID, exists := middleware.GetTenantIDFromContext(c)
	return exists && contextTenantID == tenantID
}

// canModifyTenant checks if the current user can modify the specified tenant
func canModifyTenant(c *gin.Context, tenantID uuid.UUID) bool {
	// Admins can modify any tenant
	if isAdmin(c) {
		return true
	}

	// Check if user has management role in the tenant
	// This is simplified - in a full implementation, you'd check roles/permissions
	return canAccessTenant(c, tenantID)
}
