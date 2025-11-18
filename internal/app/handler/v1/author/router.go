package author

import (
	"github.com/gin-gonic/gin"
)

// RegisterRoutes đăng ký các routes cho author
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	// Public routes (no auth required for reading)
	router.GET("", h.ListAuthors)   // GET /api/v1/authors
	router.GET("/:id", h.GetAuthor) // GET /api/v1/authors/:id

	// Protected routes (require authentication)
	// Note: Auth middleware should be applied at router group level or here
	// For now, assuming auth middleware is applied globally or at parent level
	router.POST("", h.CreateAuthor)       // POST /api/v1/authors
	router.PUT("/:id", h.UpdateAuthor)    // PUT /api/v1/authors/:id
	router.DELETE("/:id", h.DeleteAuthor) // DELETE /api/v1/authors/:id
}
