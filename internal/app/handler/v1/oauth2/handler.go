package oauth2

import (
	"context"
	"encoding/base64"
	"math/big"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"system/configs"
	"system/internal/domain"
	"system/internal/pkg/service"
	"system/pkg/util/errcode"
	"system/pkg/util/response"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid/v5"
	"github.com/ory/fosite"
	"github.com/ory/fosite/handler/openid"
	"github.com/ory/fosite/token/jwt"
	"go.uber.org/zap"
)

// OAuth2Service interface định nghĩa business logic cho OAuth2
type OAuth2Service interface {
	AuthenticateUser(ctx context.Context, identifier, password string) (*domain.User, error)
	CreateUserSession(ctx context.Context, userID uuid.UUID, ttl time.Duration, userAgent, ip string) (string, error)
	GetUserSession(ctx context.Context, sessionID string) (string, error)
	GetUserByID(ctx context.Context, userID uuid.UUID) (*domain.User, error)
	CheckUserConsent(ctx context.Context, userID, clientID uuid.UUID) (bool, error)
	CreateUserConsent(ctx context.Context, userID, clientID uuid.UUID, scopes []string) error
	LogoutUser(ctx context.Context, sessionID string, revokeTokens bool) error
	RevokeUserTokens(ctx context.Context, userID uuid.UUID) error
	GetClientInfo(ctx context.Context, clientID uuid.UUID) (*domain.OAuth2Client, error)
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	GetUserByIdentifier(ctx context.Context, identifier string) (*domain.User, error)
	GetGlobalPermissions(ctx context.Context, userID uuid.UUID) ([]string, error)
	GetOrganizationPermissions(ctx context.Context, userID, organizationID uuid.UUID) ([]string, error)
	GetGlobalRoles(ctx context.Context, userID uuid.UUID) ([]string, error)
	GetOrganizationRoles(ctx context.Context, userID, organizationID uuid.UUID) ([]string, error)
}

// Handler là struct chứa các dependencies cho OAuth2 handlers.
type Handler struct {
	config          *configs.OAuthConfig
	provider        fosite.OAuth2Provider
	oauth2Service   OAuth2Service
	authRequestRepo domain.AuthRequestRepository
	clientRepo      domain.OAuth2ClientRepository
	webauthnService service.WebAuthnService
	logger          *zap.Logger
}

// NewHandler khởi tạo một OAuth2 handler mới.
func NewHandler(
	cfg *configs.OAuthConfig,
	provider fosite.OAuth2Provider,
	oauth2Service OAuth2Service,
	authRequestRepo domain.AuthRequestRepository,
	clientRepo domain.OAuth2ClientRepository,
	webauthnService service.WebAuthnService,
	logger *zap.Logger,
) *Handler {
	return &Handler{
		config:          cfg,
		provider:        provider,
		oauth2Service:   oauth2Service,
		authRequestRepo: authRequestRepo,
		clientRepo:      clientRepo,
		webauthnService: webauthnService,
		logger:          logger,
	}
}

// writeOAuth2Error ghi OAuth2 error response theo RFC 6749
// Nếu request từ browser (Accept: text/html), hiển thị error page
// Nếu request từ API client (Accept: application/json), trả JSON
func writeOAuth2Error(c *gin.Context, err error) {
	fositeErr := fosite.ErrorToRFC6749Error(err)

	// Check if browser request (Accept header contains text/html)
	acceptHeader := c.GetHeader("Accept")
	isHtmlRequest := strings.Contains(acceptHeader, "text/html")
	isHtmxRequest := c.GetHeader("HX-Request") == "true"

	// For browser requests (not HTMX), show HTML error page
	if isHtmlRequest && !isHtmxRequest {
		c.HTML(fositeErr.CodeField, "oauth2/error.html", gin.H{
			"Error":            fositeErr.ErrorField,
			"ErrorDescription": fositeErr.DescriptionField,
			"Hint":             fositeErr.HintField,
			"RedirectURI":      c.Query("redirect_uri"),
			"ClientID":         c.Query("client_id"),
		})
		return
	}

	// For API/HTMX requests, return JSON
	c.JSON(fositeErr.CodeField, fositeErr)
}

// showErrorPage hiển thị error page cho user khi không thể redirect về client
func (h *Handler) showErrorPage(c *gin.Context, errorCode, errorDescription, hint string) {
	// Try to get redirect_uri from current request to show "Return to app" button
	redirectURI := c.Query("redirect_uri")
	clientID := c.Query("client_id")

	// Clear stale session cookie để user có thể start fresh
	// (Expired auth requests often have stale sessions)
	c.SetCookie(
		"session_id",
		"",
		-1,    // Expire immediately
		"/",   // path
		"",    // domain
		false, // secure
		true,  // httpOnly
	)

	c.HTML(http.StatusBadRequest, "oauth2/error.html", gin.H{
		"Error":            errorCode,
		"ErrorDescription": errorDescription,
		"Hint":             hint,
		"RedirectURI":      redirectURI,
		"ClientID":         clientID,
	})
}

// redirectWithError redirect về client với OAuth2 error
func (h *Handler) redirectWithError(c *gin.Context, redirectURI, state, errorCode, errorDescription string) {
	// Build error URL theo RFC 6749
	errorURL := redirectURI + "?error=" + url.QueryEscape(errorCode)
	if errorDescription != "" {
		errorURL += "&error_description=" + url.QueryEscape(errorDescription)
	}
	if state != "" {
		errorURL += "&state=" + url.QueryEscape(state)
	}

	h.logger.Warn("Redirecting to client with error",
		zap.String("error", errorCode),
		zap.String("redirect_uri", redirectURI),
	)
	c.Redirect(http.StatusFound, errorURL)
}

