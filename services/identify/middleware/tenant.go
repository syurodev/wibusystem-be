package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	m "wibusystem/pkg/common/model"
	r "wibusystem/pkg/common/response"
	"wibusystem/services/identify/repositories"
)

// TenantMiddleware resolves and enforces tenant context for multi-tenant routes.
type TenantMiddleware struct {
	repos *repositories.Repositories
}

// NewTenantMiddleware creates new tenant middleware.
func NewTenantMiddleware(repos *repositories.Repositories) *TenantMiddleware {
	return &TenantMiddleware{
		repos: repos,
	}
}

// ResolveTenant attempts to resolve the tenant from (in order):
// 1) URL path param :tenant_id
// 2) Query parameter tenant_id
// 3) Header X-Tenant-ID
// 4) Subdomain (e.g., acme.example.com -> "acme")
// If found, the tenant and tenant_id are set on the Gin context.
func (tm *TenantMiddleware) ResolveTenant() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		var tenantID uuid.UUID
		var tenant *m.Tenant
		var err error

		// 1. Try to get tenant from URL path parameter
		if tenantIDStr := c.Param("tenant_id"); tenantIDStr != "" {
			tenantID, err = uuid.Parse(tenantIDStr)
			if err == nil {
				tenant, err = tm.repos.Tenant.GetByID(ctx, tenantID)
			}
		}

		// 2. Try to get tenant from query parameter
		if tenant == nil {
			if tenantIDStr := c.Query("tenant_id"); tenantIDStr != "" {
				tenantID, err = uuid.Parse(tenantIDStr)
				if err == nil {
					tenant, err = tm.repos.Tenant.GetByID(ctx, tenantID)
				}
			}
		}

		// 3. Try to get tenant from custom header
		if tenant == nil {
			if tenantIDStr := c.GetHeader("X-Tenant-ID"); tenantIDStr != "" {
				tenantID, err = uuid.Parse(tenantIDStr)
				if err == nil {
					tenant, err = tm.repos.Tenant.GetByID(ctx, tenantID)
				}
			}
		}

		// 4. Try to get tenant from subdomain (e.g., acme.example.com)
		if tenant == nil {
			host := c.Request.Host
			if subdomain := extractSubdomain(host); subdomain != "" {
				tenant, err = tm.repos.Tenant.GetByName(ctx, subdomain)
			}
		}

		// Store tenant info in context if found
		if tenant != nil {
			c.Set("tenant", tenant)
			c.Set("tenant_id", tenant.ID)
		}

		c.Next()
	}
}

// RequireTenant ensures a tenant has been resolved and stored in context.
func (tm *TenantMiddleware) RequireTenant() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, exists := c.Get("tenant")
		if !exists {
			c.JSON(http.StatusBadRequest, r.StandardResponse{
				Success: false,
				Message: "Missing tenant",
				Data:    nil,
				Error:   &r.ErrorDetail{Code: "missing_tenant", Description: "Tenant context is required"},
				Meta:    map[string]interface{}{},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireTenantMembership ensures the current authenticated user has active
// membership in the resolved tenant.
func (tm *TenantMiddleware) RequireTenantMembership() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// Get tenant from context
		tenantID, exists := c.Get("tenant_id")
		if !exists {
			c.JSON(http.StatusBadRequest, r.StandardResponse{
				Success: false,
				Message: "Missing tenant",
				Data:    nil,
				Error:   &r.ErrorDetail{Code: "missing_tenant", Description: "Tenant context is required"},
				Meta:    map[string]interface{}{},
			})
			c.Abort()
			return
		}

		// Get user from context
		_, exists = GetUserFromContext(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, r.StandardResponse{
				Success: false,
				Message: "Unauthorized",
				Data:    nil,
				Error:   &r.ErrorDetail{Code: "unauthorized", Description: "User authentication required"},
				Meta:    map[string]interface{}{},
			})
			c.Abort()
			return
		}

		// Parse user ID from subject.
		userID, err := GetUserIDFromContext(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, r.StandardResponse{
				Success: false,
				Message: "Server error",
				Data:    nil,
				Error:   &r.ErrorDetail{Code: "server_error", Description: "Failed to parse user ID"},
				Meta:    map[string]interface{}{},
			})
			c.Abort()
			return
		}

		// Check membership
		tenantUUID, ok := tenantID.(uuid.UUID)
		if !ok {
			c.JSON(http.StatusInternalServerError, r.StandardResponse{
				Success: false,
				Message: "Server error",
				Data:    nil,
				Error:   &r.ErrorDetail{Code: "server_error", Description: "Invalid tenant ID type"},
				Meta:    map[string]interface{}{},
			})
			c.Abort()
			return
		}
		membership, err := tm.repos.Membership.GetByUserAndTenant(ctx, userID, tenantUUID)
		if err != nil || membership.Status != "active" {
			c.JSON(http.StatusForbidden, r.StandardResponse{
				Success: false,
				Message: "Access denied",
				Data:    nil,
				Error:   &r.ErrorDetail{Code: "access_denied", Description: "User is not a member of this tenant"},
				Meta:    map[string]interface{}{},
			})
			c.Abort()
			return
		}

		// Store membership info in context
		c.Set("membership", membership)
		c.Next()
	}
}

