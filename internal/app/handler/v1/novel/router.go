package novel

import (
	"github.com/gin-gonic/gin"
)

// RegisterRoutes đăng ký các routes cho novel
func (h *Handler) RegisterRoutes(router *gin.RouterGroup, middlewares ...gin.HandlerFunc) {
	// Public routes (no auth required for reading)
	router.GET("", h.ListNovels)                   // GET /api/v1/novels
	router.GET("/:id", h.GetNovel)                 // GET /api/v1/novels/:id (ID or slug)
	router.POST("/:id/view", h.IncrementViewCount) // POST /api/v1/novels/:id/view

	// Protected routes (require authentication)
	protected := router.Group("", middlewares...)
	protected.POST("", h.CreateNovel)       // POST /api/v1/novels
	protected.PUT("/:id", h.UpdateNovel)    // PUT /api/v1/novels/:id
	protected.DELETE("/:id", h.DeleteNovel) // DELETE /api/v1/novels/:id
}
