package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	d "wibusystem/pkg/common/dto"
	r "wibusystem/pkg/common/response"
	"wibusystem/pkg/i18n"
	"wibusystem/services/identify/services/interfaces"
)

// AdminHandler exposes administrative role management endpoints.
type AdminHandler struct {
	globalRoleService interfaces.GlobalRoleServiceInterface
	tenantRoleService interfaces.TenantRoleServiceInterface
	loc               *i18n.Translator
}

// NewAdminHandler constructs a new AdminHandler instance.
func NewAdminHandler(globalService interfaces.GlobalRoleServiceInterface, tenantService interfaces.TenantRoleServiceInterface, translator *i18n.Translator) *AdminHandler {
	return &AdminHandler{
		globalRoleService: globalService,
		tenantRoleService: tenantService,
		loc:               translator,
	}
}

// ListGlobalRoles handles GET /admin/global-roles
func (h *AdminHandler) ListGlobalRoles(c *gin.Context) {
	roles, err := h.globalRoleService.ListRoles(c.Request.Context())
	if err != nil {
		status, code, message, description := mapServiceError(c, err, "list_global_roles")
		c.JSON(status, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: code, Description: description},
			Meta:    map[string]interface{}{},
		})
		return
	}

	c.JSON(http.StatusOK, r.StandardResponse{
		Success: true,
		Message: i18n.Localize(c, "identify.admin.global_roles.list_success", "Global roles retrieved"),
		Data:    roles,
		Error:   nil,
		Meta:    map[string]interface{}{},
	})
}

// AssignGlobalRole handles POST /admin/global-roles/assign
func (h *AdminHandler) AssignGlobalRole(c *gin.Context) {
	var req d.AssignGlobalRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, r.StandardResponse{
			Success: false,
			Message: i18n.Localize(c, "identify.errors.validation_failed.message", "Validation failed"),
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "validation_failed", Description: err.Error()},
			Meta:    map[string]interface{}{},
		})
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, r.StandardResponse{
			Success: false,
			Message: i18n.Localize(c, "identify.errors.invalid_user_id.message", "Invalid user id"),
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "invalid_user_id", Description: err.Error()},
			Meta:    map[string]interface{}{},
		})
		return
	}

	roleID, err := uuid.Parse(req.RoleID)
	if err != nil {
		c.JSON(http.StatusBadRequest, r.StandardResponse{
			Success: false,
			Message: i18n.Localize(c, "identify.errors.invalid_role_id.message", "Invalid role id"),
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "invalid_role_id", Description: err.Error()},
			Meta:    map[string]interface{}{},
		})
		return
	}

	if err := h.globalRoleService.AssignRole(c.Request.Context(), userID, roleID); err != nil {
		status, code, message, description := mapServiceError(c, err, "assign_global_role")
		c.JSON(status, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: code, Description: description},
			Meta:    map[string]interface{}{},
		})
		return
	}

	c.JSON(http.StatusOK, r.StandardResponse{
		Success: true,
		Message: i18n.Localize(c, "identify.admin.global_roles.assign_success", "Global role assigned"),
		Data:    nil,
		Error:   nil,
		Meta:    map[string]interface{}{},
	})
}

// RemoveGlobalRole handles DELETE /admin/global-roles/:role_id/users/:user_id
func (h *AdminHandler) RemoveGlobalRole(c *gin.Context) {
	roleID, err := uuid.Parse(c.Param("role_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, r.StandardResponse{
			Success: false,
			Message: i18n.Localize(c, "identify.errors.invalid_role_id.message", "Invalid role id"),
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "invalid_role_id", Description: err.Error()},
			Meta:    map[string]interface{}{},
		})
		return
	}

	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, r.StandardResponse{
			Success: false,
			Message: i18n.Localize(c, "identify.errors.invalid_user_id.message", "Invalid user id"),
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "invalid_user_id", Description: err.Error()},
			Meta:    map[string]interface{}{},
		})
		return
	}

	if err := h.globalRoleService.RemoveRole(c.Request.Context(), userID, roleID); err != nil {
		status, code, message, description := mapServiceError(c, err, "remove_global_role")
		c.JSON(status, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: code, Description: description},
			Meta:    map[string]interface{}{},
		})
		return
	}

	c.JSON(http.StatusOK, r.StandardResponse{
		Success: true,
		Message: i18n.Localize(c, "identify.admin.global_roles.remove_success", "Global role removed"),
		Data:    nil,
		Error:   nil,
		Meta:    map[string]interface{}{},
	})
}

