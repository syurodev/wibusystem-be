package media

import (
	"net/http"
	"strconv"

	"system/pkg/util/response"

	"github.com/gin-gonic/gin"
)

// Handler handles media API endpoints
// Cung cấp APIs cho tất cả media types (anime, manga, novel) combined
type Handler struct {
	getTrendingUC GetTrendingUseCase
	getHomeDataUC GetHomeDataUseCase
}

// NewHandler creates a new media Handler instance
func NewHandler(getTrendingUC GetTrendingUseCase, getHomeDataUC GetHomeDataUseCase) *Handler {
	return &Handler{
		getTrendingUC: getTrendingUC,
		getHomeDataUC: getHomeDataUC,
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
// @Success 200 {array} mediadto.MediaSeriesResponse
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
	
	includeRankChange := c.Query("include_rank_change") == "true"

	// Call UseCase
	results, err := h.getTrendingUC.Execute(c.Request.Context(), GetTrendingInput{
		MediaType:         mediaType,
		Range:             timeRange,
		Limit:             limit,
		IncludeRankChange: includeRankChange,
	})
	if err != nil {
		// Since logic is moved to usecase, we can't distinguish serialization error easily unless we wrap errors,
		// but generic error message is fine for now.
		response.Error(c, http.StatusInternalServerError, "GET_TRENDING_FAILED", I18nAnalyticsGetTrendingFailed, err.Error())
		return
	}

	response.Success(c, http.StatusOK, I18nAnalyticsGetTrendingSuccess, results, nil)
}

// GetHomeData returns aggregated data for the home page
// @Summary Get home page data
// @Description Get aggregated data for home page (hero, trending, creators, genres) across all media types
// @Tags Media
// @Produce json
// @Success 200 {object} response.StandardResponse{data=mediadto.HomeData}
// @Failure 500 {object} response.StandardResponse
// @Router /api/v1/media/home [get]
func (h *Handler) GetHomeData(c *gin.Context) {
	homeData, err := h.getHomeDataUC.Execute(c.Request.Context(), GetHomeDataInput{})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "GET_HOME_DATA_FAILED", I18nMediaGetHomeFailed, nil)
		return
	}

	response.Success(c, http.StatusOK, I18nMediaGetHomeSuccess, homeData, nil)
}


