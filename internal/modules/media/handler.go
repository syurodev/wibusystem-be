// ============================================================================
// Media Handler
// ============================================================================
//
// Handler này cung cấp API endpoints cho Media module.
// Endpoint chính:
//   - GET /media/top: Lấy top media (anime, manga, novel) theo period (week/month/year).
//     Hiện tại chỉ hỗ trợ novel.
//
// ============================================================================

package media

import (
	"net/http"
	"strconv"

	res "system/internal/dto/media"
	analytics_module "system/internal/modules/analytics"
	"system/pkg/util/response"

	"github.com/gin-gonic/gin"
)

// Handler handles HTTP requests for media module
type Handler struct {
	analyticsSvc analytics_module.AnalyticsService
}

// NewHandler creates a new media handler
func NewHandler(analyticsSvc analytics_module.AnalyticsService) *Handler {
	return &Handler{
		analyticsSvc: analyticsSvc,
	}
}

// TopMediaResponse là response cho GET /media/top
type TopMediaResponse struct {
	Anime []res.MediaSeriesResponse `json:"anime"`
	Manga []res.MediaSeriesResponse `json:"manga"`
	Novel []res.MediaSeriesResponse `json:"novel"`
}

// GetTop retrieves top media for each type (anime, manga, novel)
// @Summary Get top media by views
// @Description Get top media for each type based on views in a calendar period. Returns empty array for types without data.
// @Tags Media
// @Accept json
// @Produce json
// @Param period query string false "Time period (day, week, month, year)" default(week)
// @Param limit query int false "Number of items per type" default(1)
// @Param offset query int false "Period offset (0=current, 1=previous)" default(0)
// @Success 200 {object} TopMediaResponse
// @Router /media/top [get]
func (h *Handler) GetTop(c *gin.Context) {
	ctx := c.Request.Context()

	period := c.DefaultQuery("period", "week")

	// Validate period
	validPeriods := map[string]bool{"day": true, "week": true, "month": true, "year": true}
	if !validPeriods[period] {
		period = "week"
	}

	limit := 1
	if l, err := strconv.Atoi(c.DefaultQuery("limit", "1")); err == nil && l > 0 && l <= 10 {
		limit = l
	}

	offset := 0
	if o, err := strconv.Atoi(c.DefaultQuery("offset", "0")); err == nil && o >= 0 && o <= 52 {
		offset = o
	}

	resp := TopMediaResponse{
		Anime: []res.MediaSeriesResponse{},
		Manga: []res.MediaSeriesResponse{},
		Novel: []res.MediaSeriesResponse{},
	}

	// Get top novels by views for the period
	novelRanks, err := h.analyticsSvc.GetTopMediaWithRankComparison(ctx, period, "novel", offset, limit)
	if err == nil && len(novelRanks) > 0 {
		for _, rank := range novelRanks {
			if mapped := mapMediaRankToResponse(&rank); mapped != nil {
				resp.Novel = append(resp.Novel, *mapped)
			}
		}
	}

	// TODO: When manga/anime are implemented, add similar calls here

	response.Success(c, http.StatusOK, I18nMediaGetTopSuccess, resp, nil)
}
