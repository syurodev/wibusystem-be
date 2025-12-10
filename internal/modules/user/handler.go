package user

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid/v5"

	"system/internal/app/middleware"
	userdto "system/internal/dto/user"
	pkgerrors "system/pkg/errors"
	"system/pkg/util/response"
)

type Handler struct {
	userService UserService
}

func NewHandler(userService UserService) *Handler {
	return &Handler{
		userService: userService,
	}
}

// GetProfile handles GET /user/profile
func (h *Handler) GetProfile(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return
	}

	user, err := h.userService.GetProfile(c.Request.Context(), userID)
	if err != nil {
		if appErr, ok := pkgerrors.AsAppError(err); ok {
			response.Error(c, appErr.StatusCode, appErr.ErrCode, appErr.I18nKey, nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", I18nFetchUserError, nil)
		return
	}

	response.Success(c, http.StatusOK, I18nProfileFetchedSuccess, user, nil)
}

// UpdateProfile handles PUT /user/profile
func (h *Handler) UpdateProfile(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return
	}

	var req userdto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_FAILED", I18nInvalidRequest, nil)
		return
	}

	user, err := h.userService.UpdateProfile(c.Request.Context(), userID, req)
	if err != nil {
		if appErr, ok := pkgerrors.AsAppError(err); ok {
			response.Error(c, appErr.StatusCode, appErr.ErrCode, appErr.I18nKey, nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", I18nUpdateProfileError, nil)
		return
	}

	response.Success(c, http.StatusOK, I18nProfileUpdatedSuccess, user, nil)
}

// GetSettings handles GET /user/settings
func (h *Handler) GetSettings(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return
	}

	settings, err := h.userService.GetSettings(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", I18nFetchUserError, nil)
		return
	}

	response.Success(c, http.StatusOK, I18nSettingsFetchedSuccess, settings, nil)
}

// UpdateSettings handles PATCH /user/settings
func (h *Handler) UpdateSettings(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return
	}

	var req map[string]any
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_FAILED", I18nInvalidRequest, nil)
		return
	}

	settings, err := h.userService.UpdateSettings(c.Request.Context(), userID, req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", I18nUpdateSettingsError, nil)
		return
	}

	response.Success(c, http.StatusOK, I18nSettingsUpdatedSuccess, settings, nil)
}

// ChangePassword handles PUT /user/settings/security/password
func (h *Handler) ChangePassword(c *gin.Context) {
	userID, err := getUserIDFromContext(c)
	if err != nil {
		return
	}

	var req userdto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_FAILED", I18nInvalidRequest, nil)
		return
	}

	if err := h.userService.ChangePassword(c.Request.Context(), userID, req.CurrentPassword, req.NewPassword); err != nil {
		if appErr, ok := pkgerrors.AsAppError(err); ok {
			response.Error(c, appErr.StatusCode, appErr.ErrCode, appErr.I18nKey, nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", I18nUpdatePasswordError, nil)
		return
	}

	response.Success(c, http.StatusOK, I18nPasswordChangedSuccess, nil, nil)
}

// GetSessions handles GET /user/sessions
func (h *Handler) GetSessions(c *gin.Context) {
	userIDStr, exists := middleware.GetUserID(c)
	if !exists {
		response.AbortWithError(c, http.StatusUnauthorized, "UNAUTHORIZED", I18nAuthUnauthorized)
		return
	}

	sessions, err := h.userService.GetSessions(c.Request.Context(), userIDStr)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", I18nFetchSessionsError, nil)
		return
	}

	currentSessionID, _ := c.Cookie("session_id")

	resp := make([]userdto.SessionResponse, len(sessions))
	for i, s := range sessions {
		resp[i] = userdto.SessionResponse{
			ID:         s.SessionID,
			IP:         s.IP,
			UserAgent:  s.UserAgent,
			Device:     s.Device,
			ClientOS:   s.ClientOS,
			Browser:    s.Browser,
			LastActive: s.LastActive.String(),
			Current:    s.SessionID == currentSessionID,
		}
	}

	response.Success(c, http.StatusOK, I18nSessionsFetchedSuccess, resp, nil)
}

// DeleteSession handles DELETE /user/sessions/:id
func (h *Handler) DeleteSession(c *gin.Context) {
	userIDStr, exists := middleware.GetUserID(c)
	if !exists {
		response.AbortWithError(c, http.StatusUnauthorized, "UNAUTHORIZED", I18nAuthUnauthorized)
		return
	}

	sessionID := c.Param("id")
	if sessionID == "" {
		response.Error(c, http.StatusBadRequest, "BAD_REQUEST", I18nInvalidSessionID, nil)
		return
	}

	if err := h.userService.DeleteSession(c.Request.Context(), userIDStr, sessionID); err != nil {
		if appErr, ok := pkgerrors.AsAppError(err); ok {
			response.Error(c, appErr.StatusCode, appErr.ErrCode, appErr.I18nKey, nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", I18nDeleteSessionError, nil)
		return
	}

	response.Success(c, http.StatusOK, I18nSessionDeletedSuccess, nil, nil)
}

// getUserIDFromContext extracts user ID from context and handles errors
func getUserIDFromContext(c *gin.Context) (uuid.UUID, error) {
	userIDStr, exists := middleware.GetUserID(c)
	if !exists {
		response.AbortWithError(c, http.StatusUnauthorized, "UNAUTHORIZED", I18nAuthUnauthorized)
		return uuid.Nil, pkgerrors.Unauthorized(I18nAuthUnauthorized, "unauthorized")
	}

	userID, err := uuid.FromString(userIDStr)
	if err != nil {
		response.AbortWithError(c, http.StatusBadRequest, "BAD_REQUEST", I18nInvalidUserID)
		return uuid.Nil, err
	}

	return userID, nil
}

