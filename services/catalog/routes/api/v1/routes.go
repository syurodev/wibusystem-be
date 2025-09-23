package v1

import (
	"github.com/gin-gonic/gin"

	"wibusystem/services/catalog/config"
	"wibusystem/services/catalog/handlers"
	"wibusystem/services/catalog/middleware"
)

// SetupRoutes registers version 1 API endpoints for the Catalog service.
func SetupRoutes(router *gin.Engine, cfg *config.Config, h *handlers.Handlers, m *middleware.Manager) {
	api := router.Group("/api/v1")
	api.Use(m.SetupPublicAPIMiddleware()...)

	// Health endpoint for API consumers (mirrors /healthz but versioned).
	api.GET("/health", h.Health.Status)
}
