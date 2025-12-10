package novel_volume

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"

	"system/internal/app/middleware"
	"system/internal/domain"
	volumedto "system/internal/dto/novel_volume"
	pkgerrors "system/pkg/errors"
	"system/pkg/util/response"
)

// Handler handles volume-related HTTP requests
type Handler struct {
	volumeService VolumeService
}

// NewHandler creates a new volume Handler instance
func NewHandler(volumeService VolumeService) *Handler {
	return &Handler{
		volumeService: volumeService,
	}
}

// CreateVolume creates a new volume
func (h *Handler) CreateVolume(c *gin.Context) {
	userIDStr, exists := middleware.GetUserID(c)
	if !exists {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "auth.unauthorized", nil)
		return
	}

	userID, uuidErr := uuid.FromString(userIDStr)
	if uuidErr != nil {
		response.Error(c, http.StatusUnauthorized, "INVALID_USER_ID", "auth.invalid_user_id", nil)
		return
	}

	var req volumedto.CreateVolumeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_FAILED", "validation.failed", err.Error())
		return
	}

	novelIDStr := c.Param("id")
	novelID, err := uuid.FromString(novelIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_NOVEL_ID", I18nNovelNotFound, nil)
		return
	}

	volume, err := h.volumeService.CreateVolume(
		c.Request.Context(),
		novelID,
		req.Title,
		req.Description,
		req.CoverImageURL,
		req.DisplayOrder,
		req.IsPublished,
		userID,
	)
	if err != nil {
		if appErr, ok := pkgerrors.AsAppError(err); ok {
			response.Error(c, appErr.StatusCode, appErr.ErrCode, appErr.I18nKey, nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "CREATE_FAILED", I18nCreateFailed, nil)
		return
	}

	resp := mapToVolumeDetailResponse(volume)
	response.Success(c, http.StatusCreated, I18nCreatedSuccess, resp, nil)
}

// UpdateVolume updates a volume
func (h *Handler) UpdateVolume(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.FromString(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", I18nInvalidID, nil)
		return
	}

	var req volumedto.UpdateVolumeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_FAILED", "validation.failed", err.Error())
		return
	}

	userIDStr, exists := middleware.GetUserID(c)
	changedBy := uuid.Nil
	if exists {
		changedBy, _ = uuid.FromString(userIDStr)
	}

	requestContext := extractRequestContext(c)

	volume, err := h.volumeService.UpdateVolume(
		c.Request.Context(),
		id,
		req.VolumeNumber,
		req.Title,
		req.Description,
		req.CoverImageURL,
		req.DisplayOrder,
		req.IsPublished,
		changedBy,
		requestContext,
	)
	if err != nil {
		if appErr, ok := pkgerrors.AsAppError(err); ok {
			response.Error(c, appErr.StatusCode, appErr.ErrCode, appErr.I18nKey, nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "UPDATE_FAILED", I18nUpdateFailed, nil)
		return
	}

	resp := mapToVolumeDetailResponse(volume)
	response.Success(c, http.StatusOK, I18nUpdatedSuccess, resp, nil)
}

// DeleteVolume deletes a volume
func (h *Handler) DeleteVolume(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.FromString(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", I18nInvalidID, nil)
		return
	}

	err = h.volumeService.DeleteVolume(c.Request.Context(), id)
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

// GetVolume retrieves volume details
func (h *Handler) GetVolume(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.FromString(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", I18nInvalidID, nil)
		return
	}

	volume, err := h.volumeService.GetVolumeByID(c.Request.Context(), id)
	if err != nil {
		if appErr, ok := pkgerrors.AsAppError(err); ok {
			response.Error(c, appErr.StatusCode, appErr.ErrCode, appErr.I18nKey, nil)
			return
		}
		if err == pgx.ErrNoRows {
			response.Error(c, http.StatusNotFound, "VOLUME_NOT_FOUND", I18nNotFound, nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "GET_FAILED", I18nGetFailed, nil)
		return
	}

	resp := mapToVolumeDetailResponse(volume)
	response.Success(c, http.StatusOK, I18nGetSuccess, resp, nil)
}

// ListVolumesByNovel retrieves all volumes for a novel
func (h *Handler) ListVolumesByNovel(c *gin.Context) {
	novelIDStr := c.Param("id")
	novelID, err := uuid.FromString(novelIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_NOVEL_ID", I18nNovelNotFound, nil)
		return
	}

	var req volumedto.ListVolumesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_FAILED", "validation.failed", err.Error())
		return
	}

	volumes, err := h.volumeService.GetVolumesByNovelID(c.Request.Context(), novelID, req.PublishedOnly)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "LIST_FAILED", I18nListFailed, nil)
		return
	}

	volumeResponses := make([]volumedto.VolumeResponse, len(volumes))
	for i, volume := range volumes {
		volumeResponses[i] = mapToVolumeResponse(volume)
	}

	response.Success(c, http.StatusOK, I18nListSuccess, volumeResponses, nil)
}

