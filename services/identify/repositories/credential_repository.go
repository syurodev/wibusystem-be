package repositories

import (
	"context"
	"fmt"
	"time"

	m "wibusystem/pkg/common/model"

	"github.com/google/uuid"

	"github.com/jackc/pgx/v5/pgxpool"
)

// CredentialRepository defines operations for credentials of various types
// (password, oauth provider, WebAuthn, etc.).
type CredentialRepository interface {
	Create(ctx context.Context, credential *m.Credential) error
	GetByID(ctx context.Context, id uuid.UUID) (*m.Credential, error)
	GetByUserIDAndType(ctx context.Context, userID uuid.UUID, authType m.AuthType) (*m.Credential, error)
	GetByProviderAndIdentifier(ctx context.Context, provider, identifier string) (*m.Credential, error)
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]*m.Credential, error)
	Update(ctx context.Context, credential *m.Credential) error
	Delete(ctx context.Context, id uuid.UUID) error
	UpdateLastUsed(ctx context.Context, id uuid.UUID) error
}

type credentialRepository struct {
	pool *pgxpool.Pool
}

// NewCredentialRepository creates a Postgres-backed CredentialRepository.
func NewCredentialRepository(pool *pgxpool.Pool) CredentialRepository {
	return &credentialRepository{pool: pool}
}

// Create inserts a new credential for a user.
func (r *credentialRepository) Create(ctx context.Context, credential *m.Credential) error {
	query := `
		INSERT INTO credentials (
			user_id, type, provider, identifier, secret_hash,
			public_key, sign_count, attestation_aaguid
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at
	`

	err := r.pool.QueryRow(
		ctx, query,
		credential.UserID,
		credential.Type,
		credential.Provider,
		credential.Identifier,
		credential.SecretHash,
		credential.PublicKey,
		credential.SignCount,
		credential.AttestationAAGUID,
	).Scan(
		&credential.ID,
		&credential.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create credential: %w", err)
	}

	return nil
}

// GetByID returns credential by ID.
func (r *credentialRepository) GetByID(ctx context.Context, id uuid.UUID) (*m.Credential, error) {
	query := `
		SELECT id, user_id, type, provider, identifier, secret_hash,
			   public_key, sign_count, attestation_aaguid, created_at, last_used_at
		FROM credentials
		WHERE id = $1
	`

	credential := &m.Credential{}
	var lastUsedAt *time.Time

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&credential.ID,
		&credential.UserID,
		&credential.Type,
		&credential.Provider,
		&credential.Identifier,
		&credential.SecretHash,
		&credential.PublicKey,
		&credential.SignCount,
		&credential.AttestationAAGUID,
		&credential.CreatedAt,
		&lastUsedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get credential by ID: %w", err)
	}

	if lastUsedAt != nil {
		credential.LastUsedAt = lastUsedAt
	}

	return credential, nil
}

// GetByUserIDAndType returns the most recent credential for a user and auth type.
func (r *credentialRepository) GetByUserIDAndType(ctx context.Context, userID uuid.UUID, authType m.AuthType) (*m.Credential, error) {
	query := `
		SELECT id, user_id, type, provider, identifier, secret_hash,
			   public_key, sign_count, attestation_aaguid, created_at, last_used_at
		FROM credentials
		WHERE user_id = $1 AND type = $2
		ORDER BY created_at DESC
		LIMIT 1
	`

	credential := &m.Credential{}
	var lastUsedAt *time.Time

	err := r.pool.QueryRow(ctx, query, userID, authType).Scan(
		&credential.ID,
		&credential.UserID,
		&credential.Type,
		&credential.Provider,
		&credential.Identifier,
		&credential.SecretHash,
		&credential.PublicKey,
		&credential.SignCount,
		&credential.AttestationAAGUID,
		&credential.CreatedAt,
		&lastUsedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get credential by user ID and type: %w", err)
	}

	if lastUsedAt != nil {
		credential.LastUsedAt = lastUsedAt
	}

	return credential, nil
}

// GetByProviderAndIdentifier finds a credential by external provider and identifier.
func (r *credentialRepository) GetByProviderAndIdentifier(ctx context.Context, provider, identifier string) (*m.Credential, error) {
	query := `
		SELECT id, user_id, type, provider, identifier, secret_hash,
			   public_key, sign_count, attestation_aaguid, created_at, last_used_at
		FROM credentials
		WHERE provider = $1 AND identifier = $2
	`

	credential := &m.Credential{}
	var lastUsedAt *time.Time

	err := r.pool.QueryRow(ctx, query, provider, identifier).Scan(
		&credential.ID,
		&credential.UserID,
		&credential.Type,
		&credential.Provider,
		&credential.Identifier,
		&credential.SecretHash,
		&credential.PublicKey,
		&credential.SignCount,
		&credential.AttestationAAGUID,
		&credential.CreatedAt,
		&lastUsedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get credential by provider and identifier: %w", err)
	}

	if lastUsedAt != nil {
		credential.LastUsedAt = lastUsedAt
	}

	return credential, nil
}

// ListByUserID returns all credentials for a user ordered by creation time.
func (r *credentialRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]*m.Credential, error) {
	query := `
		SELECT id, user_id, type, provider, identifier, secret_hash,
			   public_key, sign_count, attestation_aaguid, created_at, last_used_at
		FROM credentials
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list credentials by user ID: %w", err)
	}
	defer rows.Close()

	var credentials []*m.Credential
	for rows.Next() {
		credential := &m.Credential{}
		var lastUsedAt *time.Time

		err := rows.Scan(
			&credential.ID,
			&credential.UserID,
			&credential.Type,
			&credential.Provider,
			&credential.Identifier,
			&credential.SecretHash,
			&credential.PublicKey,
			&credential.SignCount,
			&credential.AttestationAAGUID,
			&credential.CreatedAt,
			&lastUsedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan credential: %w", err)
		}

		if lastUsedAt != nil {
			credential.LastUsedAt = lastUsedAt
		}

		credentials = append(credentials, credential)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate credentials: %w", err)
	}

	return credentials, nil
}

// Update modifies mutable fields on a credential.
func (r *credentialRepository) Update(ctx context.Context, credential *m.Credential) error {
	query := `
		UPDATE credentials
		SET provider = $2, identifier = $3, secret_hash = $4,
			public_key = $5, sign_count = $6, attestation_aaguid = $7
		WHERE id = $1
	`

	result, err := r.pool.Exec(
		ctx, query,
		credential.ID,
		credential.Provider,
		credential.Identifier,
		credential.SecretHash,
		credential.PublicKey,
		credential.SignCount,
		credential.AttestationAAGUID,
	)
	if err != nil {
		return fmt.Errorf("failed to update credential: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("credential with ID %d not found", credential.ID)
	}

	return nil
}

// Delete removes a credential by ID.
func (r *credentialRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM credentials WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete credential: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("credential with ID %d not found", id)
	}

	return nil
}

// UpdateLastUsed stamps the credential's last_used_at column with NOW().
func (r *credentialRepository) UpdateLastUsed(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE credentials SET last_used_at = NOW() WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to update last used: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("credential with ID %d not found", id)
	}

	return nil
}
