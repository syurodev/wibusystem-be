package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	d "wibusystem/pkg/common/dto"
	r "wibusystem/pkg/common/response"
	"wibusystem/pkg/i18n"
	"wibusystem/services/catalog/services/interfaces"
)

// GenreHandler handles genre management endpoints
type GenreHandler struct {
	genreService interfaces.GenreServiceInterface
	loc          *i18n.Translator
}

// NewGenreHandler creates a new genre handler
func NewGenreHandler(genreService interfaces.GenreServiceInterface, translator *i18n.Translator) *GenreHandler {
	return &GenreHandler{
		genreService: genreService,
		loc:          translator,
	}
}

// ListGenres handles GET /genres
func (h *GenreHandler) ListGenres(c *gin.Context) {
	ctx := c.Request.Context()

	// Bind query parameters
	var req d.ListGenresRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		message := i18n.Localize(c, "catalog.common.error.invalid_query_parameters", "Invalid query parameters")
		c.JSON(http.StatusBadRequest, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "validation_error", Description: err.Error()},
			Meta:    map[string]interface{}{},
		})
		return
	}

	// Get genres through service
	genres, total, err := h.genreService.ListGenres(ctx, req)
	if err != nil {
		status, code, message, description := mapServiceError(c, err, "list")
		c.JSON(status, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: code, Description: description},
			Meta:    map[string]interface{}{},
		})
		return
	}

	// Calculate pagination metadata
	totalPages := int(total) / req.PageSize
	if int(total)%req.PageSize > 0 {
		totalPages++
	}

	successMessage := i18n.Localize(c, "catalog.genres.list.success", "Genres fetched successfully")
	c.JSON(http.StatusOK, r.StandardResponse{
		Success: true,
		Message: successMessage,
		Data:    genres,
		Error:   nil,
		Meta: map[string]interface{}{
			"page":        req.Page,
			"page_size":   req.PageSize,
			"total_pages": totalPages,
			"total_items": total,
		},
	})
}

// GetGenre handles GET /genres/:id
func (h *GenreHandler) GetGenre(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse genre ID
	genreID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		message := i18n.Localize(c, "catalog.genres.error.invalid_id", "Invalid genre ID")
		detail := i18n.Localize(c, "catalog.genres.error.invalid_id_detail", "Genre ID must be a valid UUID")
		c.JSON(http.StatusBadRequest, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "invalid_id", Description: detail},
			Meta:    map[string]interface{}{},
		})
		return
	}

	// Get genre through service
	genre, err := h.genreService.GetGenreByID(ctx, genreID)
	if err != nil {
		status, code, message, description := mapServiceError(c, err, "get")
		c.JSON(status, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: code, Description: description},
			Meta:    map[string]interface{}{},
		})
		return
	}

	successMessage := i18n.Localize(c, "catalog.genres.get.success", "Genre fetched successfully")
	c.JSON(http.StatusOK, r.StandardResponse{
		Success: true,
		Message: successMessage,
		Data:    genre,
		Error:   nil,
		Meta:    map[string]interface{}{},
	})
}

// CreateGenre handles POST /genres
func (h *GenreHandler) CreateGenre(c *gin.Context) {
	ctx := c.Request.Context()

	// Bind request body
	var req d.CreateGenreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		message := i18n.Localize(c, "catalog.common.error.invalid_request_body", "Invalid request body")
		c.JSON(http.StatusBadRequest, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "validation_error", Description: err.Error()},
			Meta:    map[string]interface{}{},
		})
		return
	}

	// Create genre through service
	genre, err := h.genreService.CreateGenre(ctx, req)
	if err != nil {
		status, code, message, description := mapServiceError(c, err, "create")
		c.JSON(status, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: code, Description: description},
			Meta:    map[string]interface{}{},
		})
		return
	}

	successMessage := i18n.Localize(c, "catalog.genres.create.success", "Genre created successfully")
	c.JSON(http.StatusCreated, r.StandardResponse{
		Success: true,
		Message: successMessage,
		Data:    genre,
		Error:   nil,
		Meta:    map[string]interface{}{},
	})
}

