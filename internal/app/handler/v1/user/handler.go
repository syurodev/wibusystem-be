package user

import (
	"net/http"
	"system/internal/app/middleware"
	"system/internal/domain"
	"system/pkg/util/response"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid/v5"
)

type Handler struct {
	userRepo domain.UserRepository
}

func NewHandler(userRepo domain.UserRepository) *Handler {
	return &Handler{
		userRepo: userRepo,
	}
}

// GetSettings handles GET /user/settings
func (h *Handler) GetSettings(c *gin.Context) {
	userIDStr, exists := middleware.GetUserID(c)
	if !exists {
		response.AbortWithError(c, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}

	userID, err := uuid.FromString(userIDStr)
	if err != nil {
		response.AbortWithError(c, http.StatusUnauthorized, "INVALID_USER_ID", "invalid_user_id")
		return
	}

	user, err := h.userRepo.GetByID(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "FETCH_USER_ERROR", "failed_to_get_user", nil)
		return
	}

	// Return empty object if settings is nil
	settings := user.Settings
	if settings == nil {
		settings = make(map[string]any)
	}

	response.Success(c, http.StatusOK, "settings_fetched_success", settings, nil)
}

// UpdateSettings handles PATCH /user/settings
func (h *Handler) UpdateSettings(c *gin.Context) {
	userIDStr, exists := middleware.GetUserID(c)
	if !exists {
		response.AbortWithError(c, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}

	userID, err := uuid.FromString(userIDStr)
	if err != nil {
		response.AbortWithError(c, http.StatusUnauthorized, "INVALID_USER_ID", "invalid_user_id")
		return
	}

	var req map[string]any
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_REQUEST", "invalid_request_body", nil)
		return
	}

	user, err := h.userRepo.GetByID(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "FETCH_USER_ERROR", "failed_to_get_user", nil)
		return
	}

	// Merge settings
	if user.Settings == nil {
		user.Settings = make(map[string]any)
	}

	for k, v := range req {
		user.Settings[k] = v
	}

	if err := h.userRepo.Update(c.Request.Context(), user); err != nil {
		response.Error(c, http.StatusInternalServerError, "UPDATE_ERROR", "failed_to_update_settings", nil)
		return
	}

	response.Success(c, http.StatusOK, "settings_updated_success", user.Settings, nil)
}
