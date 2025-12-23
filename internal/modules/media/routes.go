package media

import (
	"github.com/gin-gonic/gin"
)

// RegisterRoutes đăng ký các routes cho media module
func (h *Handler) RegisterRoutes(router *gin.RouterGroup, requireAuth gin.HandlerFunc) {
	media := router.Group("/media")
	{
		// Public routes - không cần auth
		media.GET("/top", h.GetTop)

		// Protected routes - cần auth (nếu cần thêm sau)
		// protected := media.Group("")
		// protected.Use(requireAuth)
		// {
		// }
	}
}
