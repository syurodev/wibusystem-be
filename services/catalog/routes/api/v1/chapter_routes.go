package v1

import (
	"github.com/gin-gonic/gin"

	"wibusystem/services/catalog/handlers"
	"wibusystem/services/catalog/middleware"
)

// SetupChapterRoutes registers chapter management endpoints.
// This follows the API design spec from /services/catalog/api-design/novel.md sections 3.1-3.7
//
// Route structure:
//   - GET    /volumes/{volume_id}/chapters         - List chapters in a volume
//   - POST   /volumes/{volume_id}/chapters         - Create a new chapter
//   - GET    /chapters/{id}                        - Get chapter details
//   - PUT    /chapters/{id}                        - Update chapter
//   - DELETE /chapters/{id}                        - Delete chapter
//   - POST   /chapters/{id}/publish                - Publish chapter
//   - POST   /chapters/{id}/unpublish              - Unpublish chapter
func SetupChapterRoutes(router *gin.RouterGroup, h *handlers.Handlers, m *middleware.Manager) {
	// Chapter routes under /volumes/{volume_id}/chapters
	// These routes handle listing and creating chapters within a specific volume
	volumeChapters := router.Group("/volumes/:volume_id/chapters")
	volumeChapters.Use(m.SetupAdminAPIMiddleware()...)
	{
		volumeChapters.GET("", h.Chapter.ListChaptersByVolumeID)    // List chapters in volume
		volumeChapters.POST("", h.Chapter.CreateChapter)            // Create new chapter
	}

	// Direct chapter routes under /chapters/{id}
	// These routes handle operations on individual chapters
	chapters := router.Group("/chapters")
	chapters.Use(m.SetupAdminAPIMiddleware()...)
	{
		chapters.GET("/:id", h.Chapter.GetChapterByID)              // Get chapter details
		chapters.PUT("/:id", h.Chapter.UpdateChapter)               // Update chapter
	}
}