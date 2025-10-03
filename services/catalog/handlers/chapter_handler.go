package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	d "wibusystem/pkg/common/dto"
	r "wibusystem/pkg/common/response"
	"wibusystem/pkg/i18n"
	"wibusystem/services/catalog/services/interfaces"
)

// ChapterHandler handles HTTP requests for chapter management operations.
// This follows the API design spec from /services/catalog/api-design/novel.md sections 3.1-3.7
type ChapterHandler struct {
	service    interfaces.ChapterServiceInterface
	translator *i18n.Translator
}

// NewChapterHandler creates a new ChapterHandler instance with the given dependencies.
func NewChapterHandler(service interfaces.ChapterServiceInterface, translator *i18n.Translator) *ChapterHandler {
	return &ChapterHandler{
		service:    service,
		translator: translator,
	}
}

// CreateChapter handles POST /api/v1/volumes/{volume_id}/chapters
// Creates a new chapter within a volume.
func (h *ChapterHandler) CreateChapter(c *gin.Context) {
	volumeID := c.Param("volume_id")

	var req d.CreateChapterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		message := i18n.Localize(c, "catalog.chapters.error.invalid_input", "Invalid chapter data")
		c.JSON(http.StatusBadRequest, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "INVALID_INPUT", Description: err.Error()},
			Meta:    map[string]interface{}{},
		})
		return
	}

	chapter, err := h.service.CreateChapter(c.Request.Context(), volumeID, req)
	if err != nil {
		status, code, message, detail := mapChapterServiceError(c, err, "create")
		c.JSON(status, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: code, Description: detail},
			Meta:    map[string]interface{}{},
		})
		return
	}

	successMessage := i18n.Localize(c, "catalog.chapters.create.success", "Chapter created successfully")
	c.JSON(http.StatusCreated, r.StandardResponse{
		Success: true,
		Message: successMessage,
		Data:    chapter,
		Error:   nil,
		Meta:    map[string]interface{}{},
	})
}

// GetChapterByID handles GET /api/v1/chapters/{id}
// Retrieves a specific chapter by its ID.
func (h *ChapterHandler) GetChapterByID(c *gin.Context) {
	id := c.Param("id")

	// Parse include_content query parameter (default: false)
	includeContent := false
	if includeParam := c.Query("include_content"); includeParam == "true" {
		includeContent = true
	}

	chapter, err := h.service.GetChapterByID(c.Request.Context(), id, includeContent)
	if err != nil {
		status, code, message, detail := mapChapterServiceError(c, err, "get")
		c.JSON(status, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: code, Description: detail},
			Meta:    map[string]interface{}{},
		})
		return
	}

	successMessage := i18n.Localize(c, "catalog.chapters.get.success", "Chapter retrieved successfully")
	c.JSON(http.StatusOK, r.StandardResponse{
		Success: true,
		Message: successMessage,
		Data:    chapter,
		Error:   nil,
		Meta:    map[string]interface{}{},
	})
}

// ListChaptersByVolumeID handles GET /api/v1/volumes/{volume_id}/chapters
// Retrieves a paginated list of chapters in a volume.
func (h *ChapterHandler) ListChaptersByVolumeID(c *gin.Context) {
	volumeID := c.Param("volume_id")

	// Parse query parameters
	var req d.ListChaptersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		message := i18n.Localize(c, "catalog.chapters.error.invalid_query", "Invalid chapter query parameters")
		c.JSON(http.StatusBadRequest, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "INVALID_QUERY_PARAMS", Description: err.Error()},
			Meta:    map[string]interface{}{},
		})
		return
	}

	response, err := h.service.ListChaptersByVolumeID(c.Request.Context(), volumeID, req)
	if err != nil {
		status, code, message, detail := mapChapterServiceError(c, err, "list")
		c.JSON(status, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: code, Description: detail},
			Meta:    map[string]interface{}{},
		})
		return
	}

	successMessage := i18n.Localize(c, "catalog.chapters.list.success", "Chapters retrieved successfully")
	c.JSON(http.StatusOK, r.StandardResponse{
		Success: true,
		Message: successMessage,
		Data:    response.Chapters,
		Error:   nil,
		Meta: map[string]interface{}{
			"pagination": response.Pagination,
		},
	})
}

// UpdateChapter handles PUT /api/v1/chapters/{id}
// Updates an existing chapter's information.
func (h *ChapterHandler) UpdateChapter(c *gin.Context) {
	id := c.Param("id")

	var req d.UpdateChapterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		message := i18n.Localize(c, "catalog.chapters.error.invalid_input", "Invalid chapter data")
		c.JSON(http.StatusBadRequest, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "INVALID_INPUT", Description: err.Error()},
			Meta:    map[string]interface{}{},
		})
		return
	}

	chapter, err := h.service.UpdateChapter(c.Request.Context(), id, req)
	if err != nil {
		status, code, message, detail := mapChapterServiceError(c, err, "update")
		c.JSON(status, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: code, Description: detail},
			Meta:    map[string]interface{}{},
		})
		return
	}

	successMessage := i18n.Localize(c, "catalog.chapters.update.success", "Chapter updated successfully")
	c.JSON(http.StatusOK, r.StandardResponse{
		Success: true,
		Message: successMessage,
		Data:    chapter,
		Error:   nil,
		Meta:    map[string]interface{}{},
	})
}