// UpdateGenre handles PUT /genres/:id
func (h *GenreHandler) UpdateGenre(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse genre ID
	genreID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		message := i18n.Localize(c, "catalog.genres.error.invalid_id", "Invalid genre ID")
		detail := i18n.Localize(c, "catalog.genres.error.invalid_id_detail", "Genre ID must be a valid UUID")
		c.JSON(http.StatusBadRequest, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "invalid_id", Description: detail},
			Meta:    map[string]interface{}{},
		})
		return
	}

	// Bind request body
	var req d.UpdateGenreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		message := i18n.Localize(c, "catalog.common.error.invalid_request_body", "Invalid request body")
		c.JSON(http.StatusBadRequest, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "validation_error", Description: err.Error()},
			Meta:    map[string]interface{}{},
		})
		return
	}

	// Update genre through service
	genre, err := h.genreService.UpdateGenre(ctx, genreID, req)
	if err != nil {
		status, code, message, description := mapServiceError(c, err, "update")
		c.JSON(status, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: code, Description: description},
			Meta:    map[string]interface{}{},
		})
		return
	}

	successMessage := i18n.Localize(c, "catalog.genres.update.success", "Genre updated successfully")
	c.JSON(http.StatusOK, r.StandardResponse{
		Success: true,
		Message: successMessage,
		Data:    genre,
		Error:   nil,
		Meta:    map[string]interface{}{},
	})
}

// DeleteGenre handles DELETE /genres/:id
func (h *GenreHandler) DeleteGenre(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse genre ID
	genreID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		message := i18n.Localize(c, "catalog.genres.error.invalid_id", "Invalid genre ID")
		detail := i18n.Localize(c, "catalog.genres.error.invalid_id_detail", "Genre ID must be a valid UUID")
		c.JSON(http.StatusBadRequest, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "invalid_id", Description: detail},
			Meta:    map[string]interface{}{},
		})
		return
	}

	// Delete genre through service
	err = h.genreService.DeleteGenre(ctx, genreID)
	if err != nil {
		status, code, message, description := mapServiceError(c, err, "delete")
		c.JSON(status, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: code, Description: description},
			Meta:    map[string]interface{}{},
		})
		return
	}

	successMessage := i18n.Localize(c, "catalog.genres.delete.success", "Genre deleted successfully")
	c.JSON(http.StatusOK, r.StandardResponse{
		Success: true,
		Message: successMessage,
		Data:    nil,
		Error:   nil,
		Meta:    map[string]interface{}{},
	})
}

// mapServiceError maps service errors to appropriate HTTP responses
func mapServiceError(c *gin.Context, err error, operation string) (int, string, string, string) {
	errStr := err.Error()

	// Check for common error patterns
	switch {
	case strings.Contains(errStr, "not found") || strings.Contains(errStr, "no rows"):
		message := i18n.Localize(c, "catalog.common.error.not_found", "Resource not found")
		return http.StatusNotFound, "not_found", message, errStr

	case strings.Contains(errStr, "already exists") || strings.Contains(errStr, "duplicate"):
		message := i18n.Localize(c, "catalog.common.error.conflict", "Resource already exists")
		return http.StatusConflict, "conflict", message, errStr

	case strings.Contains(errStr, "validation") || strings.Contains(errStr, "invalid"):
		message := i18n.Localize(c, "catalog.common.error.validation", "Validation error")
		return http.StatusBadRequest, "validation_error", message, errStr

	case strings.Contains(errStr, "required"):
		message := i18n.Localize(c, "catalog.common.error.required_field", "Required field missing")
		return http.StatusBadRequest, "required_field", message, errStr

	default:
		message := i18n.Localize(c, "catalog.common.error.internal", "Internal server error")
		return http.StatusInternalServerError, "internal_error", message, errStr
	}
}
