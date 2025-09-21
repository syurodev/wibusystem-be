package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"

	r "wibusystem/pkg/common/response"
	"wibusystem/services/identify/handlers"
	"wibusystem/services/identify/middleware"
)

// setupAdminRoutes configures admin-only endpoints for API v1
// These routes require admin privileges and may be further restricted in production
func setupAdminRoutes(router *gin.Engine, h *handlers.Handlers, m *middleware.Manager) {
	admin := router.Group("/admin")
	admin.Use(m.SetupAdminMiddleware()...)
	{
		// System administration endpoints
		admin.GET("/stats", adminStatsHandler)
		admin.GET("/system", adminSystemHandler)

		// Dynamic Client Registration (DCR) Initial Access Token management
		// These endpoints control the issuance of tokens for OAuth2 client registration
		admin.POST("/registration/iat", h.OAuth2.AdminCreateIAT)
		admin.GET("/registration/iat", h.OAuth2.AdminListIAT)
		admin.DELETE("/registration/iat/:id", h.OAuth2.AdminRevokeIAT)
	}
}

// adminStatsHandler provides administrative statistics
func adminStatsHandler(c *gin.Context) {
	c.JSON(http.StatusOK, r.StandardResponse{
		Success: true,
		Message: "Admin statistics endpoint",
		Data:    map[string]interface{}{"admin": true},
		Error:   nil,
		Meta:    map[string]interface{}{},
	})
}

// adminSystemHandler provides system administration information
func adminSystemHandler(c *gin.Context) {
	c.JSON(http.StatusOK, r.StandardResponse{
		Success: true,
		Message: "System administration endpoint",
		Data:    map[string]interface{}{"admin": true},
		Error:   nil,
		Meta:    map[string]interface{}{},
	})
}
