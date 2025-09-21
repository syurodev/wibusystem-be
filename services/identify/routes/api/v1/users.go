package v1

import (
	"github.com/gin-gonic/gin"

	"wibusystem/services/identify/handlers"
	"wibusystem/services/identify/middleware"
)

// setupUserRoutes configures user management endpoints for API v1
// These routes handle CRUD operations for users with appropriate authorization
func setupUserRoutes(v1 *gin.RouterGroup, h *handlers.Handlers, m *middleware.Manager) {
	users := v1.Group("/users")
	users.Use(m.SetupProtectedAPIMiddleware()...)
	{
		// Admin-only routes - require admin scope
		users.GET("", m.Auth.RequireScope("admin"), h.User.ListUsers)
		users.POST("", m.Auth.RequireScope("admin"), h.User.CreateUser)

		// User-specific routes - require authentication
		// Individual user operations (self or admin)
		users.GET("/:id", h.User.GetUser)
		users.PUT("/:id", m.Auth.RequireAuth(), h.User.UpdateUser)
		users.DELETE("/:id", m.Auth.RequireAuth(), h.User.DeleteUser)

		// User tenant relationships
		users.GET("/:id/tenants", h.User.GetUserTenants)
	}
}
