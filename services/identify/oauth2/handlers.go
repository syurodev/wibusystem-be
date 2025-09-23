package oauth2

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ory/fosite"
	"github.com/ory/fosite/handler/openid"
	"golang.org/x/crypto/bcrypt"

	"crypto/sha256"
	m "wibusystem/pkg/common/model"
	"wibusystem/pkg/i18n"
	"wibusystem/services/identify/repositories"
	"wibusystem/services/identify/session"
)

// Handlers provides Gin handlers that adapt the fosite provider to HTTP
// endpoints for OAuth2/OIDC flows.
type Handlers struct {
	provider     *Provider
	repos        *repositories.Repositories
	sess         *session.Manager
	devMode      bool
	regSecret    string
	loginPageURL string
	loc          *i18n.Translator
}

// NewHandlers constructs OAuth2/OIDC HTTP handlers.
func NewHandlers(provider *Provider, repos *repositories.Repositories, sess *session.Manager, translator *i18n.Translator, devMode bool, regSecret string, loginPageURL string) *Handlers {
	return &Handlers{
		provider:     provider,
		repos:        repos,
		sess:         sess,
		devMode:      devMode,
		regSecret:    regSecret,
		loginPageURL: loginPageURL,
		loc:          translator,
	}
}

// AuthorizeHandler handles the OAuth2 authorization endpoint.
// NOTE: This implementation uses a temporary user lookup via query param.
// Integrate with your session/login system to attach real users.
func (h *Handlers) AuthorizeHandler(c *gin.Context) {
	ctx := c.Request.Context()

	fmt.Printf("DEBUG AUTHORIZE: Request URL: %s\n", c.Request.URL.String())
	fmt.Printf("DEBUG AUTHORIZE: Cookies: %+v\n", c.Request.Cookies())

	// Parse the authorization request
	ar, err := h.provider.OAuth2Provider.NewAuthorizeRequest(ctx, c.Request)
	if err != nil {
		fmt.Printf("DEBUG AUTHORIZE: NewAuthorizeRequest error: %v\n", err)
		h.provider.OAuth2Provider.WriteAuthorizeError(ctx, c.Writer, ar, err)
		return
	}

	fmt.Printf("DEBUG AUTHORIZE: Authorization request parsed successfully\n")
	fmt.Printf("DEBUG AUTHORIZE: Client ID: %s\n", ar.GetClient().GetID())
	fmt.Printf("DEBUG AUTHORIZE: Redirect URI: %s\n", ar.GetRedirectURI().String())

	// Check login via session cookie
	userID, err := h.sess.Get(c)
	fmt.Printf("DEBUG AUTHORIZE: Session check - UserID: %s, Error: %v\n", userID, err)

	if err != nil || userID == "" {
		loginURL := h.loginPageURL
		if loginURL == "" { // fallback to internal login page
			loginURL = "/login"
		}
		loginURL = loginURL + "?redirect_uri=" + url.QueryEscape(c.Request.URL.String())
		fmt.Printf("DEBUG AUTHORIZE: Redirecting to login: %s\n", loginURL)
		c.Redirect(http.StatusFound, loginURL)
		return
	}

	fmt.Printf("DEBUG AUTHORIZE: User authenticated, processing authorization...\n")

	// Get user information
	userUUID, convErr := uuid.Parse(userID)
	if convErr != nil {
		fmt.Printf("DEBUG AUTHORIZE: Failed to parse userID: %v\n", convErr)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_user_session"})
		return
	}
	// Create a new context with longer timeout for database operations
	dbCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	user, err := h.repos.User.GetByID(dbCtx, userUUID)
	if err != nil {
		fmt.Printf("DEBUG AUTHORIZE: User not found: %v\n", err)
		fmt.Printf("DEBUG AUTHORIZE: Error type: %T\n", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found", "details": err.Error()})
		return
	}

	fmt.Printf("DEBUG AUTHORIZE: User found: %s\n", user.Username)

	// Consent: auto-approve in dev; in prod require consent param (temporary)
	if !h.devMode {
		if c.Query("consent") != "approve" {
			fmt.Printf("DEBUG AUTHORIZE: Consent required but not provided\n")
			h.provider.OAuth2Provider.WriteAuthorizeError(ctx, c.Writer, ar, fosite.ErrAccessDenied.WithHint("consent_required"))
			return
		}
	}

	fmt.Printf("DEBUG AUTHORIZE: Dev mode, auto-approving consent\n")

	// Create OIDC session with user information
	session := h.provider.CreateCustomSession(userID, user.Username, user.Email, user.AvatarURL)
	// Ensure audience includes the client ID for ID Token
	if session.Claims != nil {
		session.Claims.Audience = []string{ar.GetClient().GetID()}
	}

	fmt.Printf("DEBUG AUTHORIZE: Session created for user: %s\n", user.Username)

	// Grant requested scopes and audience (consent)
	for _, s := range ar.GetRequestedScopes() {
		ar.GrantScope(s)
	}
	for _, a := range ar.GetRequestedAudience() {
		ar.GrantAudience(a)
	}

	fmt.Printf("DEBUG AUTHORIZE: Scopes and audience granted\n")

	// Handle the authorization request with a fresh context to avoid timeout issues
	authCtx, authCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer authCancel()

	response, err := h.provider.OAuth2Provider.NewAuthorizeResponse(authCtx, ar, session)
	if err != nil {
		fmt.Printf("DEBUG AUTHORIZE: NewAuthorizeResponse error: %v\n", err)
		fmt.Printf("DEBUG AUTHORIZE: Error type: %T\n", err)
		fmt.Printf("DEBUG AUTHORIZE: Error details: %+v\n", err)

		// Log additional context for better debugging
		fmt.Printf("DEBUG AUTHORIZE: Session subject: %s\n", session.GetSubject())
		fmt.Printf("DEBUG AUTHORIZE: Requested scopes: %v\n", ar.GetRequestedScopes())
		fmt.Printf("DEBUG AUTHORIZE: Client ID: %s\n", ar.GetClient().GetID())

		h.provider.OAuth2Provider.WriteAuthorizeError(authCtx, c.Writer, ar, err)
		return
	}

	fmt.Printf("DEBUG AUTHORIZE: Authorization response created successfully\n")
	fmt.Printf("DEBUG AUTHORIZE: About to write response and redirect to callback URL\n")

	// Write the successful authorization response with explicit redirect handling
	// Set headers to ensure proper redirect for browser navigation
	c.Header("Cache-Control", "no-store")
	c.Header("Pragma", "no-cache")

	h.provider.OAuth2Provider.WriteAuthorizeResponse(authCtx, c.Writer, ar, response)
}

// TokenHandler handles the OAuth2 token endpoint.
func (h *Handlers) TokenHandler(c *gin.Context) {
	fmt.Printf("DEBUG TOKEN: ===============================\n")
	fmt.Printf("DEBUG TOKEN: Request URL: %s\n", c.Request.URL.String())
	fmt.Printf("DEBUG TOKEN: Method: %s\n", c.Request.Method)
	fmt.Printf("DEBUG TOKEN: Headers: %+v\n", c.Request.Header)

	// Read and log form data
	if err := c.Request.ParseForm(); err == nil {
		fmt.Printf("DEBUG TOKEN: Form data: %+v\n", c.Request.Form)
	} else {
		fmt.Printf("DEBUG TOKEN: Failed to parse form: %v\n", err)
	}

	ctx := c.Request.Context()

	// Create a session for the request
	session := &openid.DefaultSession{}

	// Handle the token request
	ar, err := h.provider.OAuth2Provider.NewAccessRequest(ctx, c.Request, session)
	if err != nil {
		// Log details for debugging token failures
		clientID := c.PostForm("client_id")
		grantType := c.PostForm("grant_type")
		log.Printf("oauth2.token: NewAccessRequest error: %v client_id=%s grant=%s", err, clientID, grantType)
		h.provider.OAuth2Provider.WriteAccessError(ctx, c.Writer, ar, err)
		return
	}

	// Handle different grant types
	grantTypes := ar.GetGrantTypes()
	if len(grantTypes) > 0 {
		switch grantTypes[0] {
		case "authorization_code":
			err = h.handleAuthorizationCodeGrant(ctx, ar)
		case "refresh_token":
			err = h.handleRefreshTokenGrant(ctx, ar)
		case "client_credentials":
			err = h.handleClientCredentialsGrant(ctx, ar)
		case "password":
			// TODO: Password grant is disabled by default. Consider enabling only for dev/testing.
			err = fosite.ErrUnsupportedGrantType
		default:
			err = fosite.ErrUnsupportedGrantType
		}
	} else {
		err = fosite.ErrInvalidRequest.WithDescription("Grant type is required")
	}

	if err != nil {
		log.Printf("oauth2.token: grant handler error: %v grant_types=%v client=%s", err, grantTypes, ar.GetClient().GetID())
		h.provider.OAuth2Provider.WriteAccessError(ctx, c.Writer, ar, err)
		return
	}

	// Create the access response
	response, err := h.provider.OAuth2Provider.NewAccessResponse(ctx, ar)
	if err != nil {
		log.Printf("oauth2.token: NewAccessResponse error: %v client=%s", err, ar.GetClient().GetID())
		h.provider.OAuth2Provider.WriteAccessError(ctx, c.Writer, ar, err)
		return
	}

	// Write the successful token response
	h.provider.OAuth2Provider.WriteAccessResponse(ctx, c.Writer, ar, response)
}

// =================== Dynamic Client Registration (RFC 7591/7592) ===================

type registerRequest struct {
	RedirectURIs            []string `json:"redirect_uris"`
	GrantTypes              []string `json:"grant_types,omitempty"`
	ResponseTypes           []string `json:"response_types,omitempty"`
	Scope                   string   `json:"scope,omitempty"`
	TokenEndpointAuthMethod string   `json:"token_endpoint_auth_method,omitempty"`
	ClientName              string   `json:"client_name,omitempty"`
	ClientURI               string   `json:"client_uri,omitempty"`
	LogoURI                 string   `json:"logo_uri,omitempty"`
	JWKSURI                 string   `json:"jwks_uri,omitempty"`
	JWKS                    string   `json:"jwks,omitempty"`
	PolicyURI               string   `json:"policy_uri,omitempty"`
	TOSURI                  string   `json:"tos_uri,omitempty"`
	Contacts                []string `json:"contacts,omitempty"`
}

type registerResponse struct {
	ClientID                string   `json:"client_id"`
	ClientSecret            string   `json:"client_secret,omitempty"`
	ClientIDIssuedAt        int64    `json:"client_id_issued_at,omitempty"`
	ClientSecretExpiresAt   int64    `json:"client_secret_expires_at,omitempty"`
	RegistrationAccessToken string   `json:"registration_access_token"`
	RegistrationClientURI   string   `json:"registration_client_uri"`
	RedirectURIs            []string `json:"redirect_uris,omitempty"`
	GrantTypes              []string `json:"grant_types,omitempty"`
	ResponseTypes           []string `json:"response_types,omitempty"`
	Scope                   string   `json:"scope,omitempty"`
	TokenEndpointAuthMethod string   `json:"token_endpoint_auth_method,omitempty"`
	ClientName              string   `json:"client_name,omitempty"`
}

func (h *Handlers) RegisterClient(c *gin.Context) {
	// Require Initial Access Token from DB (production policy)
	if !h.validateIAT(c, c.GetHeader("Authorization")) {
		return
	}

	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}
	if len(req.RedirectURIs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_client_metadata", "error_description": "redirect_uris required"})
		return
	}
	if req.TokenEndpointAuthMethod == "" {
		// public client by default
		req.TokenEndpointAuthMethod = "none"
	}
	isPublic := req.TokenEndpointAuthMethod == "none"

	clientID := uuid.NewString()
	var clientSecret string
	var clientSecretHash *string
	if !isPublic {
		clientSecret = uuid.NewString()
		// hash secret
		hashed, _ := bcrypt.GenerateFromPassword([]byte(clientSecret), 12)
		hs := string(hashed)
		clientSecretHash = &hs
	}

	scope := req.Scope
	var scopes []string
	if scope != "" {
		// split by space
		scopes = splitScope(scope)
	}

	// Insert into DB
	_, err := h.provider.Store.pool.Exec(c.Request.Context(), `
        INSERT INTO oauth2_clients (
            id, client_secret_hash, redirect_uris, grant_types, response_types,
            scopes, audience, public, client_name
        ) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
    `, clientID, clientSecretHash, req.RedirectURIs, req.GrantTypes, req.ResponseTypes, scopes, []string{}, isPublic, req.ClientName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server_error"})
		return
	}

	// Build Registration Access Token (RAT)
	rat, err := h.buildRAT(clientID, time.Now().Add(7*24*time.Hour))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server_error"})
		return
	}

	regURI := fmt.Sprintf("%s/register/%s", h.provider.Config.AccessTokenIssuer, clientID)
	resp := registerResponse{
		ClientID:                clientID,
		ClientSecret:            clientSecret,
		ClientIDIssuedAt:        time.Now().Unix(),
		ClientSecretExpiresAt:   0,
		RegistrationAccessToken: rat,
		RegistrationClientURI:   regURI,
		RedirectURIs:            req.RedirectURIs,
		GrantTypes:              req.GrantTypes,
		ResponseTypes:           req.ResponseTypes,
		Scope:                   scope,
		TokenEndpointAuthMethod: req.TokenEndpointAuthMethod,
		ClientName:              req.ClientName,
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *Handlers) GetRegisteredClient(c *gin.Context) {
	clientID := c.Param("client_id")
	if !h.verifyRAT(c, clientID) {
		return
	}

	row := h.provider.Store.pool.QueryRow(c.Request.Context(), `
        SELECT id, redirect_uris, grant_types, response_types, scopes, public, client_name
        FROM oauth2_clients WHERE id=$1
    `, clientID)
	var id string
	var redirectURIs, grantTypes, responseTypes, scopes []string
	var public bool
	var clientName *string
	if err := row.Scan(&id, &redirectURIs, &grantTypes, &responseTypes, &scopes, &public, &clientName); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
		return
	}
	scope := joinScope(scopes)
	authMethod := "none"
	if !public {
		authMethod = "client_secret_basic"
	}
	regURI := fmt.Sprintf("%s/register/%s", h.provider.Config.AccessTokenIssuer, clientID)
	c.JSON(http.StatusOK, registerResponse{
		ClientID:                id,
		ClientSecret:            "",
		ClientIDIssuedAt:        0,
		ClientSecretExpiresAt:   0,
		RegistrationAccessToken: "",
		RegistrationClientURI:   regURI,
		RedirectURIs:            redirectURIs,
		GrantTypes:              grantTypes,
		ResponseTypes:           responseTypes,
		Scope:                   scope,
		TokenEndpointAuthMethod: authMethod,
		ClientName:              ptrStr(clientName),
	})
}

