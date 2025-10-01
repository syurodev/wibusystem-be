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

// NovelHandler handles novel management endpoints
type NovelHandler struct {
	novelService interfaces.NovelServiceInterface
	loc          *i18n.Translator
}

func NewNovelHandler(novelService interfaces.NovelServiceInterface, translator *i18n.Translator) *NovelHandler {
	return &NovelHandler{
		novelService: novelService,
		loc:          translator,
	}
}

// CreateNovel handles POST /novels
func (h *NovelHandler) CreateNovel(c *gin.Context) {
	ctx := c.Request.Context()

	// Bind request body
	var req d.CreateNovelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, r.StandardResponse{
			Success: false,
			Message: "Invalid request body",
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "validation_error", Description: err.Error()},
			Meta:    map[string]interface{}{},
		})
		return
	}

	// Create novel through service
	novel, err := h.novelService.CreateNovel(ctx, req)
	if err != nil {
		status, code, message, description := mapNovelServiceError(c, err, "create")
		c.JSON(status, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: code, Description: description},
			Meta:    map[string]interface{}{},
		})
		return
	}

	// Create simplified response according to API design
	response := d.CreateNovelResponse{
		ID:        novel.ID.String(),
		Title:     "",
		Status:    novel.Status,
		Slug:      "",
		CreatedAt: novel.CreatedAt,
		UpdatedAt: novel.UpdatedAt,
	}

	// Set title (name field)
	if novel.Name != nil {
		response.Title = *novel.Name
	}

	// Set slug
	if novel.Slug != nil {
		response.Slug = *novel.Slug
	}

	c.JSON(http.StatusCreated, r.StandardResponse{
		Success: true,
		Message: "Tạo tiểu thuyết thành công",
		Data:    response,
		Error:   nil,
		Meta:    map[string]interface{}{},
	})
}

// ListNovels handles GET /novels
func (h *NovelHandler) ListNovels(c *gin.Context) {
	ctx := c.Request.Context()

	// Bind query parameters
	var req d.ListNovelsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, r.StandardResponse{
			Success: false,
			Message: "Invalid query parameters",
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "validation_error", Description: err.Error()},
			Meta:    map[string]interface{}{},
		})
		return
	}

	// List novels through service
	response, err := h.novelService.ListNovels(ctx, req)
	if err != nil {
		status, code, message, description := mapNovelServiceError(c, err, "list")
		c.JSON(status, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: code, Description: description},
			Meta:    map[string]interface{}{},
		})
		return
	}

	c.JSON(http.StatusOK, r.StandardResponse{
		Success: true,
		Message: "Lấy danh sách tiểu thuyết thành công",
		Data:    response,
		Error:   nil,
		Meta:    map[string]interface{}{},
	})
}

// GetNovelByID handles GET /novels/{novel_id}
func (h *NovelHandler) GetNovelByID(c *gin.Context) {
	ctx := c.Request.Context()

	// Get novel ID from path
	novelID := c.Param("novel_id")
	if novelID == "" {
		c.JSON(http.StatusBadRequest, r.StandardResponse{
			Success: false,
			Message: "Novel ID is required",
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "missing_parameter", Description: "Novel ID path parameter is required"},
			Meta:    map[string]interface{}{},
		})
		return
	}

	// Parse query parameters
	includeTranslations := c.DefaultQuery("include_translations", "false") == "true"
	includeStats := c.DefaultQuery("include_stats", "false") == "true"

	// Get language preference from headers (theo API design)
	language := c.GetHeader("X-Language")
	if language == "" {
		language = c.GetHeader("Accept-Language")
		if language == "" {
			language = "vi" // Default language
		}
	}

	// Get novel through service
	novel, err := h.novelService.GetNovelByID(ctx, novelID, includeTranslations, includeStats, language)
	if err != nil {
		status, code, message, description := mapNovelServiceError(c, err, "get")
		c.JSON(status, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: code, Description: description},
			Meta:    map[string]interface{}{},
		})
		return
	}

	c.JSON(http.StatusOK, r.StandardResponse{
		Success: true,
		Message: "Lấy thông tin tiểu thuyết thành công",
		Data:    novel,
		Error:   nil,
		Meta:    map[string]interface{}{},
	})
}

