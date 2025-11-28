package volume

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"

	"system/internal/domain"
	"system/internal/pkg/service"
	pkgerrors "system/pkg/errors"
	"system/pkg/util/response"
)

type Handler struct {
	volumeService *service.VolumeService
}

func NewHandler(volumeService *service.VolumeService) *Handler {
	return &Handler{
		volumeService: volumeService,
	}
}

// CreateVolume creates a new volume
// @Summary Create a new volume
// @Tags Volumes
// @Accept json
// @Produce json
// @Param request body CreateVolumeRequest true "Create Volume Request"
// @Success 201 {object} response.StandardResponse{data=VolumeDetailResponse}
// @Failure 400 {object} response.StandardResponse
// @Failure 409 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/v1/volumes [post]
func (h *Handler) CreateVolume(c *gin.Context) {
	var req CreateVolumeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_FAILED", "validation.failed", err.Error())
		return
	}

	// Parse novel ID
	novelID, err := uuid.FromString(req.NovelID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_NOVEL_ID", "novel.invalid_id", nil)
		return
	}

	// Create volume
	volume, err := h.volumeService.CreateVolume(
		c.Request.Context(),
		novelID,
		req.VolumeNumber,
		req.Title,
		req.Description,
		req.CoverImageURL,
		req.DisplayOrder,
		req.IsPublished,
	)
	if err != nil {
		if errors.Is(err, pkgerrors.ErrInvalidInput) {
			response.Error(c, http.StatusBadRequest, "INVALID_INPUT", "volume.invalid_input", nil)
			return
		}
		if errors.Is(err, pkgerrors.ErrVolumeNumberExists) {
			response.Error(c, http.StatusConflict, "VOLUME_NUMBER_EXISTS", "volume.number_already_exists", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "CREATE_FAILED", "volume.create_failed", nil)
		return
	}

	// Map to response
	resp := mapToVolumeDetailResponse(volume)

	response.Success(c, http.StatusCreated, "volume.created_success", resp, nil)
}

// UpdateVolume updates a volume
// @Summary Update a volume
// @Tags Volumes
// @Accept json
// @Produce json
// @Param id path string true "Volume ID"
// @Param request body UpdateVolumeRequest true "Update Volume Request"
// @Success 200 {object} response.StandardResponse{data=VolumeDetailResponse}
// @Failure 400 {object} response.StandardResponse
// @Failure 404 {object} response.StandardResponse
// @Failure 409 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/v1/volumes/{id} [put]
func (h *Handler) UpdateVolume(c *gin.Context) {
	// Get volume ID from path
	idStr := c.Param("id")
	id, err := uuid.FromString(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "volume.invalid_id", nil)
		return
	}

	var req UpdateVolumeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_FAILED", "validation.failed", err.Error())
		return
	}

	// TODO: Get actual user ID from authentication context
	// For now, use a system user ID placeholder
	systemUserID := uuid.Must(uuid.FromString("00000000-0000-0000-0000-000000000000"))

	// Extract request context for history
	requestContext := extractRequestContext(c)

	// Update volume
	volume, err := h.volumeService.UpdateVolume(
		c.Request.Context(),
		id,
		req.VolumeNumber,
		req.Title,
		req.Description,
		req.CoverImageURL,
		req.DisplayOrder,
		req.IsPublished,
		systemUserID,
		requestContext,
	)
	if err != nil {
		if errors.Is(err, pkgerrors.ErrVolumeNotFound) {
			response.Error(c, http.StatusNotFound, "VOLUME_NOT_FOUND", "volume.not_found", nil)
			return
		}
		if errors.Is(err, pkgerrors.ErrInvalidInput) {
			response.Error(c, http.StatusBadRequest, "INVALID_INPUT", "volume.invalid_input", nil)
			return
		}
		if errors.Is(err, pkgerrors.ErrVolumeNumberExists) {
			response.Error(c, http.StatusConflict, "VOLUME_NUMBER_EXISTS", "volume.number_already_exists", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "UPDATE_FAILED", "volume.update_failed", nil)
		return
	}

	// Map to response
	resp := mapToVolumeDetailResponse(volume)

	response.Success(c, http.StatusOK, "volume.updated_success", resp, nil)
}

