package public

import (
	"net/http"
	"system/internal/pkg/service"
	"system/pkg/util/i18nkeys"
	"system/pkg/util/response"

	"github.com/gin-gonic/gin"
)

// Handler handles public API endpoints.
type Handler struct {
	publicService *service.PublicService
}

// NewHandler creates a new public Handler instance.
func NewHandler(publicService *service.PublicService) *Handler {
	return &Handler{
		publicService: publicService,
	}
}

// GetHomeData returns aggregated data for the home page.
// @Summary Get home page data
// @Description Get aggregated data for home page (hero, trending, creators, genres)
// @Tags Public
// @Produce json
// @Success 200 {object} response.StandardResponse{data=service.HomeData}
// @Failure 500 {object} response.StandardResponse
// @Router /api/v1/public/home [get]
func (h *Handler) GetHomeData(c *gin.Context) {
	homeData, err := h.publicService.GetHomeData(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "GET_HOME_DATA_FAILED", i18nkeys.GeneralError, nil)
		return
	}

	response.Success(c, http.StatusOK, i18nkeys.GeneralSuccess, homeData, nil)
}

// RegisterRoutes registers public routes.
func (h *Handler) RegisterRoutes(group *gin.RouterGroup) {
	group.GET("/home", h.GetHomeData)
}
