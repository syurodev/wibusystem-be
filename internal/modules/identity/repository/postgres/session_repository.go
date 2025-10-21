// Package postgres provides PostgreSQL implementations of repository interfaces.
package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"wibusystem/internal/modules/identity/domain"
	"wibusystem/internal/modules/identity/repository"
)

// SessionRepository is the PostgreSQL implementation of repository.SessionRepository.
type SessionRepository struct {
	db     *pgxpool.Pool
	schema string
}

// NewSessionRepository creates a new PostgreSQL session repository.
func NewSessionRepository(db *pgxpool.Pool, schema string) *SessionRepository {
	if schema == "" {
		schema = "identity"
	}
	return &SessionRepository{
		db:     db,
		schema: schema,
	}
}

// Create creates a new session in the database.
func (r *SessionRepository) Create(ctx context.Context, session *domain.Session) error {
	if err := session.Validate(); err != nil {
		return repository.WrapRepositoryError(err, "invalid session data", "INVALID_SESSION_DATA")
	}

	query := fmt.Sprintf(`
		INSERT INTO %s.sessions (
			id, user_id, token_hash, ip_address, user_agent,
			expires_at, created_at, last_accessed_at, revoked
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		)
	`, r.schema)

	_, err := r.db.Exec(ctx, query,
		session.ID,
		session.UserID,
		session.TokenHash,
		session.IPAddress,
		session.UserAgent,
		session.ExpiresAt,
		session.CreatedAt,
		session.LastAccessedAt,
		session.Revoked,
	)

	if err != nil {
		return repository.WrapRepositoryError(err, "failed to create session", "CREATE_FAILED")
	}

	return nil
}

// GetByID retrieves a session by its ID.
func (r *SessionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Session, error) {
	query := fmt.Sprintf(`
		SELECT id, user_id, token_hash, ip_address, user_agent,
			   expires_at, created_at, last_accessed_at, revoked
		FROM %s.sessions
		WHERE id = $1
	`, r.schema)

	session := &domain.Session{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&session.ID,
		&session.UserID,
		&session.TokenHash,
		&session.IPAddress,
		&session.UserAgent,
		&session.ExpiresAt,
		&session.CreatedAt,
		&session.LastAccessedAt,
		&session.Revoked,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrSessionNotFound
		}
		return nil, repository.WrapRepositoryError(err, "failed to get session", "GET_FAILED")
	}

	return session, nil
}

// GetByTokenHash retrieves a session by its token hash.
func (r *SessionRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*domain.Session, error) {
	query := fmt.Sprintf(`
		SELECT id, user_id, token_hash, ip_address, user_agent,
			   expires_at, created_at, last_accessed_at, revoked
		FROM %s.sessions
		WHERE token_hash = $1
	`, r.schema)

	session := &domain.Session{}
	err := r.db.QueryRow(ctx, query, tokenHash).Scan(
		&session.ID,
		&session.UserID,
		&session.TokenHash,
		&session.IPAddress,
		&session.UserAgent,
		&session.ExpiresAt,
		&session.CreatedAt,
		&session.LastAccessedAt,
		&session.Revoked,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrSessionNotFound
		}
		return nil, repository.WrapRepositoryError(err, "failed to get session by token", "GET_FAILED")
	}

	return session, nil
}

// Update updates an existing session.
func (r *SessionRepository) Update(ctx context.Context, session *domain.Session) error {
	if err := session.Validate(); err != nil {
		return repository.WrapRepositoryError(err, "invalid session data", "INVALID_SESSION_DATA")
	}

	query := fmt.Sprintf(`
		UPDATE %s.sessions
		SET user_id = $2,
			token_hash = $3,
			ip_address = $4,
			user_agent = $5,
			expires_at = $6,
			last_accessed_at = $7,
			revoked = $8
		WHERE id = $1
	`, r.schema)

	result, err := r.db.Exec(ctx, query,
		session.ID,
		session.UserID,
		session.TokenHash,
		session.IPAddress,
		session.UserAgent,
		session.ExpiresAt,
		session.LastAccessedAt,
		session.Revoked,
	)

	if err != nil {
		return repository.WrapRepositoryError(err, "failed to update session", "UPDATE_FAILED")
	}

	if result.RowsAffected() == 0 {
		return repository.ErrSessionNotFound
	}

	return nil
}

