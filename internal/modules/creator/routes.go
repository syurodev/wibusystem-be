package creator

import (
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers creator routes
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("/creators", h.ListCreators)
	router.GET("/creators/top", h.GetTopCreatorsByViews)
}
