package author

import (
	"encoding/json"
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
		response.Error(c, http.StatusBadRequest, "VALIDATION_FAILED", "validation.failed", err.Error())
		return
	}

	// Create author
	author, err := h.authorService.CreateAuthor(c.Request.Context(), req.Name, req.Biography, req.AvatarURL, req.SocialLinks)
	if err != nil {
		if errors.Is(err, pkgerrors.ErrInvalidInput) {
			response.Error(c, http.StatusBadRequest, "INVALID_INPUT", "author.invalid_input", nil)
			return
		}
		if errors.Is(err, pkgerrors.ErrSlugAlreadyExists) {
			response.Error(c, http.StatusConflict, "SLUG_EXISTS", "author.slug_already_exists", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "CREATE_FAILED", "author.create_failed", nil)
		return
	}

	// Map to response
	resp := mapToAuthorDetailResponse(author)

	response.Success(c, http.StatusCreated, "author.created_success", resp, nil)
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
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "author.invalid_id", nil)
		return
	}

	var req UpdateAuthorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_FAILED", "validation.failed", err.Error())
		return
	}

	// Update author
	author, err := h.authorService.UpdateAuthor(c.Request.Context(), id, req.Name, req.Biography, req.AvatarURL, req.SocialLinks)
	if err != nil {
		if errors.Is(err, pkgerrors.ErrAuthorNotFound) {
			response.Error(c, http.StatusNotFound, "AUTHOR_NOT_FOUND", "author.not_found", nil)
			return
		}
		if errors.Is(err, pkgerrors.ErrInvalidInput) {
			response.Error(c, http.StatusBadRequest, "INVALID_INPUT", "author.invalid_input", nil)
			return
		}
		if errors.Is(err, pkgerrors.ErrSlugAlreadyExists) {
			response.Error(c, http.StatusConflict, "SLUG_EXISTS", "author.slug_already_exists", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "UPDATE_FAILED", "author.update_failed", nil)
		return
	}

	// Map to response
	resp := mapToAuthorDetailResponse(author)

	response.Success(c, http.StatusOK, "author.updated_success", resp, nil)
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
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "author.invalid_id", nil)
		return
	}

	// Delete author
	err = h.authorService.DeleteAuthor(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, pkgerrors.ErrAuthorNotFound) {
			response.Error(c, http.StatusNotFound, "AUTHOR_NOT_FOUND", "author.not_found", nil)
			return
		}
		if errors.Is(err, pkgerrors.ErrAuthorInUse) {
			response.Error(c, http.StatusConflict, "AUTHOR_IN_USE", "author.in_use", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "DELETE_FAILED", "author.delete_failed", nil)
		return
	}

	response.Success(c, http.StatusOK, "author.deleted_success", nil, nil)
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
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "author.invalid_id", nil)
		return
	}

	// Get author
	author, err := h.authorService.GetAuthorByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, pkgerrors.ErrAuthorNotFound) || errors.Is(err, pgx.ErrNoRows) {
			response.Error(c, http.StatusNotFound, "AUTHOR_NOT_FOUND", "author.not_found", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "GET_FAILED", "author.get_failed", nil)
		return
	}

	// Map to response
	resp := mapToAuthorDetailResponse(author)

	response.Success(c, http.StatusOK, "author.get_success", resp, nil)
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
		response.Error(c, http.StatusInternalServerError, "LIST_FAILED", "author.list_failed", nil)
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

	response.Success(c, http.StatusOK, "author.list_success", authorResponses, meta)
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
		var bio map[string]interface{}
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
		NovelCount: author.NovelCount,
		TotalViews: author.TotalViews,
		CreatedAt:  author.CreatedAt.Format(timeutil.ISO8601Layout),
	}

	// Extract description from biography JSON
	if len(author.Biography) > 0 {
		var bio map[string]interface{}
		if err := json.Unmarshal(author.Biography, &bio); err == nil {
			if text, ok := bio["text"].(string); ok && text != "" {
				resp.Description = &text
			}
		}
	}

	return resp
}
