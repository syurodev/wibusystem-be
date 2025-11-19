package genre

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"

	"system/internal/domain"
	"system/internal/pkg/service"
	pkgerrors "system/pkg/errors"
	"system/pkg/util/response"
)

type Handler struct {
	genreService *service.GenreService
	logger       *zap.Logger
}

func NewHandler(genreService *service.GenreService, logger *zap.Logger) *Handler {
	return &Handler{
		genreService: genreService,
		logger:       logger,
	}
}

// CreateGenre tạo genre mới
// @Summary Create a new genre
// @Tags Genres
// @Accept json
// @Produce json
// @Param request body CreateGenreRequest true "Create Genre Request"
// @Success 201 {object} response.StandardResponse{data=GenreDetailResponse}
// @Failure 400 {object} response.StandardResponse
// @Failure 409 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/v1/genres [post]
func (h *Handler) CreateGenre(c *gin.Context) {
	var req CreateGenreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_FAILED", "validation.failed", err.Error())
		return
	}

	// Get user ID from context (assuming auth middleware sets this)
	userIDStr, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "auth.unauthorized", nil)
		return
	}

	userID, err := uuid.FromString(userIDStr.(string))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_USER_ID", "auth.invalid_user_id", nil)
		return
	}

	// Parse parent ID if provided
	var parentID *uuid.UUID
	if req.ParentID != nil && *req.ParentID != "" {
		pid, err := uuid.FromString(*req.ParentID)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "INVALID_PARENT_ID", "genre.invalid_parent_id", nil)
			return
		}
		parentID = &pid
	}

	// Create genre
	genre, err := h.genreService.CreateGenre(c.Request.Context(), req.Name, req.Description, parentID, userID)
	if err != nil {
		if errors.Is(err, pkgerrors.ErrInvalidInput) {
			response.Error(c, http.StatusBadRequest, "INVALID_INPUT", "genre.invalid_input", nil)
			return
		}
		if errors.Is(err, pkgerrors.ErrSlugAlreadyExists) {
			response.Error(c, http.StatusConflict, "SLUG_EXISTS", "genre.slug_already_exists", nil)
			return
		}
		if errors.Is(err, pkgerrors.ErrParentGenreNotFound) {
			response.Error(c, http.StatusBadRequest, "PARENT_NOT_FOUND", "genre.parent_not_found", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "CREATE_FAILED", "genre.create_failed", nil)
		return
	}

	// Map to response
	resp := mapToGenreDetailResponse(genre, "stable")

	response.Success(c, http.StatusCreated, "genre.created_success", resp, nil)
}

// UpdateGenre cập nhật genre
// @Summary Update a genre
// @Tags Genres
// @Accept json
// @Produce json
// @Param id path string true "Genre ID"
// @Param request body UpdateGenreRequest true "Update Genre Request"
// @Success 200 {object} response.StandardResponse{data=GenreDetailResponse}
// @Failure 400 {object} response.StandardResponse
// @Failure 404 {object} response.StandardResponse
// @Failure 409 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/v1/genres/{id} [put]
func (h *Handler) UpdateGenre(c *gin.Context) {
	// Get genre ID from path
	idStr := c.Param("id")
	id, err := uuid.FromString(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "genre.invalid_id", nil)
		return
	}

	var req UpdateGenreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_FAILED", "validation.failed", err.Error())
		return
	}

	// Get user ID from context
	userIDStr, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "auth.unauthorized", nil)
		return
	}

	userID, err := uuid.FromString(userIDStr.(string))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_USER_ID", "auth.invalid_user_id", nil)
		return
	}

	// Parse parent ID if provided
	var parentID *uuid.UUID
	if req.ParentID != nil && *req.ParentID != "" {
		pid, err := uuid.FromString(*req.ParentID)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "INVALID_PARENT_ID", "genre.invalid_parent_id", nil)
			return
		}
		parentID = &pid
	}

	// Update genre
	genre, err := h.genreService.UpdateGenre(c.Request.Context(), id, req.Name, req.Description, parentID, req.DisplayOrder, req.IsActive, userID)
	if err != nil {
		if errors.Is(err, pkgerrors.ErrGenreNotFound) {
			response.Error(c, http.StatusNotFound, "GENRE_NOT_FOUND", "genre.not_found", nil)
			return
		}
		if errors.Is(err, pkgerrors.ErrInvalidInput) {
			response.Error(c, http.StatusBadRequest, "INVALID_INPUT", "genre.invalid_input", nil)
			return
		}
		if errors.Is(err, pkgerrors.ErrSlugAlreadyExists) {
			response.Error(c, http.StatusConflict, "SLUG_EXISTS", "genre.slug_already_exists", nil)
			return
		}
		if errors.Is(err, pkgerrors.ErrCircularParentReference) {
			response.Error(c, http.StatusBadRequest, "CIRCULAR_REFERENCE", "genre.circular_reference", nil)
			return
		}
		if errors.Is(err, pkgerrors.ErrParentGenreNotFound) {
			response.Error(c, http.StatusBadRequest, "PARENT_NOT_FOUND", "genre.parent_not_found", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "UPDATE_FAILED", "genre.update_failed", nil)
		return
	}

	// Map to response
	resp := mapToGenreDetailResponse(genre, "stable")

	response.Success(c, http.StatusOK, "genre.updated_success", resp, nil)
}

