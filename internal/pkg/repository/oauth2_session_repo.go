package repository

import (
	"context"
	"system/internal/domain"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// oauth2SessionRepository triển khai OAuth2SessionRepository.
type oauth2SessionRepository struct {
	pool *pgxpool.Pool
}

// NewOAuth2SessionRepository tạo một instance mới của oauth2SessionRepository.
func NewOAuth2SessionRepository(pool *pgxpool.Pool) domain.OAuth2SessionRepository {
	return &oauth2SessionRepository{pool: pool}
}

// CreateSession chèn một bản ghi session mới vào bảng identify.oauth2_sessions.
func (r *oauth2SessionRepository) CreateSession(ctx context.Context, signature string, requestID string, sessionType string, sessionData []byte, expiresAt time.Time, clientID string, subject string) error {
	query := `
		INSERT INTO identify.oauth2_sessions (signature, request_id, session_type, session_data, expires_at, client_id, subject_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	// Convert []byte to string for JSONB column
	_, err := r.pool.Exec(ctx, query, signature, requestID, sessionType, string(sessionData), expiresAt, clientID, subject)
	return err
}

// GetSessionBySignature lấy dữ liệu session (dưới dạng []byte) từ DB.
func (r *oauth2SessionRepository) GetSessionBySignature(ctx context.Context, signature string) ([]byte, error) {
	var sessionData []byte
	query := `SELECT session_data FROM identify.oauth2_sessions WHERE signature = $1 AND active = TRUE`
	err := r.pool.QueryRow(ctx, query, signature).Scan(&sessionData)
	if err != nil {
		return nil, err // Trả về lỗi gốc (có thể là pgx.ErrNoRows)
	}
	return sessionData, nil
}

// DeleteSessionBySignature xóa một session khỏi DB theo signature.
func (r *oauth2SessionRepository) DeleteSessionBySignature(ctx context.Context, signature string) error {
	query := `DELETE FROM identify.oauth2_sessions WHERE signature = $1`
	_, err := r.pool.Exec(ctx, query, signature)
	return err
}

// RevokeAllUserSessions đánh dấu inactive tất cả sessions của một user (theo subject_id).
func (r *oauth2SessionRepository) RevokeAllUserSessions(ctx context.Context, subjectID string) error {
	query := `
		UPDATE identify.oauth2_sessions
		SET active = FALSE
		WHERE subject_id = $1 AND active = TRUE
	`
	_, err := r.pool.Exec(ctx, query, subjectID)
	return err
}

// GetActiveSessionsBySubject lấy tất cả signatures của sessions đang active của một user.
func (r *oauth2SessionRepository) GetActiveSessionsBySubject(ctx context.Context, subjectID string) ([]string, error) {
	query := `
		SELECT signature
		FROM identify.oauth2_sessions
		WHERE subject_id = $1 AND active = TRUE
	`
	rows, err := r.pool.Query(ctx, query, subjectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var signatures []string
	for rows.Next() {
		var signature string
		if err := rows.Scan(&signature); err != nil {
			return nil, err
		}
		signatures = append(signatures, signature)
	}

	return signatures, rows.Err()
}
