package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid/v5"
	"go.uber.org/zap"

	authdto "system/internal/dto/auth"
	"system/internal/modules/email"
	pkgerrors "system/pkg/errors"
	"system/pkg/util/response"
	"system/pkg/util/validator"
)

// OAuth2SessionCreator interface for creating sessions (subset of OAuth2Service)
type OAuth2SessionCreator interface {
	CreateUserSession(ctx context.Context, userID uuid.UUID, ttl time.Duration, userAgent, ip string) (string, error)
}

type Handler struct {
	authService      AuthService
	emailService     email.EmailService
	webauthnService  WebAuthnService
	oauth2Service    OAuth2SessionCreator
	registerUserUC   RegisterUserUseCase
	verifyEmailUC    VerifyEmailUseCase
	forgotPasswordUC ForgotPasswordUseCase
	resetPasswordUC  ResetPasswordUseCase
}

func NewHandler(
	authService AuthService,
	emailService email.EmailService,
	webauthnService WebAuthnService,
	oauth2Service OAuth2SessionCreator,
	registerUserUC RegisterUserUseCase,
	verifyEmailUC VerifyEmailUseCase,
	forgotPasswordUC ForgotPasswordUseCase,
	resetPasswordUC ResetPasswordUseCase,
) *Handler {
	return &Handler{
		authService:      authService,
		emailService:     emailService,
		webauthnService:  webauthnService,
		oauth2Service:    oauth2Service,
		registerUserUC:   registerUserUC,
		verifyEmailUC:    verifyEmailUC,
		forgotPasswordUC: forgotPasswordUC,
		resetPasswordUC:  resetPasswordUC,
	}
}

// RegisterPage hiển thị trang đăng ký
func (h *Handler) RegisterPage(c *gin.Context) {
	requestID := c.Query("request_id")
	c.HTML(http.StatusOK, "auth/register.html", gin.H{
		"RequestID": requestID,
	})
}

// Register xử lý đăng ký user mới
func (h *Handler) Register(c *gin.Context) {
	var req authdto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_FAILED", "validation.failed", err.Error())
		return
	}

	// Validate password strength (basic)
	if err := validator.ValidatePasswordStrength(req.Password); err != nil {
		response.Error(c, http.StatusBadRequest, "WEAK_PASSWORD", "auth.weak_password", nil)
		return
	}

	// Register user via UseCase
	result, err := h.registerUserUC.Execute(c.Request.Context(), RegisterUserInput{
		Email:    req.Email,
		Password: req.Password,
		FullName: req.FullName,
	})
	if err != nil {
		if errors.Is(err, pkgerrors.ErrEmailAlreadyExists) {
			response.Error(c, http.StatusConflict, "EMAIL_EXISTS", "auth.email_already_exists", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "REGISTRATION_FAILED", "auth.registration_failed", nil)
		zap.L().Error("Registration failed", zap.Error(err))
		return
	}
	
	user := result.User
	verificationToken := result.VerificationToken

	// Send verification email
	userName := req.FullName
	if userName == "" {
		userName = req.Email
	}

	if err := h.emailService.SendVerificationEmail(c.Request.Context(), user.Email, userName, verificationToken); err != nil {
		// Log error but don't fail registration
		// User can request new verification email later
		c.Error(err) // This will be logged by Gin middleware
	}

	// Auto-login: Create session for the new user
	sessionID, err := h.oauth2Service.CreateUserSession(
		c.Request.Context(),
		user.ID,
		7*24*time.Hour,
		c.Request.UserAgent(),
		c.ClientIP(),
	)
	if err != nil {
		zap.L().Error("Failed to create session after registration", zap.Error(err))
		// Still return success but without auto-login
		resp := authdto.RegisterResponse{
			UserID:  user.ID.String(),
			Email:   user.Email,
			Message: "Registration successful. Please check your email to verify your account.",
		}
		response.Success(c, http.StatusCreated, "auth.registration_success", resp, nil)
		return
	}

	// Set session cookie
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("session_id", sessionID, 86400*7, "/", "", false, true)

	resp := authdto.RegisterResponse{
		UserID:  user.ID.String(),
		Email:   user.Email,
		Message: "Registration successful. You are now logged in.",
	}

	response.Success(c, http.StatusCreated, "auth.registration_success", resp, nil)
}

// VerifyEmailPage hiển thị trang verify email
func (h *Handler) VerifyEmailPage(c *gin.Context) {
	token := c.Query("token")
	c.HTML(http.StatusOK, "auth/verify_email.html", gin.H{
		"Token": token,
	})
}

