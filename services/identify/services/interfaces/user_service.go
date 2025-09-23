package interfaces

import (
	"context"

	"github.com/google/uuid"
	d "wibusystem/pkg/common/dto"
	m "wibusystem/pkg/common/model"
)

// UserServiceInterface defines the contract for user operations
type UserServiceInterface interface {
	// CreateUser creates a new user with validation
	CreateUser(ctx context.Context, req d.CreateUserRequest) (*m.User, error)

	// GetUserByID retrieves a user by ID
	GetUserByID(ctx context.Context, userID uuid.UUID) (*m.User, error)

	// GetUserByEmail retrieves a user by email
	GetUserByEmail(ctx context.Context, email string) (*m.User, error)

	// GetUserByUsername retrieves a user by username
	GetUserByUsername(ctx context.Context, username string) (*m.User, error)

	// ListUsers retrieves paginated list of users
	ListUsers(ctx context.Context, page, pageSize int) ([]*m.User, int64, error)

	// UpdateUser updates user information
	UpdateUser(ctx context.Context, userID uuid.UUID, req d.UpdateUserRequest) (*m.User, error)

	// DeleteUser soft deletes a user
	DeleteUser(ctx context.Context, userID uuid.UUID) error

	// UpdateLastLogin updates user's last login timestamp
	UpdateLastLogin(ctx context.Context, userID uuid.UUID) error

	// ValidateUserData validates user data before creation/update
	ValidateUserData(req d.CreateUserRequest) error

	// CheckEmailExists checks if email is already taken
	CheckEmailExists(ctx context.Context, email string) (bool, error)

	// CheckUsernameExists checks if username is already taken
	CheckUsernameExists(ctx context.Context, username string) (bool, error)
}