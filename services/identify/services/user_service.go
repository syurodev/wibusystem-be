package services

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"

	d "wibusystem/pkg/common/dto"
	m "wibusystem/pkg/common/model"
	"wibusystem/services/identify/repositories"
	"wibusystem/services/identify/services/interfaces"
)

// UserService implements user-related business logic
type UserService struct {
	repos *repositories.Repositories
}

// NewUserService creates a new user service
func NewUserService(repos *repositories.Repositories) interfaces.UserServiceInterface {
	return &UserService{
		repos: repos,
	}
}

// CreateUser creates a new user with validation
func (s *UserService) CreateUser(ctx context.Context, req d.CreateUserRequest) (*m.User, error) {
	if err := s.ValidateUserData(req); err != nil {
		return nil, err
	}

	// Check if email already exists
	exists, err := s.CheckEmailExists(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check email existence: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("user with email %s already exists", req.Email)
	}

	// Check if username already exists (if provided)
	if req.Username != "" {
		exists, err := s.CheckUsernameExists(ctx, req.Username)
		if err != nil {
			return nil, fmt.Errorf("failed to check username existence: %w", err)
		}
		if exists {
			return nil, fmt.Errorf("username %s is already taken", req.Username)
		}
	}

	// Create user
	user := &m.User{
		Email:       req.Email,
		Username:    req.Username,
		DisplayName: req.DisplayName,
	}

	if err := s.repos.User.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// GetUserByID retrieves a user by ID
func (s *UserService) GetUserByID(ctx context.Context, userID uuid.UUID) (*m.User, error) {
	if userID == uuid.Nil {
		return nil, fmt.Errorf("user ID cannot be nil")
	}

	user, err := s.repos.User.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return user, nil
}

// GetUserByEmail retrieves a user by email
func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*m.User, error) {
	if email == "" {
		return nil, fmt.Errorf("email cannot be empty")
	}

	user, err := s.repos.User.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return user, nil
}

// GetUserByUsername retrieves a user by username
func (s *UserService) GetUserByUsername(ctx context.Context, username string) (*m.User, error) {
	if username == "" {
		return nil, fmt.Errorf("username cannot be empty")
	}

	user, err := s.repos.User.GetByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	return user, nil
}

// ListUsers retrieves paginated list of users
func (s *UserService) ListUsers(ctx context.Context, page, pageSize int) ([]*m.User, int64, error) {
	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	users, total, err := s.repos.User.List(ctx, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}

	return users, total, nil
}

// UpdateUser updates user information
func (s *UserService) UpdateUser(ctx context.Context, userID uuid.UUID, req d.UpdateUserRequest) (*m.User, error) {
	if userID == uuid.Nil {
		return nil, fmt.Errorf("user ID cannot be nil")
	}

	// Get existing user
	user, err := s.repos.User.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Validate email if being updated
	if req.Email != nil && *req.Email != user.Email {
		if err := s.validateEmail(*req.Email); err != nil {
			return nil, err
		}
		exists, err := s.CheckEmailExists(ctx, *req.Email)
		if err != nil {
			return nil, fmt.Errorf("failed to check email existence: %w", err)
		}
		if exists {
			return nil, fmt.Errorf("email %s is already taken", *req.Email)
		}
		user.Email = *req.Email
	}

	// Validate username if being updated
	if req.Username != nil && *req.Username != user.Username {
		if err := s.validateUsername(*req.Username); err != nil {
			return nil, err
		}
		exists, err := s.CheckUsernameExists(ctx, *req.Username)
		if err != nil {
			return nil, fmt.Errorf("failed to check username existence: %w", err)
		}
		if exists {
			return nil, fmt.Errorf("username %s is already taken", *req.Username)
		}
		user.Username = *req.Username
	}

	// Update display name if provided
	if req.DisplayName != nil {
		user.DisplayName = *req.DisplayName
	}

	if err := s.repos.User.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}

// DeleteUser soft deletes a user
func (s *UserService) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	if userID == uuid.Nil {
		return fmt.Errorf("user ID cannot be nil")
	}

	if err := s.repos.User.Delete(ctx, userID); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// UpdateLastLogin updates user's last login timestamp
func (s *UserService) UpdateLastLogin(ctx context.Context, userID uuid.UUID) error {
	if userID == uuid.Nil {
		return fmt.Errorf("user ID cannot be nil")
	}

	if err := s.repos.User.UpdateLastLogin(ctx, userID); err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}

	return nil
}

// ValidateUserData validates user data before creation/update
func (s *UserService) ValidateUserData(req d.CreateUserRequest) error {
	if err := s.validateEmail(req.Email); err != nil {
		return err
	}

	if req.Username != "" {
		if err := s.validateUsername(req.Username); err != nil {
			return err
		}
	}

	if err := s.validateDisplayName(req.DisplayName); err != nil {
		return err
	}

	return nil
}

// CheckEmailExists checks if email is already taken
func (s *UserService) CheckEmailExists(ctx context.Context, email string) (bool, error) {
	if email == "" {
		return false, fmt.Errorf("email cannot be empty")
	}

	_, err := s.repos.User.GetByEmail(ctx, email)
	if err != nil {
		// If no user found, email doesn't exist
		return false, nil
	}

	// User found, email exists
	return true, nil
}

// CheckUsernameExists checks if username is already taken
func (s *UserService) CheckUsernameExists(ctx context.Context, username string) (bool, error) {
	if username == "" {
		return false, fmt.Errorf("username cannot be empty")
	}

	_, err := s.repos.User.GetByUsername(ctx, username)
	if err != nil {
		// If no user found, username doesn't exist
		return false, nil
	}

	// User found, username exists
	return true, nil
}

// Private validation methods

func (s *UserService) validateEmail(email string) error {
	if email == "" {
		return fmt.Errorf("email is required")
	}

	if len(email) > 255 {
		return fmt.Errorf("email must not exceed 255 characters")
	}

	// Basic email validation regex
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return fmt.Errorf("invalid email format")
	}

	return nil
}

func (s *UserService) validateUsername(username string) error {
	if username == "" {
		return nil // Username is optional
	}

	if len(username) < 3 {
		return fmt.Errorf("username must be at least 3 characters long")
	}

	if len(username) > 50 {
		return fmt.Errorf("username must not exceed 50 characters")
	}

	// Username can contain letters, numbers, underscore, hyphen
	usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !usernameRegex.MatchString(username) {
		return fmt.Errorf("username can only contain letters, numbers, underscore, and hyphen")
	}

	return nil
}

func (s *UserService) validateDisplayName(displayName string) error {
	if displayName != "" {
		displayName = strings.TrimSpace(displayName)
		if len(displayName) > 100 {
			return fmt.Errorf("display name must not exceed 100 characters")
		}
	}

	return nil
}