// ... (Discovery handler is the same)

// UserInfo xử lý endpoint /oauth2/userinfo.
func (h *Handler) UserInfo(c *gin.Context) {
	// ----- THAY THẾ LOGIC GIẢ LẬP BẰNG FOSITE INTROSPECTION -----
	authHeader := c.GetHeader("Authorization")
	token := strings.TrimPrefix(authHeader, "Bearer ")

	if token == "" {
		response.Error(c, http.StatusUnauthorized, errcode.AuthInvalidToken.String(), "auth.invalid_token", nil)
		return
	}

	// IntrospectToken sẽ xác thực token, kiểm tra thời gian sống, và trả về thông tin chi tiết.
	// Tham số thứ 3 là session, Fosite sẽ điền thông tin vào đây.
	// Tham số thứ 4 là các scope bắt buộc, ở đây ta không cần, ta chỉ cần biết scope đã được cấp.
	_, ar, err := h.provider.IntrospectToken(c.Request.Context(), token, fosite.AccessToken, &openid.DefaultSession{})
	if err != nil {
		response.Error(c, http.StatusUnauthorized, errcode.AuthInvalidToken.String(), "auth.invalid_token", gin.H{"debug": err.Error()})
		return
	}

	// Lấy thông tin thật từ session đã được xác thực
	userID := ar.GetSession().GetSubject()
	scopes := ar.GetGrantedScopes()

	// Parse userID từ string sang UUID
	userUUID, err := uuid.FromString(userID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, errcode.ValidationInvalidUserID.String(), "validation.invalid_user_id", nil)
		return
	}

	// Lấy user từ service layer
	user, err := h.oauth2Service.GetUserByID(c.Request.Context(), userUUID)
	if err != nil {
		response.Error(c, http.StatusNotFound, errcode.UserNotFound.String(), "resource.not_found", gin.H{"resource": "user"})
		return
	}

	// Xây dựng claims dựa trên scope và trả về theo chuẩn OIDC (không bọc success/data)
	claims := buildUserInfoClaims(user, scopes)

	// Inject extra claims from session (roles, permissions) if internal scope is present
	if session, ok := ar.GetSession().(*openid.DefaultSession); ok {
		if session.Claims != nil && session.Claims.Extra != nil {
			for k, v := range session.Claims.Extra {
				claims[k] = v
			}
		}
	}

	c.JSON(http.StatusOK, claims)
}

// buildUserInfoClaims lọc và xây dựng các claims cho UserInfo dựa trên scope được cấp phép.
func buildUserInfoClaims(user *domain.User, scopes []string) map[string]any {
	claims := make(map[string]any)
	claims["sub"] = user.ID.String() // Subject claim luôn có theo chuẩn OpenID

	scopeMap := make(map[string]bool)
	for _, s := range scopes {
		scopeMap[s] = true
	}

	if scopeMap[string(domain.ScopeProfile)] {
		claims["name"] = user.FullName
		if user.AvatarURL != nil {
			claims["picture"] = *user.AvatarURL
		}
	}

	if scopeMap[string(domain.ScopeEmail)] {
		claims["email"] = user.Email
		claims["email_verified"] = user.EmailVerified
	}

	return claims
}

// Discovery xử lý endpoint /.well-known/openid-configuration.
// Nó trả về metadata của OpenID Connect provider.
func (h *Handler) Discovery(c *gin.Context) {
	baseURL := h.config.Issuer

	discovery := map[string]any{
		"issuer":                                baseURL,
		"authorization_endpoint":                baseURL + "/oauth2/auth",
		"token_endpoint":                        baseURL + "/oauth2/token",
		"userinfo_endpoint":                     baseURL + "/oauth2/userinfo",
		"jwks_uri":                              baseURL + "/.well-known/jwks.json",
		"introspection_endpoint":                baseURL + "/oauth2/introspect",
		"revocation_endpoint":                   baseURL + "/oauth2/revoke",
		"end_session_endpoint":                  baseURL + "/oauth2/logout",
		"response_types_supported":              []string{"code", "token", "id_token", "code id_token"},
		"grant_types_supported":                 []string{"authorization_code", "refresh_token", "client_credentials"},
		"subject_types_supported":               []string{"public"},
		"id_token_signing_alg_values_supported": []string{"RS256"},
		"scopes_supported": []string{
			"openid", "profile", "email", "offline_access", "internal",
		},
		"token_endpoint_auth_methods_supported": []string{
			"client_secret_basic",
			"client_secret_post",
		},
		"code_challenge_methods_supported": []string{"S256"},
		"claims_supported": []string{
			"sub", "iss", "aud", "exp", "iat",
			"email", "email_verified", "name", "picture",
		},
	}

	c.JSON(http.StatusOK, discovery)
}

// JWKS xử lý endpoint /.well-known/jwks.json.
// Nó trả về public key của server dưới dạng JSON Web Key Set (JWKS).
func (h *Handler) JWKS(c *gin.Context) {
	if h.config.PrivateKey == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Key not configured"})
		return
	}

	publicKey := &h.config.PrivateKey.PublicKey

	jwks := map[string]any{
		"keys": []map[string]any{
			{
				"kty": "RSA",
				"use": "sig",
				"kid": h.config.KeyID,
				"n":   base64.RawURLEncoding.EncodeToString(publicKey.N.Bytes()),
				"e":   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(publicKey.E)).Bytes()),
			},
		},
	}

	c.JSON(http.StatusOK, jwks)
}

