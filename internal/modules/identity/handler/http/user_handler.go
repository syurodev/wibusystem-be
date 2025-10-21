// Package http contains HTTP handlers for the Identity module.
package http

import (
	"log"
	"wibusystem/internal/modules/identity/repository"

	"wibusystem/internal/modules/identity/dto"
	"wibusystem/internal/modules/identity/handler/middleware"
	"wibusystem/internal/modules/identity/service"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// UserHandler handles user-related HTTP requests.
type UserHandler struct {
	userService service.UserService
}

// NewUserHandler creates a new UserHandler instance.
func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// GetProfile returns the authenticated user's profile.
// GET /api/v1/users/profile
func (h *UserHandler) GetProfile(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return middleware.NewAppError(middleware.ErrUnauthorized, "User not authenticated", fiber.StatusUnauthorized)
	}

	ctx := c.Context()
	user, err := h.userService.GetProfile(ctx, userID)
	if err != nil {
		log.Printf("[UserHandler.GetProfile] Error: %v", err)
		return middleware.NewAppError(err, "Failed to retrieve profile", fiber.StatusInternalServerError).WithCode("GET_PROFILE_FAILED")
	}

	userResponse := mapUserToResponse(user)
	return c.Status(fiber.StatusOK).JSON(dto.GetProfileResponse{
		User: userResponse,
	})
}

// UpdateProfile updates the authenticated user's profile.
// PUT /api/v1/users/profile
func (h *UserHandler) UpdateProfile(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return middleware.NewAppError(middleware.ErrUnauthorized, "User not authenticated", fiber.StatusUnauthorized)
	}

	var req dto.UpdateProfileRequest
	if err := middleware.ValidateRequest(c, &req); err != nil {
		return err
	}

	ctx := c.Context()
	user, err := h.userService.UpdateProfile(ctx, userID, req)
	if err != nil {
		log.Printf("[UserHandler.UpdateProfile] Error: %v", err)
		return middleware.NewAppError(err, "Failed to update profile", fiber.StatusBadRequest).WithCode("UPDATE_PROFILE_FAILED")
	}

	userResponse := mapUserToResponse(user)
	return c.Status(fiber.StatusOK).JSON(dto.UpdateProfileResponse{
		User:    userResponse,
		Message: "Profile updated successfully",
	})
}

// DeleteAccount deletes the authenticated user's account.
// DELETE /api/v1/users/account
func (h *UserHandler) DeleteAccount(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return middleware.NewAppError(middleware.ErrUnauthorized, "User not authenticated", fiber.StatusUnauthorized)
	}

	var req dto.DeleteAccountRequest
	if err := middleware.ValidateRequest(c, &req); err != nil {
		return err
	}

	if !req.ConfirmDeletion {
		return middleware.NewAppError(middleware.ErrBadRequest, "Account deletion must be confirmed", fiber.StatusBadRequest).WithCode("DELETION_NOT_CONFIRMED")
	}

	ctx := c.Context()
	if err := h.userService.DeleteAccount(ctx, userID, req.Password); err != nil {
		log.Printf("[UserHandler.DeleteAccount] Error: %v", err)
		return middleware.NewAppError(err, "Failed to delete account", fiber.StatusBadRequest).WithCode("DELETE_ACCOUNT_FAILED")
	}

	return c.Status(fiber.StatusOK).JSON(dto.DeleteAccountResponse{
		Message: "Account deleted successfully",
	})
}

// GetUserByID returns a user by their ID (admin only).
// GET /api/v1/users/:userId
func (h *UserHandler) GetUserByID(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Params("userId"))
	if err != nil {
		return middleware.NewAppError(middleware.ErrBadRequest, "Invalid user ID format", fiber.StatusBadRequest)
	}

	ctx := c.Context()
	user, err := h.userService.GetUserByID(ctx, userID)
	if err != nil {
		log.Printf("[UserHandler.GetUserByID] Error: %v", err)
		return middleware.NewAppError(middleware.ErrNotFound, "User not found", fiber.StatusNotFound).WithCode("USER_NOT_FOUND")
	}

	userResponse := mapUserToResponse(user)
	return c.Status(fiber.StatusOK).JSON(dto.GetProfileResponse{
		User: userResponse,
	})
}

// ListUsers returns a paginated list of users (admin only).
// GET /api/v1/users
func (h *UserHandler) ListUsers(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	pageSize := c.QueryInt("page_size", 20)
	if pageSize > 100 {
		pageSize = 100
	}

	filter := repository.UserListFilter{
		Limit:               pageSize,
		Offset:              (page - 1) * pageSize,
		DisplayNameContains: c.Query("search", ""),
	}

	ctx := c.Context()
	users, total, err := h.userService.ListUsers(ctx, filter)
	if err != nil {
		log.Printf("[UserHandler.ListUsers] Error: %v", err)
		return middleware.NewAppError(err, "Failed to list users", fiber.StatusInternalServerError).WithCode("LIST_USERS_FAILED")
	}

	userResponses := make([]dto.UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = mapUserToResponse(user)
	}

	totalPages := (total + pageSize - 1) / pageSize

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"users":       userResponses,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": totalPages,
	})
}

// SearchUsers searches for users by email or display name (admin only).
// GET /api/v1/users/search
func (h *UserHandler) SearchUsers(c *fiber.Ctx) error {
	query := c.Query("q", "")
	if query == "" {
		return middleware.NewAppError(middleware.ErrBadRequest, "Search query is required", fiber.StatusBadRequest)
	}

	limit := c.QueryInt("limit", 10)
	if limit > 50 {
		limit = 50
	}

	ctx := c.Context()
	users, err := h.userService.SearchUsers(ctx, query, limit)
	if err != nil {
		log.Printf("[UserHandler.SearchUsers] Error: %v", err)
		return middleware.NewAppError(err, "Failed to search users", fiber.StatusInternalServerError).WithCode("SEARCH_USERS_FAILED")
	}

	userResponses := make([]dto.UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = mapUserToResponse(user)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"users": userResponses,
		"query": query,
	})
}

// GetUserStats returns statistics about a user (admin only).
// GET /api/v1/users/:userId/stats
func (h *UserHandler) GetUserStats(c *fiber.Ctx) error {
	ctx := c.Context()
	stats, err := h.userService.GetUserStats(ctx)
	if err != nil {
		log.Printf("[UserHandler.GetUserStats] Error: %v", err)
		return middleware.NewAppError(err, "Failed to get user stats", fiber.StatusInternalServerError).WithCode("GET_USER_STATS_FAILED")
	}

	return c.Status(fiber.StatusOK).JSON(stats)
}

// ListSessions returns all active sessions for the authenticated user.
// GET /api/v1/users/sessions
func (h *UserHandler) ListSessions(c *fiber.Ctx) error {
	// This is a placeholder. The actual implementation is in SessionHandler.
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"message": "This endpoint is handled by the SessionHandler"})
}

// RevokeSession revokes a specific session for the authenticated user.
// DELETE /api/v1/users/sessions/:sessionId
func (h *UserHandler) RevokeSession(c *fiber.Ctx) error {
	// This is a placeholder. The actual implementation is in SessionHandler.
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"message": "This endpoint is handled by the SessionHandler"})
}

// RevokeAllSessions revokes all sessions except the current one.
// DELETE /api/v1/users/sessions
func (h *UserHandler) RevokeAllSessions(c *fiber.Ctx) error {
	// This is a placeholder. The actual implementation is in SessionHandler.
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"message": "This endpoint is handled by the SessionHandler"})
}
