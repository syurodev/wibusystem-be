package v1

import (
	"github.com/gin-gonic/gin"

	"wibusystem/services/catalog/handlers"
	"wibusystem/services/catalog/middleware"
)

// SetupCharacterRoutes registers character-related API endpoints
func SetupCharacterRoutes(router *gin.RouterGroup, h *handlers.Handlers, m *middleware.Manager) {
	// Public character endpoints (no authentication required)
	characterPublic := router.Group("/characters")
	characterPublic.GET("", h.Character.ListCharacters)      // GET /api/v1/characters
	characterPublic.GET("/:id", h.Character.GetCharacter)    // GET /api/v1/characters/:id

	// Protected character endpoints (admin authentication required)
	characterProtected := router.Group("/characters")
	characterProtected.Use(m.SetupAdminAPIMiddleware()...) // Admin required for create/update/delete
	characterProtected.POST("", h.Character.CreateCharacter)       // POST /api/v1/characters
	characterProtected.PUT("/:id", h.Character.UpdateCharacter)    // PUT /api/v1/characters/:id
	characterProtected.DELETE("/:id", h.Character.DeleteCharacter) // DELETE /api/v1/characters/:id
}