// Token xử lý endpoint /oauth2/token
// Endpoint này xử lý tất cả các grant types: authorization_code, refresh_token, client_credentials
func (h *Handler) Token(c *gin.Context) {
	ctx := c.Request.Context()

	// Tạo session mới cho request này
	session := &openid.DefaultSession{}

	// Parse và xác thực request từ client
	accessRequest, err := h.provider.NewAccessRequest(ctx, c.Request, session)
	if err != nil {
		// Log chi tiết để tìm nguyên nhân 500
		zap.L().Error("oauth2 token: failed to build access request",
			zap.Error(err),
			zap.String("grant_type", c.PostForm("grant_type")),
			zap.String("client_id", c.PostForm("client_id")),
			zap.String("code_present", strconv.FormatBool(c.PostForm("code") != "")),
			zap.String("code_verifier_present", strconv.FormatBool(c.PostForm("code_verifier") != "")),
		)
		writeOAuth2Error(c, err)
		return
	}

	// Kiểm tra grant type và xử lý tương ứng
	grantType := accessRequest.GetRequestForm().Get("grant_type")

	switch grantType {
	case "authorization_code":
		// Authorization code grant - được xử lý tự động bởi Fosite
		// Fosite sẽ:
		// 1. Validate authorization code
		// 2. Verify code verifier (PKCE)
		// 3. Check redirect_uri match
		// 4. Generate access token và refresh token

	case "refresh_token":
		// Refresh token grant - được xử lý tự động bởi Fosite
		// Fosite sẽ:
		// 1. Validate refresh token
		// 2. Check token expiration
		// 3. Generate new access token
		// 4. Optionally rotate refresh token

	case "client_credentials":
		// Client credentials grant - được xử lý tự động bởi Fosite
		// Fosite sẽ:
		// 1. Validate client credentials
		// 2. Generate access token (không có refresh token)
		// Note: Không có user context (subject) trong grant type này

	default:
		h.logger.Error("Unsupported grant type",
			zap.String("grant_type", grantType),
		)
		writeOAuth2Error(c, fosite.ErrUnsupportedGrantType)
		return
	}

	// Tạo access token response
	// Fosite sẽ tự động generate tokens dựa trên grant type
	response, err := h.provider.NewAccessResponse(ctx, accessRequest)
	if err != nil {
		clientID := ""
		if accessRequest.GetClient() != nil {
			clientID = accessRequest.GetClient().GetID()
		}
		zap.L().Error("oauth2 token: failed to build access response",
			zap.Error(err),
			zap.String("grant_type", grantType),
			zap.String("client_id", clientID),
		)
		writeOAuth2Error(c, err)
		return
	}

	h.logger.Info("Token issued successfully",
		zap.String("grant_type", grantType),
		zap.String("client_id", accessRequest.GetClient().GetID()),
	)

	// Ghi response về client (JSON format)
	h.provider.WriteAccessResponse(ctx, c.Writer, accessRequest, response)
}

// Authorize xử lý endpoint /oauth2/auth
// Đây là bước đầu tiên trong Authorization Code Flow
func (h *Handler) Authorize(c *gin.Context) {
	ctx := c.Request.Context()
	requestID := c.Query("request_id")

	var ar fosite.AuthorizeRequester
	var err error

	if requestID != "" && c.Query("client_id") == "" {
		// Quay lại từ login/consent với request_id, lấy lại request đã lưu
		ar, err = h.authRequestRepo.GetAuthRequest(ctx, requestID)
		if err != nil {
			zap.L().Error("oauth2 authorize: failed to restore request by request_id",
				zap.String("request_id", requestID),
				zap.Error(err),
			)
			writeOAuth2Error(c, fosite.ErrInvalidRequest.WithWrap(err).WithDebug("Invalid or expired authorization request"))
			return
		}
	} else {
		// Parse authorization request từ query parameters
		ar, err = h.provider.NewAuthorizeRequest(ctx, c.Request)
		if err != nil {
			// Log chi tiết để debug lỗi server_error
			zap.L().Error("oauth2 authorize: failed to build request",
				zap.Error(err),
				zap.String("client_id", c.Query("client_id")),
				zap.String("redirect_uri", c.Query("redirect_uri")),
				zap.String("scope", c.Query("scope")),
				zap.String("response_type", c.Query("response_type")),
			)
			writeOAuth2Error(c, err)
			return
		}
	}

	// Kiểm tra xem user đã đăng nhập chưa (check session cookie)
	userID, authenticated := h.checkUserAuthentication(c)

	if !authenticated {
		// User chưa đăng nhập - redirect đến login page
		h.logger.Info("User not authenticated, redirecting to login",
			zap.String("client_id", ar.GetClient().GetID()),
			zap.String("request_id", ar.GetID()),
		)
		// Lưu authorization request vào session để resume sau khi login
		h.redirectToLogin(c, ar)
		return
	}

	h.logger.Info("User authenticated",
		zap.String("user_id", userID),
		zap.String("client_id", ar.GetClient().GetID()),
	)

	// Check if we need to show passkey prompt from Redis
	showPasskeyPrompt, err := h.authRequestRepo.GetPasskeyPromptFlag(c.Request.Context(), ar.GetID())
	if err != nil {
		h.logger.Warn("Failed to get passkey prompt flag",
			zap.String("request_id", ar.GetID()),
			zap.Error(err),
		)
		showPasskeyPrompt = false
	}

	// Fix: Double check if user actually has passkeys now (maybe they just created one)
	if showPasskeyPrompt {
		if userUUID, err := uuid.FromString(userID); err == nil {
			if hasPasskeys, _ := h.webauthnService.HasPasskeys(c.Request.Context(), userUUID); hasPasskeys {
				showPasskeyPrompt = false
				// Update flag in Redis to avoid re-showing
				_ = h.authRequestRepo.SavePasskeyPromptFlag(c.Request.Context(), ar.GetID(), false, time.Minute*10)
			}
		}
	}

	// User đã đăng nhập - kiểm tra xem đã consent chưa
	consentGiven := h.checkUserConsent(c, userID, ar.GetClient().GetID())

	// Always show consent page if passkey prompt is needed, even if consent was previously given
	if !consentGiven || showPasskeyPrompt {
		// Chưa consent hoặc cần hiển thị passkey prompt - redirect đến consent page
		h.redirectToConsent(c, ar, userID)
		return
	}

	// User đã authenticate và consent - tạo authorization response
	h.finalizeAuthorization(c, ar, userID)
}