func (h *Handlers) UpdateRegisteredClient(c *gin.Context) {
	clientID := c.Param("client_id")
	if !h.verifyRAT(c, clientID) {
		return
	}
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}
	if len(req.RedirectURIs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_client_metadata", "error_description": "redirect_uris required"})
		return
	}
	_, err := h.provider.Store.pool.Exec(c.Request.Context(), `
        UPDATE oauth2_clients SET redirect_uris=$2, grant_types=$3, response_types=$4, scopes=$5, client_name=$6
        WHERE id=$1
    `, clientID, req.RedirectURIs, req.GrantTypes, req.ResponseTypes, splitScope(req.Scope), req.ClientName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server_error"})
		return
	}
	h.GetRegisteredClient(c)
}

func (h *Handlers) DeleteRegisteredClient(c *gin.Context) {
	clientID := c.Param("client_id")
	if !h.verifyRAT(c, clientID) {
		return
	}
	_, err := h.provider.Store.pool.Exec(c.Request.Context(), `DELETE FROM oauth2_clients WHERE id=$1`, clientID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server_error"})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handlers) buildRAT(clientID string, exp time.Time) (string, error) {
	// Issue DB-backed RAT token
	token := uuid.NewString()
	th := sha256.Sum256([]byte(token))
	hash := fmt.Sprintf("%x", th[:])
	_, err := h.provider.Store.pool.Exec(
		context.Background(),
		`INSERT INTO oauth2_registration_access_tokens (client_id, token_hash, expires_at, active) VALUES ($1,$2,$3,TRUE)`,
		clientID, hash, exp,
	)
	if err != nil {
		return "", err
	}
	return token, nil
}

func (h *Handlers) verifyRAT(c *gin.Context, clientID string) bool {
	auth := c.GetHeader("Authorization")
	if auth == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_token"})
		return false
	}
	token := auth
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}
	th := sha256.Sum256([]byte(token))
	hash := fmt.Sprintf("%x", th[:])
	row := h.provider.Store.pool.QueryRow(c.Request.Context(), `
        SELECT active, expires_at FROM oauth2_registration_access_tokens WHERE client_id=$1 AND token_hash=$2
    `, clientID, hash)
	var active bool
	var exp *time.Time
	if err := row.Scan(&active, &exp); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_token"})
		return false
	}
	if !active || (exp != nil && time.Now().After(*exp)) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_token"})
		return false
	}
	return true
}

