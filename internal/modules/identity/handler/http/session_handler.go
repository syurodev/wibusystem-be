// Package http contains HTTP handlers for the Identity module.
package http

import (
	"log"

	"wibusystem/internal/modules/identity/dto"
	"wibusystem/internal/modules/identity/handler/middleware"
	"wibusystem/internal/modules/identity/service"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// SessionHandler handles session-related HTTP requests.
type SessionHandler struct {
	authService service.AuthService
}

// NewSessionHandler creates a new SessionHandler instance.
func NewSessionHandler(authService service.AuthService) *SessionHandler {
	return &SessionHandler{
		authService: authService,
	}
}

// ListUserSessions returns all active sessions for the authenticated user.
// GET /api/v1/sessions
func (h *SessionHandler) ListUserSessions(c *fiber.Ctx) error {
	// Get authenticated user ID
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return middleware.NewAppError(
			middleware.ErrUnauthorized,
			"User not authenticated",
			fiber.StatusUnauthorized,
		)
	}

	// Get current session ID to mark it
	currentSessionID, _ := middleware.GetSessionID(c)

	// Parse query parameters
	activeOnly := c.QueryBool("active_only", true)

	// Call service
	ctx := c.Context()
	sessions, err := h.authService.GetUserSessions(ctx, userID)
	if err != nil {
		log.Printf("[SessionHandler.ListUserSessions] Error: %v", err)
		return middleware.NewAppError(
			err,
			"Failed to list sessions",
			fiber.StatusInternalServerError,
		).WithCode("LIST_SESSIONS_FAILED")
	}

	// Convert to response
	sessionResponses := make([]dto.SessionResponse, 0, len(sessions))
	for _, session := range sessions {
		metadata := session.ToMetadata()
		// Skip inactive sessions if activeOnly is true
		if activeOnly && !metadata.IsActive {
			continue
		}

		sessionResponse := dto.SessionResponse{
			ID:             session.ID,
			IPAddress:      session.IPAddress,
			UserAgent:      session.UserAgent,
			Browser:        metadata.Browser,
			OS:             metadata.OS,
			Device:         metadata.Device,
			CreatedAt:      session.CreatedAt,
			LastAccessedAt: session.LastAccessedAt,
			ExpiresAt:      session.ExpiresAt,
			IsActive:       metadata.IsActive,
			IsCurrent:      session.ID == currentSessionID,
		}
		sessionResponses = append(sessionResponses, sessionResponse)
	}

	// Return response
	return c.Status(fiber.StatusOK).JSON(dto.ListSessionsResponse{
		Sessions: sessionResponses,
		Total:    len(sessionResponses),
	})
}

// GetSession returns details of a specific session.
// GET /api/v1/sessions/:sessionId
func (h *SessionHandler) GetSession(c *fiber.Ctx) error {
	// Get authenticated user ID
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return middleware.NewAppError(
			middleware.ErrUnauthorized,
			"User not authenticated",
			fiber.StatusUnauthorized,
		)
	}

	// Parse session ID from params
	sessionIDParam := c.Params("sessionId")
	if sessionIDParam == "" {
		return middleware.NewAppError(
			middleware.ErrBadRequest,
			"Session ID is required",
			fiber.StatusBadRequest,
		)
	}

	sessionID, err := uuid.Parse(sessionIDParam)
	if err != nil {
		return middleware.NewAppError(
			middleware.ErrBadRequest,
			"Invalid session ID format",
			fiber.StatusBadRequest,
		)
	}

	// Call service to get session
	ctx := c.Context()
	session, err := h.authService.GetSession(ctx, sessionID)
	if err != nil {
		log.Printf("[SessionHandler.GetSession] Error: %v", err)
		return middleware.NewAppError(
			middleware.ErrNotFound,
			"Session not found",
			fiber.StatusNotFound,
		).WithCode("SESSION_NOT_FOUND")
	}

	// Verify the session belongs to the authenticated user
	if session.UserID != userID {
		return middleware.NewAppError(
			middleware.ErrForbidden,
			"You don't have permission to access this session",
			fiber.StatusForbidden,
		).WithCode("SESSION_ACCESS_DENIED")
	}

	// Get current session ID to mark it
	currentSessionID, _ := middleware.GetSessionID(c)

	// Convert to response
	metadata := session.ToMetadata()
	sessionResponse := dto.SessionResponse{
		ID:             session.ID,
		IPAddress:      session.IPAddress,
		UserAgent:      session.UserAgent,
		Browser:        metadata.Browser,
		OS:             metadata.OS,
		Device:         metadata.Device,
		CreatedAt:      session.CreatedAt,
		LastAccessedAt: session.LastAccessedAt,
		ExpiresAt:      session.ExpiresAt,
		IsActive:       metadata.IsActive,
		IsCurrent:      session.ID == currentSessionID,
	}

	// Return response
	return c.Status(fiber.StatusOK).JSON(sessionResponse)
}

