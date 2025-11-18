package novel

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"

	"system/internal/domain"
	"system/internal/pkg/service"
	pkgerrors "system/pkg/errors"
	"system/pkg/util/response"
)

type Handler struct {
	novelService *service.NovelService
}

func NewHandler(novelService *service.NovelService) *Handler {
	return &Handler{
		novelService: novelService,
	}
}

// CreateNovel tạo novel mới
// @Summary Create a new novel
// @Tags Novels
// @Accept json
// @Produce json
// @Param request body CreateNovelRequest true "Create Novel Request"
// @Success 201 {object} response.StandardResponse{data=NovelDetailResponse}
// @Failure 400 {object} response.StandardResponse
// @Failure 409 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/v1/novels [post]
func (h *Handler) CreateNovel(c *gin.Context) {
	var req CreateNovelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_FAILED", "validation.failed", err.Error())
		return
	}

	// Create novel
	novel, err := h.novelService.CreateNovel(
		c.Request.Context(),
		req.Title,
		req.Synopsis,
		req.CoverImageURL,
		req.ThumbnailURL,
		&req.Status,
		req.OriginalLanguage,
		req.OriginalTitle,
		req.Metadata,
	)
	if err != nil {
		if errors.Is(err, pkgerrors.ErrInvalidInput) {
			response.Error(c, http.StatusBadRequest, "INVALID_INPUT", "novel.invalid_input", nil)
			return
		}
		if errors.Is(err, pkgerrors.ErrSlugAlreadyExists) {
			response.Error(c, http.StatusConflict, "SLUG_EXISTS", "novel.slug_already_exists", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "CREATE_FAILED", "novel.create_failed", nil)
		return
	}

	// Map to response
	resp := mapToNovelDetailResponse(novel)

	response.Success(c, http.StatusCreated, "novel.created_success", resp, nil)
}

// UpdateNovel cập nhật novel
// @Summary Update a novel
// @Tags Novels
// @Accept json
// @Produce json
// @Param id path string true "Novel ID"
// @Param request body UpdateNovelRequest true "Update Novel Request"
// @Success 200 {object} response.StandardResponse{data=NovelDetailResponse}
// @Failure 400 {object} response.StandardResponse
// @Failure 404 {object} response.StandardResponse
// @Failure 409 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/v1/novels/{id} [put]
func (h *Handler) UpdateNovel(c *gin.Context) {
	// Get novel ID from path
	idStr := c.Param("id")
	id, err := uuid.FromString(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "novel.invalid_id", nil)
		return
	}

	var req UpdateNovelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_FAILED", "validation.failed", err.Error())
		return
	}

	// Update novel
	novel, err := h.novelService.UpdateNovel(
		c.Request.Context(),
		id,
		req.Title,
		req.Synopsis,
		req.CoverImageURL,
		req.ThumbnailURL,
		&req.Status,
		req.OriginalLanguage,
		req.OriginalTitle,
		req.Metadata,
	)
	if err != nil {
		if errors.Is(err, pkgerrors.ErrNovelNotFound) {
			response.Error(c, http.StatusNotFound, "NOVEL_NOT_FOUND", "novel.not_found", nil)
			return
		}
		if errors.Is(err, pkgerrors.ErrInvalidInput) {
			response.Error(c, http.StatusBadRequest, "INVALID_INPUT", "novel.invalid_input", nil)
			return
		}
		if errors.Is(err, pkgerrors.ErrSlugAlreadyExists) {
			response.Error(c, http.StatusConflict, "SLUG_EXISTS", "novel.slug_already_exists", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "UPDATE_FAILED", "novel.update_failed", nil)
		return
	}

	// Map to response
	resp := mapToNovelDetailResponse(novel)

	response.Success(c, http.StatusOK, "novel.updated_success", resp, nil)
}

// DeleteNovel xóa novel
// @Summary Delete a novel
// @Tags Novels
// @Produce json
// @Param id path string true "Novel ID"
// @Success 200 {object} response.StandardResponse
// @Failure 400 {object} response.StandardResponse
// @Failure 404 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/v1/novels/{id} [delete]
func (h *Handler) DeleteNovel(c *gin.Context) {
	// Get novel ID from path
	idStr := c.Param("id")
	id, err := uuid.FromString(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "novel.invalid_id", nil)
		return
	}

	// Delete novel
	err = h.novelService.DeleteNovel(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, pkgerrors.ErrNovelNotFound) {
			response.Error(c, http.StatusNotFound, "NOVEL_NOT_FOUND", "novel.not_found", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "DELETE_FAILED", "novel.delete_failed", nil)
		return
	}

	response.Success(c, http.StatusOK, "novel.deleted_success", nil, nil)
}

