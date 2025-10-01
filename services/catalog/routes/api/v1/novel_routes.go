package v1

import (
	"github.com/gin-gonic/gin"

	"wibusystem/services/catalog/handlers"
	"wibusystem/services/catalog/middleware"
)

// SetupNovelRoutes registers novel-related API endpoints
func SetupNovelRoutes(router *gin.RouterGroup, h *handlers.Handlers, m *middleware.Manager) {
	// Public novel endpoints (no authentication required)
	novelPublic := router.Group("/novels")

	// List novels - public endpoint with optional filtering
	novelPublic.GET("", h.Novel.ListNovels)

	// Get novel by ID - public endpoint
	novelPublic.GET("/:novel_id", h.Novel.GetNovelByID)

	// Protected novel endpoints (admin authentication required)
	novelProtected := router.Group("/novels")
	novelProtected.Use(m.SetupAdminAPIMiddleware()...) // Admin required for create/update/delete

	// Novel CRUD endpoints
	novelProtected.POST("", h.Novel.CreateNovel)
	novelProtected.PUT("/:novel_id", h.Novel.UpdateNovel)
	novelProtected.DELETE("/:novel_id", h.Novel.DeleteNovel)
}