// DeleteGenre xóa genre
// @Summary Delete a genre
// @Tags Genres
// @Produce json
// @Param id path string true "Genre ID"
// @Success 200 {object} response.StandardResponse
// @Failure 400 {object} response.StandardResponse
// @Failure 404 {object} response.StandardResponse
// @Failure 409 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/v1/genres/{id} [delete]
func (h *Handler) DeleteGenre(c *gin.Context) {
	// Get genre ID from path
	idStr := c.Param("id")
	id, err := uuid.FromString(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "genre.invalid_id", nil)
		return
	}

	// Get user ID from context
	userIDStr, exists := c.Get("userID")
	if !exists {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "auth.unauthorized", nil)
		return
	}

	userID, err := uuid.FromString(userIDStr.(string))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_USER_ID", "auth.invalid_user_id", nil)
		return
	}

	// Delete genre
	err = h.genreService.DeleteGenre(c.Request.Context(), id, userID)
	if err != nil {
		if errors.Is(err, pkgerrors.ErrGenreNotFound) {
			response.Error(c, http.StatusNotFound, "GENRE_NOT_FOUND", "genre.not_found", nil)
			return
		}
		if errors.Is(err, pkgerrors.ErrGenreInUse) {
			response.Error(c, http.StatusConflict, "GENRE_IN_USE", "genre.in_use", nil)
			return
		}
		if errors.Is(err, pkgerrors.ErrGenreHasChildren) {
			response.Error(c, http.StatusConflict, "GENRE_HAS_CHILDREN", "genre.has_children", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "DELETE_FAILED", "genre.delete_failed", nil)
		return
	}

	response.Success(c, http.StatusOK, "genre.deleted_success", nil, nil)
}

// GetGenre lấy thông tin chi tiết genre
// @Summary Get genre details
// @Tags Genres
// @Produce json
// @Param id path string true "Genre ID"
// @Success 200 {object} response.StandardResponse{data=GenreDetailResponse}
// @Failure 400 {object} response.StandardResponse
// @Failure 404 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/v1/genres/{id} [get]
func (h *Handler) GetGenre(c *gin.Context) {
	// Get genre ID from path
	idStr := c.Param("id")
	id, err := uuid.FromString(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "genre.invalid_id", nil)
		return
	}

	// Get genre
	genre, err := h.genreService.GetGenreByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, pkgerrors.ErrGenreNotFound) || errors.Is(err, pgx.ErrNoRows) {
			response.Error(c, http.StatusNotFound, "GENRE_NOT_FOUND", "genre.not_found", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "GET_FAILED", "genre.get_failed", nil)
		return
	}

	// Calculate trend (simple implementation)
	trend := "stable"
	if genre.ActiveReaders > 1000 {
		trend = "rising"
	} else if genre.ActiveReaders < 100 {
		trend = "falling"
	}

	// Map to response
	resp := mapToGenreDetailResponse(genre, trend)

	response.Success(c, http.StatusOK, "genre.get_success", resp, nil)
}

// ListGenres lấy danh sách genres
// @Summary List genres with pagination, search and sort
// @Tags Genres
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param search query string false "Search by name"
// @Param sort_by query string false "Sort by field (name, views, series, created, updated)"
// @Param sort_order query string false "Sort order (asc, desc)"
// @Param active_only query bool false "Filter active genres only" default(false)
// @Success 200 {object} response.StandardResponse{data=[]GenreResponse}
// @Failure 400 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/v1/genres [get]
func (h *Handler) ListGenres(c *gin.Context) {
	var req ListGenresRequest
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

	// Get genres with pagination, search and sort
	genresWithTrend, totalCount, err := h.genreService.ListGenres(
		c.Request.Context(),
		req.Page,
		req.Limit,
		req.Search,
		req.SortBy,
		req.SortOrder,
		req.ActiveOnly,
	)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "LIST_FAILED", "genre.list_failed", nil)
		return
	}

	// Map to response format
	genreResponses := make([]GenreResponse, len(genresWithTrend))
	for i, gwt := range genresWithTrend {
		genreResponses[i] = GenreResponse{
			ID:            gwt.Genre.ID.String(),
			Name:          gwt.Genre.Name,
			SeriesCount:   gwt.Genre.NovelCount,
			ActiveReaders: gwt.Genre.ActiveReaders,
			TotalViews:    gwt.Genre.TotalViews,
			Trend:         string(gwt.Trend),
			Description:   gwt.Genre.Description,
			CreatedAt:     gwt.Genre.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:     gwt.Genre.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	// Calculate pagination meta
	totalPages := (totalCount + req.Limit - 1) / req.Limit
	meta := &response.PaginationMeta{
		Page:       req.Page,
		Limit:      req.Limit,
		TotalItems: totalCount,
		TotalPages: totalPages,
	}

	response.Success(c, http.StatusOK, "genre.list_success", genreResponses, meta)
}

// Helper function to map domain model to detail response
func mapToGenreDetailResponse(genre *domain.Genre, trend string) GenreDetailResponse {
	resp := GenreDetailResponse{
		ID:            genre.ID.String(),
		Name:          genre.Name,
		Slug:          genre.Slug,
		Description:   genre.Description,
		DisplayOrder:  genre.DisplayOrder,
		IsActive:      genre.IsActive,
		SeriesCount:   genre.NovelCount,
		ActiveReaders: genre.ActiveReaders,
		TotalViews:    genre.TotalViews,
		Trend:         trend,
		CreatedAt:     genre.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:     genre.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	if genre.ParentID != nil {
		parentIDStr := genre.ParentID.String()
		resp.ParentID = &parentIDStr
	}

	return resp
}
