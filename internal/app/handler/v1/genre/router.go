package genre

import (
	"github.com/gin-gonic/gin"
)

// RegisterRoutes đăng ký các routes cho genre
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	// Public routes (no auth required for reading)
	router.GET("", h.ListGenres)       // GET /api/v1/genres
	router.GET("/:id", h.GetGenre)     // GET /api/v1/genres/:id

	// Protected routes (require authentication)
	// Note: Auth middleware should be applied at router group level or here
	// For now, assuming auth middleware is applied globally or at parent level
	router.POST("", h.CreateGenre)      // POST /api/v1/genres
	router.PUT("/:id", h.UpdateGenre)   // PUT /api/v1/genres/:id
	router.DELETE("/:id", h.DeleteGenre) // DELETE /api/v1/genres/:id
}
