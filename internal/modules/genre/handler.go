package genre

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"

	"system/internal/app/middleware"
	"system/internal/domain"
	genredto "system/internal/dto/genre"
	analytics_module "system/internal/modules/analytics"
	pkgerrors "system/pkg/errors"
	"system/pkg/util/response"
	"system/pkg/util/timeutil"
)

type Handler struct {
	createGenreUC    CreateGenreUseCase
	updateGenreUC    UpdateGenreUseCase
	deleteGenreUC    DeleteGenreUseCase
	getGenreUC       GetGenreUseCase
	listGenresUC     ListGenresUseCase
	listSelectUC     ListSelectionUseCase
	mergeGenresUC    MergeGenresUseCase
	previewMergeUC   PreviewMergeUseCase
	analyticsSvc     analytics_module.AnalyticsService
	logger           *zap.Logger
}

func NewHandler(
	createGenreUC CreateGenreUseCase,
	updateGenreUC UpdateGenreUseCase,
	deleteGenreUC DeleteGenreUseCase,
	getGenreUC GetGenreUseCase,
	listGenresUC ListGenresUseCase,
	listSelectUC ListSelectionUseCase,
	mergeGenresUC MergeGenresUseCase,
	previewMergeUC PreviewMergeUseCase,
	analyticsSvc analytics_module.AnalyticsService,
	logger *zap.Logger,
) *Handler {
	return &Handler{
		createGenreUC:    createGenreUC,
		updateGenreUC:    updateGenreUC,
		deleteGenreUC:    deleteGenreUC,
		getGenreUC:       getGenreUC,
		listGenresUC:     listGenresUC,
		listSelectUC:     listSelectUC,
		mergeGenresUC:    mergeGenresUC,
		previewMergeUC:   previewMergeUC,
		analyticsSvc:     analyticsSvc,
		logger:           logger,
	}
}

// CreateGenre tạo genre mới
func (h *Handler) CreateGenre(c *gin.Context) {
	var req genredto.CreateGenreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "ValidationFailed", I18nValidationFailed, err.Error())
		return
	}

	userIDStr, exists := middleware.GetUserID(c)
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Unauthorized", I18nAuthUnauthorized, nil)
		return
	}

	userID, err := uuid.FromString(userIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "InvalidUserID", I18nAuthInvalidUserID, nil)
		return
	}

	var parentID *uuid.UUID
	if req.ParentID != nil && *req.ParentID != "" {
		pid, err := uuid.FromString(*req.ParentID)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "BadRequest", I18nInvalidParentID, nil)
			return
		}
		parentID = &pid
	}

	genre, err := h.createGenreUC.Execute(c.Request.Context(), CreateGenreInput{
		Name:        req.Name,
		Description: req.Description,
		ParentID:    parentID,
		UserID:      userID,
	})
	if err != nil {
		if appErr, ok := pkgerrors.AsAppError(err); ok {
			response.Error(c, appErr.StatusCode, appErr.ErrCode, appErr.I18nKey, nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "InternalError", I18nCreateFailed, nil)
		return
	}

	resp := mapToGenreDetailResponse(genre, "stable")
	response.Success(c, http.StatusCreated, I18nCreatedSuccess, resp, nil)
}

// UpdateGenre cập nhật genre
func (h *Handler) UpdateGenre(c *gin.Context) {
	idStr := c.Param("identifier")
	id, err := uuid.FromString(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "BadRequest", I18nInvalidID, nil)
		return
	}

	var req genredto.UpdateGenreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "ValidationFailed", I18nValidationFailed, err.Error())
		return
	}

	userIDStr, exists := middleware.GetUserID(c)
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Unauthorized", I18nAuthUnauthorized, nil)
		return
	}

	userID, err := uuid.FromString(userIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "InvalidUserID", I18nAuthInvalidUserID, nil)
		return
	}

	var parentID *uuid.UUID
	if req.ParentID != nil && *req.ParentID != "" {
		pid, err := uuid.FromString(*req.ParentID)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "BadRequest", I18nInvalidParentID, nil)
			return
		}
		parentID = &pid
	}

	genre, err := h.updateGenreUC.Execute(c.Request.Context(), UpdateGenreInput{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
		ParentID:    parentID,
		IsActive:    req.IsActive,
		UserID:      userID,
	})
	if err != nil {
		if appErr, ok := pkgerrors.AsAppError(err); ok {
			response.Error(c, appErr.StatusCode, appErr.ErrCode, appErr.I18nKey, nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "InternalError", I18nUpdateFailed, nil)
		return
	}

	resp := mapToGenreDetailResponse(genre, "stable")
	response.Success(c, http.StatusOK, I18nUpdatedSuccess, resp, nil)
}

