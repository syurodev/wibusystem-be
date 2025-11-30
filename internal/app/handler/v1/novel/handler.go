package novel

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"

	"system/internal/app/middleware"
	"system/internal/domain"
	"system/internal/pkg/service"
	pkgerrors "system/pkg/errors"
	"system/pkg/util/response"
	"system/pkg/util/timeutil"
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
	userIDStr, exists := middleware.GetUserID(c)
	if !exists {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "auth.unauthorized", nil)
		return
	}

	userID, err := uuid.FromString(userIDStr)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "INVALID_USER_ID", "auth.invalid_user_id", nil)
		return
	}

	var req CreateNovelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_FAILED", "validation.failed", err.Error())
		return
	}

	// Create novel
	// Parse IDs
	ownerID, err := uuid.FromString(req.OwnerID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_OWNER_ID", "novel.invalid_owner_id", nil)
		return
	}

	var genreIDs []uuid.UUID
	for _, idStr := range req.GenreIDs {
		id, err := uuid.FromString(idStr)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "INVALID_GENRE_ID", "novel.invalid_genre_id", nil)
			return
		}
		genreIDs = append(genreIDs, id)
	}

	var authorIDs []uuid.UUID
	for _, idStr := range req.AuthorIDs {
		id, err := uuid.FromString(idStr)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "INVALID_AUTHOR_ID", "novel.invalid_author_id", nil)
			return
		}
		authorIDs = append(authorIDs, id)
	}

	var artistIDs []uuid.UUID
	for _, idStr := range req.ArtistIDs {
		id, err := uuid.FromString(idStr)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "INVALID_ARTIST_ID", "novel.invalid_artist_id", nil)
			return
		}
		artistIDs = append(artistIDs, id)
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
		req.IsOneshot,
		ownerID,
		req.OwnerType,
		userID,
		genreIDs,
		authorIDs,
		artistIDs,
	)
	if err != nil {
		fmt.Printf("❌ [Handler] CreateNovel Error: %v\n", err)
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

	// Parse owner UUID nếu có
	var ownerID *uuid.UUID
	if req.Owner != "" {
		id, err := uuid.FromString(req.Owner)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "INVALID_OWNER_ID", "novel.invalid_owner_id", nil)
			return
		}
		ownerID = &id
	}

	// Parse genre IDs
	var genreIDs []uuid.UUID
	for _, idStr := range req.GenreIDs {
		id, err := uuid.FromString(idStr)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "INVALID_GENRE_ID", "novel.invalid_genre_id", nil)
			return
		}
		genreIDs = append(genreIDs, id)
	}

	// Get novels with pagination, search and filters
	novels, totalCount, err := h.novelService.ListNovels(
		c.Request.Context(),
		req.Page,
		req.Limit,
		ownerID,
		req.KeySearch,
		genreIDs,
		req.Statuses,
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
		CreatedAt:        novel.CreatedAt.Format(timeutil.ISO8601Layout),
		UpdatedAt:        novel.UpdatedAt.Format(timeutil.ISO8601Layout),
	}

	// Extract synopsis from JSON
	if len(novel.Synopsis) > 0 {
		var synopsisData map[string]any
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
		publishedAt := novel.FirstPublishedAt.Format(timeutil.ISO8601Layout)
		resp.FirstPublishedAt = &publishedAt
	}

	if novel.LastChapterAt != nil {
		lastChapter := novel.LastChapterAt.Format(timeutil.ISO8601Layout)
		resp.LastChapterAt = &lastChapter
	}

	if novel.CompletedAt != nil {
		completedAt := novel.CompletedAt.Format(timeutil.ISO8601Layout)
		resp.CompletedAt = &completedAt
	}

	return resp
}

// Helper function to map domain model to list response
func mapToNovelResponse(novel *domain.Novel) NovelResponse {
	// Map owner info (đã được load từ JOIN query)
	ownerDisplayName := "Unknown Owner"
	if novel.OwnerDisplayName != nil && *novel.OwnerDisplayName != "" {
		ownerDisplayName = *novel.OwnerDisplayName
	}

	ownerUsername := "unknown"
	if novel.OwnerUsername != nil && *novel.OwnerUsername != "" {
		ownerUsername = *novel.OwnerUsername
	}

	owner := OwnerInfo{
		ID:          novel.OwnerID.String(),
		DisplayName: ownerDisplayName,
		Username:    ownerUsername,
		AvatarURL:   novel.OwnerAvatarURL,
	}

	// Map genres (đã được load từ repository)
	genres := make([]GenreInfo, 0)
	if novel.Genres != nil && len(novel.Genres) > 0 {
		for _, genre := range novel.Genres {
			genres = append(genres, GenreInfo{
				ID:   genre.ID.String(),
				Name: genre.Name,
			})
		}
	}

	// Map latest chapter nếu có
	var latestChapter *LatestChapterInfo
	if novel.LastChapterAt != nil {
		// TODO: Load actual latest chapter data from DB
		// For now, just include timestamp
		latestChapter = &LatestChapterInfo{
			ID:          "",
			Title:       "Latest Chapter",
			PublishedAt: novel.LastChapterAt.Format(timeutil.ISO8601Layout),
		}
	}

	// Convert rating from 0-5 to 0-10 scale
	rating := novel.RatingAverage * 2

	resp := NovelResponse{
		// Required fields
		ID:               novel.ID.String(),
		Title:            novel.Title,
		OriginalTitle:    novel.OriginalTitle,
		Slug:             novel.Slug,
		OriginalLanguage: novel.OriginalLanguage,

		// Content fields
		Description: make([]map[string]any, 0), // Empty for list view
		CoverURL:    novel.CoverImageURL,

		// Type và status
		Type:   "novel", // Hard-coded for novel
		Status: string(novel.Status),

		// Relations
		Genres: genres,
		Owner:  owner,

		// Stats
		Rating:    rating,
		Views:     novel.ViewCount,
		Favorites: novel.FavoriteCount,

		// Optional fields
		LatestChapter: latestChapter,

		// Timestamps
		CreatedAt: novel.CreatedAt.Format(timeutil.ISO8601Layout),
		UpdatedAt: novel.UpdatedAt.Format(timeutil.ISO8601Layout),
	}

	return resp
}
