package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	d "wibusystem/pkg/common/dto"
	m "wibusystem/pkg/common/model"
	r "wibusystem/pkg/common/response"
	"wibusystem/pkg/i18n"
	"wibusystem/services/identify/middleware"
	"wibusystem/services/identify/repositories"
)

// UserHandler handles user management endpoints
type UserHandler struct {
	repos *repositories.Repositories
	loc   *i18n.Translator
}

// NewUserHandler creates a new user handler
func NewUserHandler(repos *repositories.Repositories, translator *i18n.Translator) *UserHandler {
	return &UserHandler{
		repos: repos,
		loc:   translator,
	}
}

// ListUsers handles GET /users
func (h *UserHandler) ListUsers(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	// Get users with pagination
	users, total, err := h.repos.User.List(ctx, pageSize, offset)
	if err != nil {
		message := i18n.T(c, "identify.errors.database_error.message", "Database error", nil)
		description := i18n.T(c, "identify.errors.database_error.description", fmt.Sprintf("Failed to fetch users: %s", err.Error()), map[string]interface{}{"error": err.Error()})
		c.JSON(http.StatusInternalServerError, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "database_error", Description: description},
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
		description := i18n.T(c, "identify.errors.invalid_user_id.description", "User ID must be a valid integer", map[string]interface{}{"id": c.Param("id")})
		c.JSON(http.StatusBadRequest, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "invalid_user_id", Description: description},
			Meta:    map[string]interface{}{},
		})
		return
	}

	// Get user
	user, err := h.repos.User.GetByID(ctx, userID)
	if err != nil {
		message := i18n.T(c, "identify.errors.user_not_found.message", "User not found", nil)
		description := i18n.T(c, "identify.errors.user_not_found.description", "User not found", map[string]interface{}{"id": userID.String()})
		c.JSON(http.StatusNotFound, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "user_not_found", Description: description},
			Meta:    map[string]interface{}{},
		})
		return
	}

	// Get user tenants if requested
	includeTenants := c.Query("include_tenants") == "true"
	var tenants []*m.Tenant
	if includeTenants {
		tenants, _ = h.repos.Tenant.GetByUserID(ctx, userID)
	}

	// Get user memberships if requested
	includeMemberships := c.Query("include_memberships") == "true"
	var memberships []*m.Membership
	if includeMemberships {
		memberships, _ = h.repos.Membership.ListByUserID(ctx, userID)
	}

	responseData := map[string]interface{}{
		"user": user,
	}

	if includeTenants {
		responseData["tenants"] = tenants
	}

	if includeMemberships {
		responseData["memberships"] = memberships
	}

	message := i18n.T(c, "identify.users.get_success", "User fetched successfully", nil)
	c.JSON(http.StatusOK, r.StandardResponse{
		Success: true,
		Message: message,
		Data:    responseData,
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

	// Check if user already exists
	existingUser, _ := h.repos.User.GetByEmail(ctx, req.Email)
	if existingUser != nil {
		message := i18n.T(c, "identify.errors.user_exists.message", "User exists", nil)
		description := i18n.T(c, "identify.errors.user_exists.description", "User with this email already exists", map[string]interface{}{"email": req.Email})
		c.JSON(http.StatusConflict, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "user_exists", Description: description},
			Meta:    map[string]interface{}{},
		})
		return
	}

	// Check username uniqueness if provided
	if req.Username != "" {
		existingUser, _ = h.repos.User.GetByUsername(ctx, req.Username)
		if existingUser != nil {
			message := i18n.T(c, "identify.errors.username_taken.message", "Username taken", nil)
			description := i18n.T(c, "identify.errors.username_taken.description", "Username is already taken", map[string]interface{}{"username": req.Username})
			c.JSON(http.StatusConflict, r.StandardResponse{
				Success: false,
				Message: message,
				Data:    nil,
				Error:   &r.ErrorDetail{Code: "username_taken", Description: description},
				Meta:    map[string]interface{}{},
			})
			return
		}
	}

	// Create user
	user := &m.User{
		Email:       req.Email,
		Username:    req.Username,
		DisplayName: req.DisplayName,
	}

	if err := h.repos.User.Create(ctx, user); err != nil {
		message := i18n.T(c, "identify.errors.user_creation_failed.message", "User creation failed", nil)
		description := i18n.T(c, "identify.errors.user_creation_failed.description", fmt.Sprintf("Failed to create user: %s", err.Error()), map[string]interface{}{"error": err.Error()})
		c.JSON(http.StatusInternalServerError, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "user_creation_failed", Description: description},
			Meta:    map[string]interface{}{},
		})
		return
	}

	// Create password credential if provided
	if req.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
		if err != nil {
			message := i18n.T(c, "identify.errors.password_hash_failed.message", "Password hash failed", nil)
			description := i18n.T(c, "identify.errors.password_hash_failed.description", "Failed to hash password", nil)
			c.JSON(http.StatusInternalServerError, r.StandardResponse{
				Success: false,
				Message: message,
				Data:    nil,
				Error:   &r.ErrorDetail{Code: "password_hash_failed", Description: description},
				Meta:    map[string]interface{}{},
			})
			return
		}

		credential := &m.Credential{
			UserID:     user.ID,
			Type:       m.AuthTypePassword,
			Identifier: &req.Email,
			SecretHash: stringPtr(string(hashedPassword)),
		}

		if err := h.repos.Credential.Create(ctx, credential); err != nil {
			message := i18n.T(c, "identify.errors.credential_creation_failed.message", "Credential creation failed", nil)
			description := i18n.T(c, "identify.errors.credential_creation_failed.description", "Failed to create password credential", nil)
			c.JSON(http.StatusInternalServerError, r.StandardResponse{
				Success: false,
				Message: message,
				Data:    nil,
				Error:   &r.ErrorDetail{Code: "credential_creation_failed", Description: description},
				Meta:    map[string]interface{}{},
			})
			return
		}
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
		description := i18n.T(c, "identify.errors.invalid_user_id.description", "User ID must be a valid integer", map[string]interface{}{"id": c.Param("id")})
		c.JSON(http.StatusBadRequest, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "invalid_user_id", Description: description},
			Meta:    map[string]interface{}{},
		})
		return
	}

	// Check if user exists
	user, err := h.repos.User.GetByID(ctx, userID)
	if err != nil {
		message := i18n.T(c, "identify.errors.user_not_found.message", "User not found", nil)
		description := i18n.T(c, "identify.errors.user_not_found.description", "User not found", map[string]interface{}{"id": userID.String()})
		c.JSON(http.StatusNotFound, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "user_not_found", Description: description},
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

	// Check if the current user can update this user
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

	// Update user fields
	if req.Username != nil {
		// Check username uniqueness
		if *req.Username != user.Username {
			existingUser, _ := h.repos.User.GetByUsername(ctx, *req.Username)
			if existingUser != nil && existingUser.ID != userID {
				message := i18n.T(c, "identify.errors.username_taken.message", "Username taken", nil)
				description := i18n.T(c, "identify.errors.username_taken.description", "Username is already taken", map[string]interface{}{"username": *req.Username})
				c.JSON(http.StatusConflict, r.StandardResponse{
					Success: false,
					Message: message,
					Data:    nil,
					Error:   &r.ErrorDetail{Code: "username_taken", Description: description},
					Meta:    map[string]interface{}{},
				})
				return
			}
		}
		user.Username = *req.Username
	}

	if req.DisplayName != nil {
		user.DisplayName = *req.DisplayName
	}

	// Update user
	if err := h.repos.User.Update(ctx, user); err != nil {
		message := i18n.T(c, "identify.errors.user_update_failed.message", "User update failed", nil)
		description := i18n.T(c, "identify.errors.user_update_failed.description", fmt.Sprintf("Failed to update user: %s", err.Error()), map[string]interface{}{"error": err.Error()})
		c.JSON(http.StatusInternalServerError, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "user_update_failed", Description: description},
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
		description := i18n.T(c, "identify.errors.invalid_user_id.description", "User ID must be a valid integer", map[string]interface{}{"id": c.Param("id")})
		c.JSON(http.StatusBadRequest, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "invalid_user_id", Description: description},
			Meta:    map[string]interface{}{},
		})
		return
	}

	// Check if user exists
	_, err = h.repos.User.GetByID(ctx, userID)
	if err != nil {
		message := i18n.T(c, "identify.errors.user_not_found.message", "User not found", nil)
		description := i18n.T(c, "identify.errors.user_not_found.description", "User not found", map[string]interface{}{"id": userID.String()})
		c.JSON(http.StatusNotFound, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "user_not_found", Description: description},
			Meta:    map[string]interface{}{},
		})
		return
	}

	// Check if the current user can delete this user
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

	// Delete user (this will cascade to credentials, memberships, etc.)
	if err := h.repos.User.Delete(ctx, userID); err != nil {
		message := i18n.T(c, "identify.errors.user_deletion_failed.message", "User deletion failed", nil)
		description := i18n.T(c, "identify.errors.user_deletion_failed.description", fmt.Sprintf("Failed to delete user: %s", err.Error()), map[string]interface{}{"error": err.Error()})
		c.JSON(http.StatusInternalServerError, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "user_deletion_failed", Description: description},
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

// GetUserTenants handles GET /users/:id/tenants
func (h *UserHandler) GetUserTenants(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse user ID
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		message := i18n.T(c, "identify.errors.invalid_user_id.message", "Invalid user id", nil)
		description := i18n.T(c, "identify.errors.invalid_user_id.description", "User ID must be a valid integer", map[string]interface{}{"id": c.Param("id")})
		c.JSON(http.StatusBadRequest, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "invalid_user_id", Description: description},
			Meta:    map[string]interface{}{},
		})
		return
	}

	// Check if user exists
	_, err = h.repos.User.GetByID(ctx, userID)
	if err != nil {
		message := i18n.T(c, "identify.errors.user_not_found.message", "User not found", nil)
		description := i18n.T(c, "identify.errors.user_not_found.description", "User not found", map[string]interface{}{"id": userID.String()})
		c.JSON(http.StatusNotFound, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "user_not_found", Description: description},
			Meta:    map[string]interface{}{},
		})
		return
	}

	// Get user tenants
	tenants, err := h.repos.Tenant.GetByUserID(ctx, userID)
	if err != nil {
		message := i18n.T(c, "identify.errors.database_error.message", "Database error", nil)
		description := i18n.T(c, "identify.errors.database_error.description", fmt.Sprintf("Failed to fetch user tenants: %s", err.Error()), map[string]interface{}{"error": err.Error()})
		c.JSON(http.StatusInternalServerError, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "database_error", Description: description},
			Meta:    map[string]interface{}{},
		})
		return
	}

	message := i18n.T(c, "identify.users.tenants_success", "User tenants fetched successfully", nil)
	c.JSON(http.StatusOK, r.StandardResponse{
		Success: true,
		Message: message,
		Data:    tenants,
		Error:   nil,
		Meta:    map[string]interface{}{},
	})
}

// Helper functions

// getCurrentUserID extracts the current authenticated user ID.
//
// TODO: This is a placeholder implementation that always returns 1.
// Replace with proper extraction from middleware user info (e.g., subject).
func getCurrentUserID(c *gin.Context) (uuid.UUID, error) {
	_, exists := middleware.GetUserFromContext(c)
	if !exists {
		return uuid.Nil, fmt.Errorf("user not authenticated")
	}

	// Parse subject as user ID
	// This is simplified - in production you'd have proper user ID extraction
	// For now return a placeholder UUID
	placeholderID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	return placeholderID, nil // Placeholder
}

// isAdmin checks if the current user has admin privileges via the "admin" scope.
func isAdmin(c *gin.Context) bool {
	userInfo, exists := middleware.GetUserFromContext(c)
	if !exists {
		return false
	}

	return userInfo.Scopes.Has("admin")
}
