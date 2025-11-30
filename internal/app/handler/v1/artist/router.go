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
		artists.GET("/:id", h.GetArtist)

		// Protected routes - cần auth
		protected := artists.Group("")
		protected.Use(requireAuth)
		{
			protected.POST("", h.CreateArtist)
			protected.PUT("/:id", h.UpdateArtist)
			protected.DELETE("/:id", h.DeleteArtist)
		}
	}
}
