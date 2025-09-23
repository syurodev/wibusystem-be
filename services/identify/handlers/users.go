package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	d "wibusystem/pkg/common/dto"
	r "wibusystem/pkg/common/response"
	"wibusystem/pkg/i18n"
	"wibusystem/services/identify/services/interfaces"
)

// UserHandler handles user management endpoints
type UserHandler struct {
	userService interfaces.UserServiceInterface
	loc         *i18n.Translator
}

// NewUserHandler creates a new user handler
func NewUserHandler(userService interfaces.UserServiceInterface, translator *i18n.Translator) *UserHandler {
	return &UserHandler{
		userService: userService,
		loc:         translator,
	}
}

// ListUsers handles GET /users
func (h *UserHandler) ListUsers(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	// Get users through service
	users, total, err := h.userService.ListUsers(ctx, page, pageSize)
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
	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	message := i18n.T(c, "identify.users.list_success", "Users fetched successfully", nil)
	c.JSON(http.StatusOK, r.StandardResponse{
		Success: true,
		Message: message,
		Data:    users,
		Error:   nil,
		Meta: map[string]interface{}{
			"page":        page,
			"page_size":   pageSize,
			"total_pages": totalPages,
			"total_items": total,
		},
	})
}

// GetUser handles GET /users/:id
func (h *UserHandler) GetUser(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse user ID
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		message := i18n.T(c, "identify.errors.invalid_user_id.message", "Invalid user id", nil)
		description := i18n.T(c, "identify.errors.invalid_user_id.description", "User ID must be a valid UUID", map[string]interface{}{"id": c.Param("id")})
		c.JSON(http.StatusBadRequest, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "invalid_user_id", Description: description},
			Meta:    map[string]interface{}{},
		})
		return
	}

	// Get user through service
	user, err := h.userService.GetUserByID(ctx, userID)
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

	message := i18n.T(c, "identify.users.get_success", "User fetched successfully", nil)
	c.JSON(http.StatusOK, r.StandardResponse{
		Success: true,
		Message: message,
		Data:    user,
		Error:   nil,
		Meta:    map[string]interface{}{},
	})
}

// CreateUser handles POST /users
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req d.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		message := i18n.T(c, "identify.errors.invalid_request.message", "Invalid request", nil)
		description := i18n.T(c, "identify.errors.invalid_request.description", fmt.Sprintf("Invalid request body: %s", err.Error()), map[string]interface{}{"error": err.Error()})
		c.JSON(http.StatusBadRequest, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "invalid_request", Description: description},
			Meta:    map[string]interface{}{},
		})
		return
	}

	ctx := c.Request.Context()

	// Create user through service
	user, err := h.userService.CreateUser(ctx, req)
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

	message := i18n.T(c, "identify.users.create_success", "User created successfully", nil)
	c.JSON(http.StatusCreated, r.StandardResponse{
		Success: true,
		Message: message,
		Data:    user,
		Error:   nil,
		Meta:    map[string]interface{}{},
	})
}

// UpdateUser handles PUT /users/:id
func (h *UserHandler) UpdateUser(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse user ID
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		message := i18n.T(c, "identify.errors.invalid_user_id.message", "Invalid user id", nil)
		description := i18n.T(c, "identify.errors.invalid_user_id.description", "User ID must be a valid UUID", map[string]interface{}{"id": c.Param("id")})
		c.JSON(http.StatusBadRequest, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "invalid_user_id", Description: description},
			Meta:    map[string]interface{}{},
		})
		return
	}

	var req d.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		message := i18n.T(c, "identify.errors.invalid_request.message", "Invalid request", nil)
		description := i18n.T(c, "identify.errors.invalid_request.description", fmt.Sprintf("Invalid request body: %s", err.Error()), map[string]interface{}{"error": err.Error()})
		c.JSON(http.StatusBadRequest, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "invalid_request", Description: description},
			Meta:    map[string]interface{}{},
		})
		return
	}

	// Check authorization
	currentUserID, err := getCurrentUserID(c)
	if err != nil {
		message := i18n.T(c, "identify.errors.unauthorized.message", "Unauthorized", nil)
		description := i18n.T(c, "identify.errors.unauthorized.description", "User authentication required", nil)
		c.JSON(http.StatusUnauthorized, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "unauthorized", Description: description},
			Meta:    map[string]interface{}{},
		})
		return
	}

	// Users can only update themselves, unless they're admin
	if currentUserID != userID && !isAdmin(c) {
		message := i18n.T(c, "identify.errors.access_denied.message", "Access denied", nil)
		description := i18n.T(c, "identify.errors.access_denied.description", "You can only update your own profile", nil)
		c.JSON(http.StatusForbidden, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "access_denied", Description: description},
			Meta:    map[string]interface{}{},
		})
		return
	}

	// Update user through service
	user, err := h.userService.UpdateUser(ctx, userID, req)
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

	message := i18n.T(c, "identify.users.update_success", "User updated successfully", nil)
	c.JSON(http.StatusOK, r.StandardResponse{
		Success: true,
		Message: message,
		Data:    user,
		Error:   nil,
		Meta:    map[string]interface{}{},
	})
}

// DeleteUser handles DELETE /users/:id
func (h *UserHandler) DeleteUser(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse user ID
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		message := i18n.T(c, "identify.errors.invalid_user_id.message", "Invalid user id", nil)
		description := i18n.T(c, "identify.errors.invalid_user_id.description", "User ID must be a valid UUID", map[string]interface{}{"id": c.Param("id")})
		c.JSON(http.StatusBadRequest, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "invalid_user_id", Description: description},
			Meta:    map[string]interface{}{},
		})
		return
	}

	// Check authorization
	currentUserID, err := getCurrentUserID(c)
	if err != nil {
		message := i18n.T(c, "identify.errors.unauthorized.message", "Unauthorized", nil)
		description := i18n.T(c, "identify.errors.unauthorized.description", "User authentication required", nil)
		c.JSON(http.StatusUnauthorized, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "unauthorized", Description: description},
			Meta:    map[string]interface{}{},
		})
		return
	}

	// Users can only delete themselves, unless they're admin
	if currentUserID != userID && !isAdmin(c) {
		message := i18n.T(c, "identify.errors.access_denied.message", "Access denied", nil)
		description := i18n.T(c, "identify.errors.access_denied.description", "You can only delete your own account", nil)
		c.JSON(http.StatusForbidden, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "access_denied", Description: description},
			Meta:    map[string]interface{}{},
		})
		return
	}

	// Delete user through service
	if err := h.userService.DeleteUser(ctx, userID); err != nil {
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

	message := i18n.T(c, "identify.users.delete_success", "User deleted successfully", nil)
	c.JSON(http.StatusOK, r.StandardResponse{
		Success: true,
		Message: message,
		Data:    map[string]interface{}{},
		Error:   nil,
		Meta:    map[string]interface{}{},
	})
}

// Helper functions - using global handlers functions