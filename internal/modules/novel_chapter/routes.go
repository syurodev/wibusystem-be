package novel_chapter

import (
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers chapter routes
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	// Chapter CRUD operations
	router.GET("/:identifier", h.GetChapter)
	router.POST("", h.CreateChapter)
	router.PUT("/:identifier", h.UpdateChapter)
	router.DELETE("/:identifier", h.DeleteChapter)

	// Chapter operations
	router.POST("/:identifier/publish", h.PublishChapter)
	router.POST("/:identifier/schedule", h.ScheduleChapter)
	router.POST("/:identifier/view", h.IncrementViewCount)
	router.PUT("/:identifier/statistics", h.UpdateStatistics)
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
