package creator

import (
	"net/http"
	"strconv"
	"system/internal/domain"
	"system/pkg/util/response"
	"system/pkg/util/timeutil"
	"time"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	listCreatorsUC ListCreatorsUseCase
}

func NewHandler(listCreatorsUC ListCreatorsUseCase) *Handler {
	return &Handler{
		listCreatorsUC: listCreatorsUC,
	}
}

// CreatorResponse represents the API response for a single creator
type CreatorResponse struct {
	ID                   string  `json:"id"`
	DisplayName          string  `json:"display_name"`
	Username             string  `json:"username"`
	AvatarURL            string  `json:"avatar_url"`
	CreatedAt            string  `json:"created_at"`
	UpdatedAt            string  `json:"updated_at"`
	FollowerCount        int     `json:"follower_count"`
	WorksCount           int     `json:"works_count"`
	TotalViews           int64   `json:"total_views"`
	IsVerified           bool    `json:"is_verified"`
	Bio                  []any   `json:"bio,omitempty"`
	PopularWorkID        *string `json:"popular_work_id,omitempty"`
	PopularWorkTitle     *string `json:"popular_work_title,omitempty"`
	PopularWorkCoverURL  *string `json:"popular_work_cover_url,omitempty"`
}

// PaginationResponse represents pagination metadata
type PaginationResponse struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// ListCreatorsResponse represents the full list response
type ListCreatorsResponse struct {
	Creators   []CreatorResponse  `json:"creators"`
	Pagination PaginationResponse `json:"pagination"`
}

// ListCreators handles GET /api/v1/creators
// @Summary List creators with filters
// @Description Get paginated list of creators with optional filters
// @Tags Creators
// @Produce json
// @Param role query string false "Filter by role (default: CREATOR)"
// @Param view_period query string false "View period: day, week, month, year, all"
// @Param search query string false "Search by display_name or username"
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Items per page (default: 20, max: 100)"
// @Param sort_by query string false "Sort by: last_content_updated_at, total_views, follower_count"
// @Param sort_order query string false "Sort order: asc, desc (default: desc)"
// @Param created_from query string false "Created from date (YYYY-MM-DD)"
// @Param created_to query string false "Created to date (YYYY-MM-DD)"
// @Success 200 {object} response.StandardResponse{data=ListCreatorsResponse}
// @Failure 400 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/v1/creators [get]
func (h *Handler) ListCreators(c *gin.Context) {
	// Parse query params
	filter := domain.CreatorListFilter{
		Page:      1,
		Limit:     20,
		SortBy:    "last_content_updated_at",
		SortOrder: "desc",
	}

	// Role filter (default to CREATOR)
	if role := c.Query("role"); role != "" {
		filter.Role = &role
	} else {
		defaultRole := "CREATOR"
		filter.Role = &defaultRole
	}

	// View period
	if viewPeriod := c.Query("view_period"); viewPeriod != "" {
		filter.ViewPeriod = &viewPeriod
	}

	// Search
	if search := c.Query("search"); search != "" {
		filter.Search = &search
	}

	// Pagination
	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			filter.Page = page
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			if limit > 100 {
				limit = 100
			}
			filter.Limit = limit
		}
	}

	// Sort
	if sortBy := c.Query("sort_by"); sortBy != "" {
		filter.SortBy = sortBy
	}
	if sortOrder := c.Query("sort_order"); sortOrder != "" {
		filter.SortOrder = sortOrder
	}

	// Date filters
	if createdFrom := c.Query("created_from"); createdFrom != "" {
		if t, err := time.Parse(timeutil.DateOnlyLayout, createdFrom); err == nil {
			filter.CreatedFrom = &t
		}
	}
	if createdTo := c.Query("created_to"); createdTo != "" {
		if t, err := time.Parse(timeutil.DateOnlyLayout, createdTo); err == nil {
			endOfDay := t.Add(24*time.Hour - time.Second)
			filter.CreatedTo = &endOfDay
		}
	}

	// Call UseCase
	result, err := h.listCreatorsUC.Execute(c.Request.Context(), ListCreatorsInput{Filter: filter})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "LIST_CREATORS_ERROR", I18nCreatorListFailed, nil)
		return
	}

	// Map to response
	creators := make([]CreatorResponse, len(result.Creators))
	for i, creator := range result.Creators {
		creators[i] = CreatorResponse{
			ID:                  creator.User.ID.String(),
			DisplayName:         getValue(creator.User.DisplayName),
			Username:            getValue(creator.User.Username),
			AvatarURL:           getValue(creator.User.AvatarURL),
			CreatedAt:           creator.User.CreatedAt.Format(time.RFC3339),
			UpdatedAt:           creator.User.UpdatedAt.Format(time.RFC3339),
			FollowerCount:       creator.FollowerCount,
			WorksCount:          creator.WorksCount,
			TotalViews:          creator.TotalViews,
			IsVerified:          creator.User.IsVerified,
			Bio:                 creator.User.Bio,
			PopularWorkID:       creator.PopularWorkID,
			PopularWorkTitle:    creator.PopularWorkTitle,
			PopularWorkCoverURL: creator.PopularWorkCoverURL,
		}
	}

	// Use standard response format with Meta for pagination
	meta := &response.PaginationMeta{
		Page:       result.Page,
		Limit:      result.Limit,
		TotalItems: int(result.Total),
		TotalPages: result.TotalPages,
	}

	response.Success(c, http.StatusOK, I18nCreatorListSuccess, creators, meta)
}

func getValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

