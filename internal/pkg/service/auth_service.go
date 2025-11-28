package service

import (
	"context"
	"fmt"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"

	"system/internal/domain"
	pkgerrors "system/pkg/errors"
	"system/pkg/util/crypto"
	"system/pkg/util/random"
)

// AuthService xử lý authentication và user management logic.
type AuthService struct {
	userRepo          domain.UserRepository
	verificationRepo  domain.EmailVerificationRepository
	passwordResetRepo domain.PasswordResetRepository
}

// NewAuthService tạo instance mới của AuthService.
func NewAuthService(
	userRepo domain.UserRepository,
	verificationRepo domain.EmailVerificationRepository,
	passwordResetRepo domain.PasswordResetRepository,
) *AuthService {
	return &AuthService{
		userRepo:          userRepo,
		verificationRepo:  verificationRepo,
		passwordResetRepo: passwordResetRepo,
	}
}

// RegisterUser tạo tài khoản mới với email chưa verified.
func (s *AuthService) RegisterUser(ctx context.Context, email, password, fullName string) (*domain.User, string, error) {
	// Kiểm tra email đã tồn tại chưa
	existingUser, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil && err != pgx.ErrNoRows {
		return nil, "", fmt.Errorf("failed to check existing user: %w", err)
	}
	if existingUser != nil {
		return nil, "", pkgerrors.ErrEmailAlreadyExists
	}

	// Hash password
	passwordHash, err := crypto.HashPassword(password)
	if err != nil {
		return nil, "", fmt.Errorf("failed to hash password: %w", err)
	}

	// Tạo user mới
	user := &domain.User{
		ID:            uuid.Must(uuid.NewV7()),
		Email:         email,
		EmailVerified: false, // Chưa verified
		PasswordHash:  passwordHash,
		FullName:      &fullName,
		Status:        "active",
		Settings: map[string]any{
			"theme":                 "system",
			"language":              "en",
			"notifications_enabled": true,
			"content_filters":       []string{},
			"ui_preferences": map[string]any{
				"reduce_blur":         false,
				"auto_play_video":     false,
				"show_mature_content": false,
				"compact_view":        false,
			},
			"created_at": time.Now().Format(time.RFC3339),
			"updated_at": time.Now().Format(time.RFC3339),
		},
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, "", fmt.Errorf("failed to create user: %w", err)
	}

	// Tạo verification token
	verificationToken, err := s.CreateEmailVerificationToken(ctx, user.ID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create verification token: %w", err)
	}

	return user, verificationToken, nil
}

// CreateEmailVerificationToken tạo token để verify email.
func (s *AuthService) CreateEmailVerificationToken(ctx context.Context, userID uuid.UUID) (string, error) {
	// Xóa các tokens cũ của user này
	if err := s.verificationRepo.DeleteByUserID(ctx, userID); err != nil {
		return "", fmt.Errorf("failed to delete old tokens: %w", err)
	}

	// Generate random token (32 bytes = 64 hex characters)
	tokenStr, err := random.GenerateRandomString(32)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	// Tạo token entity
	token := &domain.EmailVerificationToken{
		ID:        uuid.Must(uuid.NewV7()),
		UserID:    userID,
		Token:     tokenStr,
		ExpiresAt: time.Now().Add(24 * time.Hour), // Hết hạn sau 24 giờ
		CreatedAt: time.Now(),
	}

	if err := s.verificationRepo.Create(ctx, token); err != nil {
		return "", fmt.Errorf("failed to save token: %w", err)
	}

	return tokenStr, nil
}

// VerifyEmail xác thực email bằng token.
func (s *AuthService) VerifyEmail(ctx context.Context, tokenStr string) error {
	// Lấy token
	token, err := s.verificationRepo.GetByToken(ctx, tokenStr)
	if err != nil {
		if err == pgx.ErrNoRows {
			return pkgerrors.ErrInvalidToken
		}
		return fmt.Errorf("failed to get token: %w", err)
	}

	// Kiểm tra đã sử dụng chưa
	if token.UsedAt != nil {
		return pkgerrors.ErrTokenAlreadyUsed
	}

	// Kiểm tra hết hạn chưa
	if time.Now().After(token.ExpiresAt) {
		return pkgerrors.ErrTokenExpired
	}

	// Cập nhật user email_verified = true
	user, err := s.userRepo.GetByID(ctx, token.UserID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	user.EmailVerified = true
	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	// Đánh dấu token đã sử dụng
	if err := s.verificationRepo.MarkAsUsed(ctx, token.ID); err != nil {
		return fmt.Errorf("failed to mark token as used: %w", err)
	}

	return nil
}

// CreatePasswordResetToken tạo token để reset password.
func (s *AuthService) CreatePasswordResetToken(ctx context.Context, email string) (string, error) {
	// Tìm user theo email
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if err == pgx.ErrNoRows {
			// Không tiết lộ email không tồn tại (security best practice)
			return "", nil
		}
		return "", fmt.Errorf("failed to get user: %w", err)
	}

	// Xóa các tokens cũ
	if err := s.passwordResetRepo.DeleteByUserID(ctx, user.ID); err != nil {
		return "", fmt.Errorf("failed to delete old tokens: %w", err)
	}

	// Generate random token
	tokenStr, err := random.GenerateRandomString(32)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	// Tạo token entity
	token := &domain.PasswordResetToken{
		ID:        uuid.Must(uuid.NewV7()),
		UserID:    user.ID,
		Token:     tokenStr,
		ExpiresAt: time.Now().Add(1 * time.Hour), // Hết hạn sau 1 giờ
		CreatedAt: time.Now(),
	}

	if err := s.passwordResetRepo.Create(ctx, token); err != nil {
		return "", fmt.Errorf("failed to save token: %w", err)
	}

	return tokenStr, nil
}

// ResetPassword reset password bằng token.
func (s *AuthService) ResetPassword(ctx context.Context, tokenStr, newPassword string) error {
	// Lấy token
	token, err := s.passwordResetRepo.GetByToken(ctx, tokenStr)
	if err != nil {
		if err == pgx.ErrNoRows {
			return pkgerrors.ErrInvalidToken
		}
		return fmt.Errorf("failed to get token: %w", err)
	}

	// Kiểm tra đã sử dụng chưa
	if token.UsedAt != nil {
		return pkgerrors.ErrTokenAlreadyUsed
	}

	// Kiểm tra hết hạn chưa
	if time.Now().After(token.ExpiresAt) {
		return pkgerrors.ErrTokenExpired
	}

	// Hash password mới
	passwordHash, err := crypto.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Cập nhật password
	user, err := s.userRepo.GetByID(ctx, token.UserID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	user.PasswordHash = passwordHash
	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	// Đánh dấu token đã sử dụng
	if err := s.passwordResetRepo.MarkAsUsed(ctx, token.ID); err != nil {
		return fmt.Errorf("failed to mark token as used: %w", err)
	}

	return nil
}

// GetUserByEmail lấy user theo email address
func (s *AuthService) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	return s.userRepo.GetByEmail(ctx, email)
}
