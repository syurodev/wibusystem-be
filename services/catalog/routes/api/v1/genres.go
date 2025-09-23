package v1

import (
	"github.com/gin-gonic/gin"

	"wibusystem/services/catalog/handlers"
	"wibusystem/services/catalog/middleware"
)

// SetupGenreRoutes registers genre-related API endpoints
func SetupGenreRoutes(router *gin.RouterGroup, h *handlers.Handlers, m *middleware.Manager) {
	// Public genre endpoints (no authentication required)
	genrePublic := router.Group("/genres")
	genrePublic.GET("", h.Genre.ListGenres)      // GET /api/v1/genres
	genrePublic.GET("/:id", h.Genre.GetGenre)    // GET /api/v1/genres/:id

	// Protected genre endpoints (admin authentication required)
	genreProtected := router.Group("/genres")
	genreProtected.Use(m.SetupAdminAPIMiddleware()...) // Admin required for create/update/delete
	genreProtected.POST("", h.Genre.CreateGenre)       // POST /api/v1/genres
	genreProtected.PUT("/:id", h.Genre.UpdateGenre)    // PUT /api/v1/genres/:id
	genreProtected.DELETE("/:id", h.Genre.DeleteGenre) // DELETE /api/v1/genres/:id
}