// ============================================================================
// OAuth2 Client Repository (Ent Implementation)
// ============================================================================

package oauth2

import (
	"context"
	"database/sql"
	"system/internal/platform/database"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/ory/fosite"

	"system/internal/domain"
	ent "system/internal/ent/generated"
	"system/internal/ent/generated/oauth2client"
	pkgerrors "system/pkg/errors"
)

// oauth2ClientEntRepository triển khai OAuth2ClientRepository sử dụng Ent
type oauth2ClientEntRepository struct {
	client *ent.Client
	db     *sql.DB
}

// NewOAuth2ClientEntRepository tạo instance mới
func NewOAuth2ClientEntRepository(client *ent.Client, db *sql.DB) domain.OAuth2ClientRepository {
	return &oauth2ClientEntRepository{client: client, db: db}
}

// GetClientByID lấy thông tin client từ database (active only)
func (r *oauth2ClientEntRepository) GetClientByID(ctx context.Context, id uuid.UUID) (*domain.OAuth2Client, error) {
	c, err := database.GetClientFromContext(ctx, r.client).OAuth2Client.Query().
		Where(
			oauth2client.IDEQ(id),
			oauth2client.ActiveEQ(true),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, pkgerrors.ErrResourceNotFound
		}
		return nil, err
	}
	return entOAuth2ClientToDomain(c), nil
}

// ClientAssertionJWTValid checks if a JTI has been used
func (r *oauth2ClientEntRepository) ClientAssertionJWTValid(ctx context.Context, jti string) error {
	var exists bool
	err := r.db.QueryRowContext(ctx,
		"SELECT EXISTS(SELECT 1 FROM identify.oauth2_jti_blacklist WHERE signature=$1)",
		jti,
	).Scan(&exists)
	if err != nil {
		return err
	}
	if exists {
		return fosite.ErrJTIKnown
	}
	return nil
}

// SetClientAssertionJWT saves a JTI to blacklist
func (r *oauth2ClientEntRepository) SetClientAssertionJWT(ctx context.Context, jti string, exp time.Time) error {
	_, err := r.db.ExecContext(ctx,
		"INSERT INTO identify.oauth2_jti_blacklist (signature, expires_at) VALUES ($1, $2)",
		jti, exp,
	)
	return err
}

// GetByID lấy thông tin client (admin API - includes inactive)
func (r *oauth2ClientEntRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.OAuth2Client, error) {
	c, err := database.GetClientFromContext(ctx, r.client).OAuth2Client.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, pkgerrors.ErrResourceNotFound
		}
		return nil, err
	}
	return entOAuth2ClientToDomain(c), nil
}

// Create tạo một OAuth2 client mới
func (r *oauth2ClientEntRepository) Create(ctx context.Context, client *domain.OAuth2Client) error {
	builder := database.GetClientFromContext(ctx, r.client).OAuth2Client.Create().
		SetID(client.ID).
		SetClientName(client.ClientName).
		SetSecretHash(client.SecretHash).
		SetRedirectUris(client.RedirectURIs).
		SetGrantTypes(client.GrantTypes).
		SetResponseTypes(client.ResponseTypes).
		SetScopes(client.Scopes).
		SetIsPublic(client.IsPublic).
		SetIsInternal(client.IsInternal).
		SetTokenEndpointAuthMethod(client.TokenEndpointAuth).
		SetActive(client.Active)

	if client.OrganizationID != nil {
		builder.SetOrganizationID(*client.OrganizationID)
	}
	if client.OwnerUserID != nil {
		builder.SetOwnerUserID(*client.OwnerUserID)
	}
	if client.ClientURI != nil {
		builder.SetClientURI(*client.ClientURI)
	}
	if client.LogoURL != nil {
		builder.SetLogoURL(*client.LogoURL)
	}
	if client.TermsOfServiceURL != nil {
		builder.SetTermsOfServiceURL(*client.TermsOfServiceURL)
	}
	if client.PolicyURL != nil {
		builder.SetPolicyURL(*client.PolicyURL)
	}

	created, err := builder.Save(ctx)
	if err != nil {
		return err
	}
	client.CreatedAt = created.CreatedAt
	client.UpdatedAt = created.UpdatedAt
	return nil
}