func splitScope(s string) []string {
	if s == "" {
		return []string{}
	}
	out := []string{}
	cur := ""
	for _, ch := range s {
		if ch == ' ' || ch == '\n' || ch == '\t' {
			if cur != "" {
				out = append(out, cur)
				cur = ""
			}
		} else {
			cur += string(ch)
		}
	}
	if cur != "" {
		out = append(out, cur)
	}
	return out
}

func joinScope(scopes []string) string {
	if len(scopes) == 0 {
		return ""
	}
	out := scopes[0]
	for i := 1; i < len(scopes); i++ {
		out += " " + scopes[i]
	}
	return out
}

func ptrStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// validateIAT validates Initial Access Token (IAT) from DB for DCR
func (h *Handlers) validateIAT(c *gin.Context, auth string) bool {
	if auth == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_token"})
		return false
	}
	token := auth
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}
	th := sha256.Sum256([]byte(token))
	hash := fmt.Sprintf("%x", th[:])
	row := h.provider.Store.pool.QueryRow(c.Request.Context(), `
        SELECT active, expires_at FROM oauth2_initial_access_tokens WHERE token_hash=$1
    `, hash)
	var active bool
	var exp *time.Time
	if err := row.Scan(&active, &exp); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_token"})
		return false
	}
	if !active || (exp != nil && time.Now().After(*exp)) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_token"})
		return false
	}
	return true
}

