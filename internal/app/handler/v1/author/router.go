package author

import (
	"github.com/gin-gonic/gin"
)

// RegisterRoutes đăng ký các routes cho author
func (h *Handler) RegisterRoutes(router *gin.RouterGroup, requireAuth gin.HandlerFunc) {
	// Public routes (no auth required for reading)
	router.GET("", h.ListAuthors)   // GET /api/v1/authors
	router.GET("/selection", h.ListSelection) // GET /api/v1/authors/selection
	router.GET("/:id", h.GetAuthor) // GET /api/v1/authors/:id

	// Protected routes (require authentication)
	protected := router.Group("")
	protected.Use(requireAuth)
	{
		protected.POST("", h.CreateAuthor)       // POST /api/v1/authors
		protected.PUT("/:id", h.UpdateAuthor)    // PUT /api/v1/authors/:id
		protected.DELETE("/:id", h.DeleteAuthor) // DELETE /api/v1/authors/:id
	}
}
