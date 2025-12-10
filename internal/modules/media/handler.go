package media

import (
	"encoding/json"
	"net/http"
	"strconv"

	mediadto "system/internal/dto/media"
	analytics_module "system/internal/modules/analytics"
	"system/pkg/util/response"

	"github.com/gin-gonic/gin"
)

// Handler handles media API endpoints
// Cung cấp APIs cho tất cả media types (anime, manga, novel) combined
type Handler struct {
	analyticsService analytics_module.AnalyticsService
	mediaService     MediaService
}

// NewHandler creates a new media Handler instance
func NewHandler(analyticsService analytics_module.AnalyticsService, mediaService MediaService) *Handler {
	return &Handler{
		analyticsService: analyticsService,
		mediaService:     mediaService,
	}
}

// GetTrending returns top trending media
// @Summary Get top trending media
// @Description Get top trending media (novel, manga, anime) based on views
// @Tags Media
// @Accept json
// @Produce json
// @Param type query string false "Media type (novel, manga, anime)"
// @Param range query string false "Time range (day, week, month)"
// @Param limit query int false "Limit (default 20)"
// @Success 200 {array} MediaSeriesResponse
// @Router /api/v1/media/trending [get]
func (h *Handler) GetTrending(c *gin.Context) {
	mediaType := c.Query("type")
	timeRange := c.Query("range")
	limitStr := c.Query("limit")

	limit := 20 // Default limit
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	results, err := h.analyticsService.GetTopTrending(c.Request.Context(), mediaType, timeRange, limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "GET_TRENDING_FAILED", I18nAnalyticsGetTrendingFailed, err.Error())
		return
	}

	// Convert map to typed response for consistency
	var resp []mediadto.MediaSeriesResponse
	bytes, err := json.Marshal(results)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "SERIALIZATION_FAILED", I18nAnalyticsSerializationFailed, nil)
		return
	}
	if err := json.Unmarshal(bytes, &resp); err != nil {
		response.Error(c, http.StatusInternalServerError, "MAPPING_FAILED", I18nAnalyticsMappingFailed, nil)
		return
	}

	response.Success(c, http.StatusOK, I18nAnalyticsGetTrendingSuccess, resp, nil)
}

// GetHomeData returns aggregated data for the home page
// @Summary Get home page data
// @Description Get aggregated data for home page (hero, trending, creators, genres) across all media types
// @Tags Media
// @Produce json
// @Success 200 {object} response.StandardResponse{data=HomeData}
// @Failure 500 {object} response.StandardResponse
// @Router /api/v1/media/home [get]
func (h *Handler) GetHomeData(c *gin.Context) {
	homeData, err := h.mediaService.GetHomeData(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "GET_HOME_DATA_FAILED", I18nMediaGetHomeFailed, nil)
		return
	}

	response.Success(c, http.StatusOK, I18nMediaGetHomeSuccess, homeData, nil)
}