// DeleteChapter handles DELETE /api/v1/chapters/{id}
// Soft-deletes a chapter.
func (h *ChapterHandler) DeleteChapter(c *gin.Context) {
	id := c.Param("id")

	err := h.service.DeleteChapter(c.Request.Context(), id)
	if err != nil {
		status, code, message, detail := mapChapterServiceError(c, err, "delete")
		c.JSON(status, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: code, Description: detail},
			Meta:    map[string]interface{}{},
		})
		return
	}

	successMessage := i18n.Localize(c, "catalog.chapters.delete.success", "Chapter deleted successfully")
	c.JSON(http.StatusOK, r.StandardResponse{
		Success: true,
		Message: successMessage,
		Data:    nil,
		Error:   nil,
		Meta:    map[string]interface{}{},
	})
}

// PublishChapter handles POST /api/v1/chapters/{id}/publish
// Publishes a chapter, making it publicly accessible.
func (h *ChapterHandler) PublishChapter(c *gin.Context) {
	id := c.Param("id")

	var req d.PublishChapterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		message := i18n.Localize(c, "catalog.chapters.error.invalid_input", "Invalid chapter data")
		c.JSON(http.StatusBadRequest, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "INVALID_INPUT", Description: err.Error()},
			Meta:    map[string]interface{}{},
		})
		return
	}

	chapter, err := h.service.PublishChapter(c.Request.Context(), id, req)
	if err != nil {
		status, code, message, detail := mapChapterServiceError(c, err, "publish")
		c.JSON(status, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: code, Description: detail},
			Meta:    map[string]interface{}{},
		})
		return
	}

	successMessage := i18n.Localize(c, "catalog.chapters.publish.success", "Chapter published successfully")
	c.JSON(http.StatusOK, r.StandardResponse{
		Success: true,
		Message: successMessage,
		Data:    chapter,
		Error:   nil,
		Meta:    map[string]interface{}{},
	})
}

// UnpublishChapter handles POST /api/v1/chapters/{id}/unpublish
// Unpublishes a chapter, removing it from public access.
func (h *ChapterHandler) UnpublishChapter(c *gin.Context) {
	id := c.Param("id")

	chapter, err := h.service.UnpublishChapter(c.Request.Context(), id)
	if err != nil {
		status, code, message, detail := mapChapterServiceError(c, err, "unpublish")
		c.JSON(status, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: code, Description: detail},
			Meta:    map[string]interface{}{},
		})
		return
	}

	successMessage := i18n.Localize(c, "catalog.chapters.unpublish.success", "Chapter unpublished successfully")
	c.JSON(http.StatusOK, r.StandardResponse{
		Success: true,
		Message: successMessage,
		Data:    chapter,
		Error:   nil,
		Meta:    map[string]interface{}{},
	})
}

// mapChapterServiceError maps service layer errors to HTTP status codes and error messages.
// This centralizes error handling logic for consistent API responses.

func mapChapterServiceError(c *gin.Context, err error, operation string) (status int, code string, message string, detail string) {
	errMsg := err.Error()
	lower := strings.ToLower(errMsg)

	switch {
	case strings.Contains(lower, "invalid chapter id format") || strings.Contains(lower, "invalid volume id format"):
		message := i18n.Localize(c, "catalog.chapters.error.invalid_id_format", "ID không hợp lệ")
		return http.StatusBadRequest, "INVALID_ID_FORMAT", message, errMsg

	case strings.Contains(lower, "chapter not found"):
		message := i18n.Localize(c, "catalog.chapters.error.not_found", "Chapter not found")
		return http.StatusNotFound, "CHAPTER_NOT_FOUND", message, errMsg

	case strings.Contains(lower, "volume not found"):
		message := i18n.Localize(c, "catalog.volumes.error.not_found", "Volume not found")
		return http.StatusNotFound, "VOLUME_NOT_FOUND", message, errMsg

	case strings.Contains(lower, "chapter number already exists"):
		message := i18n.Localize(c, "catalog.chapters.error.duplicate_number", "Chapter number already exists in this volume")
		return http.StatusConflict, "CHAPTER_NUMBER_EXISTS", message, errMsg

	case strings.Contains(lower, "cannot delete chapter with existing purchases"):
		message := i18n.Localize(c, "catalog.chapters.error.cannot_delete_purchased", "Cannot delete chapter because readers have purchased it")
		return http.StatusConflict, "CHAPTER_HAS_PURCHASES", message, errMsg

	case strings.Contains(lower, "cannot publish a draft chapter"):
		message := i18n.Localize(c, "catalog.chapters.error.cannot_publish_draft", "Cannot publish a draft chapter")
		return http.StatusBadRequest, "CANNOT_PUBLISH_DRAFT", message, errMsg

	case strings.Contains(lower, "invalid pagination") || strings.Contains(lower, "invalid query param"):
		message := i18n.Localize(c, "catalog.chapters.error.invalid_pagination", "Invalid pagination parameters")
		return http.StatusBadRequest, "INVALID_PAGINATION", message, errMsg

	default:
		defaultMsg := i18n.Localize(c, "catalog.chapters.error.internal", "Internal error while processing chapter")
		return http.StatusInternalServerError, "INTERNAL_ERROR", defaultMsg, errMsg
	}
}