// ============== Admin: Initial Access Token issuance/policy ==============
// POST /admin/registration/iat { expires_in_seconds?: int, description?: string }
func (h *Handlers) AdminCreateIAT(c *gin.Context) {
	var body struct {
		ExpiresIn   int64  `json:"expires_in_seconds"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}
	token := uuid.NewString()
	th := sha256.Sum256([]byte(token))
	hash := fmt.Sprintf("%x", th[:])
	var exp *time.Time
	if body.ExpiresIn > 0 {
		t := time.Now().Add(time.Duration(body.ExpiresIn) * time.Second)
		exp = &t
	}
	_, err := h.provider.Store.pool.Exec(c.Request.Context(), `
        INSERT INTO oauth2_initial_access_tokens (token_hash, description, expires_at, active)
        VALUES ($1,$2,$3,TRUE)
    `, hash, body.Description, exp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server_error"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"token": token, "expires_at": exp})
}

// GET /admin/registration/iat
func (h *Handlers) AdminListIAT(c *gin.Context) {
	rows, err := h.provider.Store.pool.Query(c.Request.Context(), `
        SELECT id, description, created_at, expires_at, revoked_at, active FROM oauth2_initial_access_tokens ORDER BY created_at DESC
    `)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server_error"})
		return
	}
	defer rows.Close()
	type item struct {
		ID          int64      `json:"id"`
		Description string     `json:"description"`
		CreatedAt   time.Time  `json:"created_at"`
		ExpiresAt   *time.Time `json:"expires_at"`
		RevokedAt   *time.Time `json:"revoked_at"`
		Active      bool       `json:"active"`
	}
	out := []item{}
	for rows.Next() {
		var it item
		if err := rows.Scan(&it.ID, &it.Description, &it.CreatedAt, &it.ExpiresAt, &it.RevokedAt, &it.Active); err == nil {
			out = append(out, it)
		}
	}
	c.JSON(http.StatusOK, out)
}

// DELETE /admin/registration/iat/:id
func (h *Handlers) AdminRevokeIAT(c *gin.Context) {
	id := c.Param("id")
	_, err := h.provider.Store.pool.Exec(c.Request.Context(), `
        UPDATE oauth2_initial_access_tokens SET active=FALSE, revoked_at=NOW() WHERE id=$1
    `, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server_error"})
		return
	}
	c.Status(http.StatusNoContent)
}

// =================== Consent endpoints ===================
// GET /oauth2/consent?redirect=<encoded_authorize_url>
func (h *Handlers) GetConsent(c *gin.Context) {
	// Require login
	if _, err := h.sess.Get(c); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "login_required"})
		return
	}
	red := c.Query("redirect")
	if red == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "error_description": "redirect missing"})
		return
	}
	u, err := url.Parse(red)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "error_description": "redirect invalid"})
		return
	}
	q := u.Query()
	clientID := q.Get("client_id")
	scopeStr := q.Get("scope")
	if clientID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "error_description": "client_id missing"})
		return
	}
	// Lookup client metadata
	row := h.provider.Store.pool.QueryRow(c.Request.Context(), `SELECT id, client_name, public FROM oauth2_clients WHERE id=$1`, clientID)
	var id string
	var clientName *string
	var public bool
	if err := row.Scan(&id, &clientName, &public); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "client_not_found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"client_id":   id,
		"client_name": ptrStr(clientName),
		"public":      public,
		"scopes":      splitScope(scopeStr),
		"redirect":    red,
	})
}

// POST /oauth2/consent { redirect, decision }
func (h *Handlers) PostConsent(c *gin.Context) {
	if _, err := h.sess.Get(c); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "login_required"})
		return
	}
	var body struct {
		Redirect string `json:"redirect"`
		Decision string `json:"decision"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.Redirect == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}
	red, err := url.Parse(body.Redirect)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}
	q := red.Query()
	if body.Decision == "approve" {
		q.Set("consent", "approve")
	} else {
		q.Set("error", "access_denied")
	}
	red.RawQuery = q.Encode()
	c.Redirect(http.StatusFound, red.String())
}

