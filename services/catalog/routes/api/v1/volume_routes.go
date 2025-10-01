package v1

import (
	"github.com/gin-gonic/gin"

	"wibusystem/services/catalog/handlers"
	"wibusystem/services/catalog/middleware"
)

// SetupVolumeRoutes registers volume-related API endpoints
// This follows the API design specification from /services/catalog/api-design/novel.md section 2
// Volumes can only be managed by authenticated admin users
func SetupVolumeRoutes(router *gin.RouterGroup, h *handlers.Handlers, m *middleware.Manager) {
	// Volume routes under /novels/{novel_id}/volumes
	// These endpoints are protected and require admin authentication
	novelVolumes := router.Group("/novels/:novel_id/volumes")
	novelVolumes.Use(m.SetupAdminAPIMiddleware()...) // Admin required for volume management

	// List volumes for a novel (with pagination)
	// GET /api/v1/novels/{novel_id}/volumes
	novelVolumes.GET("", h.Volume.ListVolumesByNovelID)

	// Create a new volume for a novel
	// POST /api/v1/novels/{novel_id}/volumes
	novelVolumes.POST("", h.Volume.CreateVolume)

	// Direct volume routes under /volumes/{volume_id}
	// These endpoints operate on specific volumes by ID
	volumes := router.Group("/volumes")
	volumes.Use(m.SetupAdminAPIMiddleware()...) // Admin required for volume operations

	// Get volume details by ID
	// GET /api/v1/volumes/{volume_id}
	volumes.GET("/:volume_id", h.Volume.GetVolumeByID)

	// Update a volume
	// PUT /api/v1/volumes/{volume_id}
	volumes.PUT("/:volume_id", h.Volume.UpdateVolume)

	// Delete a volume (soft delete)
	// DELETE /api/v1/volumes/{volume_id}
	volumes.DELETE("/:volume_id", h.Volume.DeleteVolume)
}