// Update cập nhật thông tin OAuth2 client
func (r *oauth2ClientEntRepository) Update(ctx context.Context, client *domain.OAuth2Client) error {
	builder := database.GetClientFromContext(ctx, r.client).OAuth2Client.UpdateOneID(client.ID).
		SetClientName(client.ClientName).
		SetSecretHash(client.SecretHash).
		SetRedirectUris(client.RedirectURIs).
		SetGrantTypes(client.GrantTypes).
		SetResponseTypes(client.ResponseTypes).
		SetScopes(client.Scopes).
		SetIsPublic(client.IsPublic).
		SetIsInternal(client.IsInternal).
		SetTokenEndpointAuthMethod(client.TokenEndpointAuth).
		SetActive(client.Active)

	if client.OrganizationID != nil {
		builder.SetOrganizationID(*client.OrganizationID)
	} else {
		builder.ClearOrganizationID()
	}
	if client.OwnerUserID != nil {
		builder.SetOwnerUserID(*client.OwnerUserID)
	} else {
		builder.ClearOwnerUserID()
	}
	if client.ClientURI != nil {
		builder.SetClientURI(*client.ClientURI)
	} else {
		builder.ClearClientURI()
	}
	if client.LogoURL != nil {
		builder.SetLogoURL(*client.LogoURL)
	} else {
		builder.ClearLogoURL()
	}
	if client.TermsOfServiceURL != nil {
		builder.SetTermsOfServiceURL(*client.TermsOfServiceURL)
	} else {
		builder.ClearTermsOfServiceURL()
	}
	if client.PolicyURL != nil {
		builder.SetPolicyURL(*client.PolicyURL)
	} else {
		builder.ClearPolicyURL()
	}

	_, err := builder.Save(ctx)
	return err
}

// Delete xóa một OAuth2 client
func (r *oauth2ClientEntRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return database.GetClientFromContext(ctx, r.client).OAuth2Client.DeleteOneID(id).Exec(ctx)
}

// List lấy danh sách OAuth2 clients với filter và pagination
func (r *oauth2ClientEntRepository) List(ctx context.Context, organizationID *uuid.UUID, active *bool, limit, offset int) ([]*domain.OAuth2Client, int, error) {
	query := database.GetClientFromContext(ctx, r.client).OAuth2Client.Query()

	if organizationID != nil {
		query = query.Where(oauth2client.OrganizationIDEQ(*organizationID))
	}
	if active != nil {
		query = query.Where(oauth2client.ActiveEQ(*active))
	}

	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	clients, err := query.
		Order(ent.Desc(oauth2client.FieldCreatedAt)).
		Limit(limit).
		Offset(offset).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}

	results := make([]*domain.OAuth2Client, len(clients))
	for i, c := range clients {
		results[i] = entOAuth2ClientToDomain(c)
	}
	return results, total, nil
}

func entOAuth2ClientToDomain(c *ent.OAuth2Client) *domain.OAuth2Client {
	return &domain.OAuth2Client{
		ID:                c.ID,
		ClientName:        c.ClientName,
		SecretHash:        c.SecretHash,
		RedirectURIs:      c.RedirectUris,
		GrantTypes:        c.GrantTypes,
		ResponseTypes:     c.ResponseTypes,
		Scopes:            c.Scopes,
		IsPublic:          c.IsPublic,
		IsInternal:        c.IsInternal,
		TokenEndpointAuth: c.TokenEndpointAuthMethod,
		OrganizationID:    c.OrganizationID,
		OwnerUserID:       c.OwnerUserID,
		ClientURI:         c.ClientURI,
		LogoURL:           c.LogoURL,
		TermsOfServiceURL: c.TermsOfServiceURL,
		PolicyURL:         c.PolicyURL,
		Active:            c.Active,
		CreatedAt:         c.CreatedAt,
		UpdatedAt:         c.UpdatedAt,
	}
}
