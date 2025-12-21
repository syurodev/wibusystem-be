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
	noveldto "system/internal/dto/novel"
	"system/internal/modules/novel_chapter"
	novel_volume "system/internal/modules/novel_volume"
	pkgerrors "system/pkg/errors"
	"system/pkg/util/response"
	"system/pkg/util/timeutil"
)

// Handler handles novel-related HTTP requests
type Handler struct {
	novelService    NovelService
	createNovelUC   CreateNovelUseCase
	updateNovelUC   UpdateNovelUseCase
	deleteNovelUC   DeleteNovelUseCase
	getNovelUC      GetNovelUseCase
	listNovelsUC    ListNovelsUseCase
	viewCountUC     IncrementViewCountUseCase
	getNovelFullUC  GetNovelFullUseCase
	volumeService   novel_volume.VolumeService
	chapterService  novel_chapter.ChapterService
}

// NewHandler creates a new novel Handler instance
func NewHandler(
	novelService NovelService,
	createNovelUC CreateNovelUseCase,
	updateNovelUC UpdateNovelUseCase,
	deleteNovelUC DeleteNovelUseCase,
	getNovelUC GetNovelUseCase,
	listNovelsUC ListNovelsUseCase,
	viewCountUC IncrementViewCountUseCase,
	getNovelFullUC GetNovelFullUseCase,
	volumeService novel_volume.VolumeService,
	chapterService novel_chapter.ChapterService,
) *Handler {
	return &Handler{
		novelService:    novelService,
		createNovelUC:   createNovelUC,
		updateNovelUC:   updateNovelUC,
		deleteNovelUC:   deleteNovelUC,
		getNovelUC:      getNovelUC,
		listNovelsUC:    listNovelsUC,
		viewCountUC:     viewCountUC,
		getNovelFullUC:  getNovelFullUC,
		volumeService:   volumeService,
		chapterService:  chapterService,
	}
}

// CreateNovel tạo novel mới
func (h *Handler) CreateNovel(c *gin.Context) {
	userIDStr, exists := middleware.GetUserID(c)
	if !exists {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", I18nAuthUnauthorized, nil)
		return
	}

	userID, err := uuid.FromString(userIDStr)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "INVALID_USER_ID", I18nAuthInvalidUserID, nil)
		return
	}

	var req noveldto.CreateNovelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_FAILED", I18nValidationFailed, err.Error())
		return
	}

	ownerID, err := uuid.FromString(req.OwnerID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_OWNER_ID", I18nInvalidOwnerID, nil)
		return
	}

	var genreIDs []uuid.UUID
	for _, idStr := range req.GenreIDs {
		id, err := uuid.FromString(idStr)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "INVALID_GENRE_ID", I18nInvalidGenreID, nil)
			return
		}
		genreIDs = append(genreIDs, id)
	}

	var authorIDs []uuid.UUID
	for _, idStr := range req.AuthorIDs {
		id, err := uuid.FromString(idStr)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "INVALID_AUTHOR_ID", I18nInvalidAuthorID, nil)
			return
		}
		authorIDs = append(authorIDs, id)
	}

	var artistIDs []uuid.UUID
	for _, idStr := range req.ArtistIDs {
		id, err := uuid.FromString(idStr)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "INVALID_ARTIST_ID", I18nInvalidArtistID, nil)
			return
		}
		artistIDs = append(artistIDs, id)
	}

	novel, err := h.createNovelUC.Execute(c.Request.Context(), CreateNovelInput{
		Title:            req.Title,
		Synopsis:         req.Synopsis,
		CoverImageURL:    req.CoverImageURL,
		ThumbnailURL:     req.ThumbnailURL,
		Status:           req.Status,
		IsOneshot:        req.IsOneshot,
		OriginalLanguage: req.OriginalLanguage,
		OriginalTitle:    req.OriginalTitle,
		MetadataJSON:     req.Metadata,
		OwnerID:          ownerID,
		OwnerType:        req.OwnerType,
		CreatedBy:        userID,
		GenreIDs:         genreIDs,
		AuthorIDs:        authorIDs,
		ArtistIDs:        artistIDs,
	})
	if err != nil {
		fmt.Printf("[DEBUG] CreateNovel failed: %v\n", err)
		if appErr, ok := pkgerrors.AsAppError(err); ok {
			response.Error(c, appErr.StatusCode, appErr.ErrCode, appErr.I18nKey, nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "CREATE_FAILED", I18nCreateFailed, nil)
		return
	}

	resp := mapToNovelDetailResponse(novel)
	response.Success(c, http.StatusCreated, I18nCreatedSuccess, resp, nil)
}

