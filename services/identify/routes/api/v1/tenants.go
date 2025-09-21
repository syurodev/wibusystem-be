package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"

	r "wibusystem/pkg/common/response"
	"wibusystem/services/identify/handlers"
	"wibusystem/services/identify/middleware"
)

// setupTenantRoutes configures tenant management endpoints for API v1
// This includes both general tenant operations and tenant-scoped routes
func setupTenantRoutes(v1 *gin.RouterGroup, h *handlers.Handlers, m *middleware.Manager) {
	// General tenant management routes
	setupGeneralTenantRoutes(v1, h, m)

	// Tenant-scoped routes (require tenant membership)
	setupTenantScopedRoutes(v1, h, m)
}

// setupGeneralTenantRoutes configures general tenant CRUD operations
func setupGeneralTenantRoutes(v1 *gin.RouterGroup, h *handlers.Handlers, m *middleware.Manager) {
	tenants := v1.Group("/tenants")
	tenants.Use(m.SetupProtectedAPIMiddleware()...)
	{
		// Tenant CRUD operations
		tenants.GET("", h.Tenant.ListTenants)
		tenants.POST("", m.Auth.RequireScope("admin"), h.Tenant.CreateTenant)
		tenants.GET("/:id", h.Tenant.GetTenant)
		tenants.PUT("/:id", h.Tenant.UpdateTenant)
		tenants.DELETE("/:id", m.Auth.RequireScope("admin"), h.Tenant.DeleteTenant)

		// Tenant member management
		tenants.GET("/:id/members", h.Tenant.GetTenantMembers)
		tenants.POST("/:id/members", h.Tenant.AddTenantMember)
		tenants.DELETE("/:id/members/:user_id", h.Tenant.RemoveTenantMember)
	}
}

// setupTenantScopedRoutes configures routes that operate within a specific tenant context
// These routes require active tenant membership and are scoped to a specific tenant
func setupTenantScopedRoutes(v1 *gin.RouterGroup, h *handlers.Handlers, m *middleware.Manager) {
	// Tenant-scoped routes pattern: /api/v1/t/:tenant_id/*
	tenantScoped := v1.Group("/t/:tenant_id")
	tenantScoped.Use(m.SetupTenantAPIMiddleware()...)
	{
		// Example tenant dashboard endpoint
		tenantScoped.GET("/dashboard", tenantDashboardHandler)

		// Add more tenant-specific endpoints here as needed
		// Examples:
		// tenantScoped.GET("/projects", h.Project.ListTenantProjects)
		// tenantScoped.POST("/projects", h.Project.CreateTenantProject)
		// tenantScoped.GET("/settings", h.Tenant.GetTenantSettings)
		// tenantScoped.PUT("/settings", h.Tenant.UpdateTenantSettings)
	}
}

// tenantDashboardHandler provides a welcome dashboard for tenant members
func tenantDashboardHandler(c *gin.Context) {
	tenant, _ := middleware.GetTenantFromContext(c)
	membership, _ := middleware.GetMembershipFromContext(c)

	c.JSON(http.StatusOK, r.StandardResponse{
		Success: true,
		Message: "Welcome to tenant dashboard",
		Data: map[string]interface{}{
			"tenant":     tenant,
			"membership": membership,
		},
		Error: nil,
		Meta:  map[string]interface{}{},
	})
}
