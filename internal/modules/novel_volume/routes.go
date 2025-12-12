package novel_volume

import (
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers volume routes
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	// Volume CRUD operations
	router.GET("/:identifier", h.GetVolume)
	router.POST("", h.CreateVolume)
	router.PUT("/:identifier", h.UpdateVolume)
	router.DELETE("/:identifier", h.DeleteVolume)

	// Volume operations
	router.PUT("/:identifier/display-order", h.UpdateDisplayOrder)
	router.POST("/:identifier/publish", h.PublishVolume)
	router.POST("/:identifier/unpublish", h.UnpublishVolume)
}

// RegisterNovelVolumesRoutes registers routes for getting volumes by novel
// This should be registered under /api/v1/novels/:novel_id/volumes
func (h *Handler) RegisterNovelVolumesRoutes(router *gin.RouterGroup) {
	router.GET("", h.ListVolumesByNovel)
}
