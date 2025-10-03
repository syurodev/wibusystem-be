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

// VolumeHandler handles volume management endpoints
// This handler manages all HTTP requests related to novel volumes,
// following the same pattern as other handlers in the catalog service.
type VolumeHandler struct {
	volumeService interfaces.VolumeServiceInterface
	loc           *i18n.Translator
}

// NewVolumeHandler creates a new volume handler instance
// Takes volume service and translator as dependencies
func NewVolumeHandler(volumeService interfaces.VolumeServiceInterface, translator *i18n.Translator) *VolumeHandler {
	return &VolumeHandler{
		volumeService: volumeService,
		loc:           translator,
	}
}

// CreateVolume handles POST /novels/{novel_id}/volumes
// Creates a new volume for a specific novel
// Returns 201 Created on success with volume details
func (h *VolumeHandler) CreateVolume(c *gin.Context) {
	ctx := c.Request.Context()

	// Get novel ID from path parameter
	novelID := c.Param("novel_id")
	if novelID == "" {
		message := i18n.Localize(c, "catalog.novels.error.id_required", "Novel ID is required")
		detail := i18n.Localize(c, "catalog.novels.error.id_required_detail", "Novel ID path parameter is required")
		c.JSON(http.StatusBadRequest, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "missing_parameter", Description: detail},
			Meta:    map[string]interface{}{},
		})
		return
	}

	// Bind and validate request body
	var req d.CreateVolumeRequest
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

	// Create volume through service
	volume, err := h.volumeService.CreateVolume(ctx, novelID, req)
	if err != nil {
		status, code, message, description := mapVolumeServiceError(c, err, "create")
		c.JSON(status, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: code, Description: description},
			Meta:    map[string]interface{}{},
		})
		return
	}

	// Return success response with created volume
	successMessage := i18n.Localize(c, "catalog.volumes.create.success", "Volume created successfully")
	c.JSON(http.StatusCreated, r.StandardResponse{
		Success: true,
		Message: successMessage,
		Data:    volume,
		Error:   nil,
		Meta:    map[string]interface{}{},
	})
}

// ListVolumesByNovelID handles GET /novels/{novel_id}/volumes
// Lists all volumes for a specific novel with pagination
// Returns 200 OK with paginated volume list
func (h *VolumeHandler) ListVolumesByNovelID(c *gin.Context) {
	ctx := c.Request.Context()

	// Get novel ID from path parameter
	novelID := c.Param("novel_id")
	if novelID == "" {
		message := i18n.Localize(c, "catalog.novels.error.id_required", "Novel ID is required")
		detail := i18n.Localize(c, "catalog.novels.error.id_required_detail", "Novel ID path parameter is required")
		c.JSON(http.StatusBadRequest, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "missing_parameter", Description: detail},
			Meta:    map[string]interface{}{},
		})
		return
	}

	// Bind and validate query parameters
	var req d.ListVolumesRequest
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

	// List volumes through service
	response, err := h.volumeService.ListVolumesByNovelID(ctx, novelID, req)
	if err != nil {
		status, code, message, description := mapVolumeServiceError(c, err, "list")
		c.JSON(status, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: code, Description: description},
			Meta:    map[string]interface{}{},
		})
		return
	}

	// Prepare response payload matching chapter list structure
	data := make([]d.VolumeResponse, 0)
	if response != nil {
		data = response.Volumes
	}

	meta := map[string]interface{}{}
	if response != nil {
		meta["pagination"] = response.Pagination
	}

	// Return success response with volume list
	successMessage := i18n.Localize(c, "catalog.volumes.list.success", "Volumes retrieved successfully")
	c.JSON(http.StatusOK, r.StandardResponse{
		Success: true,
		Message: successMessage,
		Data:    data,
		Error:   nil,
		Meta:    meta,
	})
}

// GetVolumeByID handles GET /volumes/{volume_id}
// Retrieves detailed information about a specific volume
// Returns 200 OK with volume details
func (h *VolumeHandler) GetVolumeByID(c *gin.Context) {
	ctx := c.Request.Context()

	// Get volume ID from path parameter
	volumeID := c.Param("volume_id")
	if volumeID == "" {
		message := i18n.Localize(c, "catalog.volumes.error.id_required", "Volume ID is required")
		detail := i18n.Localize(c, "catalog.volumes.error.id_required_detail", "Volume ID path parameter is required")
		c.JSON(http.StatusBadRequest, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "missing_parameter", Description: detail},
			Meta:    map[string]interface{}{},
		})
		return
	}

	// Get volume through service
	volume, err := h.volumeService.GetVolumeByID(ctx, volumeID)
	if err != nil {
		status, code, message, description := mapVolumeServiceError(c, err, "get")
		c.JSON(status, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: code, Description: description},
			Meta:    map[string]interface{}{},
		})
		return
	}

	// Return success response with volume details
	successMessage := i18n.Localize(c, "catalog.volumes.get.success", "Volume retrieved successfully")
	c.JSON(http.StatusOK, r.StandardResponse{
		Success: true,
		Message: successMessage,
		Data:    volume,
		Error:   nil,
		Meta:    map[string]interface{}{},
	})
}

