package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	d "wibusystem/pkg/common/dto"
	m "wibusystem/pkg/common/model"
	r "wibusystem/pkg/common/response"
	"wibusystem/pkg/i18n"
	"wibusystem/services/identify/oauth2"
	"wibusystem/services/identify/repositories"
	"wibusystem/services/identify/session"
)

// AuthHandler handles authentication-related endpoints
type AuthHandler struct {
	repos    *repositories.Repositories
	provider *oauth2.Provider
	sess     *session.Manager
	loc      *i18n.Translator
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(repos *repositories.Repositories, provider *oauth2.Provider, sess *session.Manager, translator *i18n.Translator) *AuthHandler {
	return &AuthHandler{
		repos:    repos,
		provider: provider,
		sess:     sess,
		loc:      translator,
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

	// Check if a user already exists
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

	// Return user without sensitive information
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

	// Find the user by email
	user, err := h.repos.User.GetByEmail(ctx, req.Email)
	if err != nil {
		message := i18n.T(c, "identify.errors.invalid_credentials.message", "Invalid credentials", nil)
		description := i18n.T(c, "identify.errors.invalid_credentials.description", "Invalid email or password", map[string]interface{}{"email": req.Email})
		c.JSON(http.StatusUnauthorized, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "invalid_credentials", Description: description},
			Meta:    map[string]interface{}{},
		})
		return
	}

	// Get password credential
	credential, err := h.repos.Credential.GetByUserIDAndType(ctx, user.ID, m.AuthTypePassword)
	if err != nil || credential.SecretHash == nil {
		message := i18n.T(c, "identify.errors.invalid_credentials.message", "Invalid credentials", nil)
		description := i18n.T(c, "identify.errors.invalid_credentials.description", "Invalid email or password", map[string]interface{}{"email": req.Email})
		c.JSON(http.StatusUnauthorized, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "invalid_credentials", Description: description},
			Meta:    map[string]interface{}{},
		})
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(*credential.SecretHash), []byte(req.Password)); err != nil {
		message := i18n.T(c, "identify.errors.invalid_credentials.message", "Invalid credentials", nil)
		description := i18n.T(c, "identify.errors.invalid_credentials.description", "Invalid email or password", map[string]interface{}{"email": req.Email})
		c.JSON(http.StatusUnauthorized, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "invalid_credentials", Description: description},
			Meta:    map[string]interface{}{},
		})
		return
	}

	// Update last login
	_ = h.repos.User.UpdateLastLogin(ctx, user.ID)
	_ = h.repos.Credential.UpdateLastUsed(ctx, credential.ID)

	// Get user tenants
	tenants, err := h.repos.Tenant.GetByUserID(ctx, user.ID)
	if err != nil {
		tenants = []*m.Tenant{} // Empty slice if no tenants
	}

	// Convert to response format
	tenantsResponse := make([]m.Tenant, len(tenants))
	for i, tenant := range tenants {
		tenantsResponse[i] = *tenant
	}

	// Set session cookie for authorize flow
	if err := h.sess.Set(c, user.ID.String()); err != nil {
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

	// Respond success and include basic info
	message := i18n.T(c, "identify.auth.login.success", "Login successful", nil)
	c.JSON(http.StatusOK, r.StandardResponse{
		Success: true,
		Message: message,
		Data:    map[string]interface{}{"user": user, "tenants": tenantsResponse},
		Error:   nil,
		Meta:    map[string]interface{}{},
	})
}

