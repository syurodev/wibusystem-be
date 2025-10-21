// Package http contains HTTP handlers for the Identity module.
package http

import (
	"log"
	"strings"
	"time"
	"wibusystem/internal/modules/identity/domain"

	"wibusystem/internal/modules/identity/dto"
	"wibusystem/internal/modules/identity/handler/middleware"
	"wibusystem/internal/modules/identity/service"

	"github.com/gofiber/fiber/v2"
)

// AuthHandler handles authentication-related HTTP requests.
type AuthHandler struct {
	authService service.AuthService
}

// NewAuthHandler creates a new AuthHandler instance.
func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Register handles user registration requests.
// POST /api/v1/auth/register
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req dto.RegisterRequest

	// Validate and parse request
	if err := middleware.ValidateRequest(c, &req); err != nil {
		return err
	}

	// Call service
	ctx := c.Context()
	user, err := h.authService.Register(ctx, req)
	if err != nil {
		log.Printf("[AuthHandler.Register] Error: %v", err)
		return middleware.NewAppError(
			err,
			"Registration failed",
			fiber.StatusBadRequest,
		).WithCode("REGISTRATION_FAILED")
	}

	// Convert domain user to response
	userResponse := mapUserToResponse(user)

	// Return response
	return c.Status(fiber.StatusCreated).JSON(dto.RegisterResponse{
		User:    userResponse,
		Message: "Registration successful. Please check your email to verify your account.",
	})
}

// Login handles user login requests.
// POST /api/v1/auth/login
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req dto.LoginRequest

	// Validate and parse request
	if err := middleware.ValidateRequest(c, &req); err != nil {
		return err
	}

	// Call service
	ctx := c.Context()
	user, sessionToken, err := h.authService.Login(ctx, req)
	if err != nil {
		log.Printf("[AuthHandler.Login] Error: %v", err)
		return middleware.NewAppError(
			err,
			"Login failed. Please check your credentials.",
			fiber.StatusUnauthorized,
		).WithCode("LOGIN_FAILED")
	}

	// Convert domain user to response
	userResponse := mapUserToResponse(user)

	// Calculate expiration time (default 7 days, or 30 days if remember me)
	expirationDuration := 7 * 24 * time.Hour
	if req.RememberMe {
		expirationDuration = 30 * 24 * time.Hour
	}
	expiresAt := time.Now().Add(expirationDuration)

	// Return response
	return c.Status(fiber.StatusOK).JSON(dto.LoginResponse{
		User:         userResponse,
		SessionToken: sessionToken,
		ExpiresAt:    expiresAt,
		Message:      "Login successful",
	})
}

// Logout handles user logout requests.
// POST /api/v1/auth/logout
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	// Extract session token from Authorization header
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return middleware.NewAppError(
			middleware.ErrUnauthorized,
			"Missing authorization header",
			fiber.StatusUnauthorized,
		)
	}

	// Extract token (support both "Bearer <token>" and "<token>")
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
		log.Printf("[AuthHandler.Logout] Error: %v", err)
		return middleware.NewAppError(
			err,
			"Logout failed",
			fiber.StatusInternalServerError,
		).WithCode("LOGOUT_FAILED")
	}

	// Return response
	return c.Status(fiber.StatusOK).JSON(dto.LogoutResponse{
		Message: "Logout successful",
	})
}

// RefreshSession handles session refresh requests.
// POST /api/v1/auth/refresh
func (h *AuthHandler) RefreshSession(c *fiber.Ctx) error {
	var req dto.RefreshSessionRequest

	// Validate and parse request
	if err := middleware.ValidateRequest(c, &req); err != nil {
		return err
	}

	// Call service
	ctx := c.Context()
	newToken, err := h.authService.RefreshSession(ctx, req.SessionToken)
	if err != nil {
		log.Printf("[AuthHandler.RefreshSession] Error: %v", err)
		return middleware.NewAppError(
			err,
			"Session refresh failed",
			fiber.StatusUnauthorized,
		).WithCode("SESSION_REFRESH_FAILED")
	}

	// Calculate new expiration time (7 days from now)
	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	// Return response
	return c.Status(fiber.StatusOK).JSON(dto.RefreshSessionResponse{
		SessionToken: newToken,
		ExpiresAt:    expiresAt,
		Message:      "Session refreshed successfully",
	})
}

// VerifyEmail handles email verification requests.
// POST /api/v1/auth/verify-email
func (h *AuthHandler) VerifyEmail(c *fiber.Ctx) error {
	var req dto.VerifyEmailRequest

	// Validate and parse request
	if err := middleware.ValidateRequest(c, &req); err != nil {
		return err
	}

	// Call service
	ctx := c.Context()
	if err := h.authService.VerifyEmail(ctx, req.Token); err != nil {
		log.Printf("[AuthHandler.VerifyEmail] Error: %v", err)
		return middleware.NewAppError(
			err,
			"Email verification failed",
			fiber.StatusBadRequest,
		).WithCode("EMAIL_VERIFICATION_FAILED")
	}

	// Return response
	return c.Status(fiber.StatusOK).JSON(dto.VerifyEmailResponse{
		Message: "Email verified successfully",
	})
}

