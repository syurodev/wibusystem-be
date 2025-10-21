// Package repository defines interfaces for data persistence in the Identity module.
package repository

import (
	"context"
	"time"

	"github.com/google/uuid"

	"wibusystem/internal/modules/identity/domain"
)

// SessionRepository defines the interface for session data persistence.
type SessionRepository interface {
	// Create creates a new session in the database.
	Create(ctx context.Context, session *domain.Session) error

	// GetByID retrieves a session by its ID.
	// Returns ErrSessionNotFound if the session doesn't exist.
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Session, error)

	// GetByTokenHash retrieves a session by its token hash.
	// Returns ErrSessionNotFound if the session doesn't exist.
	GetByTokenHash(ctx context.Context, tokenHash string) (*domain.Session, error)

	// Update updates an existing session's information.
	// Returns ErrSessionNotFound if the session doesn't exist.
	Update(ctx context.Context, session *domain.Session) error

	// Delete removes a session from the database.
	// Returns ErrSessionNotFound if the session doesn't exist.
	Delete(ctx context.Context, id uuid.UUID) error

	// DeleteByTokenHash removes a session by its token hash.
	// Returns ErrSessionNotFound if the session doesn't exist.
	DeleteByTokenHash(ctx context.Context, tokenHash string) error

	// ListByUser retrieves all sessions for a specific user.
	// Includes both active and expired sessions unless filtered.
	ListByUser(ctx context.Context, userID uuid.UUID) ([]*domain.Session, error)

	// ListActiveByUser retrieves all active (not expired, not revoked) sessions for a user.
	ListActiveByUser(ctx context.Context, userID uuid.UUID) ([]*domain.Session, error)

	// ListByUserPaginated retrieves sessions for a user with pagination.
	ListByUserPaginated(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.Session, int, error)

	// Revoke revokes a specific session.
	// Returns ErrSessionNotFound if the session doesn't exist.
	Revoke(ctx context.Context, id uuid.UUID) error

	// RevokeByTokenHash revokes a session by its token hash.
	// Returns ErrSessionNotFound if the session doesn't exist.
	RevokeByTokenHash(ctx context.Context, tokenHash string) error

	// RevokeAllByUser revokes all sessions for a specific user.
	// Returns the number of sessions revoked.
	RevokeAllByUser(ctx context.Context, userID uuid.UUID) (int, error)

	// RevokeAllExcept revokes all sessions for a user except the specified session.
	// Useful for "logout from all other devices" functionality.
	RevokeAllExcept(ctx context.Context, userID, exceptSessionID uuid.UUID) (int, error)

	// DeleteExpired deletes all expired sessions.
	// Returns the number of sessions deleted.
	DeleteExpired(ctx context.Context) (int, error)

	// DeleteExpiredBefore deletes all sessions that expired before the given time.
	// Returns the number of sessions deleted.
	DeleteExpiredBefore(ctx context.Context, before time.Time) (int, error)

	// DeleteRevoked deletes all revoked sessions.
	// Returns the number of sessions deleted.
	DeleteRevoked(ctx context.Context) (int, error)

	// DeleteByUser deletes all sessions for a specific user.
	// Returns the number of sessions deleted.
	DeleteByUser(ctx context.Context, userID uuid.UUID) (int, error)

	// CountByUser returns the total number of sessions for a user.
	CountByUser(ctx context.Context, userID uuid.UUID) (int, error)

	// CountActiveByUser returns the number of active sessions for a user.
	CountActiveByUser(ctx context.Context, userID uuid.UUID) (int, error)

	// CountExpired returns the number of expired sessions across all users.
	CountExpired(ctx context.Context) (int, error)

	// ExistsByID checks if a session with the given ID exists.
	ExistsByID(ctx context.Context, id uuid.UUID) (bool, error)

	// ExistsByTokenHash checks if a session with the given token hash exists.
	ExistsByTokenHash(ctx context.Context, tokenHash string) (bool, error)

	// UpdateLastAccessed updates the last accessed timestamp for a session.
	UpdateLastAccessed(ctx context.Context, id uuid.UUID) error

	// ExtendExpiration extends a session's expiration time.
	ExtendExpiration(ctx context.Context, id uuid.UUID, expiresAt time.Time) error

	// GetExpiringSoon retrieves sessions that are expiring within the specified duration.
	// Useful for sending expiration warnings or auto-refresh.
	GetExpiringSoon(ctx context.Context, within time.Duration, limit int) ([]*domain.Session, error)

	// CleanupOldSessions deletes sessions older than the specified duration.
	// This includes both expired and revoked sessions.
	// Returns the number of sessions deleted.
	CleanupOldSessions(ctx context.Context, olderThan time.Duration) (int, error)
}

// SessionListFilter contains filtering and pagination options for listing sessions.
type SessionListFilter struct {
	// Pagination
	Limit  int // Maximum number of results to return
	Offset int // Number of results to skip

	// Filters
	UserID         *uuid.UUID // Filter by user ID
	Revoked        *bool      // Filter by revoked status
	IncludeExpired bool       // Include expired sessions
	OnlyActive     bool       // Only return active sessions (not expired, not revoked)

	// Time filters
	CreatedAfter  *time.Time // Filter sessions created after this time
	CreatedBefore *time.Time // Filter sessions created before this time
	ExpiresAfter  *time.Time // Filter sessions expiring after this time
	ExpiresBefore *time.Time // Filter sessions expiring before this time

	// Device filters
	IPAddress *string // Filter by IP address
	UserAgent *string // Filter by user agent (partial match)

	// Sorting
	SortBy    string // Field to sort by (created_at, last_accessed_at, expires_at)
	SortOrder string // Sort order (asc, desc)
}

// Repository errors for Session
var (
	// ErrSessionNotFound is returned when a session cannot be found.
	ErrSessionNotFound = NewRepositoryError("session not found", "SESSION_NOT_FOUND")

	// ErrSessionExpired is returned when a session has expired.
	ErrSessionExpired = NewRepositoryError("session has expired", "SESSION_EXPIRED")

	// ErrSessionRevoked is returned when a session has been revoked.
	ErrSessionRevoked = NewRepositoryError("session has been revoked", "SESSION_REVOKED")

	// ErrInvalidSessionData is returned when session data is invalid.
	ErrInvalidSessionData = NewRepositoryError("invalid session data", "INVALID_SESSION_DATA")

	// ErrSessionAlreadyExists is returned when a session with the same token hash already exists.
	ErrSessionAlreadyExists = NewRepositoryError("session with this token already exists", "SESSION_ALREADY_EXISTS")
)