// UpdateNovel cập nhật novel
func (h *Handler) UpdateNovel(c *gin.Context) {
	idStr := c.Param("identifier")
	id, err := uuid.FromString(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", I18nInvalidID, nil)
		return
	}

	var req noveldto.UpdateNovelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_FAILED", I18nValidationFailed, err.Error())
		return
	}

	novel, err := h.updateNovelUC.Execute(c.Request.Context(), UpdateNovelInput{
		ID:               id,
		Title:            req.Title,
		Synopsis:         req.Synopsis,
		CoverImageURL:    req.CoverImageURL,
		ThumbnailURL:     req.ThumbnailURL,
		Status:           req.Status,
		IsOneshot:        req.IsOneshot,
		OriginalLanguage: req.OriginalLanguage,
		OriginalTitle:    req.OriginalTitle,
		MetadataJSON:     req.Metadata,
	})
	if err != nil {
		if appErr, ok := pkgerrors.AsAppError(err); ok {
			response.Error(c, appErr.StatusCode, appErr.ErrCode, appErr.I18nKey, nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "UPDATE_FAILED", I18nUpdateFailed, nil)
		return
	}

	resp := mapToNovelDetailResponse(novel)
	response.Success(c, http.StatusOK, I18nUpdatedSuccess, resp, nil)
}

// DeleteNovel xóa novel
func (h *Handler) DeleteNovel(c *gin.Context) {
	idStr := c.Param("identifier")
	id, err := uuid.FromString(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", I18nInvalidID, nil)
		return
	}

	err = h.deleteNovelUC.Execute(c.Request.Context(), DeleteNovelInput{ID: id})
	if err != nil {
		if appErr, ok := pkgerrors.AsAppError(err); ok {
			response.Error(c, appErr.StatusCode, appErr.ErrCode, appErr.I18nKey, nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "DELETE_FAILED", I18nDeleteFailed, nil)
		return
	}

	response.Success(c, http.StatusOK, I18nDeletedSuccess, nil, nil)
}

// GetNovel lấy thông tin chi tiết novel
func (h *Handler) GetNovel(c *gin.Context) {
	idOrSlug := c.Param("identifier")

	novel, err := h.getNovelUC.Execute(c.Request.Context(), GetNovelInput{
		IDOrSlug: idOrSlug,
	})

	if err != nil {
		if appErr, ok := pkgerrors.AsAppError(err); ok {
			response.Error(c, appErr.StatusCode, appErr.ErrCode, appErr.I18nKey, nil)
			return
		}
		if err == pgx.ErrNoRows {
			response.Error(c, http.StatusNotFound, "NOVEL_NOT_FOUND", I18nNotFound, nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "GET_FAILED", I18nGetFailed, nil)
		return
	}

	resp := mapToNovelDetailResponse(novel)

	// Initialize slices
	resp.GenreIDs = make([]string, 0)
	resp.AuthorIDs = make([]string, 0)
	resp.ArtistIDs = make([]string, 0)
	resp.Genres = make([]noveldto.GenreInfo, 0)
	resp.Authors = make([]noveldto.CreatorInfo, 0)
	resp.Artists = make([]noveldto.CreatorInfo, 0)

	// Load relations
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

	genres, err := h.novelService.GetNovelGenresDetails(c.Request.Context(), novel.ID)
	if err == nil {
		resp.Genres = make([]noveldto.GenreInfo, len(genres))
		for i, g := range genres {
			resp.Genres[i] = noveldto.GenreInfo{ID: g.ID.String(), Name: g.Name, Slug: g.Slug}
		}
	}

	authors, err := h.novelService.GetNovelAuthorsDetails(c.Request.Context(), novel.ID)
	if err == nil {
		resp.Authors = make([]noveldto.CreatorInfo, len(authors))
		for i, na := range authors {
			if na.Author != nil {
				resp.Authors[i] = noveldto.CreatorInfo{ID: na.Author.ID.String(), Name: na.Author.Name, Slug: na.Author.Slug}
			}
		}
	}

	artists, err := h.novelService.GetNovelArtistsDetails(c.Request.Context(), novel.ID)
	if err == nil {
		resp.Artists = make([]noveldto.CreatorInfo, len(artists))
		for i, na := range artists {
			if na.Artist != nil {
				resp.Artists[i] = noveldto.CreatorInfo{ID: na.Artist.ID.String(), Name: na.Artist.Name, Slug: na.Artist.Slug}
			}
		}
	}

	response.Success(c, http.StatusOK, I18nGetSuccess, resp, nil)
}

// ListNovels lấy danh sách novels

// @Summary List novels with pagination, search and filters
// @Tags Novels
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} response.StandardResponse{data=[]NovelResponse}
// @Failure 400 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/v1/novels [get]
func (h *Handler) ListNovels(c *gin.Context) {
	var req noveldto.ListNovelsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_FAILED", I18nValidationFailed, err.Error())
		return
	}

	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 {
		req.Limit = 20
	}

	var ownerID *uuid.UUID
	if req.Owner != "" {
		id, err := uuid.FromString(req.Owner)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "INVALID_OWNER_ID", I18nInvalidOwnerID, nil)
			return
		}
		ownerID = &id
	}

	var genreIDs []uuid.UUID
	for _, idStr := range req.GenreIDs {
		id, err := uuid.FromString(idStr)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "INVALID_GENRE_ID", I18nInvalidGenreID, nil)
			return
		}
		genreIDs = append(genreIDs, id)
	}

	novels, totalCount, err := h.listNovelsUC.Execute(c.Request.Context(), ListNovelsInput{
		Page:             req.Page,
		Limit:            req.Limit,
		OwnerID:          ownerID,
		KeySearch:        req.KeySearch,
		GenreIDs:         genreIDs,
		StatusStrs:       req.Statuses,
		OriginalLanguage: req.OriginalLanguage,
		SortBy:           req.SortBy,
		SortOrder:        req.SortOrder,
	})
	if err != nil {
		fmt.Printf("❌ [Handler] ListNovels Error: %v\n", err)
		response.Error(c, http.StatusInternalServerError, "LIST_FAILED", I18nListFailed, nil)
		return
	}

	novelResponses := make([]noveldto.NovelResponse, len(novels))
	for i, novel := range novels {
		novelResponses[i] = mapToNovelResponse(novel)
	}

	totalPages := (totalCount + req.Limit - 1) / req.Limit
	meta := &response.PaginationMeta{
		Page:       req.Page,
		Limit:      req.Limit,
		TotalItems: totalCount,
		TotalPages: totalPages,
	}

	response.Success(c, http.StatusOK, I18nListSuccess, novelResponses, meta)
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
	idStr := c.Param("identifier")
	id, err := uuid.FromString(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", I18nInvalidID, nil)
		return
	}

	err = h.viewCountUC.Execute(c.Request.Context(), IncrementViewCountInput{ID: id})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INCREMENT_FAILED", I18nIncrementViewFailed, nil)
		return
	}

	response.Success(c, http.StatusOK, I18nViewIncremented, nil, nil)
}

