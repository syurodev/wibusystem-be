package repository

import (
	"context"
	"system/internal/domain"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// consentRepository triển khai ConsentRepository sử dụng pgx
type consentRepository struct {
	pool *pgxpool.Pool
}

// NewConsentRepository tạo một instance mới của consentRepository
func NewConsentRepository(pool *pgxpool.Pool) domain.ConsentRepository {
	return &consentRepository{pool: pool}
}

// GetActiveConsent lấy active consent của user cho một client
func (r *consentRepository) GetActiveConsent(ctx context.Context, userID uuid.UUID, clientID uuid.UUID) (*domain.OAuth2Consent, error) {
	query := `
		SELECT id, user_id, client_id, granted_scopes, revoked,
		       granted_at, revoked_at, last_used_at, expires_at,
		       consent_method, ip_address, user_agent
		FROM identify.oauth2_consents
		WHERE user_id = $1
		  AND client_id = $2
		  AND revoked = FALSE
		  AND (expires_at IS NULL OR expires_at > NOW())
	`

	rows, err := r.pool.Query(ctx, query, userID, clientID)
	if err != nil {
		return nil, err
	}

	consent, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.OAuth2Consent])
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // No active consent found - this is not an error
		}
		return nil, err
	}

	return &consent, nil
}

// CreateConsent tạo consent mới
func (r *consentRepository) CreateConsent(ctx context.Context, consent *domain.OAuth2Consent) error {
	query := `
		INSERT INTO identify.oauth2_consents (
			id, user_id, client_id, granted_scopes, revoked,
			granted_at, expires_at, consent_method, ip_address, user_agent
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (user_id, client_id)
		DO UPDATE SET
			granted_scopes = EXCLUDED.granted_scopes,
			revoked = FALSE,
			granted_at = EXCLUDED.granted_at,
			revoked_at = NULL,
			expires_at = EXCLUDED.expires_at,
			consent_method = EXCLUDED.consent_method,
			ip_address = EXCLUDED.ip_address,
			user_agent = EXCLUDED.user_agent
	`

	_, err := r.pool.Exec(ctx, query,
		consent.ID,
		consent.UserID,
		consent.ClientID,
		consent.GrantedScopes,
		consent.Revoked,
		consent.GrantedAt,
		consent.ExpiresAt,
		consent.ConsentMethod,
		consent.IPAddress,
		consent.UserAgent,
	)

	return err
}

// UpdateConsent cập nhật consent (mainly for last_used_at)
func (r *consentRepository) UpdateConsent(ctx context.Context, consent *domain.OAuth2Consent) error {
	query := `
		UPDATE identify.oauth2_consents
		SET last_used_at = $3,
		    granted_scopes = $4
		WHERE user_id = $1 AND client_id = $2
	`

	_, err := r.pool.Exec(ctx, query,
		consent.UserID,
		consent.ClientID,
		consent.LastUsedAt,
		consent.GrantedScopes,
	)

	return err
}

// RevokeConsent thu hồi consent của user cho một client
func (r *consentRepository) RevokeConsent(ctx context.Context, userID uuid.UUID, clientID uuid.UUID) error {
	query := `
		UPDATE identify.oauth2_consents
		SET revoked = TRUE,
		    revoked_at = NOW()
		WHERE user_id = $1
		  AND client_id = $2
		  AND revoked = FALSE
	`

	_, err := r.pool.Exec(ctx, query, userID, clientID)
	return err
}

// RevokeAllUserConsents thu hồi tất cả consents của một user
func (r *consentRepository) RevokeAllUserConsents(ctx context.Context, userID uuid.UUID) (int, error) {
	query := `
		UPDATE identify.oauth2_consents
		SET revoked = TRUE,
		    revoked_at = NOW()
		WHERE user_id = $1
		  AND revoked = FALSE
	`

	result, err := r.pool.Exec(ctx, query, userID)
	if err != nil {
		return 0, err
	}

	return int(result.RowsAffected()), nil
}

// GetUserConsents lấy tất cả consents của một user
func (r *consentRepository) GetUserConsents(ctx context.Context, userID uuid.UUID, includeRevoked bool) ([]*domain.OAuth2Consent, error) {
	query := `
		SELECT id, user_id, client_id, granted_scopes, revoked,
		       granted_at, revoked_at, last_used_at, expires_at,
		       consent_method, ip_address, user_agent
		FROM identify.oauth2_consents
		WHERE user_id = $1
	`

	if !includeRevoked {
		query += " AND revoked = FALSE"
	}

	query += " ORDER BY granted_at DESC"

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}

	consents, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[domain.OAuth2Consent])
	if err != nil {
		return nil, err
	}

	return consents, nil
}
