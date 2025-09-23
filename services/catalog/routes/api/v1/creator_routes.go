package v1

import (
	"github.com/gin-gonic/gin"

	"wibusystem/services/catalog/handlers"
	"wibusystem/services/catalog/middleware"
)

// SetupCreatorRoutes registers creator-related API endpoints
func SetupCreatorRoutes(router *gin.RouterGroup, h *handlers.Handlers, m *middleware.Manager) {
	// Public creator endpoints (no authentication required)
	creatorPublic := router.Group("/creators")
	creatorPublic.GET("", h.Creator.ListCreators)      // GET /api/v1/creators
	creatorPublic.GET("/:id", h.Creator.GetCreator)    // GET /api/v1/creators/:id

	// Protected creator endpoints (admin authentication required)
	creatorProtected := router.Group("/creators")
	creatorProtected.Use(m.SetupAdminAPIMiddleware()...) // Admin required for create/update/delete
	creatorProtected.POST("", h.Creator.CreateCreator)       // POST /api/v1/creators
	creatorProtected.PUT("/:id", h.Creator.UpdateCreator)    // PUT /api/v1/creators/:id
	creatorProtected.DELETE("/:id", h.Creator.DeleteCreator) // DELETE /api/v1/creators/:id
}