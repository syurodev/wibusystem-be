package oauth2

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ory/fosite"
	"github.com/ory/fosite/handler/oauth2"
	"github.com/ory/fosite/handler/openid"
	"golang.org/x/crypto/bcrypt"
)

// Client decorates fosite.DefaultClient with internal metadata so we can
// distinguish trusted first-party applications without losing fosite behavior.
type Client struct {
	*fosite.DefaultClient
	Internal bool
}

// IsInternal reports whether this OAuth client is considered an internal
// first-party application.
func (c *Client) IsInternal() bool {
	if c == nil {
		return false
	}
	return c.Internal
}

// Store implements all Fosite storage interfaces
type Store struct {
	pool                *pgxpool.Pool
	clients             map[string]*Client
	mu                  sync.RWMutex
	publicKey           *rsa.PublicKey
	privateKey          *rsa.PrivateKey
	accessTokenLifespan time.Duration
}

// NewStore creates a new OAuth2 store
func NewStore(pool *pgxpool.Pool, privateKey *rsa.PrivateKey, accessTokenLifespan time.Duration) *Store {
	store := &Store{
		pool:                pool,
		clients:             make(map[string]*Client),
		privateKey:          privateKey,
		publicKey:           &privateKey.PublicKey,
		accessTokenLifespan: accessTokenLifespan,
	}

	// Health check database connection
	if err := store.healthCheck(); err != nil {
		log.Printf("oauth2.store: WARNING - Database health check failed: %v", err)
	} else {
		log.Printf("oauth2.store: Database connection healthy")
	}

	// Load clients from database
	store.loadClientsFromDatabase()

	return store
}

// healthCheck verifies the database connection is working
func (s *Store) healthCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Simple ping to verify connection
	if err := s.pool.Ping(ctx); err != nil {
		return fmt.Errorf("database ping failed: %v", err)
	}

	// Test basic query
	var count int
	query := "SELECT 1"
	if err := s.pool.QueryRow(ctx, query).Scan(&count); err != nil {
		return fmt.Errorf("database test query failed: %v", err)
	}

	// Check if oauth2 tables exist
	tableCheck := `SELECT COUNT(*) FROM information_schema.tables
                   WHERE table_name IN ('oauth2_authorization_codes', 'oauth2_clients')`
	var tableCount int
	if err := s.pool.QueryRow(ctx, tableCheck).Scan(&tableCount); err != nil {
		return fmt.Errorf("table existence check failed: %v", err)
	}

	if tableCount < 2 {
		return fmt.Errorf("required OAuth2 tables not found (found %d/2)", tableCount)
	}

	log.Printf("oauth2.store: Health check passed - all tables present")
	return nil
}

// helpers to avoid inserting NULL into NOT NULL TEXT[] columns
func toStringSlice(args fosite.Arguments) []string {
	if len(args) == 0 {
		return []string{}
	}
	return []string(args)
}

// toURLValues parses a JSON-serialized url.Values back to url.Values
func toURLValues(formJSON string) url.Values {
	if formJSON == "" {
		return url.Values{}
	}
	var m map[string][]string
	if err := json.Unmarshal([]byte(formJSON), &m); err != nil {
		return url.Values{}
	}
	v := url.Values{}
	for k, vals := range m {
		for _, s := range vals {
			v.Add(k, s)
		}
	}
	return v
}

// decodeOIDCSession reconstructs an openid.DefaultSession from stored JSON
func decodeOIDCSession(sessionJSON string) *openid.DefaultSession {
	if sessionJSON == "" {
		return &openid.DefaultSession{}
	}
	var s openid.DefaultSession
	if err := json.Unmarshal([]byte(sessionJSON), &s); err != nil {
		// Fallback to empty session to avoid panics
		return &openid.DefaultSession{}
	}
	return &s
}

func (s *Store) loadClientsFromDatabase() {
	ctx := context.Background()
	query := `
        SELECT id, client_secret_hash, redirect_uris, grant_types, response_types,
               scopes, audience, public, client_name, internal
        FROM oauth2_clients
    `

	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		log.Printf("oauth2.store: failed to query clients: %v", err)
		return
	}
	defer rows.Close()

	loaded := 0
	for rows.Next() {
		client, err := s.scanClient(rows)
		if err != nil {
			log.Printf("oauth2.store: scan client row failed: %v", err)
			continue
		}

		s.setClient(client)
		loaded++
		log.Printf("oauth2.store: loaded client id=%s public=%t grants=%v", client.ID, client.Public, client.GrantTypes)
	}

	log.Printf("oauth2.store: total clients loaded: %d", loaded)
}