// ListTenantPermissions handles GET /admin/tenant-permissions
func (h *AdminHandler) ListTenantPermissions(c *gin.Context) {
	perms, err := h.tenantRoleService.ListPermissions(c.Request.Context())
	if err != nil {
		status, code, message, description := mapServiceError(c, err, "list_tenant_permissions")
		c.JSON(status, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: code, Description: description},
			Meta:    map[string]interface{}{},
		})
		return
	}

	c.JSON(http.StatusOK, r.StandardResponse{
		Success: true,
		Message: i18n.Localize(c, "identify.admin.tenant_permissions.list_success", "Tenant permissions retrieved"),
		Data:    perms,
		Error:   nil,
		Meta:    map[string]interface{}{},
	})
}

// ListTenantRoles handles GET /admin/tenants/:tenant_id/roles
func (h *AdminHandler) ListTenantRoles(c *gin.Context) {
	tenantID, err := uuid.Parse(c.Param("tenant_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, r.StandardResponse{
			Success: false,
			Message: i18n.Localize(c, "identify.errors.invalid_tenant_id.message", "Invalid tenant id"),
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "invalid_tenant_id", Description: err.Error()},
			Meta:    map[string]interface{}{},
		})
		return
	}

	roles, err := h.tenantRoleService.ListRoles(c.Request.Context(), tenantID)
	if err != nil {
		status, code, message, description := mapServiceError(c, err, "list_tenant_roles")
		c.JSON(status, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: code, Description: description},
			Meta:    map[string]interface{}{},
		})
		return
	}

	c.JSON(http.StatusOK, r.StandardResponse{
		Success: true,
		Message: i18n.Localize(c, "identify.admin.tenant_roles.list_success", "Tenant roles retrieved"),
		Data:    roles,
		Error:   nil,
		Meta:    map[string]interface{}{},
	})
}

// CreateTenantRole handles POST /admin/tenants/:tenant_id/roles
func (h *AdminHandler) CreateTenantRole(c *gin.Context) {
	tenantID, err := uuid.Parse(c.Param("tenant_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, r.StandardResponse{
			Success: false,
			Message: i18n.Localize(c, "identify.errors.invalid_tenant_id.message", "Invalid tenant id"),
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "invalid_tenant_id", Description: err.Error()},
			Meta:    map[string]interface{}{},
		})
		return
	}

	var payload d.TenantRolePayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, r.StandardResponse{
			Success: false,
			Message: i18n.Localize(c, "identify.errors.validation_failed.message", "Validation failed"),
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "validation_failed", Description: err.Error()},
			Meta:    map[string]interface{}{},
		})
		return
	}

	role, err := h.tenantRoleService.CreateRole(c.Request.Context(), tenantID, payload)
	if err != nil {
		status, code, message, description := mapServiceError(c, err, "create_tenant_role")
		c.JSON(status, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: code, Description: description},
			Meta:    map[string]interface{}{},
		})
		return
	}

	c.JSON(http.StatusCreated, r.StandardResponse{
		Success: true,
		Message: i18n.Localize(c, "identify.admin.tenant_roles.create_success", "Tenant role created"),
		Data:    role,
		Error:   nil,
		Meta:    map[string]interface{}{},
	})
}

// UpdateTenantRole handles PUT /admin/tenants/:tenant_id/roles/:role_id
func (h *AdminHandler) UpdateTenantRole(c *gin.Context) {
	tenantID, err := uuid.Parse(c.Param("tenant_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, r.StandardResponse{
			Success: false,
			Message: i18n.Localize(c, "identify.errors.invalid_tenant_id.message", "Invalid tenant id"),
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "invalid_tenant_id", Description: err.Error()},
			Meta:    map[string]interface{}{},
		})
		return
	}

	roleID, err := uuid.Parse(c.Param("role_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, r.StandardResponse{
			Success: false,
			Message: i18n.Localize(c, "identify.errors.invalid_role_id.message", "Invalid role id"),
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "invalid_role_id", Description: err.Error()},
			Meta:    map[string]interface{}{},
		})
		return
	}

	var payload d.TenantRolePayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, r.StandardResponse{
			Success: false,
			Message: i18n.Localize(c, "identify.errors.validation_failed.message", "Validation failed"),
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "validation_failed", Description: err.Error()},
			Meta:    map[string]interface{}{},
		})
		return
	}

	role, err := h.tenantRoleService.UpdateRole(c.Request.Context(), tenantID, roleID, payload)
	if err != nil {
		status, code, message, description := mapServiceError(c, err, "update_tenant_role")
		c.JSON(status, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: code, Description: description},
			Meta:    map[string]interface{}{},
		})
		return
	}

	c.JSON(http.StatusOK, r.StandardResponse{
		Success: true,
		Message: i18n.Localize(c, "identify.admin.tenant_roles.update_success", "Tenant role updated"),
		Data:    role,
		Error:   nil,
		Meta:    map[string]interface{}{},
	})
}

