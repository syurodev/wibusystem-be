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
		c.JSON(http.StatusBadRequest, r.StandardResponse{
			Success: false,
			Message: "Novel ID is required",
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "missing_parameter", Description: "Novel ID path parameter is required"},
			Meta:    map[string]interface{}{},
		})
		return
	}

	// Bind and validate request body
	var req d.CreateVolumeRequest
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
	c.JSON(http.StatusCreated, r.StandardResponse{
		Success: true,
		Message: "Tạo volume thành công",
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
		c.JSON(http.StatusBadRequest, r.StandardResponse{
			Success: false,
			Message: "Novel ID is required",
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "missing_parameter", Description: "Novel ID path parameter is required"},
			Meta:    map[string]interface{}{},
		})
		return
	}

	// Bind and validate query parameters
	var req d.ListVolumesRequest
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

	// Return success response with volume list
	c.JSON(http.StatusOK, r.StandardResponse{
		Success: true,
		Message: "Lấy danh sách volumes thành công",
		Data:    response,
		Error:   nil,
		Meta:    map[string]interface{}{},
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
		c.JSON(http.StatusBadRequest, r.StandardResponse{
			Success: false,
			Message: "Volume ID is required",
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "missing_parameter", Description: "Volume ID path parameter is required"},
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
	c.JSON(http.StatusOK, r.StandardResponse{
		Success: true,
		Message: "Lấy chi tiết volume thành công",
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
		c.JSON(http.StatusBadRequest, r.StandardResponse{
			Success: false,
			Message: "Volume ID is required",
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "missing_parameter", Description: "Volume ID path parameter is required"},
			Meta:    map[string]interface{}{},
		})
		return
	}

	// Bind and validate request body
	var req d.UpdateVolumeRequest
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
	c.JSON(http.StatusOK, r.StandardResponse{
		Success: true,
		Message: "Cập nhật volume thành công",
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
		c.JSON(http.StatusBadRequest, r.StandardResponse{
			Success: false,
			Message: "Volume ID is required",
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "missing_parameter", Description: "Volume ID path parameter is required"},
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
	c.JSON(http.StatusOK, r.StandardResponse{
		Success: true,
		Message: "Xóa volume thành công",
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
		return http.StatusNotFound, "not_found", "Resource not found", errStr

	case strings.Contains(errStr, "already exists") || strings.Contains(errStr, "duplicate"):
		return http.StatusConflict, "conflict", "Resource already exists", errStr

	case strings.Contains(errStr, "volume number") && strings.Contains(errStr, "already exists"):
		return http.StatusConflict, "duplicate_volume_number", "Volume number already exists for this novel", errStr

	case strings.Contains(errStr, "validation") || strings.Contains(errStr, "invalid"):
		return http.StatusBadRequest, "validation_error", "Validation error", errStr

	case strings.Contains(errStr, "required"):
		return http.StatusBadRequest, "required_field", "Required field missing", errStr

	case strings.Contains(errStr, "users have purchased content"):
		return http.StatusConflict, "cannot_delete", "Cannot delete: users have purchased content from this volume", errStr

	case strings.Contains(errStr, "invalid volume ID format") || strings.Contains(errStr, "invalid novel ID format"):
		return http.StatusBadRequest, "invalid_id", "Invalid ID format", errStr

	case strings.Contains(errStr, "no fields to update"):
		return http.StatusBadRequest, "no_update_fields", "No fields provided for update", errStr

	case strings.Contains(errStr, "novel not found"):
		return http.StatusNotFound, "novel_not_found", "Parent novel not found", errStr

	default:
		return http.StatusInternalServerError, "internal_error", "Internal server error", errStr
	}
}
