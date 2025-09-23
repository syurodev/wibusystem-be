package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	d "wibusystem/pkg/common/dto"
	r "wibusystem/pkg/common/response"
	"wibusystem/pkg/i18n"
	"wibusystem/services/identify/oauth2"
	"wibusystem/services/identify/services/interfaces"
)

// AuthHandler handles authentication-related endpoints
type AuthHandler struct {
	authService interfaces.AuthServiceInterface
	provider    *oauth2.Provider
	loc         *i18n.Translator
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService interfaces.AuthServiceInterface, provider *oauth2.Provider, translator *i18n.Translator) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		provider:    provider,
		loc:         translator,
	}
}

// Register handles user registration
func (h *AuthHandler) Register(c *gin.Context) {
	var req d.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		message := i18n.T(c, "identify.errors.invalid_request.message", "Invalid request", nil)
		description := i18n.T(c, "identify.errors.invalid_request.description", fmt.Sprintf("Invalid request body: %s", err.Error()), map[string]interface{}{"Error": err.Error()})
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

	// Register user through service
	user, err := h.authService.Register(ctx, req)
	if err != nil {
		status, code, message, description := mapServiceError(c, err, "register")
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
	message := i18n.T(c, "identify.auth.register.success", "User registered successfully", nil)
	c.JSON(http.StatusCreated, r.StandardResponse{
		Success: true,
		Message: message,
		Data:    user,
		Error:   nil,
		Meta:    map[string]interface{}{},
	})
}

// Login handles user login with password
func (h *AuthHandler) Login(c *gin.Context) {
	var req d.LoginRequest
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

	// Login through service
	result, err := h.authService.Login(ctx, req)
	if err != nil {
		status, code, message, description := mapServiceError(c, err, "login")
		c.JSON(status, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: code, Description: description},
			Meta:    map[string]interface{}{},
		})
		return
	}

	// Set session cookie for authorize flow
	if err := h.authService.SetUserSession(c, result.User.ID); err != nil {
		message := i18n.T(c, "identify.errors.server_error.message", "Session error", nil)
		description := i18n.T(c, "identify.errors.server_error.description", "Failed to set session", nil)
		c.JSON(http.StatusInternalServerError, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "server_error", Description: description},
			Meta:    map[string]interface{}{},
		})
		return
	}

	// Convert tenants to response format
	tenantsResponse := make([]interface{}, len(result.Tenants))
	for i, tenant := range result.Tenants {
		tenantsResponse[i] = *tenant
	}

	// Return success response
	message := i18n.T(c, "identify.auth.login.success", "Login successful", nil)
	c.JSON(http.StatusOK, r.StandardResponse{
		Success: true,
		Message: message,
		Data:    map[string]interface{}{"user": result.User, "tenants": tenantsResponse},
		Error:   nil,
		Meta:    map[string]interface{}{},
	})
}

// Logout handles user logout
func (h *AuthHandler) Logout(c *gin.Context) {
	// Clear user session
	h.authService.ClearUserSession(c)

	// Return success response
	message := i18n.T(c, "identify.auth.logout.success", "Logout successful", nil)
	c.JSON(http.StatusOK, r.StandardResponse{
		Success: true,
		Message: message,
		Data:    map[string]interface{}{},
		Error:   nil,
		Meta:    map[string]interface{}{},
	})
}

// RefreshToken handles token refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		message := i18n.T(c, "identify.errors.invalid_request.message", "Invalid request", nil)
		description := i18n.T(c, "identify.errors.refresh_token_required.description", "Refresh token is required", nil)
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

	// Refresh token through service
	result, err := h.authService.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		status, code, message, description := mapServiceError(c, err,"refresh")
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
	message := i18n.T(c, "identify.auth.refresh.success", "Token refreshed successfully", nil)
	c.JSON(http.StatusOK, r.StandardResponse{
		Success: true,
		Message: message,
		Data:    result,
		Error:   nil,
		Meta:    map[string]interface{}{},
	})
}

// Profile returns the current user's profile
func (h *AuthHandler) Profile(c *gin.Context) {
	userID, err := getUserID(c)
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

	ctx := c.Request.Context()

	// Get profile through service
	result, err := h.authService.GetProfile(ctx, userID)
	if err != nil {
		status, code, message, description := mapServiceError(c, err,"profile")
		c.JSON(status, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: code, Description: description},
			Meta:    map[string]interface{}{},
		})
		return
	}

	// Convert tenants to response format
	tenantsResponse := make([]interface{}, len(result.Tenants))
	for i, tenant := range result.Tenants {
		tenantsResponse[i] = *tenant
	}

	// Return success response
	message := i18n.T(c, "identify.auth.profile.success", "Profile fetched successfully", nil)
	c.JSON(http.StatusOK, r.StandardResponse{
		Success: true,
		Message: message,
		Data:    map[string]interface{}{"user": result.User, "tenants": tenantsResponse},
		Error:   nil,
		Meta:    map[string]interface{}{},
	})
}

// ChangePassword handles password change
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req struct {
		CurrentPassword string `json:"current_password" binding:"required"`
		NewPassword     string `json:"new_password" binding:"required,min=8"`
	}

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

	userID, err := getUserID(c)
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

	ctx := c.Request.Context()

	// Change password through service
	if err := h.authService.ChangePassword(ctx, userID, req.CurrentPassword, req.NewPassword); err != nil {
		status, code, message, description := mapServiceError(c, err,"change_password")
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
	message := i18n.T(c, "identify.auth.change_password.success", "Password changed successfully", nil)
	c.JSON(http.StatusOK, r.StandardResponse{
		Success: true,
		Message: message,
		Data:    map[string]interface{}{},
		Error:   nil,
		Meta:    map[string]interface{}{},
	})
}

// Helper functions

// getUserID extracts user ID from context
func getUserID(c *gin.Context) (uuid.UUID, error) {
	// This is a simplified version - in reality you'd extract from JWT token
	// For now, we'll use a placeholder
	userIDStr := c.GetHeader("X-User-ID") // Temporary header-based approach
	if userIDStr == "" {
		return uuid.Nil, fmt.Errorf("user ID not found")
	}

	return uuid.Parse(userIDStr)
}

