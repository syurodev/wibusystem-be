// ============================================================================
// Artist Handler
// ============================================================================
//
// Handler này xử lý các HTTP requests cho Artist module.
//
// Luồng xử lý request:
//
//   HTTP Request
//        │
//        ▼
//   ┌─────────────────┐
//   │    Handler      │  1. Parse & validate request (BindJSON/BindQuery)
//   │  (handler.go)   │  2. Extract user context từ JWT token
//   └────────┬────────┘  3. Call UseCase với input DTO
//            │
//            ▼
//   ┌─────────────────┐
//   │    UseCase      │  4. Orchestrate business logic
//   │ (usecase_*.go)  │  5. Call Service methods
//   └────────┬────────┘
//            │
//            ▼
//   ┌─────────────────┐
//   │    Service      │  6. Business validation
//   │  (service.go)   │  7. Domain logic (slug generation, JSON handling)
//   └────────┬────────┘  8. Call Repository
//            │
//            ▼
//   ┌─────────────────┐
//   │   Repository    │  9. Execute SQL queries (embedded từ queries/)
//   │(repository.go)  │  10. Return domain models
//   └─────────────────┘
//
// Endpoints:
//   GET    /api/v1/artists           - ListArtists (public)
//   GET    /api/v1/artists/selection - ListSelection (public)
//   GET    /api/v1/artists/:id       - GetArtist (public)
//   POST   /api/v1/artists           - CreateArtist (auth required)
//   PUT    /api/v1/artists/:id       - UpdateArtist (auth required)
//   DELETE /api/v1/artists/:id       - DeleteArtist (auth required)
//   POST   /api/v1/artists/merge     - MergeArtist (auth required)
//   POST   /api/v1/artists/merge/preview - PreviewMergeArtist (auth required)
//
// ============================================================================

package artist

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid/v5"
	"go.uber.org/zap"

	"system/internal/app/middleware"
	"system/internal/domain"
	artistdto "system/internal/dto/artist"
	ent "system/internal/ent/generated"
	pkgerrors "system/pkg/errors"
	"system/pkg/util/jsonutil"
	"system/pkg/util/response"
	"system/pkg/util/timeutil"
)

type Handler struct {
	createArtistUC CreateArtistUseCase
	updateArtistUC UpdateArtistUseCase
	deleteArtistUC DeleteArtistUseCase
	getArtistUC    GetArtistUseCase
	listArtistsUC  ListArtistsUseCase
	listSelectUC   ListSelectionUseCase
	mergeArtistsUC MergeArtistsUseCase
	previewMergeUC PreviewMergeUseCase
	logger         *zap.Logger
}

func NewHandler(
	createArtistUC CreateArtistUseCase,
	updateArtistUC UpdateArtistUseCase,
	deleteArtistUC DeleteArtistUseCase,
	getArtistUC GetArtistUseCase,
	listArtistsUC ListArtistsUseCase,
	listSelectUC ListSelectionUseCase,
	mergeArtistsUC MergeArtistsUseCase,
	previewMergeUC PreviewMergeUseCase,
	logger *zap.Logger,
) *Handler {
	return &Handler{
		createArtistUC: createArtistUC,
		updateArtistUC: updateArtistUC,
		deleteArtistUC: deleteArtistUC,
		getArtistUC:    getArtistUC,
		listArtistsUC:  listArtistsUC,
		listSelectUC:   listSelectUC,
		mergeArtistsUC: mergeArtistsUC,
		previewMergeUC: previewMergeUC,
		logger:         logger,
	}
}

