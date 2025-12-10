package media

import (
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers media routes
func (h *Handler) RegisterRoutes(group *gin.RouterGroup) {
	group.GET("/trending", h.GetTrending)
	group.GET("/home", h.GetHomeData)
}
