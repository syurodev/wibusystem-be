package user

import (
	"net/http"
	"system/internal/app/middleware"
	"system/pkg/util/crypto"
	"system/pkg/util/i18nkeys"
	"system/pkg/util/response"
	"system/pkg/util/validator"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid/v5"
)

// UpdateProfileRequest defines the request body for updating user profile
type UpdateProfileRequest struct {
	FullName    *string `json:"full_name"`
	DisplayName *string `json:"display_name"`
	Username    *string `json:"username"`
	Bio         any     `json:"bio"` // PlateJS content (JSON)
	AvatarURL   *string `json:"avatar_url"`
}

// ChangePasswordRequest defines the request body for changing password
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}

// GetProfile handles GET /user/profile
func (h *Handler) GetProfile(c *gin.Context) {
	userIDStr, exists := middleware.GetUserID(c)
	if !exists {
		response.AbortWithError(c, http.StatusUnauthorized, "UNAUTHORIZED", i18nkeys.AuthUnauthorized)
		return
	}

	userID, err := uuid.FromString(userIDStr)
	if err != nil {
		response.AbortWithError(c, http.StatusUnauthorized, "INVALID_USER_ID", i18nkeys.AccountInvalidUserID)
		return
	}

	user, err := h.userRepo.GetByID(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "FETCH_USER_ERROR", i18nkeys.AccountFetchUserError, nil)
		return
	}

	// We can return the user object directly now that it has JSON tags
	// PasswordHash is excluded via json:"-"
	response.Success(c, http.StatusOK, i18nkeys.AccountProfileFetchedSuccess, user, nil)
}

// UpdateProfile handles PUT /user/profile
func (h *Handler) UpdateProfile(c *gin.Context) {
	userIDStr, exists := middleware.GetUserID(c)
	if !exists {
		response.AbortWithError(c, http.StatusUnauthorized, "UNAUTHORIZED", i18nkeys.AuthUnauthorized)
		return
	}

	userID, err := uuid.FromString(userIDStr)
	if err != nil {
		response.AbortWithError(c, http.StatusUnauthorized, "INVALID_USER_ID", i18nkeys.AccountInvalidUserID)
		return
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_REQUEST", i18nkeys.AccountInvalidRequest, nil)
		return
	}

	user, err := h.userRepo.GetByID(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "FETCH_USER_ERROR", i18nkeys.AccountFetchUserError, nil)
		return
	}

	// Update fields if provided
	if req.FullName != nil {
		user.FullName = req.FullName
	}
	if req.DisplayName != nil {
		user.DisplayName = req.DisplayName
	}
	if req.AvatarURL != nil {
		user.AvatarURL = req.AvatarURL
	}
	if req.Bio != nil {
		// Bio is []any (JSONB array)
		if bioArray, ok := req.Bio.([]any); ok {
			user.Bio = bioArray
		}
	}
	if req.Username != nil {
		newUsername := *req.Username
		
		// If user already has a username, prevent changing it
		if user.Username != nil && *user.Username != "" {
			if *user.Username != newUsername {
				response.Error(c, http.StatusForbidden, "USERNAME_IMMUTABLE", i18nkeys.AccountUsernameImmutable, nil)
				return
			}
			// If same, do nothing
		} else {
			// Basic validation
			if len(newUsername) < 3 || len(newUsername) > 30 {
				response.Error(c, http.StatusBadRequest, "INVALID_USERNAME", i18nkeys.AccountUsernameLengthInvalid, nil)
				return
			}

			// Check uniqueness
			existingUser, err := h.userRepo.GetByUsername(c.Request.Context(), newUsername)
			if err == nil && existingUser != nil {
				response.Error(c, http.StatusConflict, "USERNAME_TAKEN", i18nkeys.AccountUsernameTaken, nil)
				return
			}
			user.Username = req.Username
		}
	}

	if err := h.userRepo.Update(c.Request.Context(), user); err != nil {
		response.Error(c, http.StatusInternalServerError, "UPDATE_ERROR", i18nkeys.AccountUpdateProfileError, nil)
		return
	}

	response.Success(c, http.StatusOK, i18nkeys.AccountProfileUpdatedSuccess, user, nil)
}