// DeleteGenre xóa genre
func (h *Handler) DeleteGenre(c *gin.Context) {
	idStr := c.Param("identifier")
	id, err := uuid.FromString(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "BadRequest", I18nInvalidID, nil)
		return
	}

	userIDStr, exists := middleware.GetUserID(c)
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Unauthorized", I18nAuthUnauthorized, nil)
		return
	}

	userID, err := uuid.FromString(userIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "InvalidUserID", I18nAuthInvalidUserID, nil)
		return
	}

	err = h.deleteGenreUC.Execute(c.Request.Context(), DeleteGenreInput{
		ID:     id,
		UserID: userID,
	})
	if err != nil {
		if appErr, ok := pkgerrors.AsAppError(err); ok {
			response.Error(c, appErr.StatusCode, appErr.ErrCode, appErr.I18nKey, nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "InternalError", I18nDeleteFailed, nil)
		return
	}

	response.Success(c, http.StatusOK, I18nDeletedSuccess, nil, nil)
}

// GetGenre lấy thông tin chi tiết genre
func (h *Handler) GetGenre(c *gin.Context) {
	idOrSlug := c.Param("identifier")

	genre, err := h.getGenreUC.Execute(c.Request.Context(), GetGenreInput{
		IDOrSlug: idOrSlug,
	})
	if err != nil {
		if appErr, ok := pkgerrors.AsAppError(err); ok {
			response.Error(c, appErr.StatusCode, appErr.ErrCode, appErr.I18nKey, nil)
			return
		}
		if err == pgx.ErrNoRows {
			response.Error(c, http.StatusNotFound, "NotFound", I18nNotFound, nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "InternalError", I18nGetFailed, nil)
		return
	}

	trend := "stable"
	if genre.ActiveReaders > 1000 {
		trend = "rising"
	} else if genre.ActiveReaders < 100 {
		trend = "falling"
	}

	resp := mapToGenreDetailResponse(genre, trend)
	response.Success(c, http.StatusOK, I18nGetSuccess, resp, nil)
}

// ListGenres lấy danh sách genres
func (h *Handler) ListGenres(c *gin.Context) {
	var req genredto.ListGenresRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "ValidationFailed", I18nValidationFailed, err.Error())
		return
	}

	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 {
		req.Limit = 20
	}

	genresWithTrend, totalCount, err := h.listGenresUC.Execute(c.Request.Context(), ListGenresInput{
		Page:       req.Page,
		Limit:      req.Limit,
		Search:     req.Search,
		SortBy:     req.SortBy,
		SortOrder:  req.SortOrder,
		ActiveOnly: req.ActiveOnly,
	})
	if err != nil {
		h.logger.Error("Failed to list genres", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "InternalError", I18nListFailed, nil)
		return
	}

	genreResponses := make([]genredto.GenreResponse, len(genresWithTrend))
	for i, gwt := range genresWithTrend {
		genreResponses[i] = genredto.GenreResponse{
			ID:            gwt.Genre.ID.String(),
			Name:          gwt.Genre.Name,
			Slug:          gwt.Genre.Slug,
			IsActive:      gwt.Genre.IsActive,
			SeriesCount:   gwt.Genre.NovelCount,
			ActiveReaders: gwt.Genre.ActiveReaders,
			TotalViews:    gwt.Genre.TotalViews,
			Trend:         string(gwt.Trend),
			Description:   gwt.Genre.Description,
			CreatedAt:     gwt.Genre.CreatedAt.Format(timeutil.ISO8601Layout),
			UpdatedAt:     gwt.Genre.UpdatedAt.Format(timeutil.ISO8601Layout),
		}
	}

	totalPages := (totalCount + req.Limit - 1) / req.Limit
	meta := &response.PaginationMeta{
		Page:       req.Page,
		Limit:      req.Limit,
		TotalItems: totalCount,
		TotalPages: totalPages,
	}

	response.Success(c, http.StatusOK, I18nListSuccess, genreResponses, meta)
}

