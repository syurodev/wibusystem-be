// Package repository defines interfaces for data persistence in the Identity module.
package repository

import (
	"context"

	"github.com/google/uuid"

	"wibusystem/internal/modules/identity/domain"
)

// UserRepository defines the interface for user data persistence.
type UserRepository interface {
	// Create creates a new user in the database.
	Create(ctx context.Context, user *domain.User) error

	// GetByID retrieves a user by their ID.
	// Returns ErrUserNotFound if the user doesn't exist.
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)

	// GetByEmail retrieves a user by their email address.
	// Returns ErrUserNotFound if the user doesn't exist.
	GetByEmail(ctx context.Context, email string) (*domain.User, error)

	// Update updates an existing user's information.
	// Returns ErrUserNotFound if the user doesn't exist.
	Update(ctx context.Context, user *domain.User) error

	// Delete soft-deletes a user by their ID.
	// Returns ErrUserNotFound if the user doesn't exist.
	Delete(ctx context.Context, id uuid.UUID) error

	// HardDelete permanently deletes a user from the database.
	// Use with caution - this operation is irreversible.
	HardDelete(ctx context.Context, id uuid.UUID) error

	// List retrieves a paginated list of users based on the provided filter.
	// Returns the users, total count, and any error.
	List(ctx context.Context, filter UserListFilter) ([]*domain.User, int, error)

	// ExistsByEmail checks if a user with the given email exists.
	ExistsByEmail(ctx context.Context, email string) (bool, error)

	// ExistsByID checks if a user with the given ID exists.
	ExistsByID(ctx context.Context, id uuid.UUID) (bool, error)

	// CountAll returns the total number of users (including soft-deleted if specified).
	CountAll(ctx context.Context, includeDeleted bool) (int, error)

	// CountByStatus returns the number of users with a specific status.
	CountByStatus(ctx context.Context, status domain.UserStatus) (int, error)

	// UpdateLastLogin updates the user's last login timestamp.
	UpdateLastLogin(ctx context.Context, id uuid.UUID) error

	// UpdatePassword updates the user's password hash.
	UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error

	// VerifyEmail marks a user's email as verified.
	VerifyEmail(ctx context.Context, id uuid.UUID) error

	// UpdateStatus updates a user's status.
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.UserStatus) error

	// Search searches for users by email or display name.
	// Useful for admin interfaces or member selection.
	Search(ctx context.Context, query string, limit int) ([]*domain.User, error)

	// GetByIDs retrieves multiple users by their IDs.
	// Returns users that exist, skips non-existent ones.
	GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*domain.User, error)
}

// UserListFilter contains filtering and pagination options for listing users.
type UserListFilter struct {
	// Pagination
	Limit  int // Maximum number of results to return
	Offset int // Number of results to skip

	// Filters
	Status         *domain.UserStatus // Filter by user status
	EmailVerified  *bool              // Filter by email verification status
	IncludeDeleted bool               // Include soft-deleted users

	// Search
	EmailContains       string // Filter by email containing string
	DisplayNameContains string // Filter by display name containing string

	// Sorting
	SortBy    string // Field to sort by (created_at, email, display_name)
	SortOrder string // Sort order (asc, desc)
}

// Repository errors
var (
	// ErrUserNotFound is returned when a user cannot be found.
	ErrUserNotFound = NewRepositoryError("user not found", "USER_NOT_FOUND")

	// ErrUserAlreadyExists is returned when attempting to create a user with an email that already exists.
	ErrUserAlreadyExists = NewRepositoryError("user with this email already exists", "USER_ALREADY_EXISTS")

	// ErrInvalidUserData is returned when user data is invalid.
	ErrInvalidUserData = NewRepositoryError("invalid user data", "INVALID_USER_DATA")

	// ErrDatabaseConnection is returned when there's a database connection issue.
	ErrDatabaseConnection = NewRepositoryError("database connection error", "DB_CONNECTION_ERROR")

	// ErrTransactionFailed is returned when a database transaction fails.
	ErrTransactionFailed = NewRepositoryError("transaction failed", "TRANSACTION_FAILED")
)

// RepositoryError represents a repository-level error.
type RepositoryError struct {
	Message string
	Code    string
	Err     error
}

// Error implements the error interface.
func (e *RepositoryError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

// Unwrap returns the wrapped error.
func (e *RepositoryError) Unwrap() error {
	return e.Err
}

// NewRepositoryError creates a new RepositoryError.
func NewRepositoryError(message, code string) *RepositoryError {
	return &RepositoryError{
		Message: message,
		Code:    code,
	}
}

// WrapRepositoryError wraps an error with a RepositoryError.
func WrapRepositoryError(err error, message, code string) *RepositoryError {
	return &RepositoryError{
		Message: message,
		Code:    code,
		Err:     err,
	}
}
