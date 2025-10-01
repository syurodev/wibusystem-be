package handlers

import (
	"errors"
	"net/http"
	"strconv"

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
		c.JSON(http.StatusBadRequest, r.StandardResponse{
			Success: false,
			Message: "Dữ liệu không hợp lệ",
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

	c.JSON(http.StatusCreated, r.StandardResponse{
		Success: true,
		Message: "Tạo chapter thành công",
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

	c.JSON(http.StatusOK, r.StandardResponse{
		Success: true,
		Message: "Lấy thông tin chapter thành công",
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
		c.JSON(http.StatusBadRequest, r.StandardResponse{
			Success: false,
			Message: "Tham số không hợp lệ",
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

	c.JSON(http.StatusOK, r.StandardResponse{
		Success: true,
		Message: "Lấy danh sách chapter thành công",
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
		c.JSON(http.StatusBadRequest, r.StandardResponse{
			Success: false,
			Message: "Dữ liệu không hợp lệ",
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

	c.JSON(http.StatusOK, r.StandardResponse{
		Success: true,
		Message: "Cập nhật chapter thành công",
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

	c.JSON(http.StatusOK, r.StandardResponse{
		Success: true,
		Message: "Xóa chapter thành công",
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
		c.JSON(http.StatusBadRequest, r.StandardResponse{
			Success: false,
			Message: "Dữ liệu không hợp lệ",
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

	c.JSON(http.StatusOK, r.StandardResponse{
		Success: true,
		Message: "Xuất bản chapter thành công",
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

	c.JSON(http.StatusOK, r.StandardResponse{
		Success: true,
		Message: "Hủy xuất bản chapter thành công",
		Data:    chapter,
		Error:   nil,
		Meta:    map[string]interface{}{},
	})
}

// mapChapterServiceError maps service layer errors to HTTP status codes and error messages.
// This centralizes error handling logic for consistent API responses.
func mapChapterServiceError(c *gin.Context, err error, operation string) (status int, code string, message string, detail string) {
	errMsg := err.Error()

	// Invalid ID format
	if errors.Is(err, errors.New("invalid chapter ID format")) || errors.Is(err, errors.New("invalid volume ID format")) {
		return http.StatusBadRequest, "INVALID_ID_FORMAT", "ID không hợp lệ", errMsg
	}

	// Not found errors
	if errors.Is(err, errors.New("chapter not found")) {
		return http.StatusNotFound, "CHAPTER_NOT_FOUND", "Không tìm thấy chapter", errMsg
	}
	if errors.Is(err, errors.New("volume not found")) {
		return http.StatusNotFound, "VOLUME_NOT_FOUND", "Không tìm thấy volume", errMsg
	}

	// Conflict errors
	if errors.Is(err, errors.New("chapter number already exists in this volume")) {
		return http.StatusConflict, "CHAPTER_NUMBER_EXISTS", "Số chapter đã tồn tại trong volume này", errMsg
	}
	if errors.Is(err, errors.New("cannot delete chapter with existing purchases")) {
		return http.StatusConflict, "CHAPTER_HAS_PURCHASES", "Không thể xóa chapter đã có người mua", errMsg
	}

	// Business logic errors
	if errors.Is(err, errors.New("cannot publish a draft chapter")) {
		return http.StatusBadRequest, "CANNOT_PUBLISH_DRAFT", "Không thể xuất bản chapter ở trạng thái nháp", errMsg
	}

	// Parse pagination errors
	if _, parseErr := strconv.Atoi(errMsg); parseErr == nil {
		return http.StatusBadRequest, "INVALID_PAGINATION", "Tham số phân trang không hợp lệ", errMsg
	}

	// Default internal server error
	return http.StatusInternalServerError, "INTERNAL_ERROR", "Lỗi hệ thống khi " + operation + " chapter", errMsg
}