func mapToGenreDetailResponse(genre *domain.Genre, trend string) genredto.GenreDetailResponse {
	resp := genredto.GenreDetailResponse{
		ID:            genre.ID.String(),
		Name:          genre.Name,
		Slug:          genre.Slug,
		Description:   genre.Description,
		IsActive:      genre.IsActive,
		SeriesCount:   genre.NovelCount,
		ActiveReaders: genre.ActiveReaders,
		TotalViews:    genre.TotalViews,
		Trend:         trend,
		CreatedAt:     genre.CreatedAt.Format(timeutil.ISO8601Layout),
		UpdatedAt:     genre.UpdatedAt.Format(timeutil.ISO8601Layout),
	}

	if genre.ParentID != nil {
		parentIDStr := genre.ParentID.String()
		resp.ParentID = &parentIDStr
	}

	return resp
}

// ListSelection lấy danh sách genres rút gọn
func (h *Handler) ListSelection(c *gin.Context) {
	var req genredto.ListGenresRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "ValidationFailed", "validation.failed", err.Error())
		return
	}

	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 {
		req.Limit = 20
	}

	genres, totalCount, err := h.listSelectUC.Execute(c.Request.Context(), ListSelectionInput{
		Page:   req.Page,
		Limit:  req.Limit,
		Search: req.Search,
	})
	if err != nil {
		h.logger.Error("Failed to list genres selection", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "InternalError", "genre.list_failed", nil)
		return
	}

	type SelectionResponse struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	selectionResponses := make([]SelectionResponse, len(genres))
	for i, g := range genres {
		selectionResponses[i] = SelectionResponse{
			ID:   g.ID.String(),
			Name: g.Name,
		}
	}

	totalPages := (totalCount + req.Limit - 1) / req.Limit
	meta := &response.PaginationMeta{
		Page:       req.Page,
		Limit:      req.Limit,
		TotalItems: totalCount,
		TotalPages: totalPages,
	}

	response.Success(c, http.StatusOK, I18nListSuccess, selectionResponses, meta)
}

// MergeGenre gộp genres thành một
func (h *Handler) MergeGenre(c *gin.Context) {
	var req genredto.MergeGenreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "ValidationFailed", I18nValidationFailed, err.Error())
		return
	}

	userIDStr, exists := middleware.GetUserID(c)
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Unauthorized", I18nAuthUnauthorized, nil)
		return
	}

	userID, err := uuid.FromString(userIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "InvalidUserID", I18nAuthInvalidUserID, nil)
		return
	}

	err = h.mergeGenresUC.Execute(c.Request.Context(), MergeGenresInput{
		TargetID:  req.TargetID,
		SourceIDs: req.SourceIDs,
		MergedBy:  userID,
	})
	if err != nil {
		if appErr, ok := pkgerrors.AsAppError(err); ok {
			response.Error(c, appErr.StatusCode, appErr.ErrCode, appErr.I18nKey, nil)
			return
		}

		h.logger.Error("Failed to merge genres", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "InternalError", "genre.merge_failed", nil)
		return
	}

	response.Success(c, http.StatusOK, "genre.merge_success", nil, nil)
}

// PreviewMergeGenre xem trước kết quả gộp genres
func (h *Handler) PreviewMergeGenre(c *gin.Context) {
	var req genredto.MergeGenreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "ValidationFailed", I18nValidationFailed, err.Error())
		return
	}

	preview, err := h.previewMergeUC.Execute(c.Request.Context(), PreviewMergeInput{
		TargetID:  req.TargetID,
		SourceIDs: req.SourceIDs,
	})
	if err != nil {
		if appErr, ok := pkgerrors.AsAppError(err); ok {
			response.Error(c, appErr.StatusCode, appErr.ErrCode, appErr.I18nKey, nil)
			return
		}

		response.Error(c, http.StatusInternalServerError, "InternalError", "genre.preview_failed", err.Error())
		return
	}

	affectedNovels := make([]genredto.AffectedNovel, len(preview))
	for i, n := range preview {
		affectedNovels[i] = genredto.AffectedNovel{
			ID:            n.ID.String(),
			Title:         n.Title,
			Slug:          n.Slug,
			CoverImageURL: n.CoverImageURL,
		}
	}

	responsePayload := genredto.PreviewMergeGenreResponse{
		AffectedNovels: affectedNovels,
	}

	response.Success(c, http.StatusOK, "genre.preview_success", responsePayload, nil)
}

