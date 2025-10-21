// Package service contains business logic for the Identity module.
package service

import (
	"context"

	"github.com/google/uuid"

	"wibusystem/internal/modules/identity/domain"
	"wibusystem/internal/modules/identity/dto"
	"wibusystem/internal/modules/identity/repository"
)

// UserService defines the interface for user management operations.
type UserService interface {
	// GetProfile retrieves a user's profile by ID.
	GetProfile(ctx context.Context, userID uuid.UUID) (*domain.User, error)

	// UpdateProfile updates a user's profile information.
	UpdateProfile(ctx context.Context, userID uuid.UUID, req dto.UpdateProfileRequest) (*domain.User, error)

	// DeleteAccount soft-deletes a user account.
	DeleteAccount(ctx context.Context, userID uuid.UUID, password string) error

	// ListUsers lists all users (admin only).
	ListUsers(ctx context.Context, filter repository.UserListFilter) ([]*domain.User, int, error)

	// SearchUsers searches for users by email or display name.
	SearchUsers(ctx context.Context, query string, limit int) ([]*domain.User, error)

	// GetUserByID retrieves a user by ID (admin only).
	GetUserByID(ctx context.Context, userID uuid.UUID) (*domain.User, error)

	// GetUserByEmail retrieves a user by email (admin only).
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)

	// UpdateUserStatus updates a user's status (admin only).
	UpdateUserStatus(ctx context.Context, userID uuid.UUID, status domain.UserStatus) error

	// GetUserStats retrieves statistics about users (admin only).
	GetUserStats(ctx context.Context) (*UserStats, error)
}

// UserStats contains statistics about users.
type UserStats struct {
	TotalUsers      int
	ActiveUsers     int
	InactiveUsers   int
	SuspendedUsers  int
	PendingUsers    int
	VerifiedUsers   int
	UnverifiedUsers int
}

// userServiceImpl is the implementation of UserService.
type userServiceImpl struct {
	userRepo repository.UserRepository
}

// NewUserService creates a new UserService instance.
func NewUserService(userRepo repository.UserRepository) UserService {
	return &userServiceImpl{
		userRepo: userRepo,
	}
}

// GetProfile retrieves a user's profile.
func (s *userServiceImpl) GetProfile(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if err == repository.ErrUserNotFound {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	if user.IsDeleted() {
		return nil, ErrUserNotFound
	}

	return user, nil
}

// UpdateProfile updates a user's profile.
func (s *userServiceImpl) UpdateProfile(ctx context.Context, userID uuid.UUID, req dto.UpdateProfileRequest) (*domain.User, error) {
	// Get existing user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if err == repository.ErrUserNotFound {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	if user.IsDeleted() {
		return nil, ErrUserNotFound
	}

	// Update display name
	if req.DisplayName != nil {
		if err := user.UpdateDisplayName(*req.DisplayName); err != nil {
			return nil, err
		}
	}

	// Update avatar URL
	if req.AvatarURL != nil {
		user.UpdateAvatarURL(*req.AvatarURL)
	}

	// Save changes
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// DeleteAccount soft-deletes a user account.
func (s *userServiceImpl) DeleteAccount(ctx context.Context, userID uuid.UUID, password string) error {
	// Get user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if err == repository.ErrUserNotFound {
			return ErrUserNotFound
		}
		return err
	}

	// Verify password
	if !user.VerifyPassword(password) {
		return ErrPasswordMismatch
	}

	// Soft delete user
	if err := s.userRepo.Delete(ctx, userID); err != nil {
		return err
	}

	return nil
}

// ListUsers lists all users (admin only).
func (s *userServiceImpl) ListUsers(ctx context.Context, filter repository.UserListFilter) ([]*domain.User, int, error) {
	users, total, err := s.userRepo.List(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

// SearchUsers searches for users.
func (s *userServiceImpl) SearchUsers(ctx context.Context, query string, limit int) ([]*domain.User, error) {
	users, err := s.userRepo.Search(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	return users, nil
}

// GetUserByID retrieves a user by ID (admin only).
func (s *userServiceImpl) GetUserByID(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	return s.GetProfile(ctx, userID)
}

// GetUserByEmail retrieves a user by email (admin only).
func (s *userServiceImpl) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if err == repository.ErrUserNotFound {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	if user.IsDeleted() {
		return nil, ErrUserNotFound
	}

	return user, nil
}

// UpdateUserStatus updates a user's status (admin only).
func (s *userServiceImpl) UpdateUserStatus(ctx context.Context, userID uuid.UUID, status domain.UserStatus) error {
	// Check if user exists
	exists, err := s.userRepo.ExistsByID(ctx, userID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrUserNotFound
	}

	// Update status
	if err := s.userRepo.UpdateStatus(ctx, userID, status); err != nil {
		return err
	}

	return nil
}

// GetUserStats retrieves user statistics (admin only).
func (s *userServiceImpl) GetUserStats(ctx context.Context) (*UserStats, error) {
	stats := &UserStats{}

	// Total users
	total, err := s.userRepo.CountAll(ctx, false)
	if err != nil {
		return nil, err
	}
	stats.TotalUsers = total

	// Count by status
	active, err := s.userRepo.CountByStatus(ctx, domain.UserStatusActive)
	if err != nil {
		return nil, err
	}
	stats.ActiveUsers = active

	inactive, err := s.userRepo.CountByStatus(ctx, domain.UserStatusInactive)
	if err != nil {
		return nil, err
	}
	stats.InactiveUsers = inactive

	suspended, err := s.userRepo.CountByStatus(ctx, domain.UserStatusSuspended)
	if err != nil {
		return nil, err
	}
	stats.SuspendedUsers = suspended

	pending, err := s.userRepo.CountByStatus(ctx, domain.UserStatusPendingVerification)
	if err != nil {
		return nil, err
	}
	stats.PendingUsers = pending

	// TODO: Add verified/unverified counts if needed
	// This would require a separate query with email_verified filter

	return stats, nil
}