// DeleteVolume deletes a volume
// @Summary Delete a volume
// @Tags Volumes
// @Produce json
// @Param id path string true "Volume ID"
// @Success 200 {object} response.StandardResponse
// @Failure 400 {object} response.StandardResponse
// @Failure 404 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/v1/volumes/{id} [delete]
func (h *Handler) DeleteVolume(c *gin.Context) {
	// Get volume ID from path
	idStr := c.Param("id")
	id, err := uuid.FromString(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "volume.invalid_id", nil)
		return
	}

	// Delete volume
	err = h.volumeService.DeleteVolume(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, pkgerrors.ErrVolumeNotFound) {
			response.Error(c, http.StatusNotFound, "VOLUME_NOT_FOUND", "volume.not_found", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "DELETE_FAILED", "volume.delete_failed", nil)
		return
	}

	response.Success(c, http.StatusOK, "volume.deleted_success", nil, nil)
}

// GetVolume retrieves volume details
// @Summary Get volume details
// @Tags Volumes
// @Produce json
// @Param id path string true "Volume ID"
// @Success 200 {object} response.StandardResponse{data=VolumeDetailResponse}
// @Failure 400 {object} response.StandardResponse
// @Failure 404 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/v1/volumes/{id} [get]
func (h *Handler) GetVolume(c *gin.Context) {
	// Get volume ID from path
	idStr := c.Param("id")
	id, err := uuid.FromString(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "volume.invalid_id", nil)
		return
	}

	// Get volume
	volume, err := h.volumeService.GetVolumeByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, pkgerrors.ErrVolumeNotFound) || errors.Is(err, pgx.ErrNoRows) {
			response.Error(c, http.StatusNotFound, "VOLUME_NOT_FOUND", "volume.not_found", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "GET_FAILED", "volume.get_failed", nil)
		return
	}

	// Map to response
	resp := mapToVolumeDetailResponse(volume)

	response.Success(c, http.StatusOK, "volume.get_success", resp, nil)
}

// ListVolumesByNovel retrieves all volumes for a novel
// @Summary List volumes by novel ID
// @Tags Volumes
// @Produce json
// @Param novel_id path string true "Novel ID"
// @Param published_only query bool false "Only published volumes"
// @Success 200 {object} response.StandardResponse{data=[]VolumeResponse}
// @Failure 400 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/v1/novels/{novel_id}/volumes [get]
func (h *Handler) ListVolumesByNovel(c *gin.Context) {
	// Get novel ID from path
	novelIDStr := c.Param("id")
	novelID, err := uuid.FromString(novelIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_NOVEL_ID", "novel.invalid_id", nil)
		return
	}

	var req ListVolumesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_FAILED", "validation.failed", err.Error())
		return
	}

	// Get volumes
	volumes, err := h.volumeService.GetVolumesByNovelID(c.Request.Context(), novelID, req.PublishedOnly)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "LIST_FAILED", "volume.list_failed", nil)
		return
	}

	// Map to response format
	volumeResponses := make([]VolumeResponse, len(volumes))
	for i, volume := range volumes {
		volumeResponses[i] = mapToVolumeResponse(volume)
	}

	response.Success(c, http.StatusOK, "volume.list_success", volumeResponses, nil)
}

// UpdateDisplayOrder updates the display order of a volume
// @Summary Update volume display order
// @Tags Volumes
// @Accept json
// @Produce json
// @Param id path string true "Volume ID"
// @Param request body UpdateDisplayOrderRequest true "Update Display Order Request"
// @Success 200 {object} response.StandardResponse
// @Failure 400 {object} response.StandardResponse
// @Failure 404 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/v1/volumes/{id}/display-order [put]
func (h *Handler) UpdateDisplayOrder(c *gin.Context) {
	// Get volume ID from path
	idStr := c.Param("id")
	id, err := uuid.FromString(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "volume.invalid_id", nil)
		return
	}

	var req UpdateDisplayOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_FAILED", "validation.failed", err.Error())
		return
	}

	// Update display order
	err = h.volumeService.UpdateDisplayOrder(c.Request.Context(), id, req.DisplayOrder)
	if err != nil {
		if errors.Is(err, pkgerrors.ErrVolumeNotFound) {
			response.Error(c, http.StatusNotFound, "VOLUME_NOT_FOUND", "volume.not_found", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "UPDATE_FAILED", "volume.update_display_order_failed", nil)
		return
	}

	response.Success(c, http.StatusOK, "volume.display_order_updated", nil, nil)
}

