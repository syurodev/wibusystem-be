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
		c.JSON(http.StatusBadRequest, r.StandardResponse{
			Success: false,
			Message: "Invalid query parameters",
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

	c.JSON(http.StatusOK, r.StandardResponse{
		Success: true,
		Message: "Genres fetched successfully",
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
		c.JSON(http.StatusBadRequest, r.StandardResponse{
			Success: false,
			Message: "Invalid genre ID",
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "invalid_id", Description: "Genre ID must be a valid UUID"},
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

	c.JSON(http.StatusOK, r.StandardResponse{
		Success: true,
		Message: "Genre fetched successfully",
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
		c.JSON(http.StatusBadRequest, r.StandardResponse{
			Success: false,
			Message: "Invalid request body",
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

	c.JSON(http.StatusCreated, r.StandardResponse{
		Success: true,
		Message: "Genre created successfully",
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
		c.JSON(http.StatusBadRequest, r.StandardResponse{
			Success: false,
			Message: "Invalid genre ID",
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "invalid_id", Description: "Genre ID must be a valid UUID"},
			Meta:    map[string]interface{}{},
		})
		return
	}

	// Bind request body
	var req d.UpdateGenreRequest
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

	c.JSON(http.StatusOK, r.StandardResponse{
		Success: true,
		Message: "Genre updated successfully",
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
		c.JSON(http.StatusBadRequest, r.StandardResponse{
			Success: false,
			Message: "Invalid genre ID",
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "invalid_id", Description: "Genre ID must be a valid UUID"},
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

	c.JSON(http.StatusOK, r.StandardResponse{
		Success: true,
		Message: "Genre deleted successfully",
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
		return http.StatusNotFound, "not_found", "Resource not found", errStr

	case strings.Contains(errStr, "already exists") || strings.Contains(errStr, "duplicate"):
		return http.StatusConflict, "conflict", "Resource already exists", errStr

	case strings.Contains(errStr, "validation") || strings.Contains(errStr, "invalid"):
		return http.StatusBadRequest, "validation_error", "Validation error", errStr

	case strings.Contains(errStr, "required"):
		return http.StatusBadRequest, "required_field", "Required field missing", errStr

	default:
		return http.StatusInternalServerError, "internal_error", "Internal server error", errStr
	}
}