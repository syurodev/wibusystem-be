package repository

import (
	"context"
	"system/internal/domain"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ory/fosite"
)

// oauth2ClientRepository triển khai OAuth2ClientRepository sử dụng pgx.
type oauth2ClientRepository struct {
	pool *pgxpool.Pool
}

// NewOAuth2ClientRepository tạo một instance mới của oauth2ClientRepository.
func NewOAuth2ClientRepository(pool *pgxpool.Pool) domain.OAuth2ClientRepository {
	return &oauth2ClientRepository{pool: pool}
}

// GetClientByID lấy thông tin client từ database bằng pgx.
func (r *oauth2ClientRepository) GetClientByID(ctx context.Context, id uuid.UUID) (*domain.OAuth2Client, error) {
	query := `SELECT id, client_name, secret_hash, redirect_uris, grant_types, response_types, scopes, is_public, tenant_id FROM identify.oauth2_clients WHERE id=$1`

	rows, err := r.pool.Query(ctx, query, id)
	if err != nil {
		return nil, err
	}

	client, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.OAuth2Client])
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, err
		}
		return nil, err
	}

	return &client, nil
}

// ClientAssertionJWTValid kiểm tra xem một JTI (JWT ID) đã được sử dụng hay chưa.
func (r *oauth2ClientRepository) ClientAssertionJWTValid(ctx context.Context, jti string) error {
	query := `SELECT EXISTS(SELECT 1 FROM identify.oauth2_jti_blacklist WHERE signature=$1)`
	var exists bool
	err := r.pool.QueryRow(ctx, query, jti).Scan(&exists)
	if err != nil {
		return err
	}
	if exists {
		return fosite.ErrJTIKnown
	}
	return nil
}

// SetClientAssertionJWT lưu một JTI vào blacklist để chống tấn công phát lại (replay attacks).
func (r *oauth2ClientRepository) SetClientAssertionJWT(ctx context.Context, jti string, exp time.Time) error {
	query := `INSERT INTO identify.oauth2_jti_blacklist (signature, expires_at) VALUES ($1, $2)`
	_, err := r.pool.Exec(ctx, query, jti, exp)
	return err
}
