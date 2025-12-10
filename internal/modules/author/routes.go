package author

import (
	"github.com/gin-gonic/gin"
)

// RegisterRoutes đăng ký các routes cho author
func (h *Handler) RegisterRoutes(router *gin.RouterGroup, requireAuth gin.HandlerFunc) {
	// Public routes (no auth required for reading)
	router.GET("", h.ListAuthors)   // GET /api/v1/authors
	router.GET("/selection", h.ListSelection) // GET /api/v1/authors/selection
	router.GET("/:identifier", h.GetAuthor) // GET /api/v1/authors/:identifier

	// Protected routes (require authentication)
	protected := router.Group("")
	protected.Use(requireAuth)
	{
		protected.POST("", h.CreateAuthor)       // POST /api/v1/authors
		protected.PUT("/:identifier", h.UpdateAuthor)    // PUT /api/v1/authors/:identifier
		protected.DELETE("/:identifier", h.DeleteAuthor) // DELETE /api/v1/authors/:identifier

		// Merge routes
		protected.POST("/merge", h.MergeAuthor)          // POST /api/v1/authors/merge
		protected.POST("/merge/preview", h.PreviewMergeAuthor) // POST /api/v1/authors/merge/preview
	}
}