// handleAuthorizationCodeGrant handles authorization code grant.
func (h *Handlers) handleAuthorizationCodeGrant(ctx context.Context, ar fosite.AccessRequester) error {
	fmt.Printf("DEBUG TOKEN: handleAuthorizationCodeGrant called\n")

	// The session should be retrieved from the authorization code
	// For OAuth2 authorization code flow, fosite automatically handles the session retrieval
	// from the stored authorization code, but we need to make sure the session is properly set

	session := ar.GetSession()
	if session == nil {
		fmt.Printf("DEBUG TOKEN: No session found in AccessRequester\n")
		return fosite.ErrServerError.WithHint("No session data found")
	}

	// Log session details for debugging
	fmt.Printf("DEBUG TOKEN: Session subject: %s\n", session.GetSubject())

	if oidcSession, ok := session.(*openid.DefaultSession); ok {
		fmt.Printf("DEBUG TOKEN: OIDC session found\n")
		if oidcSession.Claims != nil {
			fmt.Printf("DEBUG TOKEN: Claims subject: %s\n", oidcSession.Claims.Subject)
			fmt.Printf("DEBUG TOKEN: Claims audience: %v\n", oidcSession.Claims.Audience)
		} else {
			fmt.Printf("DEBUG TOKEN: No claims in session\n")
		}
	} else {
		fmt.Printf("DEBUG TOKEN: Session is not OIDC session: %T\n", session)
	}

	return nil
}

