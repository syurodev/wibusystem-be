package artist

import (
	"github.com/gin-gonic/gin"
)

// RegisterRoutes đăng ký các routes cho artist
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	// Public routes (no auth required for reading)
	router.GET("", h.ListArtists)   // GET /api/v1/artists
	router.GET("/:id", h.GetArtist) // GET /api/v1/artists/:id

	// Protected routes (require authentication)
	// Note: Auth middleware should be applied at router group level or here
	// For now, assuming auth middleware is applied globally or at parent level
	router.POST("", h.CreateArtist)       // POST /api/v1/artists
	router.PUT("/:id", h.UpdateArtist)    // PUT /api/v1/artists/:id
	router.DELETE("/:id", h.DeleteArtist) // DELETE /api/v1/artists/:id
}