// checkUserAuthentication kiểm tra xem user đã đăng nhập chưa
func (h *Handler) checkUserAuthentication(c *gin.Context) (string, bool) {
	// Kiểm tra session cookie
	sessionID, err := c.Cookie("session_id")
	if err != nil || sessionID == "" {
		return "", false
	}

	// Validate session với Redis thông qua service
	userID, err := h.oauth2Service.GetUserSession(c.Request.Context(), sessionID)
	if err != nil {
		return "", false
	}

	return userID, true
}

// checkUserConsent kiểm tra xem user đã consent cho client này chưa
func (h *Handler) checkUserConsent(c *gin.Context, userID string, clientID string) bool {
	// Parse UUID
	userUUID, err := uuid.FromString(userID)
	if err != nil {
		return false
	}

	clientUUID, err := uuid.FromString(clientID)
	if err != nil {
		return false
	}

	// Kiểm tra consent thông qua service
	hasConsent, err := h.oauth2Service.CheckUserConsent(c.Request.Context(), userUUID, clientUUID)
	if err != nil {
		return false
	}

	return hasConsent
}

// redirectToLogin redirect user đến login page
func (h *Handler) redirectToLogin(c *gin.Context, ar fosite.AuthorizeRequester) {
	// Save authorization request to Redis để resume sau khi login
	requestID := ar.GetID()

	// Save full auth request object
	err := h.authRequestRepo.SaveAuthRequest(c.Request.Context(), requestID, ar, time.Minute*10)
	if err != nil {
		zap.L().Error("oauth2 authorize: failed to save auth request before login",
			zap.String("request_id", requestID),
			zap.Error(err),
		)
		writeOAuth2Error(c, fosite.ErrServerError.WithWrap(err).WithDebug("Failed to save auth request"))
		return
	}

	// Also save query params for potential use in other pages
	originalParams := c.Request.URL.RawQuery
	err = h.authRequestRepo.SaveQueryParams(c.Request.Context(), requestID, originalParams, time.Minute*10)
	if err != nil {
		h.logger.Warn("Failed to save query params",
			zap.String("error", err.Error()),
			zap.String("request_id", requestID),
		)
		// Don't fail if query params save fails, auth request is already saved
	}

	// Redirect đến login page với request_id
	loginURL := "/oauth2/login?request_id=" + requestID
	h.logger.Info("Redirecting to login page",
		zap.String("request_id", requestID),
		zap.String("login_url", loginURL),
	)
	c.Redirect(http.StatusFound, loginURL)
}

// redirectToConsent redirect user đến consent page
func (h *Handler) redirectToConsent(c *gin.Context, ar fosite.AuthorizeRequester, userID string) {
	requestID := ar.GetID()

	// Lưu authorization request với userID vào Redis
	err := h.authRequestRepo.SaveAuthRequestWithUserID(c.Request.Context(), requestID, ar, userID, time.Minute*10)
	if err != nil {
		zap.L().Error("oauth2 authorize: failed to save auth request before consent",
			zap.String("request_id", requestID),
			zap.Error(err),
		)
		writeOAuth2Error(c, fosite.ErrServerError.WithWrap(err).WithDebug("Failed to save auth request"))
		return
	}

	// IMPORTANT: Chỉ lưu query params nếu đây là request gốc (có client_id)
	// Nếu đây là resume request (chỉ có request_id), ta không nên overwrite params cũ trong Redis
	if c.Query("client_id") != "" {
		originalParams := c.Request.URL.RawQuery
		err = h.authRequestRepo.SaveQueryParams(c.Request.Context(), requestID, originalParams, time.Minute*10)
		if err != nil {
			h.logger.Error("Failed to save query params for consent",
				zap.String("error", err.Error()),
				zap.String("request_id", requestID),
			)
			writeOAuth2Error(c, fosite.ErrServerError.WithWrap(err).WithDebug("Failed to save query params"))
			return
		}
	}

	// Redirect đến consent page
	consentURL := "/oauth2/consent?request_id=" + requestID
	c.Redirect(http.StatusFound, consentURL)
}