// PublishVolume publishes a volume
// @Summary Publish a volume
// @Tags Volumes
// @Produce json
// @Param id path string true "Volume ID"
// @Success 200 {object} response.StandardResponse
// @Failure 400 {object} response.StandardResponse
// @Failure 404 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/v1/volumes/{id}/publish [post]
func (h *Handler) PublishVolume(c *gin.Context) {
	// Get volume ID from path
	idStr := c.Param("id")
	id, err := uuid.FromString(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "volume.invalid_id", nil)
		return
	}

	// TODO: Get actual user ID from authentication context
	systemUserID := uuid.Must(uuid.FromString("00000000-0000-0000-0000-000000000000"))
	requestContext := extractRequestContext(c)

	// Publish volume
	err = h.volumeService.PublishVolume(c.Request.Context(), id, systemUserID, requestContext)
	if err != nil {
		if errors.Is(err, pkgerrors.ErrVolumeNotFound) {
			response.Error(c, http.StatusNotFound, "VOLUME_NOT_FOUND", "volume.not_found", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "PUBLISH_FAILED", "volume.publish_failed", nil)
		return
	}

	response.Success(c, http.StatusOK, "volume.published_success", nil, nil)
}

// UnpublishVolume unpublishes a volume
// @Summary Unpublish a volume
// @Tags Volumes
// @Produce json
// @Param id path string true "Volume ID"
// @Success 200 {object} response.StandardResponse
// @Failure 400 {object} response.StandardResponse
// @Failure 404 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /api/v1/volumes/{id}/unpublish [post]
func (h *Handler) UnpublishVolume(c *gin.Context) {
	// Get volume ID from path
	idStr := c.Param("id")
	id, err := uuid.FromString(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "volume.invalid_id", nil)
		return
	}

	// TODO: Get actual user ID from authentication context
	systemUserID := uuid.Must(uuid.FromString("00000000-0000-0000-0000-000000000000"))
	requestContext := extractRequestContext(c)

	// Unpublish volume
	err = h.volumeService.UnpublishVolume(c.Request.Context(), id, systemUserID, requestContext)
	if err != nil {
		if errors.Is(err, pkgerrors.ErrVolumeNotFound) {
			response.Error(c, http.StatusNotFound, "VOLUME_NOT_FOUND", "volume.not_found", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "UNPUBLISH_FAILED", "volume.unpublish_failed", nil)
		return
	}

	response.Success(c, http.StatusOK, "volume.unpublished_success", nil, nil)
}

// Helper function to map domain model to detail response
func mapToVolumeDetailResponse(volume *domain.Volume) VolumeDetailResponse {
	resp := VolumeDetailResponse{
		ID:           volume.ID.String(),
		NovelID:      volume.NovelID.String(),
		VolumeNumber: volume.VolumeNumber,
		Title:        volume.Title,
		Slug:         volume.Slug,
		Description:  volume.Description,
		CoverImageURL: volume.CoverImageURL,
		ChapterCount: volume.ChapterCount,
		WordCount:    volume.WordCount,
		DisplayOrder: volume.DisplayOrder,
		IsPublished:  volume.IsPublished,
		CreatedAt:    volume.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:    volume.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	// Format optional dates
	if volume.PublishedAt != nil {
		publishedAt := volume.PublishedAt.Format("2006-01-02T15:04:05Z07:00")
		resp.PublishedAt = &publishedAt
	}

	return resp
}

// Helper function to map domain model to list response
func mapToVolumeResponse(volume *domain.Volume) VolumeResponse {
	resp := VolumeResponse{
		ID:           volume.ID.String(),
		NovelID:      volume.NovelID.String(),
		VolumeNumber: volume.VolumeNumber,
		Title:        volume.Title,
		Slug:         volume.Slug,
		CoverImageURL: volume.CoverImageURL,
		ChapterCount: volume.ChapterCount,
		WordCount:    volume.WordCount,
		DisplayOrder: volume.DisplayOrder,
		IsPublished:  volume.IsPublished,
		CreatedAt:    volume.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:    volume.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	// Format optional dates
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