// GetTopGenresByViews returns top genres by view count for a time period
// @Summary Get top genres by views
// @Description Get genres with highest view counts for a calendar-based time period. Week starts Monday, Month starts 1st.
// @Tags Genres
// @Accept json
// @Produce json
// @Param period query string false "Time period (day, week, month, year)" default(week)
// @Param offset query int false "0 = current period, 1 = previous period" default(0)
// @Param limit query int false "Limit (default 10)" default(10)
// @Success 200 {object} response.StandardResponse{data=[]genredto.GenreResponse}
// @Failure 500 {object} response.StandardResponse
// @Router /api/v1/genres/top [get]
func (h *Handler) GetTopGenresByViews(c *gin.Context) {
	period := c.DefaultQuery("period", "week")
	offsetStr := c.DefaultQuery("offset", "0")
	limitStr := c.DefaultQuery("limit", "10")

	// Validate period
	validPeriods := map[string]bool{"day": true, "week": true, "month": true, "year": true}
	if !validPeriods[period] {
		period = "week"
	}

	// Parse offset (0 = current, 1 = previous, etc.)
	offset := 0
	if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 && o <= 52 {
		offset = o
	}

	limit := 10
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
		limit = l
	}

	// Call analytics service
	includeRankChange := c.Query("include_rank_change") == "true"
	
	if includeRankChange {
		genresWithRank, err := h.analyticsSvc.GetTopGenresWithRankComparison(c.Request.Context(), period, limit)
		if err != nil {
			h.logger.Error("Failed to get top genres with rank", zap.Error(err))
			response.Error(c, http.StatusInternalServerError, "InternalError", I18nListFailed, nil)
			return
		}

		genreResponses := make([]genredto.GenreResponse, len(genresWithRank))
		for i, gwt := range genresWithRank {
			// Calculate Trend based on existing logic or rank change
			trend := "stable"
			if gwt.Genre.ActiveReaders > 1000 {
				trend = "rising"
			} else if gwt.Genre.ActiveReaders < 100 {
				trend = "falling"
			}
			
			// Use pointers for optional fields
			currentRank := gwt.Stats.CurrentRank
			var prevRank *int
			if gwt.Stats.PreviousRank != nil {
				pr := *gwt.Stats.PreviousRank
				prevRank = &pr
			}
			var rankChange *int
			if gwt.Stats.RankChange != nil {
				rc := *gwt.Stats.RankChange
				rankChange = &rc
			}

			genreResponses[i] = genredto.GenreResponse{
				ID:            gwt.Genre.ID.String(),
				Name:          gwt.Genre.Name,
				Slug:          gwt.Genre.Slug,
				IsActive:      gwt.Genre.IsActive,
				SeriesCount:   gwt.Genre.NovelCount,
				ActiveReaders: gwt.Genre.ActiveReaders,
				TotalViews:    gwt.Genre.TotalViews,
				Trend:         trend,
				Description:   gwt.Genre.Description,
				CreatedAt:     gwt.Genre.CreatedAt.Format(timeutil.ISO8601Layout),
				UpdatedAt:     gwt.Genre.UpdatedAt.Format(timeutil.ISO8601Layout),
				
				// Rank info
				CurrentRank:  &currentRank,
				PreviousRank: prevRank,
				RankChange:   rankChange,
			}
		}
		response.Success(c, http.StatusOK, I18nListSuccess, genreResponses, nil)
		return
	}

	genres, err := h.analyticsSvc.GetTopGenresByViews(c.Request.Context(), period, offset, limit)
	if err != nil {
		h.logger.Error("Failed to get top genres by views", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "InternalError", I18nListFailed, nil)
		return
	}

	// Map to GenreResponse format (same as ListGenres)
	genreResponses := make([]genredto.GenreResponse, len(genres))
	for i, g := range genres {
		trend := "stable"
		if g.ActiveReaders > 1000 {
			trend = "rising"
		} else if g.ActiveReaders < 100 {
			trend = "falling"
		}

		genreResponses[i] = genredto.GenreResponse{
			ID:            g.ID.String(),
			Name:          g.Name,
			Slug:          g.Slug,
			IsActive:      g.IsActive,
			SeriesCount:   g.NovelCount,
			ActiveReaders: g.ActiveReaders,
			TotalViews:    g.TotalViews,
			Trend:         trend,
			Description:   g.Description,
			CreatedAt:     g.CreatedAt.Format(timeutil.ISO8601Layout),
			UpdatedAt:     g.UpdatedAt.Format(timeutil.ISO8601Layout),
		}
	}

	response.Success(c, http.StatusOK, I18nListSuccess, genreResponses, nil)
}