// scanClient builds an internal-aware client from a query row
func (s *Store) scanClient(row interface{ Scan(...any) error }) (*Client, error) {
	var (
		id            string
		secret        pgtype.Text
		redirectURIs  []string
		grantTypes    []string
		responseTypes []string
		scopes        []string
		audience      []string
		public        bool
		clientName    pgtype.Text // currently unused, reserved for future metadata
		internal      bool
	)

	if err := row.Scan(
		&id,
		&secret,
		&redirectURIs,
		&grantTypes,
		&responseTypes,
		&scopes,
		&audience,
		&public,
		&clientName,
		&internal,
	); err != nil {
		return nil, err
	}

	defaultClient := &fosite.DefaultClient{
		ID:            id,
		RedirectURIs:  redirectURIs,
		GrantTypes:    grantTypes,
		ResponseTypes: responseTypes,
		Scopes:        scopes,
		Audience:      audience,
		Public:        public,
	}

	if secret.Valid {
		defaultClient.Secret = []byte(secret.String)
	}

	return &Client{DefaultClient: defaultClient, Internal: internal}, nil
}

func (s *Store) setClient(client *Client) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clients[client.ID] = client
}

func (s *Store) deleteClient(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.clients, id)
}

// ReloadClient refreshes a single client from the database and updates the in-memory cache.
func (s *Store) ReloadClient(ctx context.Context, id string) error {
	if s.pool == nil {
		return fmt.Errorf("database pool unavailable for reloading client")
	}

	if ctx == nil {
		ctx = context.Background()
	}

	query := `
		SELECT id, client_secret_hash, redirect_uris, grant_types, response_types,
		       scopes, audience, public, client_name, internal
		FROM oauth2_clients
		WHERE id = $1
	`

	row := s.pool.QueryRow(ctx, query, id)
	client, err := s.scanClient(row)
	if err != nil {
		return err
	}

	s.setClient(client)
	return nil
}

// DeleteClient removes a client from the in-memory cache.
func (s *Store) DeleteClient(id string) {
	s.deleteClient(id)
}

// Client management methods

func (s *Store) GetClient(ctx context.Context, id string) (fosite.Client, error) {
	s.mu.RLock()
	client, ok := s.clients[id]
	s.mu.RUnlock()
	if ok {
		return client, nil
	}

	log.Printf("oauth2.store: GetClient miss for id=%s (cached=%d) - attempting reload", id, len(s.clients))
	if err := s.ReloadClient(ctx, id); err == nil {
		s.mu.RLock()
		client = s.clients[id]
		s.mu.RUnlock()
		if client != nil {
			return client, nil
		}
	}

	return nil, fosite.ErrNotFound
}

// Authorization Code methods