// UpdateDisplayOrder updates the display order of a volume
func (h *Handler) UpdateDisplayOrder(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.FromString(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", I18nInvalidID, nil)
		return
	}

	var req volumedto.UpdateDisplayOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_FAILED", "validation.failed", err.Error())
		return
	}

	err = h.volumeService.UpdateDisplayOrder(c.Request.Context(), id, req.DisplayOrder)
	if err != nil {
		if appErr, ok := pkgerrors.AsAppError(err); ok {
			response.Error(c, appErr.StatusCode, appErr.ErrCode, appErr.I18nKey, nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "UPDATE_FAILED", I18nReorderFailed, nil)
		return
	}

	response.Success(c, http.StatusOK, I18nReorderSuccess, nil, nil)
}

// PublishVolume publishes a volume
func (h *Handler) PublishVolume(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.FromString(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", I18nInvalidID, nil)
		return
	}

	userIDStr, exists := middleware.GetUserID(c)
	changedBy := uuid.Nil
	if exists {
		changedBy, _ = uuid.FromString(userIDStr)
	}

	requestContext := extractRequestContext(c)

	err = h.volumeService.PublishVolume(c.Request.Context(), id, changedBy, requestContext)
	if err != nil {
		if appErr, ok := pkgerrors.AsAppError(err); ok {
			response.Error(c, appErr.StatusCode, appErr.ErrCode, appErr.I18nKey, nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "PUBLISH_FAILED", "volume.publish_failed", nil)
		return
	}

	response.Success(c, http.StatusOK, "volume.published_success", nil, nil)
}

// UnpublishVolume unpublishes a volume
func (h *Handler) UnpublishVolume(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.FromString(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", I18nInvalidID, nil)
		return
	}

	userIDStr, exists := middleware.GetUserID(c)
	changedBy := uuid.Nil
	if exists {
		changedBy, _ = uuid.FromString(userIDStr)
	}

	requestContext := extractRequestContext(c)

	err = h.volumeService.UnpublishVolume(c.Request.Context(), id, changedBy, requestContext)
	if err != nil {
		if appErr, ok := pkgerrors.AsAppError(err); ok {
			response.Error(c, appErr.StatusCode, appErr.ErrCode, appErr.I18nKey, nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "UNPUBLISH_FAILED", "volume.unpublish_failed", nil)
		return
	}

	response.Success(c, http.StatusOK, "volume.unpublished_success", nil, nil)
}

// Helper function to map domain model to detail response

func mapToVolumeDetailResponse(volume *domain.NovelVolume) volumedto.VolumeDetailResponse {
	resp := volumedto.VolumeDetailResponse{
		ID:            volume.ID.String(),
		NovelID:       volume.NovelID.String(),
		VolumeNumber:  volume.VolumeNumber,
		Title:         volume.Title,
		Slug:          volume.Slug,
		Description:   volume.Description,
		CoverImageURL: volume.CoverImageURL,
		ChapterCount:  volume.ChapterCount,
		WordCount:     volume.WordCount,
		DisplayOrder:  volume.DisplayOrder,
		IsPublished:   volume.IsPublished,
		CreatedAt:     volume.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:     volume.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	if volume.PublishedAt != nil {
		publishedAt := volume.PublishedAt.Format("2006-01-02T15:04:05Z07:00")
		resp.PublishedAt = &publishedAt
	}

	return resp
}

// Helper function to map domain model to list response
func mapToVolumeResponse(volume *domain.NovelVolume) volumedto.VolumeResponse {
	resp := volumedto.VolumeResponse{
		ID:            volume.ID.String(),
		NovelID:       volume.NovelID.String(),
		VolumeNumber:  volume.VolumeNumber,
		Title:         volume.Title,
		Slug:          volume.Slug,
		CoverImageURL: volume.CoverImageURL,
		ChapterCount:  volume.ChapterCount,
		WordCount:     volume.WordCount,
		DisplayOrder:  volume.DisplayOrder,
		IsPublished:   volume.IsPublished,
		CreatedAt:     volume.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:     volume.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	if volume.Novel != nil {
		resp.NovelTitle = volume.Novel.Title
	}

	if volume.PublishedAt != nil {
		publishedAt := volume.PublishedAt.Format("2006-01-02T15:04:05Z07:00")
		resp.PublishedAt = &publishedAt
	}

	return resp
}

// Helper function to extract request context for history logging
func extractRequestContext(c *gin.Context) map[string]any {
	return map[string]any{
		"request_id": c.GetHeader("X-Request-ID"),
		"ip_address": c.ClientIP(),
		"user_agent": c.GetHeader("User-Agent"),
	}
}