// Logout handles user logout
func (h *AuthHandler) Logout(c *gin.Context) {
	fmt.Printf("DEBUG: Logout endpoint called - Method: %s, Path: %s\n", c.Request.Method, c.Request.URL.Path)
	fmt.Printf("DEBUG: Headers: %+v\n", c.Request.Header)
	fmt.Printf("DEBUG: Cookies: %+v\n", c.Request.Cookies())

	// In a full implementation, this would:
	// 1. Revoke the access token
	// 2. Revoke the refresh token
	// 3. Clear session cookies
	// 4. Log the logout event

	fmt.Printf("DEBUG: Calling session Clear\n")
	h.sess.Clear(c)
	fmt.Printf("DEBUG: Session cleared, responding with success\n")
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

	// In a full implementation, this would:
	// 1. Validate the refresh token
	// 2. Check if it's not expired or revoked
	// 3. Generate new access token
	// 4. Optionally rotate refresh token

	message := i18n.T(c, "identify.auth.refresh.success", "Token refreshed successfully", nil)
	c.JSON(http.StatusOK, r.StandardResponse{
		Success: true,
		Message: message,
		Data: map[string]interface{}{
			"access_token":  "new-access-token",
			"token_type":    "Bearer",
			"expires_in":    3600,
			"refresh_token": "new-refresh-token",
		},
		Error: nil,
		Meta:  map[string]interface{}{},
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
	user, err := h.repos.User.GetByID(ctx, userID)
	if err != nil {
		message := i18n.T(c, "identify.errors.user_not_found.message", "User not found", nil)
		description := i18n.T(c, "identify.errors.user_not_found.description", "User not found", nil)
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
	tenants, err := h.repos.Tenant.GetByUserID(ctx, user.ID)
	if err != nil {
		tenants = []*m.Tenant{}
	}

	// Convert to response format
	tenantsResponse := make([]m.Tenant, len(tenants))
	for i, tenant := range tenants {
		tenantsResponse[i] = *tenant
	}

	message := i18n.T(c, "identify.auth.profile.success", "Profile fetched successfully", nil)
	c.JSON(http.StatusOK, r.StandardResponse{
		Success: true,
		Message: message,
		Data:    map[string]interface{}{"user": user, "tenants": tenantsResponse},
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

	// Get current password credential
	credential, err := h.repos.Credential.GetByUserIDAndType(ctx, userID, m.AuthTypePassword)
	if err != nil || credential.SecretHash == nil {
		message := i18n.T(c, "identify.errors.no_password_set.message", "No password set", nil)
		description := i18n.T(c, "identify.errors.no_password_set.description", "No password is currently set for this user", nil)
		c.JSON(http.StatusBadRequest, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "no_password_set", Description: description},
			Meta:    map[string]interface{}{},
		})
		return
	}

	// Verify current password
	if err := bcrypt.CompareHashAndPassword([]byte(*credential.SecretHash), []byte(req.CurrentPassword)); err != nil {
		message := i18n.T(c, "identify.errors.invalid_current_password.message", "Invalid current password", nil)
		description := i18n.T(c, "identify.errors.invalid_current_password.description", "Current password is incorrect", nil)
		c.JSON(http.StatusUnauthorized, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "invalid_current_password", Description: description},
			Meta:    map[string]interface{}{},
		})
		return
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), 12)
	if err != nil {
		message := i18n.T(c, "identify.errors.password_hash_failed.message", "Password hash failed", nil)
		description := i18n.T(c, "identify.errors.password_hash_failed.description", "Failed to hash new password", nil)
		c.JSON(http.StatusInternalServerError, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "password_hash_failed", Description: description},
			Meta:    map[string]interface{}{},
		})
		return
	}

	// Update credential
	credential.SecretHash = stringPtr(string(hashedPassword))
	if err := h.repos.Credential.Update(ctx, credential); err != nil {
		message := i18n.T(c, "identify.errors.password_update_failed.message", "Password update failed", nil)
		description := i18n.T(c, "identify.errors.password_update_failed.description", "Failed to update password", nil)
		c.JSON(http.StatusInternalServerError, r.StandardResponse{
			Success: false,
			Message: message,
			Data:    nil,
			Error:   &r.ErrorDetail{Code: "password_update_failed", Description: description},
			Meta:    map[string]interface{}{},
		})
		return
	}

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

// stringPtr returns a pointer to string
func stringPtr(s string) *string {
	return &s
}