func (s *Store) CreateAuthorizeCodeSession(ctx context.Context, code string, req fosite.Requester) error {
	start := time.Now()
	fmt.Printf("DEBUG STORE: CreateAuthorizeCodeSession called\n")
	fmt.Printf("DEBUG STORE: Code: %s\n", code)
	fmt.Printf("DEBUG STORE: Request ID: %s\n", req.GetID())
	fmt.Printf("DEBUG STORE: Client ID: %s\n", req.GetClient().GetID())
	fmt.Printf("DEBUG STORE: Subject: %s\n", req.GetSession().GetSubject())

	defer func() {
		duration := time.Since(start)
		fmt.Printf("DEBUG STORE: CreateAuthorizeCodeSession took %v\n", duration)
		if duration > 2*time.Second {
			fmt.Printf("DEBUG STORE: WARNING - Slow operation detected: %v\n", duration)
		}
	}()

	// Create a new context with timeout for database operations
	dbCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	sessionData, sessionErr := json.Marshal(req.GetSession())
	if sessionErr != nil {
		fmt.Printf("DEBUG STORE: Failed to marshal session data: %v\n", sessionErr)
		return sessionErr
	}

	formData, formErr := json.Marshal(req.GetRequestForm())
	if formErr != nil {
		fmt.Printf("DEBUG STORE: Failed to marshal form data: %v\n", formErr)
		return formErr
	}

	query := `
        INSERT INTO oauth2_authorization_codes (
            signature, request_id, requested_at, client_id, scopes, granted_scopes,
            form_data, session_data, subject, requested_audience, granted_audience
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
    `

	// Log connection pool stats before operation
	stats := s.pool.Stat()
	fmt.Printf("DEBUG STORE: Pool stats - Total: %d, Idle: %d, Used: %d\n",
		stats.TotalConns(), stats.IdleConns(), stats.TotalConns()-stats.IdleConns())

	// Use the new database context with timeout instead of the original context
	execStart := time.Now()
	_, err := s.pool.Exec(dbCtx, query,
		code,
		req.GetID(),
		req.GetRequestedAt(),
		req.GetClient().GetID(),
		toStringSlice(req.GetRequestedScopes()),
		toStringSlice(req.GetGrantedScopes()),
		string(formData),
		string(sessionData),
		req.GetSession().GetSubject(),
		toStringSlice(req.GetRequestedAudience()),
		toStringSlice(req.GetGrantedAudience()),
	)
	execDuration := time.Since(execStart)
	fmt.Printf("DEBUG STORE: Database exec took %v\n", execDuration)

	if err != nil {
		fmt.Printf("DEBUG STORE: Failed to insert authorization code: %v\n", err)
		fmt.Printf("DEBUG STORE: Error type: %T\n", err)
		// Add retry logic for transient errors
		if isRetryableError(err) {
			fmt.Printf("DEBUG STORE: Retrying with new context...\n")
			retryCtx, retryCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer retryCancel()

			_, retryErr := s.pool.Exec(retryCtx, query,
				code, req.GetID(), req.GetRequestedAt(), req.GetClient().GetID(),
				toStringSlice(req.GetRequestedScopes()), toStringSlice(req.GetGrantedScopes()),
				string(formData), string(sessionData), req.GetSession().GetSubject(),
				toStringSlice(req.GetRequestedAudience()), toStringSlice(req.GetGrantedAudience()),
			)
			if retryErr == nil {
				fmt.Printf("DEBUG STORE: Authorization code inserted successfully on retry\n")
				return nil
			} else {
				fmt.Printf("DEBUG STORE: Retry also failed: %v\n", retryErr)
			}
		}
	} else {
		fmt.Printf("DEBUG STORE: Authorization code inserted successfully\n")
	}

	return err
}

// isRetryableError checks if the error is retryable (e.g., connection issues, timeouts)
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "context canceled") ||
		strings.Contains(errStr, "context deadline exceeded") ||
		strings.Contains(errStr, "connection") ||
		strings.Contains(errStr, "timeout")
}

func (s *Store) GetAuthorizeCodeSession(ctx context.Context, code string, session fosite.Session) (fosite.Requester, error) {
	query := `
		SELECT request_id, requested_at, client_id, scopes, granted_scopes,
			   form_data, session_data, subject, requested_audience, granted_audience
		FROM oauth2_authorization_codes
		WHERE signature = $1 AND active = true
	`

	var requestID string
	var requestedAt time.Time
	var clientID string
	var scopes, grantedScopes []string
	var formData, sessionData string
	var subject string
	var requestedAudience, grantedAudience []string

	err := s.pool.QueryRow(ctx, query, code).Scan(
		&requestID, &requestedAt, &clientID, &scopes, &grantedScopes,
		&formData, &sessionData, &subject, &requestedAudience, &grantedAudience,
	)
	if err != nil {
		return nil, fosite.ErrNotFound
	}

	client, err := s.GetClient(ctx, clientID)
	if err != nil {
		return nil, err
	}

	req := fosite.NewRequest()
	req.ID = requestID
	req.RequestedAt = requestedAt
	req.Client = client
	req.RequestedScope = []string(scopes)
	req.GrantedScope = []string(grantedScopes)
	req.RequestedAudience = []string(requestedAudience)
	req.GrantedAudience = []string(grantedAudience)
	// Replace with stored session for OIDC (contains claims)
	req.Session = decodeOIDCSession(sessionData)
	req.Form = toURLValues(formData)

	return req, nil
}

func (s *Store) InvalidateAuthorizeCodeSession(ctx context.Context, code string) error {
	query := `UPDATE oauth2_authorization_codes SET active = false WHERE signature = $1`
	_, err := s.pool.Exec(ctx, query, code)
	return err
}

// Access Token methods