// finalizeAuthorization hoàn tất quá trình authorization và tạo response
// finalizeAuthorization hoàn tất quá trình authorization và tạo response
func (h *Handler) finalizeAuthorization(c *gin.Context, ar fosite.AuthorizeRequester, userID string) {
	ctx := c.Request.Context()

	// Parse userID
	userUUID, err := uuid.FromString(userID)
	if err != nil {
		zap.L().Error("oauth2 authorize: invalid user id", zap.Error(err))
		writeOAuth2Error(c, fosite.ErrServerError.WithDebug("Invalid user ID"))
		return
	}

	// Lấy thông tin client để check IsInternal
	clientID := ar.GetClient().GetID()
	clientUUID, err := uuid.FromString(clientID)
	if err != nil {
		zap.L().Error("oauth2 authorize: invalid client id", zap.Error(err))
		writeOAuth2Error(c, fosite.ErrServerError.WithDebug("Invalid client ID"))
		return
	}

	client, err := h.oauth2Service.GetClientInfo(ctx, clientUUID)
	if err != nil {
		zap.L().Error("oauth2 authorize: failed to get client info", zap.Error(err))
		writeOAuth2Error(c, fosite.ErrServerError.WithDebug("Failed to get client info"))
		return
	}

	// Tạo session với user info
	now := time.Now().UTC()
	session := &openid.DefaultSession{
		Subject: userID,
		Claims: &jwt.IDTokenClaims{
			Subject:     userID,
			AuthTime:    now,
			RequestedAt: ar.GetRequestedAt(),
			Extra:       make(map[string]any),
		},
		Headers: &jwt.Headers{},
		ExpiresAt: map[fosite.TokenType]time.Time{
			fosite.AuthorizeCode: now.Add(10 * time.Minute),
			fosite.AccessToken:   now.Add(time.Hour),
			fosite.RefreshToken:  now.Add(24 * time.Hour * 30),
			fosite.IDToken:       now.Add(time.Hour),
		},
	}

	// Populate session với các claims cần thiết
	ar.SetSession(session)

	// V1: chấp thuận toàn bộ scopes đã request (đã kiểm tra consent trước đó)
	// Check scope "internal" logic
	for _, scope := range ar.GetRequestedScopes() {
		if scope == string(domain.ScopeInternal) {
			// Chỉ cấp scope internal nếu client là internal
			if !client.IsInternal {
				h.logger.Warn("Client requested internal scope but is not internal",
					zap.String("client_id", clientID),
				)
				continue // Skip granting this scope
			}

			// Inject roles & permissions
			// 1. Global Permissions
			perms, err := h.oauth2Service.GetGlobalPermissions(ctx, userUUID)
			if err == nil && len(perms) > 0 {
				session.Claims.Extra["permissions"] = perms
			}

			// 2. Global Roles
			roles, err := h.oauth2Service.GetGlobalRoles(ctx, userUUID)
			if err == nil && len(roles) > 0 {
				session.Claims.Extra["roles"] = roles
			}

			// 3. Tenant Permissions & Roles (if client is tenant-specific)
			if client.OrganizationID != nil {
				organizationPerms, err := h.oauth2Service.GetOrganizationPermissions(ctx, userUUID, *client.OrganizationID)
				if err == nil && len(organizationPerms) > 0 {
					session.Claims.Extra["tenant_permissions"] = organizationPerms
				}

				organizationRoles, err := h.oauth2Service.GetOrganizationRoles(ctx, userUUID, *client.OrganizationID)
				if err == nil && len(organizationRoles) > 0 {
					session.Claims.Extra["tenant_roles"] = organizationRoles
				}
			}
		}
		ar.GrantScope(scope)
	}

	// Tạo authorization response (authorization code)
	response, err := h.provider.NewAuthorizeResponse(ctx, ar, session)
	if err != nil {
		zap.L().Error("oauth2 authorize: failed to create authorize response",
			zap.String("request_id", ar.GetID()),
			zap.Error(err),
		)
		writeOAuth2Error(c, err)
		return
	}

	h.logger.Info("Authorization code issued successfully",
		zap.String("user_id", userID),
		zap.String("client_id", ar.GetClient().GetID()),
	)

	// Write response - redirect user về client với authorization code
	h.provider.WriteAuthorizeResponse(ctx, c.Writer, ar, response)
}

// LoginPage hiển thị trang đăng nhập
func (h *Handler) LoginPage(c *gin.Context) {
	requestID := c.Query("request_id")

	// TODO: Validate requestID và lấy thông tin client từ stored request
	// ar, err := h.store.GetAuthRequest(c.Request.Context(), requestID)

	c.HTML(http.StatusOK, "oauth2/login.html", gin.H{
		"RequestID": requestID,
	})
}

