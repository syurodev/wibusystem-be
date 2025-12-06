package genre

import (
	"github.com/gin-gonic/gin"
)

// RegisterRoutes đăng ký các routes cho genre
// RegisterRoutes đăng ký các routes cho genre
func (h *Handler) RegisterRoutes(router *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	// Public routes (no auth required for reading)
	router.GET("", h.ListGenres)       // GET /api/v1/genres
	router.GET("/selection", h.ListSelection) // GET /api/v1/genres/selection
	router.GET("/:id", h.GetGenre)     // GET /api/v1/genres/:id

	// Protected routes (require authentication)
	protected := router.Group("")
	if authMiddleware != nil {
		protected.Use(authMiddleware)
	}
	protected.POST("", h.CreateGenre)      // POST /api/v1/genres
	protected.POST("/merge", h.MergeGenre) // POST /api/v1/genres/merge
	protected.POST("/merge/preview", h.PreviewMergeGenre) // POST /api/v1/genres/merge/preview
	protected.PUT("/:id", h.UpdateGenre)   // PUT /api/v1/genres/:id
	protected.DELETE("/:id", h.DeleteGenre) // DELETE /api/v1/genres/:id
}
