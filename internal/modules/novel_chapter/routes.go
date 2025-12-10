package novel_chapter

import (
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers chapter routes
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	// Chapter CRUD operations
	router.GET("/:id", h.GetChapter)
	router.POST("", h.CreateChapter)
	router.PUT("/:id", h.UpdateChapter)
	router.DELETE("/:id", h.DeleteChapter)

	// Chapter operations
	router.POST("/:id/publish", h.PublishChapter)
	router.POST("/:id/schedule", h.ScheduleChapter)
	router.POST("/:id/view", h.IncrementViewCount)
	router.PUT("/:id/statistics", h.UpdateStatistics)
}

// RegisterNovelChaptersRoutes registers routes for getting chapters by novel
// This should be registered under /api/v1/novels/:novel_id/chapters
func (h *Handler) RegisterNovelChaptersRoutes(router *gin.RouterGroup) {
	router.GET("", h.ListChaptersByNovel)
}

// RegisterVolumeChaptersRoutes registers routes for getting chapters by volume
// This should be registered under /api/v1/volumes/:volume_id/chapters
func (h *Handler) RegisterVolumeChaptersRoutes(router *gin.RouterGroup) {
	router.GET("", h.ListChaptersByVolume)
}
