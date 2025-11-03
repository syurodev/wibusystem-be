package oauth2_admin

import "github.com/gin-gonic/gin"

// RegisterRoutes đăng ký các routes cho OAuth2 Admin API.
func RegisterRoutes(router *gin.RouterGroup, handler *Handler) {
	oauth2Admin := router.Group("/admin/oauth2")
	{
		// Client management endpoints
		oauth2Admin.POST("/clients", handler.CreateClient)
		oauth2Admin.GET("/clients", handler.ListClients)
		oauth2Admin.GET("/clients/:id", handler.GetClient)
		oauth2Admin.PUT("/clients/:id", handler.UpdateClient)
		oauth2Admin.DELETE("/clients/:id", handler.DeleteClient)
		oauth2Admin.POST("/clients/:id/regenerate-secret", handler.RegenerateSecret)
	}
}
