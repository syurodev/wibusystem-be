package genre

import (
	"github.com/gin-gonic/gin"
)

func (h *Handler) RegisterRoutes(router *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	// Public routes (no auth required for reading)
	router.GET("", h.ListGenres)       // GET /api/v1/genres
	router.GET("/selection", h.ListSelection) // GET /api/v1/genres/selection
	router.GET("/:identifier", h.GetGenre)     // GET /api/v1/genres/:identifier

	// Protected routes (require authentication)
	protected := router.Group("")
	if authMiddleware != nil {
		protected.Use(authMiddleware)
	}
	protected.POST("", h.CreateGenre)      // POST /api/v1/genres
	protected.POST("/merge", h.MergeGenre) // POST /api/v1/genres/merge
	protected.POST("/merge/preview", h.PreviewMergeGenre) // POST /api/v1/genres/merge/preview
	protected.PUT("/:identifier", h.UpdateGenre)   // PUT /api/v1/genres/:identifier
	protected.DELETE("/:identifier", h.DeleteGenre) // DELETE /api/v1/genres/:identifier
}