// LoginSubmit xử lý form submit từ trang đăng nhập
func (h *Handler) LoginSubmit(c *gin.Context) {
	requestID := c.PostForm("request_id")
	email := c.PostForm("email")
	password := c.PostForm("password")

	// Validate input
	if email == "" || password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email and password are required", "message": "Email and password are required"})
		return
	}

	// Authenticate user thông qua service
	user, err := h.oauth2Service.AuthenticateUser(c.Request.Context(), email, password)
	if err != nil {
		h.logger.Warn("Login authentication failed",
			zap.String("email", email),
			zap.String("request_id", requestID),
			zap.Error(err),
		)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password", "message": "Invalid email or password"})
		return
	}

	// Create session in Redis thông qua service
	sessionID, err := h.oauth2Service.CreateUserSession(
		c.Request.Context(),
		user.ID,
		time.Hour,
		c.Request.UserAgent(),
		c.ClientIP(),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session", "message": "Failed to create session"})
		return
	}

	// Set secure session cookie
	c.SetCookie(
		"session_id",
		sessionID,
		3600,  // 1 hour
		"/",   // path
		"",    // domain
		false, // secure (set true in production with HTTPS)
		true,  // httpOnly
	)

	// Check if user has any passkeys
	hasPasskeys, err := h.webauthnService.HasPasskeys(c.Request.Context(), user.ID)
	if err != nil {
		h.logger.Warn("Failed to check user passkeys",
			zap.String("user_id", user.ID.String()),
			zap.Error(err),
		)
		// Continue with normal flow even if check fails
		hasPasskeys = true
	}

	// If user doesn't have passkeys, save flag to Redis
	if !hasPasskeys {
		err := h.authRequestRepo.SavePasskeyPromptFlag(c.Request.Context(), requestID, true, time.Minute*10)
		if err != nil {
			h.logger.Warn("Failed to save passkey prompt flag",
				zap.String("request_id", requestID),
				zap.Error(err),
			)
		} else {
			h.logger.Info("Login successful, will show passkey setup prompt",
				zap.String("request_id", requestID),
				zap.String("user_id", user.ID.String()),
			)
		}
	}

	// Redirect back to /oauth2/auth with just request_id
	// The Authorize handler will detect this and load full params from Redis
	redirectURL := "/oauth2/auth?request_id=" + requestID

	h.logger.Info("Login successful, redirecting to authorization endpoint",
		zap.String("request_id", requestID),
		zap.String("user_id", user.ID.String()),
		zap.String("redirect_url", redirectURL),
	)

	// Check if request is from HTMX - return JSON response for client-side redirect
	if c.GetHeader("HX-Request") == "true" {
		c.JSON(http.StatusOK, gin.H{
			"success":      true,
			"redirect_url": redirectURL,
		})
		return
	}

	// Regular form submit - do server-side redirect
	c.Redirect(http.StatusFound, redirectURL)
}

// LoginCheck kiểm tra email/username và trạng thái passkey của user
func (h *Handler) LoginCheck(c *gin.Context) {
	identifier := c.PostForm("email") // form field still named 'email' for compatibility
	if identifier == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email or username is required"})
		return
	}

	// Check if user exists by email or username
	user, err := h.oauth2Service.GetUserByIdentifier(c.Request.Context(), identifier)
	if err != nil {
		// User not found - return generic response to avoid enumeration
		c.JSON(http.StatusOK, gin.H{
			"exists":       false,
			"has_passkeys": false,
		})
		return
	}

	// Check if user has passkeys
	hasPasskeys, err := h.webauthnService.HasPasskeys(c.Request.Context(), user.ID)
	if err != nil {
		h.logger.Error("Failed to check passkeys", zap.Error(err))
		// Default to false on error
		hasPasskeys = false
	}

	c.JSON(http.StatusOK, gin.H{
		"exists":       true,
		"has_passkeys": hasPasskeys,
	})
}


// ScopeInfo chứa thông tin về scope
type ScopeInfo struct {
	Name        string
	Description string
}

// ConsentPage hiển thị trang consent (chấp thuận quyền)
func (h *Handler) ConsentPage(c *gin.Context) {
	requestID := c.Query("request_id")

	// 1. Check user authentication - MUST be logged in to see consent page
	sessionID, err := c.Cookie("session_id")
	if err != nil || sessionID == "" {
		h.logger.Warn("Unauthenticated access to consent page",
			zap.String("request_id", requestID),
		)
		// Redirect to login with request_id to resume flow
		c.Redirect(http.StatusFound, "/oauth2/login?request_id="+requestID)
		return
	}

	// 2. Validate session
	userID, err := h.oauth2Service.GetUserSession(c.Request.Context(), sessionID)
	if err != nil {
		h.logger.Warn("Invalid or expired session on consent page",
			zap.String("error", err.Error()),
			zap.String("request_id", requestID),
		)
		// Session expired - redirect to login
		c.Redirect(http.StatusFound, "/oauth2/login?request_id="+requestID)
		return
	}

	// 3. Load query params từ Redis
	queryParams, err := h.authRequestRepo.GetQueryParams(c.Request.Context(), requestID)
	if err != nil {
		h.logger.Error("Failed to load query params for consent page",
			zap.String("error", err.Error()),
			zap.String("request_id", requestID),
			zap.String("user_id", userID),
		)
		// Authorization request expired
		h.showErrorPage(c,
			"invalid_request",
			"Your authorization request has expired. Please try again.",
			"Authorization requests are valid for 10 minutes. Click below to start over.",
		)
		return
	}

	// 4. Parse query params
	values, err := url.ParseQuery(queryParams)
	if err != nil {
		h.logger.Error("Failed to parse query params",
			zap.String("error", err.Error()),
			zap.String("request_id", requestID),
		)
		h.showErrorPage(c, "invalid_request", "Invalid authorization request", "")
		return
	}

	clientID := values.Get("client_id")
	requestedScopes := strings.Split(values.Get("scope"), " ")

	// 5. Load client info từ database
	clientUUID, err := uuid.FromString(clientID)
	if err != nil {
		h.showErrorPage(c, "invalid_request", "Invalid client ID", "")
		return
	}

	// Get full client info from service to access ClientName
	domainClient, err := h.oauth2Service.GetClientInfo(c.Request.Context(), clientUUID)
	if err != nil {
		h.logger.Error("Failed to load client for consent page",
			zap.String("error", err.Error()),
			zap.String("client_id", clientID),
		)
		h.showErrorPage(c, "unauthorized_client", "Client not found or inactive", "")
		return
	}

	// 6. Build scope info list với real scopes
	scopes := h.buildScopeInfoList(requestedScopes)

	// 7. Check if we should show passkey setup prompt from Redis
	showPasskeyPrompt, err := h.authRequestRepo.GetPasskeyPromptFlag(c.Request.Context(), requestID)
	if err != nil {
		h.logger.Warn("Failed to get passkey prompt flag",
			zap.String("request_id", requestID),
			zap.Error(err),
		)
		showPasskeyPrompt = false
	}

	// Fix: Double check if user actually has passkeys now
	if showPasskeyPrompt {
		if userUUID, err := uuid.FromString(userID); err == nil {
			if hasPasskeys, _ := h.webauthnService.HasPasskeys(c.Request.Context(), userUUID); hasPasskeys {
				showPasskeyPrompt = false
				// Update flag in Redis
				_ = h.authRequestRepo.SavePasskeyPromptFlag(c.Request.Context(), requestID, false, time.Minute*10)
			}
		}
	}

	// 8. Show consent page với real data
	h.logger.Info("Showing consent page",
		zap.String("user_id", userID),
		zap.String("client_id", clientID),
		zap.String("client_name", domainClient.ClientName),
		zap.Strings("scopes", requestedScopes),
		zap.Bool("show_passkey_prompt", showPasskeyPrompt),
	)

	c.HTML(http.StatusOK, "oauth2/consent.html", gin.H{
		"RequestID":          requestID,
		"ClientName":         domainClient.ClientName,
		"ClientID":           clientUUID.String(),
		"Scopes":             scopes,
		"ShowPasskeyPrompt":  showPasskeyPrompt,
	})
}

