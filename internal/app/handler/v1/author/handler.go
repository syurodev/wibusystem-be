package author

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"

	"system/internal/app/middleware"
	"system/internal/domain"
	"system/internal/pkg/service"
	pkgerrors "system/pkg/errors"
	"system/pkg/util/i18nkeys"
	"system/pkg/util/response"
	"system/pkg/util/timeutil"
)

type Handler struct {
	authorService *service.AuthorService
	logger        *zap.Logger
}

func NewHandler(authorService *service.AuthorService, logger *zap.Logger) *Handler {
	return &Handler{
		authorService: authorService,
		logger:        logger,
	}
}

// CreateAuthor tạo author mới
// @Summary Create a new author
// @Tags Authors
// @Accept json
// @Produce json
// @Param request body CreateAuthorRequest true "Create Author Request"
// @Success 201 {object} response.StandardResponse{data=AuthorDetailResponse}
// @Failure 400 {object} response.StandardResponse
// @Failure 409 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/v1/authors [post]
func (h *Handler) CreateAuthor(c *gin.Context) {
	var req CreateAuthorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind JSON",
			zap.Error(err),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
		)
		response.Error(c, http.StatusBadRequest, "VALIDATION_FAILED", i18nkeys.ValidationFailed, err.Error())
		return
	}

	// Get user ID from token
	userIDStr, exists := middleware.GetUserID(c)
	if !exists {
		h.logger.Error("Failed to get user ID from token")
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", i18nkeys.AuthUnauthorized, nil)
		return
	}

	userID, err := uuid.FromString(userIDStr)
	if err != nil {
		h.logger.Error("Invalid user ID format", zap.Error(err), zap.String("user_id", userIDStr))
		response.Error(c, http.StatusBadRequest, "INVALID_USER_ID", i18nkeys.AuthInvalidUserID, nil)
		return
	}

	h.logger.Debug("Creating author",
		zap.String("name", req.Name),
		zap.String("biography", req.Biography),
		zap.Any("avatar_url", req.AvatarURL),
		zap.Any("social_links", req.SocialLinks),
		zap.String("created_by", userID.String()),
	)

	// Create author
	author, err := h.authorService.CreateAuthor(c.Request.Context(), req.Name, req.Biography, req.AvatarURL, req.SocialLinks, userID)
	if err != nil {
		h.logger.Error("Failed to create author",
			zap.Error(err),
			zap.String("name", req.Name),
			zap.String("biography", req.Biography),
			zap.Any("avatar_url", req.AvatarURL),
			zap.Any("social_links", req.SocialLinks),
		)
		if errors.Is(err, pkgerrors.ErrInvalidInput) {
			response.Error(c, http.StatusBadRequest, "INVALID_INPUT", i18nkeys.AuthorInvalidInput, nil)
			return
		}
		if errors.Is(err, pkgerrors.ErrSlugAlreadyExists) {
			response.Error(c, http.StatusConflict, "SLUG_EXISTS", i18nkeys.AuthorSlugAlreadyExists, nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "CREATE_FAILED", i18nkeys.AuthorCreateFailed, nil)
		return
	}

	h.logger.Info("Author created successfully",
		zap.String("author_id", author.ID.String()),
		zap.String("name", author.Name),
		zap.String("slug", author.Slug),
	)

	// Map to response
	resp := mapToAuthorDetailResponse(author)

	response.Success(c, http.StatusCreated, i18nkeys.AuthorCreatedSuccess, resp, nil)
}

// UpdateAuthor cập nhật author
// @Summary Update an author
// @Tags Authors
// @Accept json
// @Produce json
// @Param id path string true "Author ID"
// @Param request body UpdateAuthorRequest true "Update Author Request"
// @Success 200 {object} response.StandardResponse{data=AuthorDetailResponse}
// @Failure 400 {object} response.StandardResponse
// @Failure 404 {object} response.StandardResponse
// @Failure 409 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/v1/authors/{id} [put]
func (h *Handler) UpdateAuthor(c *gin.Context) {
	// Get author ID from path
	idStr := c.Param("id")
	id, err := uuid.FromString(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", i18nkeys.AuthorInvalidID, nil)
		return
	}

	var req UpdateAuthorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_FAILED", i18nkeys.ValidationFailed, err.Error())
		return
	}

	// Update author
	author, err := h.authorService.UpdateAuthor(c.Request.Context(), id, req.Name, req.Biography, req.AvatarURL, req.SocialLinks)
	if err != nil {
		if errors.Is(err, pkgerrors.ErrAuthorNotFound) {
			response.Error(c, http.StatusNotFound, "AUTHOR_NOT_FOUND", i18nkeys.AuthorNotFound, nil)
			return
		}
		if errors.Is(err, pkgerrors.ErrInvalidInput) {
			response.Error(c, http.StatusBadRequest, "INVALID_INPUT", i18nkeys.AuthorInvalidInput, nil)
			return
		}
		if errors.Is(err, pkgerrors.ErrSlugAlreadyExists) {
			response.Error(c, http.StatusConflict, "SLUG_EXISTS", i18nkeys.AuthorSlugAlreadyExists, nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "UPDATE_FAILED", i18nkeys.AuthorUpdateFailed, nil)
		return
	}

	// Map to response
	resp := mapToAuthorDetailResponse(author)

	response.Success(c, http.StatusOK, i18nkeys.AuthorUpdatedSuccess, resp, nil)
}

