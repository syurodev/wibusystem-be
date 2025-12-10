package user

import (
	"context"
	"time"

	"github.com/gofrs/uuid/v5"

	"system/internal/domain"
	userdto "system/internal/dto/user"
	pkgerrors "system/pkg/errors"
	"system/pkg/util/crypto"
	"system/pkg/util/validator"
)

// UserService cung cấp business logic cho user operations
type userServiceImpl struct {
	userRepo    domain.UserRepository
	sessionRepo domain.SessionRepository
}

// NewUserService tạo một instance mới của UserService
func NewService(userRepo domain.UserRepository, sessionRepo domain.SessionRepository) *userServiceImpl {
	return &userServiceImpl{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
	}
}

// GetUserByID lấy thông tin user theo ID
func (s *userServiceImpl) GetUserByID(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}

// GetSettings lấy settings của user
func (s *userServiceImpl) GetSettings(ctx context.Context, userID uuid.UUID) (map[string]any, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	settings := user.Settings
	if settings == nil {
		settings = make(map[string]any)
	}
	return settings, nil
}

// UpdateSettings cập nhật settings của user
func (s *userServiceImpl) UpdateSettings(ctx context.Context, userID uuid.UUID, newSettings map[string]any) (map[string]any, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Merge settings
	if user.Settings == nil {
		user.Settings = make(map[string]any)
	}
	for k, v := range newSettings {
		user.Settings[k] = v
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user.Settings, nil
}

// GetProfile lấy profile của user
func (s *userServiceImpl) GetProfile(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}

// UpdateProfile cập nhật profile của user
func (s *userServiceImpl) UpdateProfile(ctx context.Context, userID uuid.UUID, req userdto.UpdateProfileRequest) (*domain.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
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
		if bioArray, ok := req.Bio.([]any); ok {
			user.Bio = bioArray
		}
	}
	if req.Username != nil {
		newUsername := *req.Username

		// If user already has a username, prevent changing it
		if user.Username != nil && *user.Username != "" {
			if *user.Username != newUsername {
				return nil, pkgerrors.BadRequest(I18nUsernameImmutable, "username cannot be changed")
			}
		} else {
			// Basic validation
			if len(newUsername) < 3 || len(newUsername) > 30 {
				return nil, pkgerrors.BadRequest(I18nUsernameLengthInvalid, "username must be between 3 and 30 characters")
			}

			// Check uniqueness
			existingUser, err := s.userRepo.GetByUsername(ctx, newUsername)
			if err == nil && existingUser != nil {
				return nil, pkgerrors.Conflict(I18nUsernameTaken, "username is already taken")
			}
			user.Username = req.Username
		}
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// ChangePassword thay đổi password của user
func (s *userServiceImpl) ChangePassword(ctx context.Context, userID uuid.UUID, currentPassword, newPassword string) error {
	// Validate new password strength
	if err := validator.ValidatePasswordStrength(newPassword); err != nil {
		return pkgerrors.BadRequest(I18nWeakPassword, "password is too weak")
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	// Verify current password
	if !crypto.VerifyPassword(user.PasswordHash, currentPassword) {
		return pkgerrors.BadRequest(I18nInvalidPassword, "current password is incorrect")
	}

	// Hash new password
	hashedPassword, err := crypto.HashPassword(newPassword)
	if err != nil {
		return err
	}

	user.PasswordHash = hashedPassword

	return s.userRepo.Update(ctx, user)
}

// GetSessions lấy danh sách sessions của user
func (s *userServiceImpl) GetSessions(ctx context.Context, userIDStr string) ([]*domain.UserSession, error) {
	return s.sessionRepo.GetUserSessions(ctx, userIDStr)
}

// DeleteSession xoá một session của user
func (s *userServiceImpl) DeleteSession(ctx context.Context, userIDStr, sessionID string) error {
	// Verify ownership
	session, err := s.sessionRepo.GetSession(ctx, sessionID)
	if err != nil {
		// Session not found, consider it already deleted
		return nil
	}

	if session.UserID != userIDStr {
		return pkgerrors.Forbidden(I18nForbidden, "not allowed to delete this session")
	}

	return s.sessionRepo.DeleteSession(ctx, sessionID)
}

// GetUserByEmail lấy thông tin user theo email
func (s *userServiceImpl) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	return s.userRepo.GetByEmail(ctx, email)
}

// GetUserByUsername lấy thông tin user theo username
func (s *userServiceImpl) GetUserByUsername(ctx context.Context, username string) (*domain.User, error) {
	return s.userRepo.GetByUsername(ctx, username)
}

// CreateUser tạo user mới
func (s *userServiceImpl) CreateUser(ctx context.Context, user *domain.User) error {
	return s.userRepo.Create(ctx, user)
}

// UpdateUser cập nhật thông tin user (low-level update)
func (s *userServiceImpl) UpdateUser(ctx context.Context, user *domain.User) error {
	return s.userRepo.Update(ctx, user)
}
// CreateSession creates a new session
func (s *userServiceImpl) CreateSession(ctx context.Context, session *domain.UserSession, ttl time.Duration) error {
	return s.sessionRepo.CreateSession(ctx, session, ttl)
}

// GetSession retrieves a session by ID
func (s *userServiceImpl) GetSession(ctx context.Context, sessionID string) (*domain.UserSession, error) {
	return s.sessionRepo.GetSession(ctx, sessionID)
}
// UpdateLastLogin updates the last login timestamp
func (s *userServiceImpl) UpdateLastLogin(ctx context.Context, userID uuid.UUID) error {
	return s.userRepo.UpdateLastLogin(ctx, userID)
}

// GetGlobalPermissions gets all global permissions for a user
func (s *userServiceImpl) GetGlobalPermissions(ctx context.Context, userID uuid.UUID) ([]string, error) {
	return s.userRepo.GetGlobalPermissions(ctx, userID)
}

// GetOrganizationPermissions gets all permissions for a user in an organization
func (s *userServiceImpl) GetOrganizationPermissions(ctx context.Context, userID, organizationID uuid.UUID) ([]string, error) {
	return s.userRepo.GetOrganizationPermissions(ctx, userID, organizationID)
}

// GetGlobalRoles gets all global roles for a user
func (s *userServiceImpl) GetGlobalRoles(ctx context.Context, userID uuid.UUID) ([]string, error) {
	return s.userRepo.GetGlobalRoles(ctx, userID)
}

// GetOrganizationRoles gets all roles for a user in an organization
func (s *userServiceImpl) GetOrganizationRoles(ctx context.Context, userID, organizationID uuid.UUID) ([]string, error) {
	return s.userRepo.GetOrganizationRoles(ctx, userID, organizationID)
}
