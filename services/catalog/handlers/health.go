package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	r "wibusystem/pkg/common/response"
	"wibusystem/pkg/i18n"
	"wibusystem/services/catalog/repositories"
)

// HealthHandler exposes health check endpoints.
type HealthHandler struct {
	repos      *repositories.Repositories
	translator *i18n.Translator
}

// NewHealthHandler creates a new HealthHandler instance.
func NewHealthHandler(repos *repositories.Repositories, translator *i18n.Translator) *HealthHandler {
	return &HealthHandler{repos: repos, translator: translator}
}

// Status returns the health of the catalog service and backing resources.
func (h *HealthHandler) Status(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	var issues []string
	if err := h.repos.Health.Ping(ctx); err != nil {
		issues = append(issues, err.Error())
	}

	if len(issues) > 0 {
		message := i18n.T(c, "catalog.health.degraded", "Catalog service is experiencing issues", nil)
		c.JSON(http.StatusServiceUnavailable, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    map[string]any{"status": "degraded"},
			Error:   &r.ErrorDetail{Code: "service_degraded", Description: issues[0]},
			Meta:    map[string]any{"issues": issues},
		})
		return
	}

	message := i18n.T(c, "catalog.health.ok", "Catalog service is healthy", nil)
	c.JSON(http.StatusOK, r.StandardResponse{
		Success: true,
		Message: message,
		Data: map[string]any{
			"status": "ok",
		},
		Error: nil,
		Meta:  map[string]any{},
	})
}