// Delete removes a session.
func (r *SessionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := fmt.Sprintf(`DELETE FROM %s.sessions WHERE id = $1`, r.schema)
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return repository.WrapRepositoryError(err, "failed to delete session", "DELETE_FAILED")
	}
	if result.RowsAffected() == 0 {
		return repository.ErrSessionNotFound
	}
	return nil
}

// DeleteByTokenHash removes a session by token hash.
func (r *SessionRepository) DeleteByTokenHash(ctx context.Context, tokenHash string) error {
	query := fmt.Sprintf(`DELETE FROM %s.sessions WHERE token_hash = $1`, r.schema)
	result, err := r.db.Exec(ctx, query, tokenHash)
	if err != nil {
		return repository.WrapRepositoryError(err, "failed to delete session", "DELETE_FAILED")
	}
	if result.RowsAffected() == 0 {
		return repository.ErrSessionNotFound
	}
	return nil
}

// ListByUser retrieves all sessions for a user.
func (r *SessionRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]*domain.Session, error) {
	query := fmt.Sprintf(`
		SELECT id, user_id, token_hash, ip_address, user_agent,
			   expires_at, created_at, last_accessed_at, revoked
		FROM %s.sessions
		WHERE user_id = $1
		ORDER BY created_at DESC
	`, r.schema)

	return r.querySessions(ctx, query, userID)
}

// ListActiveByUser retrieves all active sessions for a user.
func (r *SessionRepository) ListActiveByUser(ctx context.Context, userID uuid.UUID) ([]*domain.Session, error) {
	now := time.Now().UTC()
	query := fmt.Sprintf(`
		SELECT id, user_id, token_hash, ip_address, user_agent,
			   expires_at, created_at, last_accessed_at, revoked
		FROM %s.sessions
		WHERE user_id = $1 AND revoked = false AND expires_at > $2
		ORDER BY created_at DESC
	`, r.schema)

	return r.querySessions(ctx, query, userID, now)
}

// ListByUserPaginated retrieves sessions with pagination.
func (r *SessionRepository) ListByUserPaginated(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.Session, int, error) {
	// Count total
	var total int
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM %s.sessions WHERE user_id = $1`, r.schema)
	if err := r.db.QueryRow(ctx, countQuery, userID).Scan(&total); err != nil {
		return nil, 0, repository.WrapRepositoryError(err, "failed to count sessions", "COUNT_FAILED")
	}

	// Get sessions
	query := fmt.Sprintf(`
		SELECT id, user_id, token_hash, ip_address, user_agent,
			   expires_at, created_at, last_accessed_at, revoked
		FROM %s.sessions
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, r.schema)

	sessions, err := r.querySessions(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	return sessions, total, nil
}

// Revoke revokes a session.
func (r *SessionRepository) Revoke(ctx context.Context, id uuid.UUID) error {
	query := fmt.Sprintf(`UPDATE %s.sessions SET revoked = true WHERE id = $1`, r.schema)
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return repository.WrapRepositoryError(err, "failed to revoke session", "REVOKE_FAILED")
	}
	if result.RowsAffected() == 0 {
		return repository.ErrSessionNotFound
	}
	return nil
}

// RevokeByTokenHash revokes a session by token hash.
func (r *SessionRepository) RevokeByTokenHash(ctx context.Context, tokenHash string) error {
	query := fmt.Sprintf(`UPDATE %s.sessions SET revoked = true WHERE token_hash = $1`, r.schema)
	result, err := r.db.Exec(ctx, query, tokenHash)
	if err != nil {
		return repository.WrapRepositoryError(err, "failed to revoke session", "REVOKE_FAILED")
	}
	if result.RowsAffected() == 0 {
		return repository.ErrSessionNotFound
	}
	return nil
}

// RevokeAllByUser revokes all sessions for a user.
func (r *SessionRepository) RevokeAllByUser(ctx context.Context, userID uuid.UUID) (int, error) {
	query := fmt.Sprintf(`UPDATE %s.sessions SET revoked = true WHERE user_id = $1 AND revoked = false`, r.schema)
	result, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		return 0, repository.WrapRepositoryError(err, "failed to revoke sessions", "REVOKE_FAILED")
	}
	return int(result.RowsAffected()), nil
}

