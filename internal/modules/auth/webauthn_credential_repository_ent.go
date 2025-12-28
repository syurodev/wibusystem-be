// ============================================================================
// WebAuthn Credential Repository (Ent Implementation)
// ============================================================================

package auth

import (
	"context"
	"system/internal/platform/database"
	"time"

	"github.com/gofrs/uuid/v5"

	"system/internal/domain"
	ent "system/internal/ent/generated"
	"system/internal/ent/generated/webauthncredential"
)

// webauthnCredentialEntRepository triển khai WebAuthnCredentialRepository sử dụng Ent
type webauthnCredentialEntRepository struct {
	client *ent.Client
}

// NewWebAuthnCredentialEntRepository tạo một instance mới của webauthnCredentialEntRepository
func NewWebAuthnCredentialEntRepository(client *ent.Client) domain.WebAuthnCredentialRepository {
	return &webauthnCredentialEntRepository{client: client}
}

// GetByID lấy credential từ database theo ID
func (r *webauthnCredentialEntRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.WebAuthnCredential, error) {
	cred, err := database.GetClientFromContext(ctx, r.client).WebAuthnCredential.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return entWebAuthnCredentialToDomain(cred), nil
}

// GetByCredentialID lấy credential từ database theo credential ID
func (r *webauthnCredentialEntRepository) GetByCredentialID(ctx context.Context, credentialID string) (*domain.WebAuthnCredential, error) {
	cred, err := database.GetClientFromContext(ctx, r.client).WebAuthnCredential.Query().
		Where(webauthncredential.CredentialIDEQ(credentialID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return entWebAuthnCredentialToDomain(cred), nil
}

// GetByUserID lấy tất cả credentials của một user
func (r *webauthnCredentialEntRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.WebAuthnCredential, error) {
	creds, err := database.GetClientFromContext(ctx, r.client).WebAuthnCredential.Query().
		Where(webauthncredential.UserIDEQ(userID)).
		Order(ent.Desc(webauthncredential.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*domain.WebAuthnCredential, len(creds))
	for i, cred := range creds {
		result[i] = entWebAuthnCredentialToDomain(cred)
	}
	return result, nil
}

// Create tạo credential mới trong database
func (r *webauthnCredentialEntRepository) Create(ctx context.Context, credential *domain.WebAuthnCredential) error {
	builder := database.GetClientFromContext(ctx, r.client).WebAuthnCredential.Create().
		SetID(credential.ID).
		SetUserID(credential.UserID).
		SetCredentialID(credential.CredentialID).
		SetPublicKey(credential.PublicKey).
		SetAttestationType(webauthncredential.AttestationType(credential.AttestationType)).
		SetSignCount(credential.SignCount).
		SetBackupEligible(credential.BackupEligible).
		SetBackupState(credential.BackupState)

	if credential.AAGUID != nil {
		builder.SetAaguid(credential.AAGUID)
	}
	if credential.Transports != nil {
		builder.SetTransports(credential.Transports)
	}
	if credential.CredentialName != nil {
		builder.SetCredentialName(*credential.CredentialName)
	}

	_, err := builder.Save(ctx)
	return err
}

// Update cập nhật thông tin credential
func (r *webauthnCredentialEntRepository) Update(ctx context.Context, credential *domain.WebAuthnCredential) error {
	builder := database.GetClientFromContext(ctx, r.client).WebAuthnCredential.UpdateOneID(credential.ID).
		SetSignCount(credential.SignCount).
		SetBackupEligible(credential.BackupEligible).
		SetBackupState(credential.BackupState)

	if credential.Transports != nil {
		builder.SetTransports(credential.Transports)
	}
	if credential.CredentialName != nil {
		builder.SetCredentialName(*credential.CredentialName)
	} else {
		builder.ClearCredentialName()
	}
	if credential.LastUsedAt != nil {
		builder.SetLastUsedAt(*credential.LastUsedAt)
	}

	_, err := builder.Save(ctx)
	return err
}

// Delete xóa credential khỏi database
func (r *webauthnCredentialEntRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return database.GetClientFromContext(ctx, r.client).WebAuthnCredential.DeleteOneID(id).Exec(ctx)
}

// UpdateSignCount cập nhật sign count và last used time
func (r *webauthnCredentialEntRepository) UpdateSignCount(ctx context.Context, credentialID string, signCount int32) error {
	_, err := database.GetClientFromContext(ctx, r.client).WebAuthnCredential.Update().
		Where(webauthncredential.CredentialIDEQ(credentialID)).
		SetSignCount(signCount).
		SetLastUsedAt(time.Now()).
		Save(ctx)
	return err
}

// Helper function
func entWebAuthnCredentialToDomain(c *ent.WebAuthnCredential) *domain.WebAuthnCredential {
	return &domain.WebAuthnCredential{
		ID:              c.ID,
		UserID:          c.UserID,
		CredentialID:    c.CredentialID,
		PublicKey:       c.PublicKey,
		AttestationType: domain.AttestationType(c.AttestationType),
		AAGUID:          c.Aaguid,
		SignCount:       c.SignCount,
		Transports:      c.Transports,
		BackupEligible:  c.BackupEligible,
		BackupState:     c.BackupState,
		CredentialName:  c.CredentialName,
		CreatedAt:       c.CreatedAt,
		UpdatedAt:       c.UpdatedAt,
		LastUsedAt:      c.LastUsedAt,
	}
}