// handleRefreshTokenGrant handles refresh token grant.
func (h *Handlers) handleRefreshTokenGrant(_ context.Context, _ fosite.AccessRequester) error {
	// Fosite handles refresh token validation automatically
	return nil
}

// handleClientCredentialsGrant handles client credentials grant.
func (h *Handlers) handleClientCredentialsGrant(_ context.Context, ar fosite.AccessRequester) error {
	// For client credentials, we create a session with the client as the subject
	clientID := ar.GetClient().GetID()
	session := h.provider.CreateCustomSession(clientID, clientID, "", nil)
	ar.SetSession(session)
	return nil
}

// handlePasswordGrant handles resource owner password credentials grant.
func (h *Handlers) handlePasswordGrant(ctx context.Context, ar fosite.AccessRequester) error {
	username := ar.GetRequestForm().Get("username")
	password := ar.GetRequestForm().Get("password")

	if username == "" || password == "" {
		return fosite.ErrInvalidRequest.WithDescription("Username and password required")
	}

	// Authenticate user
	user, err := h.authenticateUser(ctx, username, password)
	if err != nil {
		return fosite.ErrInvalidGrant.WithDescription("Invalid username or password")
	}

	// Create session with user information
	session := h.provider.CreateCustomSession(
		user.ID.String(),
		user.Username,
		user.Email,
		user.AvatarURL,
	)
	ar.SetSession(session)

	return nil
}