// RevokeAllExcept revokes all sessions except one.
func (r *SessionRepository) RevokeAllExcept(ctx context.Context, userID, exceptSessionID uuid.UUID) (int, error) {
	query := fmt.Sprintf(`UPDATE %s.sessions SET revoked = true WHERE user_id = $1 AND id != $2 AND revoked = false`, r.schema)
	result, err := r.db.Exec(ctx, query, userID, exceptSessionID)
	if err != nil {
		return 0, repository.WrapRepositoryError(err, "failed to revoke sessions", "REVOKE_FAILED")
	}
	return int(result.RowsAffected()), nil
}

// DeleteExpired deletes all expired sessions.
func (r *SessionRepository) DeleteExpired(ctx context.Context) (int, error) {
	now := time.Now().UTC()
	query := fmt.Sprintf(`DELETE FROM %s.sessions WHERE expires_at < $1`, r.schema)
	result, err := r.db.Exec(ctx, query, now)
	if err != nil {
		return 0, repository.WrapRepositoryError(err, "failed to delete expired sessions", "DELETE_FAILED")
	}
	return int(result.RowsAffected()), nil
}

// DeleteExpiredBefore deletes sessions expired before a time.
func (r *SessionRepository) DeleteExpiredBefore(ctx context.Context, before time.Time) (int, error) {
	query := fmt.Sprintf(`DELETE FROM %s.sessions WHERE expires_at < $1`, r.schema)
	result, err := r.db.Exec(ctx, query, before)
	if err != nil {
		return 0, repository.WrapRepositoryError(err, "failed to delete expired sessions", "DELETE_FAILED")
	}
	return int(result.RowsAffected()), nil
}

// DeleteRevoked deletes all revoked sessions.
func (r *SessionRepository) DeleteRevoked(ctx context.Context) (int, error) {
	query := fmt.Sprintf(`DELETE FROM %s.sessions WHERE revoked = true`, r.schema)
	result, err := r.db.Exec(ctx, query)
	if err != nil {
		return 0, repository.WrapRepositoryError(err, "failed to delete revoked sessions", "DELETE_FAILED")
	}
	return int(result.RowsAffected()), nil
}

// DeleteByUser deletes all sessions for a user.
func (r *SessionRepository) DeleteByUser(ctx context.Context, userID uuid.UUID) (int, error) {
	query := fmt.Sprintf(`DELETE FROM %s.sessions WHERE user_id = $1`, r.schema)
	result, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		return 0, repository.WrapRepositoryError(err, "failed to delete user sessions", "DELETE_FAILED")
	}
	return int(result.RowsAffected()), nil
}

// CountByUser returns total sessions for a user.
func (r *SessionRepository) CountByUser(ctx context.Context, userID uuid.UUID) (int, error) {
	query := fmt.Sprintf(`SELECT COUNT(*) FROM %s.sessions WHERE user_id = $1`, r.schema)
	var count int
	if err := r.db.QueryRow(ctx, query, userID).Scan(&count); err != nil {
		return 0, repository.WrapRepositoryError(err, "failed to count sessions", "COUNT_FAILED")
	}
	return count, nil
}

// CountActiveByUser returns active sessions for a user.
func (r *SessionRepository) CountActiveByUser(ctx context.Context, userID uuid.UUID) (int, error) {
	now := time.Now().UTC()
	query := fmt.Sprintf(`SELECT COUNT(*) FROM %s.sessions WHERE user_id = $1 AND revoked = false AND expires_at > $2`, r.schema)
	var count int
	if err := r.db.QueryRow(ctx, query, userID, now).Scan(&count); err != nil {
		return 0, repository.WrapRepositoryError(err, "failed to count active sessions", "COUNT_FAILED")
	}
	return count, nil
}

// CountExpired returns number of expired sessions.
func (r *SessionRepository) CountExpired(ctx context.Context) (int, error) {
	now := time.Now().UTC()
	query := fmt.Sprintf(`SELECT COUNT(*) FROM %s.sessions WHERE expires_at < $1`, r.schema)
	var count int
	if err := r.db.QueryRow(ctx, query, now).Scan(&count); err != nil {
		return 0, repository.WrapRepositoryError(err, "failed to count expired sessions", "COUNT_FAILED")
	}
	return count, nil
}

