// ============================================================================
// WebAuthn Session Repository (Ent Implementation)
// ============================================================================

package auth

import (
	"context"
	"system/internal/platform/database"
	"time"

	"github.com/gofrs/uuid/v5"

	"system/internal/domain"
	ent "system/internal/ent/generated"
	"system/internal/ent/generated/webauthnsession"
)

// webauthnSessionEntRepository triển khai WebAuthnSessionRepository sử dụng Ent
type webauthnSessionEntRepository struct {
	client *ent.Client
}

// NewWebAuthnSessionEntRepository tạo một instance mới của webauthnSessionEntRepository
func NewWebAuthnSessionEntRepository(client *ent.Client) domain.WebAuthnSessionRepository {
	return &webauthnSessionEntRepository{client: client}
}

// GetByChallenge lấy session từ database theo challenge
func (r *webauthnSessionEntRepository) GetByChallenge(ctx context.Context, challenge string) (*domain.WebAuthnSession, error) {
	sess, err := database.GetClientFromContext(ctx, r.client).WebAuthnSession.Query().
		Where(
			webauthnsession.ChallengeEQ(challenge),
			webauthnsession.ExpiresAtGT(time.Now()),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return entWebAuthnSessionToDomain(sess), nil
}

// Create tạo session mới trong database
func (r *webauthnSessionEntRepository) Create(ctx context.Context, session *domain.WebAuthnSession) error {
	builder := database.GetClientFromContext(ctx, r.client).WebAuthnSession.Create().
		SetID(session.ID).
		SetChallenge(session.Challenge).
		SetSessionType(webauthnsession.SessionType(session.SessionType)).
		SetExpiresAt(session.ExpiresAt)

	if session.UserID != nil {
		builder.SetUserID(*session.UserID)
	}
	if session.UserAgent != nil {
		builder.SetUserAgent(*session.UserAgent)
	}
	if session.IPAddress != nil {
		builder.SetIPAddress(*session.IPAddress)
	}

	_, err := builder.Save(ctx)
	return err
}

// Delete xóa session khỏi database
func (r *webauthnSessionEntRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return database.GetClientFromContext(ctx, r.client).WebAuthnSession.DeleteOneID(id).Exec(ctx)
}

// DeleteExpired xóa các sessions đã hết hạn
func (r *webauthnSessionEntRepository) DeleteExpired(ctx context.Context) error {
	_, err := database.GetClientFromContext(ctx, r.client).WebAuthnSession.Delete().
		Where(webauthnsession.ExpiresAtLTE(time.Now())).
		Exec(ctx)
	return err
}

// Helper function
func entWebAuthnSessionToDomain(s *ent.WebAuthnSession) *domain.WebAuthnSession {
	return &domain.WebAuthnSession{
		ID:          s.ID,
		UserID:      s.UserID,
		Challenge:   s.Challenge,
		SessionType: domain.SessionType(s.SessionType),
		UserAgent:   s.UserAgent,
		IPAddress:   s.IPAddress,
		CreatedAt:   s.CreatedAt,
		ExpiresAt:   s.ExpiresAt,
	}
}
