package repository

import (
	"context"
	"strconv"
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
	query := `
		SELECT id, client_name, secret_hash, redirect_uris, grant_types, response_types,
		       scopes, is_public, is_internal, token_endpoint_auth_method, tenant_id,
		       client_uri, logo_url, created_at, updated_at
		FROM identify.oauth2_clients
		WHERE id=$1
	`

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

// GetByID lấy thông tin client từ database bằng ID (alias for GetClientByID for admin API).
func (r *oauth2ClientRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.OAuth2Client, error) {
	query := `
		SELECT id, client_name, secret_hash, redirect_uris, grant_types, response_types,
		       scopes, is_public, is_internal, token_endpoint_auth_method, tenant_id, client_uri,
		       logo_url, active, created_at, updated_at
		FROM identify.oauth2_clients
		WHERE id = $1
	`

	rows, err := r.pool.Query(ctx, query, id)
	if err != nil {
		return nil, err
	}

	client, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.OAuth2Client])
	if err != nil {
		return nil, err
	}

	return &client, nil
}

// Create tạo một OAuth2 client mới.
func (r *oauth2ClientRepository) Create(ctx context.Context, client *domain.OAuth2Client) error {
	query := `
		INSERT INTO identify.oauth2_clients (
			id, client_name, secret_hash, redirect_uris, grant_types, response_types,
			scopes, is_public, is_internal, token_endpoint_auth_method, tenant_id, client_uri,
			logo_url, active, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	`

	_, err := r.pool.Exec(ctx, query,
		client.ID,
		client.ClientName,
		client.SecretHash,
		client.RedirectURIs,
		client.GrantTypes,
		client.ResponseTypes,
		client.Scopes,
		client.IsPublic,
		client.IsInternal,
		client.TokenEndpointAuth,
		client.TenantID,
		client.ClientURI,
		client.LogoURL,
		client.Active,
		client.CreatedAt,
		client.UpdatedAt,
	)

	return err
}

// Update cập nhật thông tin OAuth2 client.
func (r *oauth2ClientRepository) Update(ctx context.Context, client *domain.OAuth2Client) error {
	query := `
		UPDATE identify.oauth2_clients
		SET client_name = $2,
		    secret_hash = $3,
		    redirect_uris = $4,
		    grant_types = $5,
		    response_types = $6,
		    scopes = $7,
		    is_public = $8,
		    is_internal = $9,
		    token_endpoint_auth_method = $10,
		    tenant_id = $11,
		    client_uri = $12,
		    logo_url = $13,
		    active = $14,
		    updated_at = $15
		WHERE id = $1
	`

	_, err := r.pool.Exec(ctx, query,
		client.ID,
		client.ClientName,
		client.SecretHash,
		client.RedirectURIs,
		client.GrantTypes,
		client.ResponseTypes,
		client.Scopes,
		client.IsPublic,
		client.IsInternal,
		client.TokenEndpointAuth,
		client.TenantID,
		client.ClientURI,
		client.LogoURL,
		client.Active,
		client.UpdatedAt,
	)

	return err
}

// Delete xóa một OAuth2 client.
func (r *oauth2ClientRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM identify.oauth2_clients WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

// List lấy danh sách OAuth2 clients với filter và pagination.
func (r *oauth2ClientRepository) List(ctx context.Context, tenantID *uuid.UUID, active *bool, limit, offset int) ([]*domain.OAuth2Client, int, error) {
	// Build dynamic query with filters
	baseQuery := `
		SELECT id, client_name, secret_hash, redirect_uris, grant_types, response_types,
		       scopes, is_public, is_internal, token_endpoint_auth_method, tenant_id, client_uri,
		       logo_url, active, created_at, updated_at
		FROM identify.oauth2_clients
		WHERE 1=1
	`
	countQuery := `SELECT COUNT(*) FROM identify.oauth2_clients WHERE 1=1`

	args := []any{}
	argIndex := 1

	// Add filters
	if tenantID != nil {
		baseQuery += ` AND tenant_id = $` + strconv.Itoa(argIndex)
		countQuery += ` AND tenant_id = $` + strconv.Itoa(argIndex)
		args = append(args, *tenantID)
		argIndex++
	}

	if active != nil {
		baseQuery += ` AND active = $` + strconv.Itoa(argIndex)
		countQuery += ` AND active = $` + strconv.Itoa(argIndex)
		args = append(args, *active)
		argIndex++
	}

	// Get total count
	var total int
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Add pagination
	baseQuery += ` ORDER BY created_at DESC LIMIT $` + strconv.Itoa(argIndex) + ` OFFSET $` + strconv.Itoa(argIndex+1)
	args = append(args, limit, offset)

	// Execute query
	rows, err := r.pool.Query(ctx, baseQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	clients, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[domain.OAuth2Client])
	if err != nil {
		return nil, 0, err
	}

	return clients, total, nil
}
