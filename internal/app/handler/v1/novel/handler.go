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
	novelService   *service.NovelService
	volumeService  *service.VolumeService
	chapterService *service.ChapterService
}

func NewHandler(
	novelService *service.NovelService,
	volumeService *service.VolumeService,
	chapterService *service.ChapterService,
) *Handler {
	return &Handler{
		novelService:   novelService,
		volumeService:  volumeService,
		chapterService: chapterService,
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
		req.IsOneshot,
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

	// Initialize slices to ensure JSON array [] instead of null
	resp.GenreIDs = make([]string, 0)
	resp.AuthorIDs = make([]string, 0)
	resp.ArtistIDs = make([]string, 0)
	resp.Genres = make([]GenreInfo, 0)
	resp.Authors = make([]OwnerInfo, 0)
	resp.Artists = make([]OwnerInfo, 0)

	// Load relations for edit screen
	genreIDs, err := h.novelService.GetNovelGenres(c.Request.Context(), novel.ID)
	if err == nil {
		resp.GenreIDs = make([]string, len(genreIDs))
		for i, id := range genreIDs {
			resp.GenreIDs[i] = id.String()
		}
	}

	authorIDs, err := h.novelService.GetNovelAuthors(c.Request.Context(), novel.ID)
	if err == nil {
		resp.AuthorIDs = make([]string, len(authorIDs))
		for i, id := range authorIDs {
			resp.AuthorIDs[i] = id.String()
		}
	}

	artistIDs, err := h.novelService.GetNovelArtists(c.Request.Context(), novel.ID)
	if err == nil {
		resp.ArtistIDs = make([]string, len(artistIDs))
		for i, id := range artistIDs {
			resp.ArtistIDs[i] = id.String()
		}
	}

	// Load detailed relations for UI display
	genres, err := h.novelService.GetNovelGenresDetails(c.Request.Context(), novel.ID)
	if err == nil {
		resp.Genres = make([]GenreInfo, len(genres))
		for i, g := range genres {
			resp.Genres[i] = GenreInfo{
				ID:   g.ID.String(),
				Name: g.Name,
			}
		}
	}

	authors, err := h.novelService.GetNovelAuthorsDetails(c.Request.Context(), novel.ID)
	if err == nil {
		fmt.Printf("GetNovelAuthorsDetails found %d authors\n", len(authors))
		resp.Authors = make([]OwnerInfo, len(authors))
		for i, na := range authors {
			if na.Author != nil {
				resp.Authors[i] = OwnerInfo{
					ID:          na.Author.ID.String(),
					DisplayName: na.Author.Name,
				}
			}
		}
	} else {
		fmt.Printf("GetNovelAuthorsDetails Error: %v\n", err)
	}

	artists, err := h.novelService.GetNovelArtistsDetails(c.Request.Context(), novel.ID)
	if err == nil {
		fmt.Printf("GetNovelArtistsDetails found %d artists\n", len(artists))
		resp.Artists = make([]OwnerInfo, len(artists))
		for i, na := range artists {
			if na.Artist != nil {
				resp.Artists[i] = OwnerInfo{
					ID:          na.Artist.ID.String(),
					DisplayName: na.Artist.Name,
				}
			}
		}
	} else {
		fmt.Printf("GetNovelArtistsDetails Error: %v\n", err)
	}


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
		fmt.Printf("❌ [Handler] ListNovels Error: %v\n", err)
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
		Synopsis:         json.RawMessage("{}"),
		CoverImageURL:    novel.CoverImageURL,
		ThumbnailURL:     novel.ThumbnailURL,
		Status:           string(novel.Status),
		IsOneshot:        novel.IsOneshot,
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

	// Return synopsis as JSON object
	if len(novel.Synopsis) > 0 && string(novel.Synopsis) != "null" {
		resp.Synopsis = novel.Synopsis
	} else {
		resp.Synopsis = json.RawMessage("{}")
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
		Synopsis:    json.RawMessage("[]"), // Empty array if synopsis not in list view
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

// GetNovelFull lấy thông tin đầy đủ novel (public API)
// @Summary Get full novel details with volumes and chapters
// @Tags Novels
// @Produce json
// @Param slug path string true "Novel Slug"
// @Success 200 {object} response.StandardResponse{data=NovelFullResponse}
// @Failure 404 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/v1/novels/{id}/full [get]
func (h *Handler) GetNovelFull(c *gin.Context) {
	idOrSlug := c.Param("id")

	var novel *domain.Novel
	var err error

	// Try to parse as UUID first
	id, parseErr := uuid.FromString(idOrSlug)
	if parseErr == nil {
		novel, err = h.novelService.GetNovelByID(c.Request.Context(), id)
	} else {
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

	// Build base response
	resp := NovelFullResponse{
		NovelDetailResponse: mapToNovelDetailResponse(novel),
	}

	// Initialize slices
	resp.GenreIDs = make([]string, 0)
	resp.AuthorIDs = make([]string, 0)
	resp.ArtistIDs = make([]string, 0)
	resp.Genres = make([]GenreInfo, 0)
	resp.Authors = make([]OwnerInfo, 0)
	resp.Artists = make([]OwnerInfo, 0)
	resp.Volumes = make([]VolumeInfoResponse, 0)
	resp.Chapters = make([]ChapterSummaryResponse, 0)

	// Set owner info
	ownerDisplayName := "Unknown Owner"
	if novel.OwnerDisplayName != nil && *novel.OwnerDisplayName != "" {
		ownerDisplayName = *novel.OwnerDisplayName
	}
	ownerUsername := "unknown"
	if novel.OwnerUsername != nil && *novel.OwnerUsername != "" {
		ownerUsername = *novel.OwnerUsername
	}
	resp.Owner = OwnerInfo{
		ID:          novel.OwnerID.String(),
		DisplayName: ownerDisplayName,
		Username:    ownerUsername,
		AvatarURL:   novel.OwnerAvatarURL,
	}

	// Load genres
	genres, err := h.novelService.GetNovelGenresDetails(c.Request.Context(), novel.ID)
	if err == nil {
		resp.Genres = make([]GenreInfo, len(genres))
		for i, g := range genres {
			resp.Genres[i] = GenreInfo{ID: g.ID.String(), Name: g.Name}
		}
	}

	// Load authors
	authors, err := h.novelService.GetNovelAuthorsDetails(c.Request.Context(), novel.ID)
	if err == nil {
		resp.Authors = make([]OwnerInfo, len(authors))
		for i, na := range authors {
			if na.Author != nil {
				resp.Authors[i] = OwnerInfo{ID: na.Author.ID.String(), DisplayName: na.Author.Name}
			}
		}
	}

	// Load artists
	artists, err := h.novelService.GetNovelArtistsDetails(c.Request.Context(), novel.ID)
	if err == nil {
		resp.Artists = make([]OwnerInfo, len(artists))
		for i, na := range artists {
			if na.Artist != nil {
				resp.Artists[i] = OwnerInfo{ID: na.Artist.ID.String(), DisplayName: na.Artist.Name}
			}
		}
	}

	// Load published volumes
	volumes, err := h.volumeService.GetVolumesByNovelID(c.Request.Context(), novel.ID, true)
	if err == nil {
		for _, vol := range volumes {
			volResp := VolumeInfoResponse{
				ID:            vol.ID.String(),
				VolumeNumber:  vol.VolumeNumber,
				Title:         vol.Title,
				Slug:          vol.Slug,
				CoverImageURL: vol.CoverImageURL,
				DisplayOrder:  vol.DisplayOrder,
				IsPublished:   vol.IsPublished,
				Chapters:      make([]ChapterSummaryResponse, 0),
			}
			if vol.PublishedAt != nil {
				publishedAt := vol.PublishedAt.Format(timeutil.ISO8601Layout)
				volResp.PublishedAt = &publishedAt
			}

			// Load published chapters for this volume
			chapters, chErr := h.chapterService.GetChaptersByVolumeID(c.Request.Context(), vol.ID, true)
			if chErr == nil {
				for _, ch := range chapters {
					chResp := mapToChapterSummary(ch)
					volResp.Chapters = append(volResp.Chapters, chResp)
				}
			}

			resp.Volumes = append(resp.Volumes, volResp)
		}
	}

	// Load published chapters without volume (volume_id = null)
	filter := domain.ChapterFilter{
		PublishedOnly: true,
		SortBy:        "chapter_number",
		SortOrder:     "asc",
	}
	allChapters, err := h.chapterService.GetChaptersByNovelID(c.Request.Context(), novel.ID, filter)
	if err == nil {
		for _, ch := range allChapters {
			if ch.VolumeID == nil {
				resp.Chapters = append(resp.Chapters, mapToChapterSummary(ch))
			}
		}
	}

	response.Success(c, http.StatusOK, "novel.get_full_success", resp, nil)
}

// mapToChapterSummary maps domain.Chapter to ChapterSummaryResponse
func mapToChapterSummary(ch *domain.Chapter) ChapterSummaryResponse {
	resp := ChapterSummaryResponse{
		ID:            ch.ID.String(),
		ChapterNumber: ch.ChapterNumber,
		Title:         ch.Title,
		Slug:          ch.Slug,
		DisplayOrder:  ch.DisplayOrder,
		Status:        string(ch.Status),
	}
	if ch.VolumeID != nil {
		volIDStr := ch.VolumeID.String()
		resp.VolumeID = &volIDStr
	}
	if ch.PublishedAt != nil {
		publishedAt := ch.PublishedAt.Format(timeutil.ISO8601Layout)
		resp.PublishedAt = &publishedAt
	}
	return resp
}
