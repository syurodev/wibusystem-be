package analytics

import (
	"encoding/json"
	"net/http"
	"strconv"
	"system/internal/pkg/service"
	"system/pkg/util/i18nkeys"
	"system/pkg/util/response"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	analyticsService *service.AnalyticsService
}

func NewHandler(analyticsService *service.AnalyticsService) *Handler {
	return &Handler{
		analyticsService: analyticsService,
	}
}

// GetTrending returns top trending media
// @Summary Get top trending media
// @Description Get top trending media (novel, manga, anime) based on views
// @Tags analytics
// @Accept json
// @Produce json
// @Param type query string false "Media type (novel, manga, anime)"
// @Param range query string false "Time range (day, week, month)"
// @Param limit query int false "Limit (default 20)"
// @Success 200 {array} MediaSeriesResponse
// @Router /api/v1/analytics/trending [get]
func (h *Handler) GetTrending(c *gin.Context) {
	mediaType := c.Query("type")
	timeRange := c.Query("range")
	limitStr := c.Query("limit")

	limit := 20 // Default limit as requested
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	results, err := h.analyticsService.GetTopTrending(c.Request.Context(), mediaType, timeRange, limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "GET_TRENDING_FAILED", i18nkeys.AnalyticsGetTrendingFailed, err.Error())
		return
	}

	// Convert map to typed response for consistency
	var resp []MediaSeriesResponse
	// Use JSON round-trip for simplicity since structure matches
	bytes, err := json.Marshal(results)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "SERIALIZATION_FAILED", i18nkeys.AnalyticsSerializationFailed, nil)
		return
	}
	if err := json.Unmarshal(bytes, &resp); err != nil {
		response.Error(c, http.StatusInternalServerError, "MAPPING_FAILED", i18nkeys.AnalyticsMappingFailed, nil)
		return
	}

	response.Success(c, http.StatusOK, i18nkeys.AnalyticsGetTrendingSuccess, resp, nil)
}
