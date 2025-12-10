package artist

import (
	"github.com/gin-gonic/gin"
)

// RegisterRoutes đăng ký các routes cho artist
func (h *Handler) RegisterRoutes(router *gin.RouterGroup, requireAuth gin.HandlerFunc) {
	artists := router.Group("/artists")
	{
		// Public routes - không cần auth
		artists.GET("", h.ListArtists)
		artists.GET("/selection", h.ListSelection) // New selection endpoint
		artists.GET("/:identifier", h.GetArtist)

		// Protected routes - cần auth
		protected := artists.Group("")
		protected.Use(requireAuth)
		{
			protected.POST("", h.CreateArtist)
			protected.PUT("/:identifier", h.UpdateArtist)
			protected.DELETE("/:identifier", h.DeleteArtist)

			// Merge routes
			protected.POST("/merge", h.MergeArtist)          // POST /api/v1/artists/merge
			protected.POST("/merge/preview", h.PreviewMergeArtist) // POST /api/v1/artists/merge/preview
		}
	}
}