// CreateArtist tạo artist mới
// @Summary Create a new artist
// @Tags Artists
// @Accept json
// @Produce json
// @Param request body CreateArtistRequest true "Create Artist Request"
// @Success 201 {object} response.StandardResponse{data=ArtistDetailResponse}
// @Failure 400 {object} response.StandardResponse
// @Failure 409 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/v1/artists [post]
func (h *Handler) CreateArtist(c *gin.Context) {
	var req artistdto.CreateArtistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "ValidationFailed", I18nValidationFailed, err.Error())
		return
	}

	// Get user ID from token
	userIDStr, exists := middleware.GetUserID(c)
	if !exists {
		response.Error(c, http.StatusUnauthorized, "Unauthorized", I18nAuthUnauthorized, nil)
		return
	}

	createdBy, err := uuid.FromString(userIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "InvalidUserID", I18nAuthInvalidUserID, nil)
		return
	}

	// Create artist
	artist, err := h.createArtistUC.Execute(c.Request.Context(), CreateArtistInput{
		Name:            req.Name,
		Biography:       req.Biography,
		AvatarURL:       req.AvatarURL,
		SocialLinksJSON: req.SocialLinks,
		Specialization:  req.Specialization,
		PortfolioURL:    req.PortfolioURL,
		CreatedBy:       createdBy,
	})
	if err != nil {
		if appErr, ok := pkgerrors.AsAppError(err); ok {
			response.Error(c, appErr.StatusCode, appErr.ErrCode, appErr.I18nKey, nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "InternalError", I18nCreateFailed, nil)
		return
	}

	// Map to response
	resp := mapToArtistDetailResponse(artist)

	response.Success(c, http.StatusCreated, I18nCreatedSuccess, resp, nil)
}

// UpdateArtist cập nhật artist
// @Summary Update an artist
// @Tags Artists
// @Accept json
// @Produce json
// @Param id path string true "Artist ID"
// @Param request body UpdateArtistRequest true "Update Artist Request"
// @Success 200 {object} response.StandardResponse{data=ArtistDetailResponse}
// @Failure 400 {object} response.StandardResponse
// @Failure 404 {object} response.StandardResponse
// @Failure 409 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/v1/artists/{id} [put]
func (h *Handler) UpdateArtist(c *gin.Context) {
	// Get artist ID from path
	idStr := c.Param("identifier")
	id, err := uuid.FromString(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "BadRequest", I18nInvalidID, nil)
		return
	}

	var req artistdto.UpdateArtistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "ValidationFailed", I18nValidationFailed, err.Error())
		return
	}

	// Update artist
	artist, err := h.updateArtistUC.Execute(c.Request.Context(), UpdateArtistInput{
		ID:              id,
		Name:            req.Name,
		Biography:       req.Biography,
		AvatarURL:       req.AvatarURL,
		SocialLinksJSON: req.SocialLinks,
		Specialization:  req.Specialization,
		PortfolioURL:    req.PortfolioURL,
	})
	if err != nil {
		if appErr, ok := pkgerrors.AsAppError(err); ok {
			response.Error(c, appErr.StatusCode, appErr.ErrCode, appErr.I18nKey, nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "InternalError", I18nUpdateFailed, nil)
		return
	}

	// Map to response
	resp := mapToArtistDetailResponse(artist)

	response.Success(c, http.StatusOK, I18nUpdatedSuccess, resp, nil)
}