// DeleteAuthor xóa author
// @Summary Delete an author
// @Tags Authors
// @Produce json
// @Param id path string true "Author ID"
// @Success 200 {object} response.StandardResponse
// @Failure 400 {object} response.StandardResponse
// @Failure 404 {object} response.StandardResponse
// @Failure 409 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/v1/authors/{id} [delete]
func (h *Handler) DeleteAuthor(c *gin.Context) {
	// Get author ID from path
	idStr := c.Param("id")
	id, err := uuid.FromString(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", i18nkeys.AuthorInvalidID, nil)
		return
	}

	// Delete author
	err = h.authorService.DeleteAuthor(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, pkgerrors.ErrAuthorNotFound) {
			response.Error(c, http.StatusNotFound, "AUTHOR_NOT_FOUND", i18nkeys.AuthorNotFound, nil)
			return
		}
		if errors.Is(err, pkgerrors.ErrAuthorInUse) {
			response.Error(c, http.StatusConflict, "AUTHOR_IN_USE", i18nkeys.AuthorInUse, nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "DELETE_FAILED", i18nkeys.AuthorDeleteFailed, nil)
		return
	}

	response.Success(c, http.StatusOK, i18nkeys.AuthorDeletedSuccess, nil, nil)
}

// GetAuthor lấy thông tin chi tiết author
// @Summary Get author details
// @Tags Authors
// @Produce json
// @Param id path string true "Author ID"
// @Success 200 {object} response.StandardResponse{data=AuthorDetailResponse}
// @Failure 400 {object} response.StandardResponse
// @Failure 404 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/v1/authors/{id} [get]
func (h *Handler) GetAuthor(c *gin.Context) {
	// Get author ID from path
	idStr := c.Param("id")
	id, err := uuid.FromString(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", i18nkeys.AuthorInvalidID, nil)
		return
	}

	// Get author
	author, err := h.authorService.GetAuthorByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, pkgerrors.ErrAuthorNotFound) || errors.Is(err, pgx.ErrNoRows) {
			response.Error(c, http.StatusNotFound, "AUTHOR_NOT_FOUND", i18nkeys.AuthorNotFound, nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "GET_FAILED", i18nkeys.AuthorGetFailed, nil)
		return
	}

	// Map to response
	resp := mapToAuthorDetailResponse(author)

	response.Success(c, http.StatusOK, i18nkeys.AuthorGetSuccess, resp, nil)
}

// ListAuthors lấy danh sách authors
// @Summary List authors with pagination, search and sort
// @Tags Authors
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param search query string false "Search by name"
// @Param sort_by query string false "Sort by field (name, views, novels, created)"
// @Param sort_order query string false "Sort order (asc, desc)"
// @Param is_verified query bool false "Filter verified authors only"
// @Success 200 {object} response.StandardResponse{data=[]AuthorResponse}
// @Failure 400 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/v1/authors [get]
func (h *Handler) ListAuthors(c *gin.Context) {
	var req ListAuthorsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_FAILED", i18nkeys.ValidationFailed, err.Error())
		return
	}

	// Set defaults
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 {
		req.Limit = 20
	}

	// Get authors with pagination, search and sort
	authors, totalCount, err := h.authorService.ListAuthors(
		c.Request.Context(),
		req.Page,
		req.Limit,
		req.Search,
		req.SortBy,
		req.SortOrder,
		req.IsVerified,
	)
	if err != nil {
		h.logger.Error("Failed to list authors",
			zap.Error(err),
			zap.Int("page", req.Page),
			zap.Int("limit", req.Limit),
			zap.String("search", req.Search),
			zap.String("sort_by", req.SortBy),
			zap.String("sort_order", req.SortOrder),
			zap.Any("is_verified", req.IsVerified),
		)
		response.Error(c, http.StatusInternalServerError, "LIST_FAILED", i18nkeys.AuthorListFailed, nil)
		return
	}

	// Map to response format
	authorResponses := make([]AuthorResponse, len(authors))
	for i, author := range authors {
		authorResponses[i] = mapToAuthorResponse(author)
	}

	// Calculate pagination meta
	totalPages := (totalCount + req.Limit - 1) / req.Limit
	meta := &response.PaginationMeta{
		Page:       req.Page,
		Limit:      req.Limit,
		TotalItems: totalCount,
		TotalPages: totalPages,
	}

	response.Success(c, http.StatusOK, i18nkeys.AuthorListSuccess, authorResponses, meta)
}

