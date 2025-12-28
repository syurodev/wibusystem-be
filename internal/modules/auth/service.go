package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/gofrs/uuid/v5"
	"go.uber.org/zap"

	"system/internal/domain"
	ent "system/internal/ent/generated"
	"system/internal/modules/user"
	pkgerrors "system/pkg/errors"
	"system/pkg/util/crypto"
	"system/pkg/util/stringutil"
)

// AuthService xử lý authentication và user management logic.
type authServiceImpl struct {
	userService       user.UserService
	verificationRepo  domain.EmailVerificationRepository
	passwordResetRepo domain.PasswordResetRepository
	roleRepo          domain.RoleRepository
}

// NewAuthService tạo instance mới của AuthService.
func NewService(
	userService user.UserService,
	verificationRepo domain.EmailVerificationRepository,
	passwordResetRepo domain.PasswordResetRepository,
	roleRepo domain.RoleRepository,
) AuthService {
	return &authServiceImpl{
		userService:       userService,
		verificationRepo:  verificationRepo,
		passwordResetRepo: passwordResetRepo,
		roleRepo:          roleRepo,
	}
}

// RegisterUser tạo tài khoản mới với email chưa verified.
func (s *authServiceImpl) RegisterUser(ctx context.Context, email, password, fullName string) (*domain.User, string, error) {
	// Kiểm tra email đã tồn tại chưa
	existingUser, err := s.userService.GetUserByEmail(ctx, email)
	if err != nil && !ent.IsNotFound(err) {
		return nil, "", fmt.Errorf("failed to check existing user: %w", err)
	}
	if existingUser != nil {
		return nil, "", pkgerrors.Conflict(I18nEmailAlreadyExists, "email already exists")
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
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.userService.CreateUser(ctx, user); err != nil {
		return nil, "", fmt.Errorf("failed to create user: %w", err)
	}

	// Gán role USER mặc định cho user mới
	if err := s.assignDefaultRole(ctx, user.ID); err != nil {
		// Log error nhưng không fail registration
		// User vẫn có thể được gán role sau
		zap.L().Error("failed to assign default role to new user",
			zap.String("user_id", user.ID.String()),
			zap.Error(err),
		)
	}

	// Tạo verification token
	verificationToken, err := s.CreateEmailVerificationToken(ctx, user.ID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create verification token: %w", err)
	}

	return user, verificationToken, nil
}

// assignDefaultRole gán role mặc định (USER) cho user mới
func (s *authServiceImpl) assignDefaultRole(ctx context.Context, userID uuid.UUID) error {
	defaultRole := domain.GetDefaultRole()

	// Lấy role ID từ database
	roleID, err := s.roleRepo.GetRoleIDByName(ctx, defaultRole)
	if err != nil {
		return fmt.Errorf("failed to get default role ID: %w", err)
	}

	// Gán role cho user
	if err := s.roleRepo.AssignGlobalRole(ctx, userID, roleID); err != nil {
		return fmt.Errorf("failed to assign default role: %w", err)
	}

	zap.L().Info("assigned default role to new user",
		zap.String("user_id", userID.String()),
		zap.String("role", defaultRole.String()),
	)

	return nil
}

// CreateEmailVerificationToken tạo token để verify email.
func (s *authServiceImpl) CreateEmailVerificationToken(ctx context.Context, userID uuid.UUID) (string, error) {
	// Xóa các tokens cũ của user này
	if err := s.verificationRepo.DeleteByUserID(ctx, userID); err != nil {
		return "", fmt.Errorf("failed to delete old tokens: %w", err)
	}

	// Generate random token (32 bytes = 64 hex characters)
	tokenStr, err := stringutil.GenerateRandomString(32)
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
func (s *authServiceImpl) VerifyEmail(ctx context.Context, tokenStr string) error {
	// Lấy token
	token, err := s.verificationRepo.GetByToken(ctx, tokenStr)
	if err != nil {
		if ent.IsNotFound(err) {
			return pkgerrors.BadRequest(I18nInvalidToken, "invalid token")
		}
		return fmt.Errorf("failed to get token: %w", err)
	}

	// Kiểm tra đã sử dụng chưa
	if token.UsedAt != nil {
		return pkgerrors.BadRequest(I18nTokenAlreadyUsed, "token already used")
	}

	// Kiểm tra hết hạn chưa
	if time.Now().After(token.ExpiresAt) {
		return pkgerrors.BadRequest(I18nTokenExpired, "token expired")
	}

	// Cập nhật user email_verified = true
	user, err := s.userService.GetUserByID(ctx, token.UserID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	user.EmailVerified = true
	user.UpdatedAt = time.Now()

	if err := s.userService.UpdateUser(ctx, user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	// Đánh dấu token đã sử dụng
	if err := s.verificationRepo.MarkAsUsed(ctx, token.ID); err != nil {
		return fmt.Errorf("failed to mark token as used: %w", err)
	}

	return nil
}

// CreatePasswordResetToken tạo token để reset password.
func (s *authServiceImpl) CreatePasswordResetToken(ctx context.Context, email string) (string, error) {
	// Tìm user theo email
	user, err := s.userService.GetUserByEmail(ctx, email)
	if err != nil {
		if ent.IsNotFound(err) {
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
	tokenStr, err := stringutil.GenerateRandomString(32)
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
func (s *authServiceImpl) ResetPassword(ctx context.Context, tokenStr, newPassword string) error {
	// Lấy token
	token, err := s.passwordResetRepo.GetByToken(ctx, tokenStr)
	if err != nil {
		if ent.IsNotFound(err) {
			return pkgerrors.BadRequest(I18nInvalidToken, "invalid token")
		}
		return fmt.Errorf("failed to get token: %w", err)
	}

	// Kiểm tra đã sử dụng chưa
	if token.UsedAt != nil {
		return pkgerrors.BadRequest(I18nTokenAlreadyUsed, "token already used")
	}

	// Kiểm tra hết hạn chưa
	if time.Now().After(token.ExpiresAt) {
		return pkgerrors.BadRequest(I18nTokenExpired, "token expired")
	}

	// Hash password mới
	passwordHash, err := crypto.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Cập nhật password
	user, err := s.userService.GetUserByID(ctx, token.UserID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	user.PasswordHash = passwordHash
	user.UpdatedAt = time.Now()

	if err := s.userService.UpdateUser(ctx, user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	// Đánh dấu token đã sử dụng
	if err := s.passwordResetRepo.MarkAsUsed(ctx, token.ID); err != nil {
		return fmt.Errorf("failed to mark token as used: %w", err)
	}

	return nil
}

// GetUserByEmail lấy user theo email address
func (s *authServiceImpl) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	return s.userService.GetUserByEmail(ctx, email)
}

// GetUserByUsername lấy user theo username
func (s *authServiceImpl) GetUserByUsername(ctx context.Context, username string) (*domain.User, error) {
	return s.userService.GetUserByUsername(ctx, username)
}
