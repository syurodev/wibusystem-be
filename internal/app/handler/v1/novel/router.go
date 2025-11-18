package novel

import (
	"github.com/gin-gonic/gin"
)

// RegisterRoutes đăng ký các routes cho novel
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	// Public routes (no auth required for reading)
	router.GET("", h.ListNovels)                   // GET /api/v1/novels
	router.GET("/:id", h.GetNovel)                 // GET /api/v1/novels/:id (ID or slug)
	router.POST("/:id/view", h.IncrementViewCount) // POST /api/v1/novels/:id/view

	// Protected routes (require authentication)
	// Note: Auth middleware should be applied at router group level or here
	// For now, assuming auth middleware is applied globally or at parent level
	router.POST("", h.CreateNovel)       // POST /api/v1/novels
	router.PUT("/:id", h.UpdateNovel)    // PUT /api/v1/novels/:id
	router.DELETE("/:id", h.DeleteNovel) // DELETE /api/v1/novels/:id
}