// buildScopeInfoList tạo danh sách ScopeInfo từ scope strings
func (h *Handler) buildScopeInfoList(scopes []string) []ScopeInfo {
	scopeDescriptions := map[string]ScopeInfo{
		"openid":         {"OpenID Connect", "Verify your identity"},
		"profile":        {"Profile Information", "Access your name and profile picture"},
		"email":          {"Email Address", "Access your email address"},
		"offline_access": {"Offline Access", "Access your data when you're not using the app"},
	}

	result := make([]ScopeInfo, 0, len(scopes))
	for _, scope := range scopes {
		info, ok := scopeDescriptions[scope]
		if !ok {
			info = ScopeInfo{
				Name:        scope,
				Description: "Access " + scope,
			}
		}
		result = append(result, info)
	}
	return result
}

// ConsentSubmit xử lý form submit từ trang consent
func (h *Handler) ConsentSubmit(c *gin.Context) {
	requestID := c.PostForm("request_id")
	action := c.PostForm("action")

	// Get user from session
	sessionID, err := c.Cookie("session_id")
	if err != nil || sessionID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	// Get userID from session store thông qua service
	userID, err := h.oauth2Service.GetUserSession(c.Request.Context(), sessionID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Session expired or invalid"})
		return
	}

	// Load original OAuth2 query params từ Redis
	queryParams, err := h.authRequestRepo.GetQueryParams(c.Request.Context(), requestID)
	if err != nil {
		h.logger.Error("Failed to load query params for consent - request expired",
			zap.String("error", err.Error()),
			zap.String("request_id", requestID),
			zap.String("user_id", userID),
		)

		// Authorization request expired (TTL exceeded)
		// Show error page với option để user thử lại
		h.showErrorPage(c,
			"invalid_request",
			"Your authorization request has expired. Please try again.",
			"Authorization requests are valid for 10 minutes. Click below to start over.",
		)
		return
	}

	// Parse query params to get client_id, redirect_uri, state, scopes
	values, err := url.ParseQuery(queryParams)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query parameters"})
		return
	}

	clientID := values.Get("client_id")
	redirectURI := values.Get("redirect_uri")
	state := values.Get("state")
	scopes := strings.Split(values.Get("scope"), " ")

	if action == "deny" {
		// User denied consent - redirect back to client with error
		errorURL := redirectURI + "?error=access_denied&error_description=User+denied+consent&state=" + state
		c.Redirect(http.StatusFound, errorURL)
		return
	}

	// User allowed - save consent thông qua service
	userUUID, _ := uuid.FromString(userID)
	clientUUID, _ := uuid.FromString(clientID)

	err = h.oauth2Service.CreateUserConsent(c.Request.Context(), userUUID, clientUUID, scopes)
	if err != nil {
		h.logger.Error("Failed to save consent",
			zap.String("error", err.Error()),
			zap.String("user_id", userID),
			zap.String("client_id", clientID),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save consent"})
		return
	}

	// Redirect back to /oauth2/auth with request_id only
	// Authorize handler will load full params from Redis
	// User is now authenticated and has given consent

	// Clear passkey prompt flag to prevent redirect loop
	_ = h.authRequestRepo.SavePasskeyPromptFlag(c.Request.Context(), requestID, false, time.Minute*10)

	h.logger.Info("Consent granted, redirecting to authorization endpoint",
		zap.String("request_id", requestID),
		zap.String("user_id", userID),
		zap.String("client_id", clientID),
	)
	c.Redirect(http.StatusFound, "/oauth2/auth?request_id="+requestID)
}