// ResendVerificationEmail handles requests to resend verification email.
// POST /api/v1/auth/resend-verification
func (h *AuthHandler) ResendVerificationEmail(c *fiber.Ctx) error {
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
	if err := h.authService.SendVerificationEmail(ctx, userID); err != nil {
		log.Printf("[AuthHandler.ResendVerificationEmail] Error: %v", err)
		return middleware.NewAppError(
			err,
			"Failed to send verification email",
			fiber.StatusInternalServerError,
		).WithCode("SEND_VERIFICATION_FAILED")
	}

	// Return response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Verification email sent successfully",
	})
}

// ForgotPassword handles password reset requests.
// POST /api/v1/auth/forgot-password
func (h *AuthHandler) ForgotPassword(c *fiber.Ctx) error {
	var req dto.ForgotPasswordRequest

	// Validate and parse request
	if err := middleware.ValidateRequest(c, &req); err != nil {
		return err
	}

	// Call service
	ctx := c.Context()
	if err := h.authService.RequestPasswordReset(ctx, req.Email); err != nil {
		log.Printf("[AuthHandler.ForgotPassword] Error: %v", err)
		// Don't reveal if email exists or not for security reasons
		// Always return success
	}

	// Return generic success response (don't reveal if email exists)
	return c.Status(fiber.StatusOK).JSON(dto.ForgotPasswordResponse{
		Message: "If the email exists, a password reset link has been sent",
	})
}

// ResetPassword handles password reset confirmation.
// POST /api/v1/auth/reset-password
func (h *AuthHandler) ResetPassword(c *fiber.Ctx) error {
	var req dto.ResetPasswordRequest

	// Validate and parse request
	if err := middleware.ValidateRequest(c, &req); err != nil {
		return err
	}

	// Call service
	ctx := c.Context()
	if err := h.authService.ResetPassword(ctx, req.Token, req.NewPassword); err != nil {
		log.Printf("[AuthHandler.ResetPassword] Error: %v", err)
		return middleware.NewAppError(
			err,
			"Password reset failed",
			fiber.StatusBadRequest,
		).WithCode("PASSWORD_RESET_FAILED")
	}

	// Return response
	return c.Status(fiber.StatusOK).JSON(dto.ResetPasswordResponse{
		Message: "Password reset successful. You can now login with your new password.",
	})
}

// ChangePassword handles password change requests (authenticated).
// POST /api/v1/auth/change-password
func (h *AuthHandler) ChangePassword(c *fiber.Ctx) error {
	// Get authenticated user ID
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return middleware.NewAppError(
			middleware.ErrUnauthorized,
			"User not authenticated",
			fiber.StatusUnauthorized,
		)
	}

	var req dto.ChangePasswordRequest

	// Validate and parse request
	if err := middleware.ValidateRequest(c, &req); err != nil {
		return err
	}

	// Call service
	ctx := c.Context()
	if err := h.authService.ChangePassword(ctx, userID, req.CurrentPassword, req.NewPassword); err != nil {
		log.Printf("[AuthHandler.ChangePassword] Error: %v", err)
		return middleware.NewAppError(
			err,
			"Password change failed",
			fiber.StatusBadRequest,
		).WithCode("PASSWORD_CHANGE_FAILED")
	}

	// Return response
	return c.Status(fiber.StatusOK).JSON(dto.ChangePasswordResponse{
		Message: "Password changed successfully",
	})
}

// ValidateSession validates the current session.
// GET /api/v1/auth/validate
func (h *AuthHandler) ValidateSession(c *fiber.Ctx) error {
	// Extract session token from Authorization header
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return middleware.NewAppError(
			middleware.ErrUnauthorized,
			"Missing authorization header",
			fiber.StatusUnauthorized,
		)
	}

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
	session, err := h.authService.ValidateSession(ctx, token)
	if err != nil {
		return middleware.NewAppError(
			middleware.ErrUnauthorized,
			"Invalid or expired session",
			fiber.StatusUnauthorized,
		).WithCode("INVALID_SESSION")
	}

	// Return session info
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"valid":      true,
		"user_id":    session.UserID,
		"session_id": session.ID,
		"expires_at": session.ExpiresAt,
	})
}

// GetMe returns the current authenticated user's profile.
// GET /api/v1/auth/me
func (h *AuthHandler) GetMe(c *fiber.Ctx) error {
	// Get authenticated user ID
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return middleware.NewAppError(
			middleware.ErrUnauthorized,
			"User not authenticated",
			fiber.StatusUnauthorized,
		)
	}

	// Note: This would typically call UserService.GetProfile
	// For now, return just the user ID
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"user_id": userID,
		"message": "Authenticated",
	})
}

// Helper functions

// extractToken extracts the token from the Authorization header
func extractToken(authHeader string) string {
	// Split by space
	parts := strings.Split(authHeader, " ")
	if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
		return parts[1]
	}
	// If no "Bearer" prefix, assume the entire header is the token
	return authHeader
}

// mapUserToResponse converts a domain User to a UserResponse DTO
func mapUserToResponse(user *domain.User) dto.UserResponse {
	return dto.UserResponse{
		ID:            user.ID,
		Email:         user.Email,
		EmailVerified: user.EmailVerified,
		DisplayName:   user.DisplayName,
		AvatarURL:     user.AvatarURL,
		Status:        string(user.Status),
		LastLoginAt:   user.LastLoginAt,
		CreatedAt:     user.CreatedAt,
		UpdatedAt:     user.UpdatedAt,
	}
}