// RevokeSession revokes a specific session.
// DELETE /api/v1/sessions/:sessionId
func (h *SessionHandler) RevokeSession(c *fiber.Ctx) error {
	// Get authenticated user ID
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return middleware.NewAppError(
			middleware.ErrUnauthorized,
			"User not authenticated",
			fiber.StatusUnauthorized,
		)
	}

	// Parse session ID from params
	sessionIDParam := c.Params("sessionId")
	if sessionIDParam == "" {
		return middleware.NewAppError(
			middleware.ErrBadRequest,
			"Session ID is required",
			fiber.StatusBadRequest,
		)
	}

	sessionID, err := uuid.Parse(sessionIDParam)
	if err != nil {
		return middleware.NewAppError(
			middleware.ErrBadRequest,
			"Invalid session ID format",
			fiber.StatusBadRequest,
		)
	}

	// First, verify the session belongs to the user
	ctx := c.Context()
	session, err := h.authService.GetSession(ctx, sessionID)
	if err != nil {
		log.Printf("[SessionHandler.RevokeSession] Error getting session: %v", err)
		return middleware.NewAppError(
			middleware.ErrNotFound,
			"Session not found",
			fiber.StatusNotFound,
		).WithCode("SESSION_NOT_FOUND")
	}

	// Verify ownership
	if session.UserID != userID {
		return middleware.NewAppError(
			middleware.ErrForbidden,
			"You don't have permission to revoke this session",
			fiber.StatusForbidden,
		).WithCode("SESSION_REVOKE_DENIED")
	}

	// Call service to revoke session
	if err := h.authService.RevokeSession(ctx, sessionID); err != nil {
		log.Printf("[SessionHandler.RevokeSession] Error: %v", err)
		return middleware.NewAppError(
			err,
			"Failed to revoke session",
			fiber.StatusInternalServerError,
		).WithCode("REVOKE_SESSION_FAILED")
	}

	// Return response
	return c.Status(fiber.StatusOK).JSON(dto.RevokeSessionResponse{
		Message: "Session revoked successfully",
	})
}

// RevokeAllSessions revokes all sessions except the current one.
// DELETE /api/v1/sessions
func (h *SessionHandler) RevokeAllSessions(c *fiber.Ctx) error {
	// Get authenticated user ID
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return middleware.NewAppError(
			middleware.ErrUnauthorized,
			"User not authenticated",
			fiber.StatusUnauthorized,
		)
	}

	// Get current session ID to keep it active
	currentSessionID, hasCurrentSession := middleware.GetSessionID(c)

	// Call service to revoke all sessions
	ctx := c.Context()
	var revokedCount int
	var err error

	if hasCurrentSession {
		revokedCount, err = h.authService.RevokeAllUserSessionsExcept(ctx, userID, currentSessionID)
	} else {
		// If no current session (shouldn't happen in authenticated context), revoke all
		revokedCount, err = h.authService.RevokeAllUserSessions(ctx, userID)
	}

	if err != nil {
		log.Printf("[SessionHandler.RevokeAllSessions] Error: %v", err)
		return middleware.NewAppError(
			err,
			"Failed to revoke sessions",
			fiber.StatusInternalServerError,
		).WithCode("REVOKE_ALL_SESSIONS_FAILED")
	}

	// Return response
	return c.Status(fiber.StatusOK).JSON(dto.RevokeAllSessionsResponse{
		RevokedCount: revokedCount,
		Message:      "All other sessions revoked successfully",
	})
}

// RevokeCurrentSession revokes the current session (logout).
// DELETE /api/v1/sessions/current
func (h *SessionHandler) RevokeCurrentSession(c *fiber.Ctx) error {
	// Extract session token from Authorization header
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return middleware.NewAppError(
			middleware.ErrUnauthorized,
			"Missing authorization header",
			fiber.StatusUnauthorized,
		)
	}

	// Extract token
	token := extractToken(authHeader)
	if token == "" {
		return middleware.NewAppError(
			middleware.ErrUnauthorized,
			"Invalid authorization header format",
			fiber.StatusUnauthorized,
		)
	}

	// Call service
	ctx := c.Context()
	if err := h.authService.Logout(ctx, token); err != nil {
		log.Printf("[SessionHandler.RevokeCurrentSession] Error: %v", err)
		return middleware.NewAppError(
			err,
			"Failed to revoke current session",
			fiber.StatusInternalServerError,
		).WithCode("REVOKE_CURRENT_SESSION_FAILED")
	}

	// Return response
	return c.Status(fiber.StatusOK).JSON(dto.RevokeSessionResponse{
		Message: "Current session revoked successfully (logged out)",
	})
}

// GetActiveSessionsCount returns the count of active sessions for the user.
// GET /api/v1/sessions/count
func (h *SessionHandler) GetActiveSessionsCount(c *fiber.Ctx) error {
	// Get authenticated user ID
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return middleware.NewAppError(
			middleware.ErrUnauthorized,
			"User not authenticated",
			fiber.StatusUnauthorized,
		)
	}

	// Call service
	ctx := c.Context()
	sessions, err := h.authService.GetUserSessions(ctx, userID)
	if err != nil {
		log.Printf("[SessionHandler.GetActiveSessionsCount] Error: %v", err)
		return middleware.NewAppError(
			err,
			"Failed to get sessions count",
			fiber.StatusInternalServerError,
		).WithCode("GET_SESSIONS_COUNT_FAILED")
	}

	// Count active sessions
	activeCount := 0
	for _, session := range sessions {
		if session.ToMetadata().IsActive {
			activeCount++
		}
	}

	// Return response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"active_sessions": activeCount,
		"total_sessions":  len(sessions),
	})
}