// Revoke xử lý endpoint /oauth2/revoke
// Token Revocation theo RFC 7009
func (h *Handler) Revoke(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse revocation request từ POST body
	// Fosite sẽ tự động:
	// 1. Xác thực client credentials (client_id, client_secret)
	// 2. Parse token parameter
	// 3. Validate token type hint (nếu có)
	err := h.provider.NewRevocationRequest(ctx, c.Request)
	if err != nil {
		writeOAuth2Error(c, err)
		return
	}

	// Revocation thành công - trả về 200 OK theo RFC 7009
	// RFC 7009 Section 2.2: The authorization server responds with HTTP status code 200
	// if the token has been revoked successfully or if the client submitted an invalid token.
	c.JSON(http.StatusOK, gin.H{})
}

// Introspect xử lý endpoint /oauth2/introspect
// Token Introspection theo RFC 7662
func (h *Handler) Introspect(c *gin.Context) {
	ctx := c.Request.Context()

	// Tạo session cho introspection
	session := &openid.DefaultSession{}

	// Parse và xác thực introspection request
	// Fosite sẽ tự động:
	// 1. Xác thực client credentials (chỉ authorized clients mới được introspect)
	// 2. Parse token parameter
	// 3. Validate token type hint (nếu có)
	// 4. Kiểm tra token có tồn tại và còn valid không
	tokenUse, ar, err := h.provider.IntrospectToken(ctx, c.Request.FormValue("token"), fosite.AccessToken, session)
	if err != nil {
		// Token không valid hoặc có lỗi - trả về inactive
		c.JSON(http.StatusOK, gin.H{
			"active": false,
		})
		return
	}

	// Token valid - trả về thông tin chi tiết
	response := gin.H{
		"active":     true,
		"scope":      strings.Join(ar.GetGrantedScopes(), " "),
		"client_id":  ar.GetClient().GetID(),
		"token_type": string(tokenUse),
		"exp":        ar.GetSession().GetExpiresAt(tokenUse).Unix(),
		"iat":        ar.GetRequestedAt().Unix(),
	}

	// Thêm subject nếu có (không có trong client_credentials grant)
	if subject := ar.GetSession().GetSubject(); subject != "" {
		response["sub"] = subject
	}

	// Thêm audience nếu có
	if audiences := ar.GetGrantedAudience(); len(audiences) > 0 {
		response["aud"] = audiences
	}

	c.JSON(http.StatusOK, response)
}

// Logout xử lý endpoint /oauth2/logout
// Implements OpenID Connect RP-Initiated Logout (https://openid.net/specs/openid-connect-rpinitiated-1_0.html)
func (h *Handler) Logout(c *gin.Context) {
	ctx := c.Request.Context()

	// 1. Extract parameters (support both GET and POST)
	idTokenHint := c.Query("id_token_hint")
	if idTokenHint == "" {
		idTokenHint = c.PostForm("id_token_hint")
	}

	postLogoutRedirectURI := c.Query("post_logout_redirect_uri")
	if postLogoutRedirectURI == "" {
		postLogoutRedirectURI = c.PostForm("post_logout_redirect_uri")
	}

	state := c.Query("state")
	if state == "" {
		state = c.PostForm("state")
	}

	// 2. Get session from cookie
	sessionID, err := c.Cookie("session_id")
	var userID string
	revokeTokens := false // Default: không revoke tokens khi logout (chỉ xóa session)

	// 3. Logout user nếu có session
	if err == nil && sessionID != "" {
		// Lấy userID từ session trước khi logout
		if uid, err := h.oauth2Service.GetUserSession(ctx, sessionID); err == nil {
			userID = uid
		}

		// Logout user - xóa session và optionally revoke tokens
		if err := h.oauth2Service.LogoutUser(ctx, sessionID, revokeTokens); err != nil {
			zap.L().Error("oauth2 logout: failed to logout user",
				zap.String("session_id", sessionID),
				zap.Error(err),
			)
		}
	}

	// 4. Optionally revoke all tokens (nếu được enable)
	if userID != "" && revokeTokens {
		userUUID, err := uuid.FromString(userID)
		if err == nil {
			if err := h.oauth2Service.RevokeUserTokens(ctx, userUUID); err != nil {
				zap.L().Error("oauth2 logout: failed to revoke user tokens",
					zap.String("user_id", userID),
					zap.Error(err),
				)
			}
		}
	}

	// 5. Clear session cookie
	c.SetCookie(
		"session_id",
		"",
		-1,    // MaxAge -1 means delete cookie
		"/",   // path
		"",    // domain
		false, // secure (set true in production with HTTPS)
		true,  // httpOnly
	)

	// 6. Redirect về client (nếu có post_logout_redirect_uri)
	if postLogoutRedirectURI != "" {
		// TODO: SECURITY - Validate redirect URI against client's registered post_logout_redirect_uris
		// For now, we trust it but should add validation in production

		redirectURL := postLogoutRedirectURI
		if state != "" {
			// Preserve state parameter
			if strings.Contains(redirectURL, "?") {
				redirectURL += "&state=" + state
			} else {
				redirectURL += "?state=" + state
			}
		}

		zap.L().Info("oauth2 logout: redirecting to post_logout_redirect_uri",
			zap.String("redirect_uri", redirectURL),
		)

		c.Redirect(http.StatusFound, redirectURL)
		return
	}

	// 7. No redirect URI - return success response
	c.JSON(http.StatusOK, gin.H{
		"message": "Logged out successfully",
	})
}
