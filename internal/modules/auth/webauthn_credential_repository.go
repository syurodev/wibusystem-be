package auth

import (
	"context"
	"system/internal/domain"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// webauthnCredentialRepository triển khai WebAuthnCredentialRepository sử dụng pgx
type webauthnCredentialRepository struct {
	pool *pgxpool.Pool
}

// NewWebAuthnCredentialRepository tạo một instance mới của webauthnCredentialRepository
func NewWebAuthnCredentialRepository(pool *pgxpool.Pool) domain.WebAuthnCredentialRepository {
	return &webauthnCredentialRepository{pool: pool}
}

// GetByID lấy credential từ database theo ID
func (r *webauthnCredentialRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.WebAuthnCredential, error) {
	query := `
		SELECT id, user_id, credential_id, public_key, attestation_type,
		       aaguid, sign_count, transports, backup_eligible, backup_state,
		       credential_name, created_at, updated_at, last_used_at
		FROM identify.webauthn_credentials
		WHERE id = $1
	`

	rows, err := r.pool.Query(ctx, query, id)
	if err != nil {
		return nil, err
	}

	credential, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.WebAuthnCredential])
	if err != nil {
		return nil, err
	}

	return &credential, nil
}

// GetByCredentialID lấy credential từ database theo credential ID
func (r *webauthnCredentialRepository) GetByCredentialID(ctx context.Context, credentialID string) (*domain.WebAuthnCredential, error) {
	query := `
		SELECT id, user_id, credential_id, public_key, attestation_type,
		       aaguid, sign_count, transports, backup_eligible, backup_state,
		       credential_name, created_at, updated_at, last_used_at
		FROM identify.webauthn_credentials
		WHERE credential_id = $1
	`

	rows, err := r.pool.Query(ctx, query, credentialID)
	if err != nil {
		return nil, err
	}

	credential, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.WebAuthnCredential])
	if err != nil {
		return nil, err
	}

	return &credential, nil
}

// GetByUserID lấy tất cả credentials của một user
func (r *webauthnCredentialRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.WebAuthnCredential, error) {
	query := `
		SELECT id, user_id, credential_id, public_key, attestation_type,
		       aaguid, sign_count, transports, backup_eligible, backup_state,
		       credential_name, created_at, updated_at, last_used_at
		FROM identify.webauthn_credentials
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}

	credentials, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[domain.WebAuthnCredential])
	if err != nil {
		return nil, err
	}

	return credentials, nil
}

// Create tạo credential mới trong database
func (r *webauthnCredentialRepository) Create(ctx context.Context, credential *domain.WebAuthnCredential) error {
	query := `
		INSERT INTO identify.webauthn_credentials (
			id, user_id, credential_id, public_key, attestation_type,
			aaguid, sign_count, transports, backup_eligible, backup_state,
			credential_name
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := r.pool.Exec(ctx, query,
		credential.ID,
		credential.UserID,
		credential.CredentialID,
		credential.PublicKey,
		credential.AttestationType,
		credential.AAGUID,
		credential.SignCount,
		credential.Transports,
		credential.BackupEligible,
		credential.BackupState,
		credential.CredentialName,
	)

	return err
}

// Update cập nhật thông tin credential
func (r *webauthnCredentialRepository) Update(ctx context.Context, credential *domain.WebAuthnCredential) error {
	query := `
		UPDATE identify.webauthn_credentials
		SET sign_count = $2,
		    transports = $3,
		    backup_eligible = $4,
		    backup_state = $5,
		    credential_name = $6,
		    last_used_at = $7,
		    updated_at = NOW()
		WHERE id = $1
	`

	_, err := r.pool.Exec(ctx, query,
		credential.ID,
		credential.SignCount,
		credential.Transports,
		credential.BackupEligible,
		credential.BackupState,
		credential.CredentialName,
		credential.LastUsedAt,
	)

	return err
}

// Delete xóa credential khỏi database
func (r *webauthnCredentialRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		DELETE FROM identify.webauthn_credentials
		WHERE id = $1
	`

	_, err := r.pool.Exec(ctx, query, id)
	return err
}

// UpdateSignCount cập nhật sign count và last used time
func (r *webauthnCredentialRepository) UpdateSignCount(ctx context.Context, credentialID string, signCount int32) error {
	query := `
		UPDATE identify.webauthn_credentials
		SET sign_count = $2,
		    last_used_at = NOW(),
		    updated_at = NOW()
		WHERE credential_id = $1
	`

	_, err := r.pool.Exec(ctx, query, credentialID, signCount)
	return err
}