// DeleteArtist xóa artist
// @Summary Delete an artist
// @Tags Artists
// @Produce json
// @Param id path string true "Artist ID"
// @Success 200 {object} response.StandardResponse
// @Failure 400 {object} response.StandardResponse
// @Failure 404 {object} response.StandardResponse
// @Failure 409 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/v1/artists/{id} [delete]
func (h *Handler) DeleteArtist(c *gin.Context) {
	// Get artist ID from path
	idStr := c.Param("identifier")
	id, err := uuid.FromString(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "BadRequest", I18nInvalidID, nil)
		return
	}

	// Delete artist
	err = h.deleteArtistUC.Execute(c.Request.Context(), DeleteArtistInput{
		ID: id,
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

// GetArtist lấy thông tin chi tiết artist
// @Summary Get artist details
// @Tags Artists
// @Produce json
// @Param id path string true "Artist ID"
// @Success 200 {object} response.StandardResponse{data=ArtistDetailResponse}
// @Failure 400 {object} response.StandardResponse
// @Failure 404 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/v1/artists/{id} [get]
func (h *Handler) GetArtist(c *gin.Context) {
	// Get artist ID or slug from path
	idOrSlug := c.Param("identifier")

	// Get artist
	artist, err := h.getArtistUC.Execute(c.Request.Context(), GetArtistInput{
		IDOrSlug: idOrSlug,
	})
	if err != nil {
		if appErr, ok := pkgerrors.AsAppError(err); ok {
			response.Error(c, appErr.StatusCode, appErr.ErrCode, appErr.I18nKey, nil)
			return
		}
		if ent.IsNotFound(err) {
			response.Error(c, http.StatusNotFound, "NotFound", I18nNotFound, nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "InternalError", I18nGetFailed, nil)
		return
	}

	// Map to response
	resp := mapToArtistDetailResponse(artist)

	response.Success(c, http.StatusOK, I18nGetSuccess, resp, nil)
}

// ListArtists lấy danh sách artists
// @Summary List artists with pagination, search and sort
// @Tags Artists
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param search query string false "Search by name"
// @Param sort_by query string false "Sort by field (name, novels, created)"
// @Param sort_order query string false "Sort order (asc, desc)"
// @Param specialization query string false "Filter by specialization"
// @Param is_verified query bool false "Filter verified artists only"
// @Success 200 {object} response.StandardResponse{data=[]ArtistResponse}
// @Failure 400 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/v1/artists [get]
func (h *Handler) ListArtists(c *gin.Context) {
	var req artistdto.ListArtistsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "ValidationFailed", I18nValidationFailed, err.Error())
		return
	}

	// Set defaults
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 {
		req.Limit = 20
	}

	// Get artists with pagination, search and sort
	artists, totalCount, err := h.listArtistsUC.Execute(c.Request.Context(), ListArtistsInput{
		Page:           req.Page,
		Limit:          req.Limit,
		Search:         req.Search,
		SortBy:         req.SortBy,
		SortOrder:      req.SortOrder,
		Specialization: req.Specialization,
		IsVerified:     req.IsVerified,
	})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "InternalError", I18nListFailed, nil)
		return
	}

	// Map to response format
	artistResponses := make([]artistdto.ArtistResponse, len(artists))
	for i, artist := range artists {
		artistResponses[i] = mapToArtistResponse(artist)
	}

	// Calculate pagination meta
	totalPages := (totalCount + req.Limit - 1) / req.Limit
	meta := &response.PaginationMeta{
		Page:       req.Page,
		Limit:      req.Limit,
		TotalItems: totalCount,
		TotalPages: totalPages,
	}

	response.Success(c, http.StatusOK, I18nListSuccess, artistResponses, meta)
}

// Helper function to map domain model to detail response
func mapToArtistDetailResponse(artist *domain.Artist) artistdto.ArtistDetailResponse {
	resp := artistdto.ArtistDetailResponse{
		ID:             artist.ID.String(),
		Name:           artist.Name,
		Slug:           artist.Slug,
		Specialization: artist.Specialization,
		PortfolioURL:   artist.PortfolioURL,
		NovelCount:     artist.NovelCount,
		ArtworkCount:   artist.ArtworkCount,
		FollowerCount:  artist.FollowerCount,
		IsVerified:     artist.IsVerified,
		CreatedAt:      artist.CreatedAt.Format(timeutil.ISO8601Layout),
		UpdatedAt:      artist.UpdatedAt.Format(timeutil.ISO8601Layout),
	}

	// Extract description from biography JSON
	resp.Description = jsonutil.ExtractTextField(artist.Biography)

	// Convert social links to string
	if len(artist.SocialLinks) > 0 {
		socialLinksStr := string(artist.SocialLinks)
		resp.SocialLinks = &socialLinksStr
	}

	if artist.AvatarURL != nil {
		resp.AvatarURL = artist.AvatarURL
	}

	return resp
}

