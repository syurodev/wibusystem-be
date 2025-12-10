package auth

import (
	"context"
	"system/internal/domain"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// webauthnSessionRepository triển khai WebAuthnSessionRepository sử dụng pgx
type webauthnSessionRepository struct {
	pool *pgxpool.Pool
}

// NewWebAuthnSessionRepository tạo một instance mới của webauthnSessionRepository
func NewWebAuthnSessionRepository(pool *pgxpool.Pool) domain.WebAuthnSessionRepository {
	return &webauthnSessionRepository{pool: pool}
}

// GetByChallenge lấy session từ database theo challenge
func (r *webauthnSessionRepository) GetByChallenge(ctx context.Context, challenge string) (*domain.WebAuthnSession, error) {
	query := `
		SELECT id, user_id, challenge, session_type, user_agent, ip_address,
		       created_at, expires_at
		FROM identify.webauthn_sessions
		WHERE challenge = $1 AND expires_at > NOW()
	`

	rows, err := r.pool.Query(ctx, query, challenge)
	if err != nil {
		return nil, err
	}

	session, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.WebAuthnSession])
	if err != nil {
		return nil, err
	}

	return &session, nil
}

// Create tạo session mới trong database
func (r *webauthnSessionRepository) Create(ctx context.Context, session *domain.WebAuthnSession) error {
	query := `
		INSERT INTO identify.webauthn_sessions (
			id, user_id, challenge, session_type, user_agent, ip_address, expires_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.pool.Exec(ctx, query,
		session.ID,
		session.UserID,
		session.Challenge,
		session.SessionType,
		session.UserAgent,
		session.IPAddress,
		session.ExpiresAt,
	)

	return err
}

// Delete xóa session khỏi database
func (r *webauthnSessionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		DELETE FROM identify.webauthn_sessions
		WHERE id = $1
	`

	_, err := r.pool.Exec(ctx, query, id)
	return err
}

// DeleteExpired xóa các sessions đã hết hạn
func (r *webauthnSessionRepository) DeleteExpired(ctx context.Context) error {
	query := `
		DELETE FROM identify.webauthn_sessions
		WHERE expires_at <= NOW()
	`

	_, err := r.pool.Exec(ctx, query)
	return err
}