// GetNovelFull lấy thông tin đầy đủ novel (public API)
// @Summary Get full novel details with volumes and chapters
// @Tags Novels
// @Produce json
// @Param identifier path string true "Novel Slug"
// @Success 200 {object} response.StandardResponse{data=NovelFullResponse}
// @Failure 404 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/v1/novels/{identifier}/full [get]
func (h *Handler) GetNovelFull(c *gin.Context) {
	identifier := c.Param("identifier")

	data, err := h.getNovelFullUC.Execute(c.Request.Context(), GetNovelFullInput{
		Slug: identifier,
	})
	if err != nil {
		if errors.Is(err, pkgerrors.ErrNovelNotFound) || errors.Is(err, pgx.ErrNoRows) {
			response.Error(c, http.StatusNotFound, "NOVEL_NOT_FOUND", I18nNotFound, nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "GET_FAILED", I18nGetFailed, nil)
		return
	}

	resp := mapNovelFullDataToResponse(data)
	response.Success(c, http.StatusOK, I18nGetSuccess, resp, nil)
}

// mapNovelFullDataToResponse chuyển đổi NovelFullData thành NovelFullResponse
func mapNovelFullDataToResponse(data *NovelFullData) noveldto.NovelFullResponse {
	novel := data.Novel

	resp := noveldto.NovelFullResponse{
		NovelDetailResponse: mapToNovelDetailResponse(novel),
	}

	// Initialize slices
	resp.GenreIDs = make([]string, 0)
	resp.AuthorIDs = make([]string, 0)
	resp.ArtistIDs = make([]string, 0)
	resp.Genres = make([]noveldto.GenreInfo, 0)
	resp.Authors = make([]noveldto.CreatorInfo, 0)
	resp.Artists = make([]noveldto.CreatorInfo, 0)
	resp.Volumes = make([]noveldto.VolumeInfoResponseWithChapters, 0)
	resp.Chapters = make([]noveldto.ChapterSummaryResponse, 0)

	// Set owner info
	ownerDisplayName := "Unknown Owner"
	if novel.OwnerDisplayName != nil && *novel.OwnerDisplayName != "" {
		ownerDisplayName = *novel.OwnerDisplayName
	}
	ownerUsername := "unknown"
	if novel.OwnerUsername != nil && *novel.OwnerUsername != "" {
		ownerUsername = *novel.OwnerUsername
	}
	resp.Owner = noveldto.OwnerInfo{
		ID:          novel.OwnerID.String(),
		DisplayName: ownerDisplayName,
		Username:    ownerUsername,
		AvatarURL:   novel.OwnerAvatarURL,
	}

	// Map genres
	for _, g := range data.Genres {
		resp.Genres = append(resp.Genres, noveldto.GenreInfo{ID: g.ID.String(), Name: g.Name, Slug: g.Slug})
	}

	// Map authors
	for _, na := range data.Authors {
		if na.Author != nil {
			resp.Authors = append(resp.Authors, noveldto.CreatorInfo{ID: na.Author.ID.String(), Name: na.Author.Name, Slug: na.Author.Slug})
		}
	}

	// Map artists
	for _, na := range data.Artists {
		if na.Artist != nil {
			resp.Artists = append(resp.Artists, noveldto.CreatorInfo{ID: na.Artist.ID.String(), Name: na.Artist.Name, Slug: na.Artist.Slug})
		}
	}

	// Map volumes with chapters
	for _, vwc := range data.VolumesWithChapters {
		vol := vwc.Volume
		volResp := noveldto.VolumeInfoResponseWithChapters{
			ID:            vol.ID.String(),
			VolumeNumber:  vol.VolumeNumber,
			Title:         vol.Title,
			Slug:          vol.Slug,
			CoverImageURL: vol.CoverImageURL,
			DisplayOrder:  vol.DisplayOrder,
			IsPublished:   vol.IsPublished,
			Chapters:      make([]noveldto.ChapterSummaryResponse, 0),
		}
		if vol.PublishedAt != nil {
			publishedAt := vol.PublishedAt.Format(timeutil.ISO8601Layout)
			volResp.PublishedAt = &publishedAt
		}

		// Map chapters for this volume
		for _, ch := range vwc.Chapters {
			volResp.Chapters = append(volResp.Chapters, mapToChapterSummary(ch))
		}

		resp.Volumes = append(resp.Volumes, volResp)
	}

	// Map chapters without volume
	for _, ch := range data.ChaptersWithoutVol {
		resp.Chapters = append(resp.Chapters, mapToChapterSummary(ch))
	}

	return resp
}

// Helper functions

func mapToNovelDetailResponse(novel *domain.Novel) noveldto.NovelDetailResponse {
	resp := noveldto.NovelDetailResponse{
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

	if len(novel.Synopsis) > 0 && string(novel.Synopsis) != "null" {
		resp.Synopsis = novel.Synopsis
	}

	if len(novel.Metadata) > 0 {
		metadataStr := string(novel.Metadata)
		resp.Metadata = &metadataStr
	}

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

func mapToNovelResponse(novel *domain.Novel) noveldto.NovelResponse {
	ownerDisplayName := "Unknown Owner"
	if novel.OwnerDisplayName != nil && *novel.OwnerDisplayName != "" {
		ownerDisplayName = *novel.OwnerDisplayName
	}
	ownerUsername := "unknown"
	if novel.OwnerUsername != nil && *novel.OwnerUsername != "" {
		ownerUsername = *novel.OwnerUsername
	}

	owner := noveldto.OwnerInfo{
		ID:          novel.OwnerID.String(),
		DisplayName: ownerDisplayName,
		Username:    ownerUsername,
		AvatarURL:   novel.OwnerAvatarURL,
	}

	genres := make([]noveldto.GenreInfo, 0)
	if novel.Genres != nil && len(novel.Genres) > 0 {
		for _, genre := range novel.Genres {
			genres = append(genres, noveldto.GenreInfo{ID: genre.ID.String(), Name: genre.Name, Slug: genre.Slug})
		}
	}

	var latestChapter *noveldto.LatestChapterInfo
	if novel.LastChapterAt != nil {
		latestChapter = &noveldto.LatestChapterInfo{
			ID:          "",
			Title:       "Latest Chapter",
			PublishedAt: novel.LastChapterAt.Format(timeutil.ISO8601Layout),
		}
	}

	rating := novel.RatingAverage * 2

	return noveldto.NovelResponse{
		ID:               novel.ID.String(),
		Title:            novel.Title,
		OriginalTitle:    novel.OriginalTitle,
		Slug:             novel.Slug,
		OriginalLanguage: novel.OriginalLanguage,
		Synopsis:         json.RawMessage("[]"),
		CoverURL:         novel.CoverImageURL,
		Type:             "novel",
		Status:           string(novel.Status),
		Genres:           genres,
		Owner:            owner,
		Rating:           rating,
		Views:            novel.ViewCount,
		Favorites:        novel.FavoriteCount,
		LatestChapter:    latestChapter,
		CreatedAt:        novel.CreatedAt.Format(timeutil.ISO8601Layout),
		UpdatedAt:        novel.UpdatedAt.Format(timeutil.ISO8601Layout),
	}
}

func mapToChapterSummary(ch *domain.NovelChapter) noveldto.ChapterSummaryResponse {
	resp := noveldto.ChapterSummaryResponse{
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