// ExistsByID checks if a session exists.
func (r *SessionRepository) ExistsByID(ctx context.Context, id uuid.UUID) (bool, error) {
	query := fmt.Sprintf(`SELECT EXISTS(SELECT 1 FROM %s.sessions WHERE id = $1)`, r.schema)
	var exists bool
	if err := r.db.QueryRow(ctx, query, id).Scan(&exists); err != nil {
		return false, repository.WrapRepositoryError(err, "failed to check session", "EXISTS_CHECK_FAILED")
	}
	return exists, nil
}

// ExistsByTokenHash checks if a session exists by token.
func (r *SessionRepository) ExistsByTokenHash(ctx context.Context, tokenHash string) (bool, error) {
	query := fmt.Sprintf(`SELECT EXISTS(SELECT 1 FROM %s.sessions WHERE token_hash = $1)`, r.schema)
	var exists bool
	if err := r.db.QueryRow(ctx, query, tokenHash).Scan(&exists); err != nil {
		return false, repository.WrapRepositoryError(err, "failed to check session", "EXISTS_CHECK_FAILED")
	}
	return exists, nil
}

// UpdateLastAccessed updates last accessed time.
func (r *SessionRepository) UpdateLastAccessed(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()
	query := fmt.Sprintf(`UPDATE %s.sessions SET last_accessed_at = $2 WHERE id = $1`, r.schema)
	result, err := r.db.Exec(ctx, query, id, now)
	if err != nil {
		return repository.WrapRepositoryError(err, "failed to update last accessed", "UPDATE_FAILED")
	}
	if result.RowsAffected() == 0 {
		return repository.ErrSessionNotFound
	}
	return nil
}

// ExtendExpiration extends expiration time.
func (r *SessionRepository) ExtendExpiration(ctx context.Context, id uuid.UUID, expiresAt time.Time) error {
	query := fmt.Sprintf(`UPDATE %s.sessions SET expires_at = $2 WHERE id = $1`, r.schema)
	result, err := r.db.Exec(ctx, query, id, expiresAt)
	if err != nil {
		return repository.WrapRepositoryError(err, "failed to extend expiration", "UPDATE_FAILED")
	}
	if result.RowsAffected() == 0 {
		return repository.ErrSessionNotFound
	}
	return nil
}

// GetExpiringSoon retrieves sessions expiring soon.
func (r *SessionRepository) GetExpiringSoon(ctx context.Context, within time.Duration, limit int) ([]*domain.Session, error) {
	threshold := time.Now().UTC().Add(within)
	query := fmt.Sprintf(`
		SELECT id, user_id, token_hash, ip_address, user_agent,
			   expires_at, created_at, last_accessed_at, revoked
		FROM %s.sessions
		WHERE revoked = false AND expires_at > NOW() AND expires_at < $1
		ORDER BY expires_at ASC
		LIMIT $2
	`, r.schema)

	return r.querySessions(ctx, query, threshold, limit)
}

// CleanupOldSessions deletes old sessions.
func (r *SessionRepository) CleanupOldSessions(ctx context.Context, olderThan time.Duration) (int, error) {
	threshold := time.Now().UTC().Add(-olderThan)
	query := fmt.Sprintf(`DELETE FROM %s.sessions WHERE created_at < $1`, r.schema)
	result, err := r.db.Exec(ctx, query, threshold)
	if err != nil {
		return 0, repository.WrapRepositoryError(err, "failed to cleanup sessions", "DELETE_FAILED")
	}
	return int(result.RowsAffected()), nil
}

// querySessions is a helper to query multiple sessions.
func (r *SessionRepository) querySessions(ctx context.Context, query string, args ...any) ([]*domain.Session, error) {
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, repository.WrapRepositoryError(err, "failed to query sessions", "QUERY_FAILED")
	}
	defer rows.Close()

	sessions := []*domain.Session{}
	for rows.Next() {
		session := &domain.Session{}
		if err := rows.Scan(
			&session.ID,
			&session.UserID,
			&session.TokenHash,
			&session.IPAddress,
			&session.UserAgent,
			&session.ExpiresAt,
			&session.CreatedAt,
			&session.LastAccessedAt,
			&session.Revoked,
		); err != nil {
			return nil, repository.WrapRepositoryError(err, "failed to scan session", "SCAN_FAILED")
		}
		sessions = append(sessions, session)
	}

	return sessions, rows.Err()
}
