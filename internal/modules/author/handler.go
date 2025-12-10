package author

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"

	"system/internal/app/middleware"
	"system/internal/domain"
	authordto "system/internal/dto/author"
	pkgerrors "system/pkg/errors"
	"system/pkg/util/response"
	"system/pkg/util/timeutil"
)

type Handler struct {
	authorService AuthorService
	logger        *zap.Logger
}

func NewHandler(authorService AuthorService, logger *zap.Logger) *Handler {
	return &Handler{
		authorService: authorService,
		logger:        logger,
	}
}

// CreateAuthor tạo author mới
func (h *Handler) CreateAuthor(c *gin.Context) {
	var req authordto.CreateAuthorRequest
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

	author, err := h.authorService.CreateAuthor(c.Request.Context(), req.Name, req.Biography, req.AvatarURL, req.SocialLinks, userID)
	if err != nil {
		if appErr, ok := pkgerrors.AsAppError(err); ok {
			response.Error(c, appErr.StatusCode, appErr.ErrCode, appErr.I18nKey, nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "InternalError", I18nCreateFailed, nil)
		return
	}

	resp := mapToAuthorDetailResponse(author)
	response.Success(c, http.StatusCreated, I18nCreatedSuccess, resp, nil)
}

// UpdateAuthor cập nhật author
func (h *Handler) UpdateAuthor(c *gin.Context) {
	idStr := c.Param("identifier")
	id, err := uuid.FromString(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "BadRequest", I18nInvalidID, nil)
		return
	}

	var req authordto.UpdateAuthorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "ValidationFailed", I18nValidationFailed, err.Error())
		return
	}

	author, err := h.authorService.UpdateAuthor(c.Request.Context(), id, req.Name, req.Biography, req.AvatarURL, req.SocialLinks)
	if err != nil {
		if appErr, ok := pkgerrors.AsAppError(err); ok {
			response.Error(c, appErr.StatusCode, appErr.ErrCode, appErr.I18nKey, nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "InternalError", I18nUpdateFailed, nil)
		return
	}

	resp := mapToAuthorDetailResponse(author)
	response.Success(c, http.StatusOK, I18nUpdatedSuccess, resp, nil)
}

// DeleteAuthor xóa author
func (h *Handler) DeleteAuthor(c *gin.Context) {
	idStr := c.Param("identifier")
	id, err := uuid.FromString(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "BadRequest", I18nInvalidID, nil)
		return
	}

	err = h.authorService.DeleteAuthor(c.Request.Context(), id)
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

// GetAuthor lấy thông tin chi tiết author
func (h *Handler) GetAuthor(c *gin.Context) {
	idStr := c.Param("identifier")
	id, err := uuid.FromString(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "BadRequest", I18nInvalidID, nil)
		return
	}

	author, err := h.authorService.GetAuthorByID(c.Request.Context(), id)
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

	resp := mapToAuthorDetailResponse(author)
	response.Success(c, http.StatusOK, I18nGetSuccess, resp, nil)
}

// ListAuthors lấy danh sách authors
func (h *Handler) ListAuthors(c *gin.Context) {
	var req authordto.ListAuthorsRequest
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
		h.logger.Error("Failed to list authors", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "InternalError", I18nListFailed, nil)
		return
	}

	authorResponses := make([]authordto.AuthorResponse, len(authors))
	for i, author := range authors {
		authorResponses[i] = mapToAuthorResponse(author)
	}

	totalPages := (totalCount + req.Limit - 1) / req.Limit
	meta := &response.PaginationMeta{
		Page:       req.Page,
		Limit:      req.Limit,
		TotalItems: totalCount,
		TotalPages: totalPages,
	}

	response.Success(c, http.StatusOK, I18nListSuccess, authorResponses, meta)
}

func mapToAuthorDetailResponse(author *domain.Author) authordto.AuthorDetailResponse {
	resp := authordto.AuthorDetailResponse{
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

	if len(author.Biography) > 0 {
		var bio map[string]any
		if err := json.Unmarshal(author.Biography, &bio); err == nil {
			if text, ok := bio["text"].(string); ok && text != "" {
				resp.Description = &text
			}
		}
	}

	if len(author.SocialLinks) > 0 {
		socialLinksStr := string(author.SocialLinks)
		resp.SocialLinks = &socialLinksStr
	}

	if author.AvatarURL != nil {
		resp.AvatarURL = author.AvatarURL
	}

	return resp
}

func mapToAuthorResponse(author *domain.Author) authordto.AuthorResponse {
	resp := authordto.AuthorResponse{
		ID:         author.ID.String(),
		Name:       author.Name,
		Slug:       author.Slug,
		NovelCount: author.NovelCount,
		TotalViews: author.TotalViews,
		CreatedAt:  author.CreatedAt.Format(timeutil.ISO8601Layout),
	}

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

// ListSelection lấy danh sách authors rút gọn
func (h *Handler) ListSelection(c *gin.Context) {
	var req authordto.ListAuthorsRequest
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

	authors, totalCount, err := h.authorService.ListSelection(
		c.Request.Context(),
		req.Page,
		req.Limit,
		req.Search,
	)
	if err != nil {
		h.logger.Error("Failed to list authors selection", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "InternalError", I18nListFailed, nil)
		return
	}

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

	totalPages := (totalCount + int64(req.Limit) - 1) / int64(req.Limit)
	meta := &response.PaginationMeta{
		Page:       req.Page,
		Limit:      req.Limit,
		TotalItems: int(totalCount),
		TotalPages: int(totalPages),
	}

	response.Success(c, http.StatusOK, I18nListSuccess, selectionResponses, meta)
}

// MergeAuthor gộp authors
func (h *Handler) MergeAuthor(c *gin.Context) {
	var req authordto.MergeAuthorRequest
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

	err = h.authorService.MergeAuthors(c.Request.Context(), req.TargetID, req.SourceIDs, userID)
	if err != nil {
		if appErr, ok := pkgerrors.AsAppError(err); ok {
			response.Error(c, appErr.StatusCode, appErr.ErrCode, appErr.I18nKey, nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "InternalError", I18nMergeFailed, err.Error())
		return
	}

	response.Success(c, http.StatusOK, I18nMergeSuccess, nil, nil)
}

// PreviewMergeAuthor xem trước kết quả gộp authors
func (h *Handler) PreviewMergeAuthor(c *gin.Context) {
	var req authordto.MergeAuthorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "ValidationFailed", I18nValidationFailed, err.Error())
		return
	}

	preview, err := h.authorService.PreviewMergeAuthors(c.Request.Context(), req.TargetID, req.SourceIDs)
	if err != nil {
		if appErr, ok := pkgerrors.AsAppError(err); ok {
			response.Error(c, appErr.StatusCode, appErr.ErrCode, appErr.I18nKey, nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "InternalError", I18nPreviewFailed, err.Error())
		return
	}

	affectedNovels := make([]authordto.AffectedNovel, len(preview))
	for i, novel := range preview {
		affectedNovels[i] = authordto.AffectedNovel{
			ID:            novel.ID.String(),
			Title:         novel.Title,
			Slug:          novel.Slug,
			CoverImageURL: novel.CoverImageURL,
		}
	}

	resp := authordto.PreviewMergeAuthorResponse{
		AffectedNovels: affectedNovels,
		SourceAuthors:  nil,
	}

	response.Success(c, http.StatusOK, I18nPreviewSuccess, resp, nil)
}
