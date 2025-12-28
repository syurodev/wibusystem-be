// ============================================================================
// OAuth2 Consent Repository (Ent Implementation)
// ============================================================================

package oauth2

import (
	"context"
	"system/internal/platform/database"
	"time"

	"github.com/gofrs/uuid/v5"

	"system/internal/domain"
	ent "system/internal/ent/generated"
	"system/internal/ent/generated/consent"
)

// consentEntRepository triển khai ConsentRepository sử dụng Ent
type consentEntRepository struct {
	client *ent.Client
}

// NewConsentEntRepository tạo instance mới
func NewConsentEntRepository(client *ent.Client) domain.ConsentRepository {
	return &consentEntRepository{client: client}
}

// GetActiveConsent lấy active consent của user cho một client
func (r *consentEntRepository) GetActiveConsent(ctx context.Context, userID uuid.UUID, clientID uuid.UUID) (*domain.OAuth2Consent, error) {
	c, err := database.GetClientFromContext(ctx, r.client).Consent.Query().
		Where(
			consent.UserIDEQ(userID),
			consent.ClientIDEQ(clientID),
			consent.RevokedEQ(false),
			consent.Or(
				consent.ExpiresAtIsNil(),
				consent.ExpiresAtGT(time.Now()),
			),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return entConsentToDomain(c), nil
}

// CreateConsent tạo consent mới (upsert)
func (r *consentEntRepository) CreateConsent(ctx context.Context, consentData *domain.OAuth2Consent) error {
	// Check existing
	existing, err := database.GetClientFromContext(ctx, r.client).Consent.Query().
		Where(
			consent.UserIDEQ(consentData.UserID),
			consent.ClientIDEQ(consentData.ClientID),
		).
		Only(ctx)

	if err != nil && !ent.IsNotFound(err) {
		return err
	}

	if existing != nil {
		// Update existing
		builder := database.GetClientFromContext(ctx, r.client).Consent.UpdateOne(existing).
			SetGrantedScopes(consentData.GrantedScopes).
			SetRevoked(false).
			SetGrantedAt(consentData.GrantedAt).
			SetConsentMethod(consent.ConsentMethod(consentData.ConsentMethod)).
			ClearRevokedAt()

		if consentData.ExpiresAt != nil {
			builder.SetExpiresAt(*consentData.ExpiresAt)
		}
		if consentData.IPAddress != nil {
			builder.SetIPAddress(*consentData.IPAddress)
		}
		if consentData.UserAgent != nil {
			builder.SetUserAgent(*consentData.UserAgent)
		}

		_, err = builder.Save(ctx)
		return err
	}

	// Create new
	builder := database.GetClientFromContext(ctx, r.client).Consent.Create().
		SetID(consentData.ID).
		SetUserID(consentData.UserID).
		SetClientID(consentData.ClientID).
		SetGrantedScopes(consentData.GrantedScopes).
		SetRevoked(false).
		SetGrantedAt(consentData.GrantedAt).
		SetConsentMethod(consent.ConsentMethod(consentData.ConsentMethod))

	if consentData.ExpiresAt != nil {
		builder.SetExpiresAt(*consentData.ExpiresAt)
	}
	if consentData.IPAddress != nil {
		builder.SetIPAddress(*consentData.IPAddress)
	}
	if consentData.UserAgent != nil {
		builder.SetUserAgent(*consentData.UserAgent)
	}

	_, err = builder.Save(ctx)
	return err
}

// UpdateConsent cập nhật consent
func (r *consentEntRepository) UpdateConsent(ctx context.Context, consentData *domain.OAuth2Consent) error {
	builder := database.GetClientFromContext(ctx, r.client).Consent.Update().
		Where(
			consent.UserIDEQ(consentData.UserID),
			consent.ClientIDEQ(consentData.ClientID),
		).
		SetGrantedScopes(consentData.GrantedScopes)

	if consentData.LastUsedAt != nil {
		builder.SetLastUsedAt(*consentData.LastUsedAt)
	}

	_, err := builder.Save(ctx)
	return err
}

// RevokeConsent thu hồi consent của user cho một client
func (r *consentEntRepository) RevokeConsent(ctx context.Context, userID uuid.UUID, clientID uuid.UUID) error {
	now := time.Now()
	_, err := database.GetClientFromContext(ctx, r.client).Consent.Update().
		Where(
			consent.UserIDEQ(userID),
			consent.ClientIDEQ(clientID),
			consent.RevokedEQ(false),
		).
		SetRevoked(true).
		SetRevokedAt(now).
		Save(ctx)
	return err
}

// RevokeAllUserConsents thu hồi tất cả consents của một user
func (r *consentEntRepository) RevokeAllUserConsents(ctx context.Context, userID uuid.UUID) (int, error) {
	now := time.Now()
	count, err := database.GetClientFromContext(ctx, r.client).Consent.Update().
		Where(
			consent.UserIDEQ(userID),
			consent.RevokedEQ(false),
		).
		SetRevoked(true).
		SetRevokedAt(now).
		Save(ctx)
	return count, err
}

// GetUserConsents lấy tất cả consents của một user
func (r *consentEntRepository) GetUserConsents(ctx context.Context, userID uuid.UUID, includeRevoked bool) ([]*domain.OAuth2Consent, error) {
	query := database.GetClientFromContext(ctx, r.client).Consent.Query().
		Where(consent.UserIDEQ(userID)).
		Order(ent.Desc(consent.FieldGrantedAt))

	if !includeRevoked {
		query = query.Where(consent.RevokedEQ(false))
	}

	consents, err := query.All(ctx)
	if err != nil {
		return nil, err
	}

	results := make([]*domain.OAuth2Consent, len(consents))
	for i, c := range consents {
		results[i] = entConsentToDomain(c)
	}
	return results, nil
}

func entConsentToDomain(c *ent.Consent) *domain.OAuth2Consent {
	return &domain.OAuth2Consent{
		ID:            c.ID,
		UserID:        c.UserID,
		ClientID:      c.ClientID,
		GrantedScopes: c.GrantedScopes,
		Revoked:       c.Revoked,
		GrantedAt:     c.GrantedAt,
		RevokedAt:     c.RevokedAt,
		LastUsedAt:    c.LastUsedAt,
		ExpiresAt:     c.ExpiresAt,
		ConsentMethod: domain.ConsentMethod(c.ConsentMethod),
		IPAddress:     c.IPAddress,
		UserAgent:     c.UserAgent,
	}
}