// DeleteTenantRole handles DELETE /admin/tenants/:tenant_id/roles/:role_id
func (h *AdminHandler) DeleteTenantRole(c *gin.Context) {
	tenantID, err := uuid.Parse(c.Param("tenant_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, r.StandardResponse{
			Success: false,
			Message: i18n.Localize(c, "identify.errors.invalid_tenant_id.message", "Invalid tenant id"),
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "invalid_tenant_id", Description: err.Error()},
			Meta:    map[string]interface{}{},
		})
		return
	}

	roleID, err := uuid.Parse(c.Param("role_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, r.StandardResponse{
			Success: false,
			Message: i18n.Localize(c, "identify.errors.invalid_role_id.message", "Invalid role id"),
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "invalid_role_id", Description: err.Error()},
			Meta:    map[string]interface{}{},
		})
		return
	}

	if err := h.tenantRoleService.DeleteRole(c.Request.Context(), tenantID, roleID); err != nil {
		status, code, message, description := mapServiceError(c, err, "delete_tenant_role")
		c.JSON(status, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: code, Description: description},
			Meta:    map[string]interface{}{},
		})
		return
	}

	c.JSON(http.StatusOK, r.StandardResponse{
		Success: true,
		Message: i18n.Localize(c, "identify.admin.tenant_roles.delete_success", "Tenant role deleted"),
		Data:    nil,
		Error:   nil,
		Meta:    map[string]interface{}{},
	})
}

// AssignTenantRole handles POST /admin/tenants/:tenant_id/roles/:role_id/assign
func (h *AdminHandler) AssignTenantRole(c *gin.Context) {
	tenantID, err := uuid.Parse(c.Param("tenant_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, r.StandardResponse{
			Success: false,
			Message: i18n.Localize(c, "identify.errors.invalid_tenant_id.message", "Invalid tenant id"),
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "invalid_tenant_id", Description: err.Error()},
			Meta:    map[string]interface{}{},
		})
		return
	}

	roleID, err := uuid.Parse(c.Param("role_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, r.StandardResponse{
			Success: false,
			Message: i18n.Localize(c, "identify.errors.invalid_role_id.message", "Invalid role id"),
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "invalid_role_id", Description: err.Error()},
			Meta:    map[string]interface{}{},
		})
		return
	}

	var req d.AssignTenantRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, r.StandardResponse{
			Success: false,
			Message: i18n.Localize(c, "identify.errors.validation_failed.message", "Validation failed"),
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "validation_failed", Description: err.Error()},
			Meta:    map[string]interface{}{},
		})
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, r.StandardResponse{
			Success: false,
			Message: i18n.Localize(c, "identify.errors.invalid_user_id.message", "Invalid user id"),
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "invalid_user_id", Description: err.Error()},
			Meta:    map[string]interface{}{},
		})
		return
	}

	if err := h.tenantRoleService.AssignRoleToUser(c.Request.Context(), tenantID, roleID, userID); err != nil {
		status, code, message, description := mapServiceError(c, err, "assign_tenant_role")
		c.JSON(status, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: code, Description: description},
			Meta:    map[string]interface{}{},
		})
		return
	}

	c.JSON(http.StatusOK, r.StandardResponse{
		Success: true,
		Message: i18n.Localize(c, "identify.admin.tenant_roles.assign_success", "Tenant role assigned"),
		Data:    nil,
		Error:   nil,
		Meta:    map[string]interface{}{},
	})
}

// RemoveTenantRole handles DELETE /admin/tenants/:tenant_id/roles/:role_id/users/:user_id
func (h *AdminHandler) RemoveTenantRole(c *gin.Context) {
	tenantID, err := uuid.Parse(c.Param("tenant_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, r.StandardResponse{
			Success: false,
			Message: i18n.Localize(c, "identify.errors.invalid_tenant_id.message", "Invalid tenant id"),
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "invalid_tenant_id", Description: err.Error()},
			Meta:    map[string]interface{}{},
		})
		return
	}

	roleID, err := uuid.Parse(c.Param("role_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, r.StandardResponse{
			Success: false,
			Message: i18n.Localize(c, "identify.errors.invalid_role_id.message", "Invalid role id"),
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "invalid_role_id", Description: err.Error()},
			Meta:    map[string]interface{}{},
		})
		return
	}

	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, r.StandardResponse{
			Success: false,
			Message: i18n.Localize(c, "identify.errors.invalid_user_id.message", "Invalid user id"),
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "invalid_user_id", Description: err.Error()},
			Meta:    map[string]interface{}{},
		})
		return
	}

	if err := h.tenantRoleService.RemoveRoleFromUser(c.Request.Context(), tenantID, roleID, userID); err != nil {
		status, code, message, description := mapServiceError(c, err, "remove_tenant_role")
		c.JSON(status, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: code, Description: description},
			Meta:    map[string]interface{}{},
		})
		return
	}

	c.JSON(http.StatusOK, r.StandardResponse{
		Success: true,
		Message: i18n.Localize(c, "identify.admin.tenant_roles.remove_success", "Tenant role removed"),
		Data:    nil,
		Error:   nil,
		Meta:    map[string]interface{}{},
	})
}
