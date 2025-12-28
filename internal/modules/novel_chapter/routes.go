package novel_chapter

import (
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers chapter routes
// authMiddleware is optional - if nil, protected routes won't be registered
func (h *Handler) RegisterRoutes(router *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	// More specific routes MUST come before wildcard routes in Gin
	// Public routes
	router.GET("/:identifier/full", h.GetChapterFull) // Get full chapter with novel/volume/owner info
	router.GET("/:identifier", h.GetChapter)
	router.POST("/:identifier/view", h.IncrementViewCount)

	// Protected routes (require auth)
	if authMiddleware != nil {
		protected := router.Group("", authMiddleware)
		protected.POST("", h.CreateChapter)
		protected.PUT("/:identifier", h.UpdateChapter)
		protected.DELETE("/:identifier", h.DeleteChapter)
		protected.POST("/:identifier/publish", h.PublishChapter)
		protected.POST("/:identifier/schedule", h.ScheduleChapter)
		protected.PUT("/:identifier/statistics", h.UpdateStatistics)
	}
}

// RegisterNovelChaptersRoutes registers routes for getting chapters by novel
// This should be registered under /api/v1/novels/:novel_id/chapters
func (h *Handler) RegisterNovelChaptersRoutes(router *gin.RouterGroup) {
	router.GET("", h.ListChaptersByNovel)
}

// RegisterVolumeChaptersRoutes registers routes for getting chapters by volume
// This should be registered under /api/v1/volumes/:volume_id/chapters
func (h *Handler) RegisterVolumeChaptersRoutes(router *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	router.GET("", h.ListChaptersByVolume)
	if authMiddleware != nil {
		router.POST("", authMiddleware, h.CreateChapter)
	}
}