// Helper function to map domain model to list response
func mapToArtistResponse(artist *domain.Artist) artistdto.ArtistResponse {
	resp := artistdto.ArtistResponse{
		ID:             artist.ID.String(),
		Name:           artist.Name,
		Slug:           artist.Slug,
		NovelCount:     artist.NovelCount,
		Specialization: artist.Specialization,
		PortfolioURL:   artist.PortfolioURL,
		CreatedAt:      artist.CreatedAt.Format(timeutil.ISO8601Layout),
	}

	// Extract description from biography JSON
	resp.Description = jsonutil.ExtractTextField(artist.Biography)

	return resp
}

// ListSelection lấy danh sách artists rút gọn (selection)
// @Summary List artists for selection (ID and Name only)
// @Tags Artists
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param search query string false "Search by name"
// @Success 200 {object} response.StandardResponse{data=[]SelectionResponse}
// @Failure 400 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/v1/artists/selection [get]
func (h *Handler) ListSelection(c *gin.Context) {
	var req artistdto.ListArtistsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "ValidationFailed", I18nValidationFailed, err.Error())
		return
	}

	// Set defaults
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 {
		req.Limit = 20
	}

	// Get artists selection
	artists, totalCount, err := h.listSelectUC.Execute(c.Request.Context(), ListSelectionInput{
		Page:   req.Page,
		Limit:  req.Limit,
		Search: req.Search,
	})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "InternalError", I18nListFailed, nil)
		return
	}

	// Map to response format using DTO type
	selectionResponses := make([]artistdto.SelectionResponse, len(artists))
	for i, a := range artists {
		selectionResponses[i] = artistdto.SelectionResponse{
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

	response.Success(c, http.StatusOK, I18nListSuccess, selectionResponses, meta)
}

// MergeArtist gộp artists
// @Summary Merge artists
// @Tags Artists
// @Accept json
// @Produce json
// @Param request body MergeArtistRequest true "Merge Artist Request"
// @Success 200 {object} response.StandardResponse
// @Failure 400 {object} response.StandardResponse
// @Failure 404 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/v1/artists/merge [post]
func (h *Handler) MergeArtist(c *gin.Context) {
	var req artistdto.MergeArtistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "ValidationFailed", I18nValidationFailed, err.Error())
		return
	}

	// Get user ID from token
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

	err = h.mergeArtistsUC.Execute(c.Request.Context(), MergeArtistsInput{
		TargetID:  req.TargetID,
		SourceIDs: req.SourceIDs,
		MergedBy:  userID,
	})
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

// PreviewMergeArtist xem trước kết quả gộp artists
// @Summary Preview merge artists
// @Tags Artists
// @Accept json
// @Produce json
// @Param request body MergeArtistRequest true "Merge Artist Request"
// @Success 200 {object} response.StandardResponse{data=PreviewMergeArtistResponse}
// @Failure 400 {object} response.StandardResponse
// @Failure 404 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/v1/artists/merge/preview [post]
func (h *Handler) PreviewMergeArtist(c *gin.Context) {
	var req artistdto.MergeArtistRequest
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

		response.Error(c, http.StatusInternalServerError, "InternalError", I18nPreviewFailed, err.Error())
		return
	}

	// Transform to response DTO
	affectedNovels := make([]artistdto.AffectedNovel, len(preview))
	for i, novel := range preview {
		affectedNovels[i] = artistdto.AffectedNovel{
			ID:            novel.ID.String(),
			Title:         novel.Title,
			Slug:          novel.Slug,
			CoverImageURL: novel.CoverImageURL,
		}
	}

	resp := artistdto.PreviewMergeArtistResponse{
		AffectedNovels: affectedNovels,
		SourceArtists:  nil, // Optional
	}

	response.Success(c, http.StatusOK, I18nPreviewSuccess, resp, nil)
}