// GetNovel lấy thông tin chi tiết novel
// @Summary Get novel details
// @Tags Novels
// @Produce json
// @Param id path string true "Novel ID or Slug"
// @Success 200 {object} response.StandardResponse{data=NovelDetailResponse}
// @Failure 400 {object} response.StandardResponse
// @Failure 404 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/v1/novels/{id} [get]
func (h *Handler) GetNovel(c *gin.Context) {
	// Get novel ID or slug from path
	idOrSlug := c.Param("id")

	var novel *domain.Novel
	var err error

	// Try to parse as UUID first
	id, parseErr := uuid.FromString(idOrSlug)
	if parseErr == nil {
		// It's a valid UUID, get by ID
		novel, err = h.novelService.GetNovelByID(c.Request.Context(), id)
	} else {
		// It's not a UUID, treat as slug
		novel, err = h.novelService.GetNovelBySlug(c.Request.Context(), idOrSlug)
	}

	if err != nil {
		if errors.Is(err, pkgerrors.ErrNovelNotFound) || errors.Is(err, pgx.ErrNoRows) {
			response.Error(c, http.StatusNotFound, "NOVEL_NOT_FOUND", "novel.not_found", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "GET_FAILED", "novel.get_failed", nil)
		return
	}

	// Map to response
	resp := mapToNovelDetailResponse(novel)

	response.Success(c, http.StatusOK, "novel.get_success", resp, nil)
}

// ListNovels lấy danh sách novels
// @Summary List novels with pagination, search and filters
// @Tags Novels
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param search query string false "Search in title and synopsis"
// @Param status query string false "Filter by status (draft, ongoing, completed, hiatus, dropped)"
// @Param original_language query string false "Filter by original language (ISO 639-1)"
// @Param sort_by query string false "Sort by field (created_at, rating, views, last_chapter)"
// @Param sort_order query string false "Sort order (asc, desc)"
// @Success 200 {object} response.StandardResponse{data=[]NovelResponse}
// @Failure 400 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/v1/novels [get]
func (h *Handler) ListNovels(c *gin.Context) {
	var req ListNovelsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_FAILED", "validation.failed", err.Error())
		return
	}

	// Set defaults
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 {
		req.Limit = 20
	}

	// Get novels with pagination, search and filters
	novels, totalCount, err := h.novelService.ListNovels(
		c.Request.Context(),
		req.Page,
		req.Limit,
		req.Search,
		req.Status,
		req.OriginalLanguage,
		req.SortBy,
		req.SortOrder,
	)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "LIST_FAILED", "novel.list_failed", nil)
		return
	}

	// Map to response format
	novelResponses := make([]NovelResponse, len(novels))
	for i, novel := range novels {
		novelResponses[i] = mapToNovelResponse(novel)
	}

	// Calculate pagination meta
	totalPages := (totalCount + req.Limit - 1) / req.Limit
	meta := &response.PaginationMeta{
		Page:       req.Page,
		Limit:      req.Limit,
		TotalItems: totalCount,
		TotalPages: totalPages,
	}

	response.Success(c, http.StatusOK, "novel.list_success", novelResponses, meta)
}

// IncrementViewCount tăng view count của novel
// @Summary Increment novel view count
// @Tags Novels
// @Produce json
// @Param id path string true "Novel ID"
// @Success 200 {object} response.StandardResponse
// @Failure 400 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/v1/novels/{id}/view [post]
func (h *Handler) IncrementViewCount(c *gin.Context) {
	// Get novel ID from path
	idStr := c.Param("id")
	id, err := uuid.FromString(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "novel.invalid_id", nil)
		return
	}

	// Increment view count
	err = h.novelService.IncrementViewCount(c.Request.Context(), id)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INCREMENT_FAILED", "novel.increment_view_failed", nil)
		return
	}

	response.Success(c, http.StatusOK, "novel.view_incremented", nil, nil)
}

// Helper function to map domain model to detail response
func mapToNovelDetailResponse(novel *domain.Novel) NovelDetailResponse {
	resp := NovelDetailResponse{
		ID:               novel.ID.String(),
		Title:            novel.Title,
		Slug:             novel.Slug,
		Synopsis:         "",
		CoverImageURL:    novel.CoverImageURL,
		ThumbnailURL:     novel.ThumbnailURL,
		Status:           string(novel.Status),
		OriginalLanguage: novel.OriginalLanguage,
		OriginalTitle:    novel.OriginalTitle,
		TotalVolumes:     novel.TotalVolumes,
		TotalChapters:    novel.TotalChapters,
		TotalWords:       novel.TotalWords,
		ViewCount:        novel.ViewCount,
		FavoriteCount:    novel.FavoriteCount,
		RatingAverage:    novel.RatingAverage,
		RatingCount:      novel.RatingCount,
		CreatedAt:        novel.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:        novel.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	// Extract synopsis from JSON
	if len(novel.Synopsis) > 0 {
		var synopsisData map[string]interface{}
		if err := json.Unmarshal(novel.Synopsis, &synopsisData); err == nil {
			if content, ok := synopsisData["content"].(string); ok && content != "" {
				resp.Synopsis = content
			}
		}
	}

	// Convert metadata to string
	if len(novel.Metadata) > 0 {
		metadataStr := string(novel.Metadata)
		resp.Metadata = &metadataStr
	}

	// Format optional dates
	if novel.FirstPublishedAt != nil {
		publishedAt := novel.FirstPublishedAt.Format("2006-01-02T15:04:05Z07:00")
		resp.FirstPublishedAt = &publishedAt
	}

	if novel.LastChapterAt != nil {
		lastChapter := novel.LastChapterAt.Format("2006-01-02T15:04:05Z07:00")
		resp.LastChapterAt = &lastChapter
	}

	if novel.CompletedAt != nil {
		completedAt := novel.CompletedAt.Format("2006-01-02T15:04:05Z07:00")
		resp.CompletedAt = &completedAt
	}

	return resp
}

// Helper function to map domain model to list response
func mapToNovelResponse(novel *domain.Novel) NovelResponse {
	resp := NovelResponse{
		ID:            novel.ID.String(),
		Title:         novel.Title,
		Slug:          novel.Slug,
		CoverImageURL: novel.CoverImageURL,
		ThumbnailURL:  novel.ThumbnailURL,
		Status:        string(novel.Status),
		TotalChapters: novel.TotalChapters,
		ViewCount:     novel.ViewCount,
		FavoriteCount: novel.FavoriteCount,
		RatingAverage: novel.RatingAverage,
		RatingCount:   novel.RatingCount,
		CreatedAt:     novel.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:     novel.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	// Format last chapter date
	if novel.LastChapterAt != nil {
		lastChapter := novel.LastChapterAt.Format("2006-01-02T15:04:05Z07:00")
		resp.LastChapterAt = &lastChapter
	}

	return resp
}
