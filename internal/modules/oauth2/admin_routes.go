package oauth2

import "github.com/gin-gonic/gin"

// RegisterAdminRoutes đăng ký các routes cho OAuth2 Admin API.
func RegisterAdminRoutes(router *gin.RouterGroup, handler *AdminHandler) {
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