func (s *Store) CreateAccessTokenSession(ctx context.Context, signature string, req fosite.Requester) error {
	sessionData, _ := json.Marshal(req.GetSession())
	formData, _ := json.Marshal(req.GetRequestForm())
	expiresAt := req.GetRequestedAt().Add(s.accessTokenLifespan)

	query := `
		INSERT INTO oauth2_access_tokens (
			signature, request_id, requested_at, client_id, scopes, granted_scopes,
			form_data, session_data, subject, requested_audience, granted_audience, expires_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	_, err := s.pool.Exec(ctx, query,
		signature,
		req.GetID(),
		req.GetRequestedAt(),
		req.GetClient().GetID(),
		toStringSlice(req.GetRequestedScopes()),
		toStringSlice(req.GetGrantedScopes()),
		string(formData),
		string(sessionData),
		req.GetSession().GetSubject(),
		toStringSlice(req.GetRequestedAudience()),
		toStringSlice(req.GetGrantedAudience()),
		expiresAt,
	)

	return err
}

func (s *Store) GetAccessTokenSession(ctx context.Context, signature string, session fosite.Session) (fosite.Requester, error) {
	query := `
		SELECT request_id, requested_at, client_id, scopes, granted_scopes,
			   form_data, session_data, subject, requested_audience, granted_audience
		FROM oauth2_access_tokens
		WHERE signature = $1 AND active = true AND expires_at > NOW()
	`

	var requestID string
	var requestedAt time.Time
	var clientID string
	var scopes, grantedScopes []string
	var formData, sessionData string
	var subject string
	var requestedAudience, grantedAudience []string

	err := s.pool.QueryRow(ctx, query, signature).Scan(
		&requestID, &requestedAt, &clientID, &scopes, &grantedScopes,
		&formData, &sessionData, &subject, &requestedAudience, &grantedAudience,
	)
	if err != nil {
		return nil, fosite.ErrNotFound
	}

	client, err := s.GetClient(ctx, clientID)
	if err != nil {
		return nil, err
	}

	req := fosite.NewRequest()
	req.ID = requestID
	req.RequestedAt = requestedAt
	req.Client = client
	req.RequestedScope = []string(scopes)
	req.GrantedScope = []string(grantedScopes)
	req.RequestedAudience = []string(requestedAudience)
	req.GrantedAudience = []string(grantedAudience)
	req.Session = decodeOIDCSession(sessionData)
	req.Form = toURLValues(formData)

	return req, nil
}

func (s *Store) DeleteAccessTokenSession(ctx context.Context, signature string) error {
	query := `UPDATE oauth2_access_tokens SET active = false WHERE signature = $1`
	_, err := s.pool.Exec(ctx, query, signature)
	return err
}

// Refresh Token methods

func (s *Store) CreateRefreshTokenSession(ctx context.Context, signature string, requestID string, req fosite.Requester) error {
	sessionData, _ := json.Marshal(req.GetSession())
	formData, _ := json.Marshal(req.GetRequestForm())

	query := `
		INSERT INTO oauth2_refresh_tokens (
			signature, request_id, requested_at, client_id, scopes, granted_scopes,
			form_data, session_data, subject, requested_audience, granted_audience
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := s.pool.Exec(ctx, query,
		signature,
		req.GetID(),
		req.GetRequestedAt(),
		req.GetClient().GetID(),
		toStringSlice(req.GetRequestedScopes()),
		toStringSlice(req.GetGrantedScopes()),
		string(formData),
		string(sessionData),
		req.GetSession().GetSubject(),
		toStringSlice(req.GetRequestedAudience()),
		toStringSlice(req.GetGrantedAudience()),
	)

	return err
}

func (s *Store) GetRefreshTokenSession(ctx context.Context, signature string, session fosite.Session) (fosite.Requester, error) {
	query := `
		SELECT request_id, requested_at, client_id, scopes, granted_scopes,
			   form_data, session_data, subject, requested_audience, granted_audience
		FROM oauth2_refresh_tokens
		WHERE signature = $1 AND active = true
	`

	var requestID string
	var requestedAt time.Time
	var clientID string
	var scopes, grantedScopes []string
	var formData, sessionData string
	var subject string
	var requestedAudience, grantedAudience []string

	err := s.pool.QueryRow(ctx, query, signature).Scan(
		&requestID, &requestedAt, &clientID, &scopes, &grantedScopes,
		&formData, &sessionData, &subject, &requestedAudience, &grantedAudience,
	)
	if err != nil {
		return nil, fosite.ErrNotFound
	}

	client, err := s.GetClient(ctx, clientID)
	if err != nil {
		return nil, err
	}

	req := fosite.NewRequest()
	req.ID = requestID
	req.RequestedAt = requestedAt
	req.Client = client
	req.RequestedScope = []string(scopes)
	req.GrantedScope = []string(grantedScopes)
	req.RequestedAudience = []string(requestedAudience)
	req.GrantedAudience = []string(grantedAudience)
	req.Session = decodeOIDCSession(sessionData)
	req.Form = toURLValues(formData)

	return req, nil
}

func (s *Store) DeleteRefreshTokenSession(ctx context.Context, signature string) error {
	query := `UPDATE oauth2_refresh_tokens SET active = false WHERE signature = $1`
	_, err := s.pool.Exec(ctx, query, signature)
	return err
}

func (s *Store) RotateRefreshToken(ctx context.Context, requestID string, newSignature string) error {
	// This method is called when refresh tokens need to be rotated
	// We'll mark the old token as inactive - the new token will be created separately
	query := `UPDATE oauth2_refresh_tokens SET active = false WHERE request_id = $1`
	_, err := s.pool.Exec(ctx, query, requestID)
	return err
}

// PKCE methods

func (s *Store) CreatePKCERequestSession(ctx context.Context, signature string, req fosite.Requester) error {
	sessionData, _ := json.Marshal(req.GetSession())
	formData, _ := json.Marshal(req.GetRequestForm())

	query := `
		INSERT INTO oauth2_pkce (
			signature, request_id, requested_at, client_id, scopes, granted_scopes,
			form_data, session_data, subject, requested_audience, granted_audience
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := s.pool.Exec(ctx, query,
		signature,
		req.GetID(),
		req.GetRequestedAt(),
		req.GetClient().GetID(),
		toStringSlice(req.GetRequestedScopes()),
		toStringSlice(req.GetGrantedScopes()),
		string(formData),
		string(sessionData),
		req.GetSession().GetSubject(),
		toStringSlice(req.GetRequestedAudience()),
		toStringSlice(req.GetGrantedAudience()),
	)

	return err
}

func (s *Store) GetPKCERequestSession(ctx context.Context, signature string, session fosite.Session) (fosite.Requester, error) {
	query := `
		SELECT request_id, requested_at, client_id, scopes, granted_scopes,
			   form_data, session_data, subject, requested_audience, granted_audience
		FROM oauth2_pkce
		WHERE signature = $1 AND active = true
	`

	var requestID string
	var requestedAt time.Time
	var clientID string
	var scopes, grantedScopes []string
	var formData, sessionData string
	var subject string
	var requestedAudience, grantedAudience []string

	err := s.pool.QueryRow(ctx, query, signature).Scan(
		&requestID, &requestedAt, &clientID, &scopes, &grantedScopes,
		&formData, &sessionData, &subject, &requestedAudience, &grantedAudience,
	)
	if err != nil {
		return nil, fosite.ErrNotFound
	}

	client, err := s.GetClient(ctx, clientID)
	if err != nil {
		return nil, err
	}

	req := fosite.NewRequest()
	req.ID = requestID
	req.RequestedAt = requestedAt
	req.Client = client
	req.RequestedScope = []string(scopes)
	req.GrantedScope = []string(grantedScopes)
	req.RequestedAudience = []string(requestedAudience)
	req.GrantedAudience = []string(grantedAudience)
	req.Session = decodeOIDCSession(sessionData)
	req.Form = toURLValues(formData)

	return req, nil
}

func (s *Store) DeletePKCERequestSession(ctx context.Context, signature string) error {
	query := `UPDATE oauth2_pkce SET active = false WHERE signature = $1`
	_, err := s.pool.Exec(ctx, query, signature)
	return err
}

// OpenID Connect methods

func (s *Store) CreateOpenIDConnectSession(ctx context.Context, authorizeCode string, req fosite.Requester) error {
	sessionData, _ := json.Marshal(req.GetSession())
	formData, _ := json.Marshal(req.GetRequestForm())

	query := `
		INSERT INTO oauth2_oidc_sessions (
			signature, request_id, requested_at, client_id, scopes, granted_scopes,
			form_data, session_data, subject, requested_audience, granted_audience
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := s.pool.Exec(ctx, query,
		authorizeCode,
		req.GetID(),
		req.GetRequestedAt(),
		req.GetClient().GetID(),
		toStringSlice(req.GetRequestedScopes()),
		toStringSlice(req.GetGrantedScopes()),
		string(formData),
		string(sessionData),
		req.GetSession().GetSubject(),
		toStringSlice(req.GetRequestedAudience()),
		toStringSlice(req.GetGrantedAudience()),
	)

	return err
}

func (s *Store) GetOpenIDConnectSession(ctx context.Context, authorizeCode string, req fosite.Requester) (fosite.Requester, error) {
	query := `
		SELECT request_id, requested_at, client_id, scopes, granted_scopes,
			   form_data, session_data, subject, requested_audience, granted_audience
		FROM oauth2_oidc_sessions
		WHERE signature = $1 AND active = true
	`

	var requestID string
	var requestedAt time.Time
	var clientID string
	var scopes, grantedScopes []string
	var formData, sessionData string
	var subject string
	var requestedAudience, grantedAudience []string

	err := s.pool.QueryRow(ctx, query, authorizeCode).Scan(
		&requestID, &requestedAt, &clientID, &scopes, &grantedScopes,
		&formData, &sessionData, &subject, &requestedAudience, &grantedAudience,
	)
	if err != nil {
		return nil, fosite.ErrNotFound
	}

	client, err := s.GetClient(ctx, clientID)
	if err != nil {
		return nil, err
	}

	request := fosite.NewRequest()
	request.ID = requestID
	request.RequestedAt = requestedAt
	request.Client = client
	request.RequestedScope = []string(scopes)
	request.GrantedScope = []string(grantedScopes)
	request.RequestedAudience = []string(requestedAudience)
	request.GrantedAudience = []string(grantedAudience)
	request.Session = decodeOIDCSession(sessionData)
	request.Form = toURLValues(formData)

	return request, nil
}

func (s *Store) DeleteOpenIDConnectSession(ctx context.Context, authorizeCode string) error {
	query := `UPDATE oauth2_oidc_sessions SET active = false WHERE signature = $1`
	_, err := s.pool.Exec(ctx, query, authorizeCode)
	return err
}

// JTI Blacklist methods

func (s *Store) IsJWTUsed(ctx context.Context, jti string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM oauth2_blacklisted_jtis WHERE signature = $1)`
	var exists bool
	err := s.pool.QueryRow(ctx, query, jti).Scan(&exists)
	return exists, err
}

func (s *Store) MarkJWTUsedForTime(ctx context.Context, jti string, exp time.Time) error {
	query := `
		INSERT INTO oauth2_blacklisted_jtis (signature, expires_at)
		VALUES ($1, $2)
		ON CONFLICT (signature) DO NOTHING
	`
	_, err := s.pool.Exec(ctx, query, jti, exp)
	return err
}

// Helper methods

func (s *Store) Authenticate(ctx context.Context, name string, secret string) error {
	// This would implement client authentication
	client, err := s.GetClient(ctx, name)
	if err != nil {
		return err
	}

	if client.IsPublic() {
		return nil
	}

	// In a real implementation, you would hash and compare the secret
	hashedSecret := client.GetHashedSecret()
	if len(hashedSecret) == 0 {
		return fosite.ErrInvalidClient
	}

	if err := bcrypt.CompareHashAndPassword(hashedSecret, []byte(secret)); err != nil {
		return fosite.ErrInvalidClient
	}

	return nil
}

// RevokeRefreshToken Satisfy interface requirements (these are unused but required by fosite interfaces)
func (s *Store) RevokeRefreshToken(ctx context.Context, requestID string) error {
	return s.revokeSessionByRequestID(ctx, "oauth2_refresh_tokens", requestID)
}

func (s *Store) RevokeAccessToken(ctx context.Context, requestID string) error {
	return s.revokeSessionByRequestID(ctx, "oauth2_access_tokens", requestID)
}

func (s *Store) revokeSessionByRequestID(ctx context.Context, table, requestID string) error {
	query := fmt.Sprintf("UPDATE %s SET active = false WHERE request_id = $1", table)
	_, err := s.pool.Exec(ctx, query, requestID)
	return err
}

// ClientAssertionJWTValid checks if a JWT has been used before
func (s *Store) ClientAssertionJWTValid(ctx context.Context, jti string) error {
	used, err := s.IsJWTUsed(ctx, jti)
	if err != nil {
		return err
	}
	if used {
		return fosite.ErrJTIKnown
	}
	return nil
}

// SetClientAssertionJWT marks a JWT as used
func (s *Store) SetClientAssertionJWT(ctx context.Context, jti string, exp time.Time) error {
	return s.MarkJWTUsedForTime(ctx, jti, exp)
}

// Ensure Store implements required interfaces
var _ fosite.ClientManager = (*Store)(nil)
var _ fosite.Storage = (*Store)(nil)
var _ oauth2.CoreStorage = (*Store)(nil)
