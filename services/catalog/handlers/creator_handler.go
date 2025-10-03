package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	d "wibusystem/pkg/common/dto"
	r "wibusystem/pkg/common/response"
	"wibusystem/pkg/i18n"
	"wibusystem/services/catalog/services/interfaces"
)

// CreatorHandler handles creator management endpoints
type CreatorHandler struct {
	creatorService interfaces.CreatorServiceInterface
	loc            *i18n.Translator
}

// NewCreatorHandler creates a new creator handler
func NewCreatorHandler(creatorService interfaces.CreatorServiceInterface, translator *i18n.Translator) *CreatorHandler {
	return &CreatorHandler{
		creatorService: creatorService,
		loc:            translator,
	}
}

// ListCreators handles GET /creators
func (h *CreatorHandler) ListCreators(c *gin.Context) {
	ctx := c.Request.Context()

	// Bind query parameters
	var req d.ListCreatorsRequest
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

	// Get creators through service
	creators, total, err := h.creatorService.ListCreators(ctx, req)
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

	successMessage := i18n.Localize(c, "catalog.creators.list.success", "Creators fetched successfully")
	c.JSON(http.StatusOK, r.StandardResponse{
		Success: true,
		Message: successMessage,
		Data:    creators,
		Error:   nil,
		Meta: map[string]interface{}{
			"page":        req.Page,
			"page_size":   req.PageSize,
			"total_pages": totalPages,
			"total_items": total,
		},
	})
}

// GetCreator handles GET /creators/:id
func (h *CreatorHandler) GetCreator(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse creator ID
	creatorID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		message := i18n.Localize(c, "catalog.creators.error.invalid_id", "Invalid creator ID")
		detail := i18n.Localize(c, "catalog.creators.error.invalid_id_detail", "Creator ID must be a valid UUID")
		c.JSON(http.StatusBadRequest, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "invalid_id", Description: detail},
			Meta:    map[string]interface{}{},
		})
		return
	}

	// Get creator through service
	creator, err := h.creatorService.GetCreatorByID(ctx, creatorID)
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

	successMessage := i18n.Localize(c, "catalog.creators.get.success", "Creator fetched successfully")
	c.JSON(http.StatusOK, r.StandardResponse{
		Success: true,
		Message: successMessage,
		Data:    creator,
		Error:   nil,
		Meta:    map[string]interface{}{},
	})
}

// CreateCreator handles POST /creators
func (h *CreatorHandler) CreateCreator(c *gin.Context) {
	ctx := c.Request.Context()

	// Bind request body
	var req d.CreateCreatorRequest
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

	// Create creator through service
	creator, err := h.creatorService.CreateCreator(ctx, req)
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

	successMessage := i18n.Localize(c, "catalog.creators.create.success", "Creator created successfully")
	c.JSON(http.StatusCreated, r.StandardResponse{
		Success: true,
		Message: successMessage,
		Data:    creator,
		Error:   nil,
		Meta:    map[string]interface{}{},
	})
}

// UpdateCreator handles PUT /creators/:id
func (h *CreatorHandler) UpdateCreator(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse creator ID
	creatorID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		message := i18n.Localize(c, "catalog.creators.error.invalid_id", "Invalid creator ID")
		detail := i18n.Localize(c, "catalog.creators.error.invalid_id_detail", "Creator ID must be a valid UUID")
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
	var req d.UpdateCreatorRequest
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

	// Update creator through service
	creator, err := h.creatorService.UpdateCreator(ctx, creatorID, req)
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

	successMessage := i18n.Localize(c, "catalog.creators.update.success", "Creator updated successfully")
	c.JSON(http.StatusOK, r.StandardResponse{
		Success: true,
		Message: successMessage,
		Data:    creator,
		Error:   nil,
		Meta:    map[string]interface{}{},
	})
}

// DeleteCreator handles DELETE /creators/:id
func (h *CreatorHandler) DeleteCreator(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse creator ID
	creatorID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, r.StandardResponse{
			Success: false,
			Message: "Invalid creator ID",
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "invalid_id", Description: "Creator ID must be a valid UUID"},
			Meta:    map[string]interface{}{},
		})
		return
	}

	// Delete creator through service
	err = h.creatorService.DeleteCreator(ctx, creatorID)
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

	successMessage := i18n.Localize(c, "catalog.creators.delete.success", "Creator deleted successfully")
	c.JSON(http.StatusOK, r.StandardResponse{
		Success: true,
		Message: successMessage,
		Data:    nil,
		Error:   nil,
		Meta:    map[string]interface{}{},
	})
}