// VerifyEmail xác thực email bằng token
func (h *Handler) VerifyEmail(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		var req authdto.VerifyEmailRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, http.StatusBadRequest, "VALIDATION_FAILED", "validation.failed", err.Error())
			return
		}
		token = req.Token
	}

	if err := h.verifyEmailUC.Execute(c.Request.Context(), VerifyEmailInput{Token: token}); err != nil {
		if errors.Is(err, pkgerrors.ErrInvalidToken) {
			response.Error(c, http.StatusBadRequest, "INVALID_TOKEN", "auth.invalid_token", nil)
			return
		}
		if errors.Is(err, pkgerrors.ErrTokenExpired) {
			response.Error(c, http.StatusBadRequest, "TOKEN_EXPIRED", "auth.token_expired", nil)
			return
		}
		if errors.Is(err, pkgerrors.ErrTokenAlreadyUsed) {
			response.Error(c, http.StatusBadRequest, "TOKEN_USED", "auth.token_already_used", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "VERIFICATION_FAILED", "auth.verification_failed", nil)
		return
	}

	response.Success(c, http.StatusOK, "auth.email_verified", nil, nil)
}

// ForgotPasswordPage hiển thị trang quên mật khẩu
func (h *Handler) ForgotPasswordPage(c *gin.Context) {
	c.HTML(http.StatusOK, "auth/forgot_password.html", gin.H{})
}

// ForgotPassword tạo token reset password và gửi email
func (h *Handler) ForgotPassword(c *gin.Context) {
	var req authdto.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_FAILED", "validation.failed", err.Error())
		return
	}

	resetToken, err := h.forgotPasswordUC.Execute(c.Request.Context(), ForgotPasswordInput{Email: req.Email})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "RESET_FAILED", "auth.reset_failed", nil)
		return
	}

	// Send password reset email (only if token was created, i.e., user exists)
	if resetToken != "" {
		if err := h.emailService.SendPasswordResetEmail(c.Request.Context(), req.Email, req.Email, resetToken); err != nil {
			// Log error but don't fail request (security: don't reveal if email exists)
			c.Error(err)
		}
	}

	// Always return success để không tiết lộ email có tồn tại hay không
	response.Success(c, http.StatusOK, "auth.reset_email_sent", nil, nil)
}

// ResetPasswordPage hiển thị trang reset password
func (h *Handler) ResetPasswordPage(c *gin.Context) {
	token := c.Query("token")
	c.HTML(http.StatusOK, "auth/reset_password.html", gin.H{
		"Token": token,
	})
}

// ResetPassword reset password bằng token
func (h *Handler) ResetPassword(c *gin.Context) {
	var req authdto.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_FAILED", "validation.failed", err.Error())
		return
	}

	// Validate password strength
	if err := validator.ValidatePasswordStrength(req.NewPassword); err != nil {
		response.Error(c, http.StatusBadRequest, "WEAK_PASSWORD", "auth.weak_password", nil)
		return
	}

	if err := h.resetPasswordUC.Execute(c.Request.Context(), ResetPasswordInput{
		Token:       req.Token,
		NewPassword: req.NewPassword,
	}); err != nil {
		if errors.Is(err, pkgerrors.ErrInvalidToken) {
			response.Error(c, http.StatusBadRequest, "INVALID_TOKEN", "auth.invalid_token", nil)
			return
		}
		if errors.Is(err, pkgerrors.ErrTokenExpired) {
			response.Error(c, http.StatusBadRequest, "TOKEN_EXPIRED", "auth.token_expired", nil)
			return
		}
		if errors.Is(err, pkgerrors.ErrTokenAlreadyUsed) {
			response.Error(c, http.StatusBadRequest, "TOKEN_USED", "auth.token_already_used", nil)
			return
		}
		response.Error(c, http.StatusInternalServerError, "RESET_FAILED", "auth.reset_failed", nil)
		return
	}

	response.Success(c, http.StatusOK, "auth.password_reset", nil, nil)
}

// PasskeySetupPage hiển thị trang đề xuất tạo passkey
func (h *Handler) PasskeySetupPage(c *gin.Context) {
	requestID := c.Query("request_id")
	c.HTML(http.StatusOK, "auth/passkey_setup.html", gin.H{
		"RequestID": requestID,
	})
}

// PasskeyRegisterPage hiển thị trang đăng ký passkey
func (h *Handler) PasskeyRegisterPage(c *gin.Context) {
	requestID := c.Query("request_id")
	c.HTML(http.StatusOK, "auth/passkey_register_htmx.html", gin.H{
		"RequestID": requestID,
	})
}

// PasskeyManagePage hiển thị trang quản lý passkey
func (h *Handler) PasskeyManagePage(c *gin.Context) {
	c.HTML(http.StatusOK, "auth/passkey_manage_htmx.html", gin.H{})
}

// Helper functions

// normalizeEmail normalizes email address
func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