// ChangePassword handles PUT /user/settings/security/password
func (h *Handler) ChangePassword(c *gin.Context) {
	userIDStr, exists := middleware.GetUserID(c)
	if !exists {
		response.AbortWithError(c, http.StatusUnauthorized, "UNAUTHORIZED", i18nkeys.AuthUnauthorized)
		return
	}

	userID, err := uuid.FromString(userIDStr)
	if err != nil {
		response.AbortWithError(c, http.StatusUnauthorized, "INVALID_USER_ID", i18nkeys.AccountInvalidUserID)
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_REQUEST", i18nkeys.AccountInvalidRequest, nil)
		return
	}

	// Validate new password strength
	if err := validator.ValidatePasswordStrength(req.NewPassword); err != nil {
		response.Error(c, http.StatusBadRequest, "WEAK_PASSWORD", i18nkeys.AccountWeakPassword, nil)
		return
	}

	user, err := h.userRepo.GetByID(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "FETCH_USER_ERROR", i18nkeys.AccountFetchUserError, nil)
		return
	}

	// Verify current password
	if !crypto.VerifyPassword(user.PasswordHash, req.CurrentPassword) {
		response.Error(c, http.StatusUnauthorized, "INVALID_PASSWORD", i18nkeys.AccountInvalidPassword, nil)
		return
	}

	// Hash new password
	hashedPassword, err := crypto.HashPassword(req.NewPassword)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "HASH_ERROR", i18nkeys.AccountHashError, nil)
		return
	}

	user.PasswordHash = hashedPassword

	if err := h.userRepo.Update(c.Request.Context(), user); err != nil {
		response.Error(c, http.StatusInternalServerError, "UPDATE_ERROR", i18nkeys.AccountUpdatePasswordError, nil)
		return
	}

	response.Success(c, http.StatusOK, i18nkeys.AccountPasswordChangedSuccess, nil, nil)
}

// GetSessions handles GET /user/sessions
func (h *Handler) GetSessions(c *gin.Context) {
	userIDStr, exists := middleware.GetUserID(c)
	if !exists {
		response.AbortWithError(c, http.StatusUnauthorized, "UNAUTHORIZED", i18nkeys.AuthUnauthorized)
		return
	}

	sessions, err := h.sessionRepo.GetUserSessions(c.Request.Context(), userIDStr)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "FETCH_SESSIONS_ERROR", i18nkeys.AccountFetchSessionsError, nil)
		return
	}

	// Calculate current session
	currentSessionID, _ := c.Cookie("session_id")

	type SessionResponse struct {
		ID         string `json:"id"`
		IP         string `json:"ip"`
		UserAgent  string `json:"user_agent"`
		Device     string `json:"device"` // Extracted from UA if possible, or just raw
		ClientOS   string `json:"client_os"`
		Browser    string `json:"browser"`
		LastActive string `json:"last_active"`
		Current    bool   `json:"current"`
	}

	resp := []SessionResponse{}
	for _, s := range sessions {
		resp = append(resp, SessionResponse{
			ID:         s.SessionID,
			IP:         s.IP,
			UserAgent:  s.UserAgent,
			Device:     s.Device,
			ClientOS:   s.ClientOS,
			Browser:    s.Browser,
			LastActive: s.LastActive.String(),
			Current:    s.SessionID == currentSessionID,
		})
	}

	response.Success(c, http.StatusOK, i18nkeys.AccountSessionsFetchedSuccess, resp, nil)
}

// DeleteSession handles DELETE /user/sessions/:id
func (h *Handler) DeleteSession(c *gin.Context) {
	userIDStr, exists := middleware.GetUserID(c)
	if !exists {
		response.AbortWithError(c, http.StatusUnauthorized, "UNAUTHORIZED", i18nkeys.AuthUnauthorized)
		return
	}

	sessionID := c.Param("id")
	if sessionID == "" {
		response.Error(c, http.StatusBadRequest, "INVALID_REQUEST", i18nkeys.AccountInvalidSessionID, nil)
		return
	}

	// Verify ownership
	session, err := h.sessionRepo.GetSession(c.Request.Context(), sessionID)
	if err != nil {
		// If session not found, return success (already deleted)
		response.Success(c, http.StatusOK, i18nkeys.AccountSessionDeletedSuccess, nil, nil)
		return
	}

	if session.UserID != userIDStr {
		response.Error(c, http.StatusForbidden, "FORBIDDEN", i18nkeys.AuthForbidden, nil)
		return
	}

	if err := h.sessionRepo.DeleteSession(c.Request.Context(), sessionID); err != nil {
		response.Error(c, http.StatusInternalServerError, "DELETE_ERROR", i18nkeys.AccountDeleteSessionError, nil)
		return
	}

	response.Success(c, http.StatusOK, i18nkeys.AccountSessionDeletedSuccess, nil, nil)
}