// RequireTenantRole enforces that the user has one of the required roles
// within the current tenant. Role lookup is simplified and should be
// replaced with a proper role repository/lookup.
func (tm *TenantMiddleware) RequireTenantRole(requiredRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// Get membership from context (should be set by RequireTenantMembership)
		membership, exists := c.Get("membership")
		if !exists {
			c.JSON(http.StatusForbidden, r.StandardResponse{
				Success: false,
				Message: "Access denied",
				Data:    nil,
				Error:   &r.ErrorDetail{Code: "access_denied", Description: "Tenant membership required"},
				Meta:    map[string]interface{}{},
			})
			c.Abort()
			return
		}

		membershipInfo, ok := membership.(*m.Membership)
		if !ok {
			c.JSON(http.StatusInternalServerError, r.StandardResponse{
				Success: false,
				Message: "Server error",
				Data:    nil,
				Error:   &r.ErrorDetail{Code: "server_error", Description: "Invalid membership information"},
				Meta:    map[string]interface{}{},
			})
			c.Abort()
			return
		}

		// Get user roles in this tenant through role assignments
		roleAssignments, err := tm.repos.Membership.ListByUserID(ctx, membershipInfo.UserID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, r.StandardResponse{
				Success: false,
				Message: "Server error",
				Data:    nil,
				Error:   &r.ErrorDetail{Code: "server_error", Description: "Failed to fetch user roles"},
				Meta:    map[string]interface{}{},
			})
			c.Abort()
			return
		}

		// Check if a user has any of the required roles
		hasRequiredRole := false
		userRoles := make([]string, 0)

		for _, assignment := range roleAssignments {
			if assignment.TenantID == membershipInfo.TenantID && assignment.Status == "active" {
				// Get role information (this would require a role lookup)
				// For now, we'll assume the role names are stored somewhere accessibly
				// In a full implementation, you'd query the role table
				for _, required := range requiredRoles {
					// This is simplified - in reality you'd look up a role by ID
					if assignment.RoleID != nil {
						// Role lookup would go here
						hasRequiredRole = true
						userRoles = append(userRoles, required)
					}
				}
			}
		}

		if !hasRequiredRole && len(requiredRoles) > 0 {
			c.JSON(http.StatusForbidden, r.StandardResponse{
				Success: false,
				Message: "Insufficient permissions",
				Data:    nil,
				Error:   &r.ErrorDetail{Code: "insufficient_permissions", Description: "User does not have required role in this tenant"},
				Meta:    map[string]interface{}{},
			})
			c.Abort()
			return
		}

		// Store user roles in context
		c.Set("user_roles", userRoles)
		c.Next()
	}
}

// RequireTenantPermission enforces specific permissions within the tenant.
// TODO: Implement permission lookups via roles->permissions mapping.
func (tm *TenantMiddleware) RequireTenantPermission(requiredPermissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// This would check if the user has specific permissions within the tenant
		// Implementation would involve:
		// 1. Get user's roles in the tenant
		// 2. Get permissions for those roles
		// 3. Check if any of the required permissions are granted

		// For now, we'll implement a basic version
		c.JSON(http.StatusNotImplemented, r.StandardResponse{
			Success: false,
			Message: "Not implemented",
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "not_implemented", Description: "Permission-based authorization not yet implemented"},
			Meta:    map[string]interface{}{},
		})
		c.Abort()
	}
}

// GetTenantFromContext extracts tenant information from Gin context.
func GetTenantFromContext(c *gin.Context) (*m.Tenant, bool) {
	tenantInfo, exists := c.Get("tenant")
	if !exists {
		return nil, false
	}

	tenant, ok := tenantInfo.(*m.Tenant)
	return tenant, ok
}

// GetTenantIDFromContext extracts tenant ID from Gin context.
func GetTenantIDFromContext(c *gin.Context) (uuid.UUID, bool) {
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		return uuid.Nil, false
	}

	id, ok := tenantID.(uuid.UUID)
	return id, ok
}

// GetMembershipFromContext extracts membership information from Gin context.
func GetMembershipFromContext(c *gin.Context) (*m.Membership, bool) {
	membershipInfo, exists := c.Get("membership")
	if !exists {
		return nil, false
	}

	membership, ok := membershipInfo.(*m.Membership)
	return membership, ok
}

// extractSubdomain extracts subdomain from host header
// Example: "acme.example.com" -> "acme"
func extractSubdomain(host string) string {
	// Remove port if present
	if colonIndex := strings.Index(host, ":"); colonIndex != -1 {
		host = host[:colonIndex]
	}

	// Split by dots
	parts := strings.Split(host, ".")

	// If we have at least 3 parts (subdomain.domain.tld), return the first part
	if len(parts) >= 3 {
		return parts[0]
	}

	return ""
}