// UpdateVolume handles PUT /volumes/{volume_id}
// Updates an existing volume
// Returns 200 OK with updated volume details
func (h *VolumeHandler) UpdateVolume(c *gin.Context) {
	ctx := c.Request.Context()

	// Get volume ID from path parameter
	volumeID := c.Param("volume_id")
	if volumeID == "" {
		message := i18n.Localize(c, "catalog.volumes.error.id_required", "Volume ID is required")
		detail := i18n.Localize(c, "catalog.volumes.error.id_required_detail", "Volume ID path parameter is required")
		c.JSON(http.StatusBadRequest, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "missing_parameter", Description: detail},
			Meta:    map[string]interface{}{},
		})
		return
	}

	// Bind and validate request body
	var req d.UpdateVolumeRequest
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

	// Update volume through service
	response, err := h.volumeService.UpdateVolume(ctx, volumeID, req)
	if err != nil {
		status, code, message, description := mapVolumeServiceError(c, err, "update")
		c.JSON(status, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: code, Description: description},
			Meta:    map[string]interface{}{},
		})
		return
	}

	// Return success response with updated volume
	successMessage := i18n.Localize(c, "catalog.volumes.update.success", "Volume updated successfully")
	c.JSON(http.StatusOK, r.StandardResponse{
		Success: true,
		Message: successMessage,
		Data:    response,
		Error:   nil,
		Meta:    map[string]interface{}{},
	})
}

// DeleteVolume handles DELETE /volumes/{volume_id}
// Soft-deletes a volume and its chapters
// Returns 200 OK on successful deletion
func (h *VolumeHandler) DeleteVolume(c *gin.Context) {
	ctx := c.Request.Context()

	// Get volume ID from path parameter
	volumeID := c.Param("volume_id")
	if volumeID == "" {
		message := i18n.Localize(c, "catalog.volumes.error.id_required", "Volume ID is required")
		detail := i18n.Localize(c, "catalog.volumes.error.id_required_detail", "Volume ID path parameter is required")
		c.JSON(http.StatusBadRequest, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "missing_parameter", Description: detail},
			Meta:    map[string]interface{}{},
		})
		return
	}

	// Delete volume through service
	err := h.volumeService.DeleteVolume(ctx, volumeID)
	if err != nil {
		status, code, message, description := mapVolumeServiceError(c, err, "delete")
		c.JSON(status, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: code, Description: description},
			Meta:    map[string]interface{}{},
		})
		return
	}

	// Return success response
	successMessage := i18n.Localize(c, "catalog.volumes.delete.success", "Volume deleted successfully")
	c.JSON(http.StatusOK, r.StandardResponse{
		Success: true,
		Message: successMessage,
		Data:    nil,
		Error:   nil,
		Meta:    map[string]interface{}{},
	})
}

// mapVolumeServiceError maps service errors to appropriate HTTP responses for volume operations
// This helper function provides consistent error handling across all volume endpoints
// Returns: HTTP status code, error code, user message, and detailed description
func mapVolumeServiceError(c *gin.Context, err error, operation string) (int, string, string, string) {
	errStr := err.Error()

	// Check for common error patterns and map to appropriate HTTP responses
	switch {
	case strings.Contains(errStr, "not found") || strings.Contains(errStr, "no rows"):
		message := i18n.Localize(c, "catalog.common.error.not_found", "Resource not found")
		return http.StatusNotFound, "not_found", message, errStr

	case strings.Contains(errStr, "already exists") || strings.Contains(errStr, "duplicate"):
		message := i18n.Localize(c, "catalog.common.error.conflict", "Resource already exists")
		return http.StatusConflict, "conflict", message, errStr

	case strings.Contains(errStr, "volume number") && strings.Contains(errStr, "already exists"):
		message := i18n.Localize(c, "catalog.volumes.error.duplicate_number", "Volume number already exists for this novel")
		return http.StatusConflict, "duplicate_volume_number", message, errStr

	case strings.Contains(errStr, "validation") || strings.Contains(errStr, "invalid"):
		message := i18n.Localize(c, "catalog.common.error.validation", "Validation error")
		return http.StatusBadRequest, "validation_error", message, errStr

	case strings.Contains(errStr, "required"):
		message := i18n.Localize(c, "catalog.common.error.required_field", "Required field missing")
		return http.StatusBadRequest, "required_field", message, errStr

	case strings.Contains(errStr, "users have purchased content"):
		message := i18n.Localize(c, "catalog.volumes.error.cannot_delete_purchased", "Cannot delete: users have purchased content from this volume")
		return http.StatusConflict, "cannot_delete", message, errStr

	case strings.Contains(errStr, "invalid volume ID format") || strings.Contains(errStr, "invalid novel ID format"):
		message := i18n.Localize(c, "catalog.common.error.invalid_id_format", "Invalid ID format")
		return http.StatusBadRequest, "invalid_id", message, errStr

	case strings.Contains(errStr, "no fields to update"):
		message := i18n.Localize(c, "catalog.common.error.no_update_fields", "No fields provided for update")
		return http.StatusBadRequest, "no_update_fields", message, errStr

	case strings.Contains(errStr, "novel not found"):
		message := i18n.Localize(c, "catalog.volumes.error.parent_novel_not_found", "Parent novel not found")
		return http.StatusNotFound, "novel_not_found", message, errStr

	default:
		message := i18n.Localize(c, "catalog.common.error.internal", "Internal server error")
		return http.StatusInternalServerError, "internal_error", message, errStr
	}
}
