package v1

import (
	"github.com/gin-gonic/gin"

	"wibusystem/services/identify/handlers"
	"wibusystem/services/identify/middleware"
)

// setupAuthRoutes configures authentication-related endpoints for API v1
// This includes both public routes (register, login, refresh, logout) and
// protected routes (profile, change-password) with appropriate middleware
func setupAuthRoutes(v1 *gin.RouterGroup, h *handlers.Handlers, m *middleware.Manager) {
	// Public authentication routes - no auth required
	auth := v1.Group("/auth")
	{
		auth.POST("/register", h.Auth.Register)
		auth.POST("/login", h.Auth.Login)
		auth.POST("/refresh", h.Auth.RefreshToken)
		auth.POST("/logout", h.Auth.Logout) // Logout should work even with expired auth
	}

	// Protected authentication routes - require valid authentication
	authProtected := v1.Group("/auth")
	authProtected.Use(m.Auth.RequireAuth())
	{
		authProtected.GET("/profile", h.Auth.Profile)
		authProtected.POST("/change-password", h.Auth.ChangePassword)
	}
}