// Helper function to map domain model to detail response
func mapToAuthorDetailResponse(author *domain.Author) AuthorDetailResponse {
	resp := AuthorDetailResponse{
		ID:            author.ID.String(),
		Name:          author.Name,
		Slug:          author.Slug,
		NovelCount:    author.NovelCount,
		TotalChapters: author.TotalChapters,
		TotalViews:    author.TotalViews,
		FollowerCount: author.FollowerCount,
		IsVerified:    author.IsVerified,
		CreatedAt:     author.CreatedAt.Format(timeutil.ISO8601Layout),
		UpdatedAt:     author.UpdatedAt.Format(timeutil.ISO8601Layout),
	}

	// Extract description from biography JSON
	if len(author.Biography) > 0 {
		var bio map[string]any
		if err := json.Unmarshal(author.Biography, &bio); err == nil {
			if text, ok := bio["text"].(string); ok && text != "" {
				resp.Description = &text
			}
		}
	}

	// Convert social links to string
	if len(author.SocialLinks) > 0 {
		socialLinksStr := string(author.SocialLinks)
		resp.SocialLinks = &socialLinksStr
	}

	if author.AvatarURL != nil {
		resp.AvatarURL = author.AvatarURL
	}

	return resp
}

// Helper function to map domain model to list response
func mapToAuthorResponse(author *domain.Author) AuthorResponse {
	resp := AuthorResponse{
		ID:         author.ID.String(),
		Name:       author.Name,
		Slug:       author.Slug,
		NovelCount: author.NovelCount,
		TotalViews: author.TotalViews,
		CreatedAt:  author.CreatedAt.Format(timeutil.ISO8601Layout),
	}

	// Extract description from biography JSON
	if len(author.Biography) > 0 {
		var bio map[string]any
		if err := json.Unmarshal(author.Biography, &bio); err == nil {
			if text, ok := bio["text"].(string); ok && text != "" {
				resp.Description = &text
			}
		}
	}

	return resp
}

// ListSelection lấy danh sách authors rút gọn (selection)
// @Summary List authors for selection (ID and Name only)
// @Tags Authors
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param search query string false "Search by name"
// @Success 200 {object} response.StandardResponse{data=[]SelectionResponse}
// @Failure 400 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/v1/authors/selection [get]
func (h *Handler) ListSelection(c *gin.Context) {
	var req ListAuthorsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_FAILED", i18nkeys.ValidationFailed, err.Error())
		return
	}

	// Set defaults
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 {
		req.Limit = 20
	}

	// Get authors selection
	authors, totalCount, err := h.authorService.ListSelection(
		c.Request.Context(),
		req.Page,
		req.Limit,
		req.Search,
	)
	if err != nil {
		h.logger.Error("Failed to list authors selection", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "LIST_FAILED", i18nkeys.AuthorListFailed, nil)
		return
	}

	// Map to response format
	type SelectionResponse struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	selectionResponses := make([]SelectionResponse, len(authors))
	for i, a := range authors {
		selectionResponses[i] = SelectionResponse{
			ID:   a.ID.String(),
			Name: a.Name,
		}
	}

	// Calculate pagination meta
	totalPages := (totalCount + int64(req.Limit) - 1) / int64(req.Limit)
	meta := &response.PaginationMeta{
		Page:       req.Page,
		Limit:      req.Limit,
		TotalItems: int(totalCount),
		TotalPages: int(totalPages),
	}

	response.Success(c, http.StatusOK, i18nkeys.AuthorListSuccess, selectionResponses, meta)
}


