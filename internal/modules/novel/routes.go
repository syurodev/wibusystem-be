package novel

import (
	"github.com/gin-gonic/gin"
)

// RegisterRoutes đăng ký các routes cho novel
func (h *Handler) RegisterRoutes(router *gin.RouterGroup, middlewares ...gin.HandlerFunc) {
	// Public routes (no auth required for reading)
	router.GET("", h.ListNovels)                           // GET /api/v1/novels
	router.GET("/top", h.GetTop)                           // GET /api/v1/novels/top
	router.GET("/:identifier", h.GetNovel)                 // GET /api/v1/novels/:identifier (ID or slug)
	router.GET("/:identifier/full", h.GetNovelFull)        // GET /api/v1/novels/:identifier/full (slug)
	router.POST("/:identifier/view", h.IncrementViewCount) // POST /api/v1/novels/:identifier/view

	// Protected routes (require authentication)
	protected := router.Group("", middlewares...)
	protected.POST("", h.CreateNovel)               // POST /api/v1/novels
	protected.PUT("/:identifier", h.UpdateNovel)    // PUT /api/v1/novels/:identifier
	protected.DELETE("/:identifier", h.DeleteNovel) // DELETE /api/v1/novels/:id
}