// authenticateUser authenticates a user with username/email and password
// by verifying a stored bcrypt hash for the password credential.
func (h *Handlers) authenticateUser(ctx context.Context, identifier, password string) (*m.User, error) {
	// Try to find user by email first, then username
	var user *m.User
	var err error

	user, err = h.repos.User.GetByEmail(ctx, identifier)
	if err != nil {
		// Try username if email lookup failed
		user, err = h.repos.User.GetByUsername(ctx, identifier)
		if err != nil {
			return nil, err
		}
	}

	// Get password credential for the user
	credential, err := h.repos.Credential.GetByUserIDAndType(ctx, user.ID, m.AuthTypePassword)
	if err != nil {
		return nil, err
	}

	// Verify password
	if credential.SecretHash == nil {
		return nil, fmt.Errorf("no password set for user")
	}

	err = bcrypt.CompareHashAndPassword([]byte(*credential.SecretHash), []byte(password))
	if err != nil {
		return nil, err
	}

	// Update last login and credential usage
	_ = h.repos.User.UpdateLastLogin(ctx, user.ID)
	_ = h.repos.Credential.UpdateLastUsed(ctx, credential.ID)

	return user, nil
}

// IntrospectHandler handles token introspection.
func (h *Handlers) IntrospectHandler(c *gin.Context) {
	ctx := c.Request.Context()

	// Handle introspection request
	ir, err := h.provider.OAuth2Provider.NewIntrospectionRequest(ctx, c.Request, &openid.DefaultSession{})
	if err != nil {
		h.provider.OAuth2Provider.WriteIntrospectionError(ctx, c.Writer, err)
		return
	}

	// Write introspection response
	h.provider.OAuth2Provider.WriteIntrospectionResponse(ctx, c.Writer, ir)
}

// RevokeHandler handles token revocation.
func (h *Handlers) RevokeHandler(c *gin.Context) {
	ctx := c.Request.Context()

	// Handle revocation request
	err := h.provider.OAuth2Provider.NewRevocationRequest(ctx, c.Request)
	if err != nil {
		h.provider.OAuth2Provider.WriteRevocationResponse(ctx, c.Writer, err)
		return
	}

	// Write successful revocation response
	h.provider.OAuth2Provider.WriteRevocationResponse(ctx, c.Writer, err)
}

// JWKSHandler returns the JSON Web Key Set used to validate OIDC ID tokens.
func (h *Handlers) JWKSHandler(c *gin.Context) {
	jwks, err := h.provider.GetJWKS()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate JWKS"})
		return
	}

	c.JSON(http.StatusOK, jwks)
}

// UserInfoHandler handles the OpenID Connect UserInfo endpoint.
func (h *Handlers) UserInfoHandler(c *gin.Context) {
	ctx := c.Request.Context()

	// Extract token from the Authorization header
	token := fosite.AccessTokenFromRequest(c.Request)
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No access token provided"})
		return
	}

	// Validate the access token
	_, ar, err := h.provider.OAuth2Provider.IntrospectToken(ctx, token, fosite.AccessToken, &openid.DefaultSession{})
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid access token"})
		return
	}

	// Check if the token has openid scope
	if !ar.GetGrantedScopes().Has("openid") {
		c.JSON(http.StatusForbidden, gin.H{"error": "Token does not have openid scope"})
		return
	}

	// Get user info from the session
	session := ar.GetSession().(*openid.DefaultSession)
	userInfo := map[string]interface{}{
		"sub":                session.GetSubject(),
		"email":              session.Claims.Extra["email"],
		"email_verified":     session.Claims.Extra["email_verified"],
		"preferred_username": session.Claims.Extra["preferred_username"],
		"name":               session.Claims.Extra["name"],
		"updated_at":         session.Claims.IssuedAt.Unix(),
	}

	// Add image field if picture is available in claims
	if picture, exists := session.Claims.Extra["picture"]; exists && picture != nil {
		userInfo["image"] = picture
	}

	c.JSON(http.StatusOK, userInfo)
}
