package novel_volume

import (
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers volume routes
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	// Volume CRUD operations
	router.GET("/:id", h.GetVolume)
	router.POST("", h.CreateVolume)
	router.PUT("/:id", h.UpdateVolume)
	router.DELETE("/:id", h.DeleteVolume)

	// Volume operations
	router.PUT("/:id/display-order", h.UpdateDisplayOrder)
	router.POST("/:id/publish", h.PublishVolume)
	router.POST("/:id/unpublish", h.UnpublishVolume)
}

// RegisterNovelVolumesRoutes registers routes for getting volumes by novel
// This should be registered under /api/v1/novels/:novel_id/volumes
func (h *Handler) RegisterNovelVolumesRoutes(router *gin.RouterGroup) {
	router.GET("", h.ListVolumesByNovel)
}