// UpdateNovel handles PUT /novels/{novel_id}
func (h *NovelHandler) UpdateNovel(c *gin.Context) {
	ctx := c.Request.Context()

	// Get novel ID from path
	novelID := c.Param("novel_id")
	if novelID == "" {
		c.JSON(http.StatusBadRequest, r.StandardResponse{
			Success: false,
			Message: "Novel ID is required",
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "missing_parameter", Description: "Novel ID path parameter is required"},
			Meta:    map[string]interface{}{},
		})
		return
	}

	// Bind request body
	var req d.UpdateNovelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, r.StandardResponse{
			Success: false,
			Message: "Invalid request body",
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "validation_error", Description: err.Error()},
			Meta:    map[string]interface{}{},
		})
		return
	}

	// Update novel through service
	response, err := h.novelService.UpdateNovel(ctx, novelID, req)
	if err != nil {
		status, code, message, description := mapNovelServiceError(c, err, "update")
		c.JSON(status, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: code, Description: description},
			Meta:    map[string]interface{}{},
		})
		return
	}

	c.JSON(http.StatusOK, r.StandardResponse{
		Success: true,
		Message: "Cập nhật tiểu thuyết thành công",
		Data:    response,
		Error:   nil,
		Meta:    map[string]interface{}{},
	})
}

// DeleteNovel handles DELETE /novels/{novel_id}
func (h *NovelHandler) DeleteNovel(c *gin.Context) {
	ctx := c.Request.Context()

	// Get novel ID from path
	novelID := c.Param("novel_id")
	if novelID == "" {
		c.JSON(http.StatusBadRequest, r.StandardResponse{
			Success: false,
			Message: "Novel ID is required",
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "missing_parameter", Description: "Novel ID path parameter is required"},
			Meta:    map[string]interface{}{},
		})
		return
	}

	// Get user ID from context (set by middleware)
	// TODO: Get actual user ID from JWT token/middleware
	userID := "00000000-0000-0000-0000-000000000000" // Placeholder

	// Delete novel through service
	err := h.novelService.DeleteNovel(ctx, novelID, userID)
	if err != nil {
		status, code, message, description := mapNovelServiceError(c, err, "delete")
		c.JSON(status, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: code, Description: description},
			Meta:    map[string]interface{}{},
		})
		return
	}

	c.JSON(http.StatusOK, r.StandardResponse{
		Success: true,
		Message: "Xóa tiểu thuyết thành công",
		Data:    nil,
		Error:   nil,
		Meta:    map[string]interface{}{},
	})
}

// mapNovelServiceError maps service errors to appropriate HTTP responses for novel operations
func mapNovelServiceError(c *gin.Context, err error, operation string) (int, string, string, string) {
	errStr := err.Error()

	// Check for common error patterns
	switch {
	case strings.Contains(errStr, "not found") || strings.Contains(errStr, "no rows"):
		return http.StatusNotFound, "not_found", "Resource not found", errStr

	case strings.Contains(errStr, "already exists") || strings.Contains(errStr, "duplicate"):
		return http.StatusConflict, "conflict", "Resource already exists", errStr

	case strings.Contains(errStr, "validation") || strings.Contains(errStr, "invalid"):
		return http.StatusBadRequest, "validation_error", "Validation error", errStr

	case strings.Contains(errStr, "required"):
		return http.StatusBadRequest, "required_field", "Required field missing", errStr

	case strings.Contains(errStr, "genres do not exist"):
		return http.StatusBadRequest, "invalid_genre", "One or more genres do not exist", errStr

	case strings.Contains(errStr, "users have purchased content"):
		return http.StatusConflict, "cannot_delete", "Cannot delete: users have purchased content", errStr

	case strings.Contains(errStr, "invalid novel ID format") || strings.Contains(errStr, "invalid user ID format"):
		return http.StatusBadRequest, "invalid_id", "Invalid ID format", errStr

	case strings.Contains(errStr, "no fields to update"):
		return http.StatusBadRequest, "no_update_fields", "No fields provided for update", errStr

	default:
		return http.StatusInternalServerError, "internal_error", "Internal server error", errStr
	}
}
