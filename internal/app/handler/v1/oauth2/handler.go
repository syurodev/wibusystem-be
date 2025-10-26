package oauth2

import (
	"encoding/base64"
	"math/big"
	"net/http"
	"strings"
	"system/configs"
	"system/internal/domain"
	"system/pkg/util/response"

	"github.com/gin-gonic/gin"
	"github.com/ory/fosite"
)

// Handler là struct chứa các dependencies cho OAuth2 handlers.
type Handler struct {
	config   *configs.OAuthConfig
	provider fosite.OAuth2Provider
	store    Store
}

// NewHandler khởi tạo một OAuth2 handler mới.
func NewHandler(cfg *configs.OAuthConfig, provider fosite.OAuth2Provider, store Store) *Handler {
	return &Handler{
		config:   cfg,
		provider: provider,
		store:    store,
	}
}

// ... (Discovery handler is the same)

// UserInfo xử lý endpoint /oauth2/userinfo.
func (h *Handler) UserInfo(c *gin.Context) {
	// ----- THAY THẾ LOGIC GIẢ LẬP BẰNG FOSITE INTROSPECTION -----
	authHeader := c.GetHeader("Authorization")
	token := strings.TrimPrefix(authHeader, "Bearer ")

	if token == "" {
		response.Error(c, http.StatusUnauthorized, "E4011", "auth.invalid_token", nil)
		return
	}

	// IntrospectToken sẽ xác thực token, kiểm tra thời gian sống, và trả về thông tin chi tiết.
	// Tham số thứ 3 là session, Fosite sẽ điền thông tin vào đây.
	// Tham số thứ 4 là các scope bắt buộc, ở đây ta không cần, ta chỉ cần biết scope đã được cấp.
	_, ar, err := h.provider.IntrospectToken(c.Request.Context(), token, fosite.AccessToken, &fosite.DefaultSession{})
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "E4011", "auth.invalid_token", gin.H{"debug": err.Error()})
		return
	}

	// Lấy thông tin thật từ session đã được xác thực
	userID := ar.GetSession().GetSubject()
	scopes := ar.GetGrantedScopes()
	// ----- KẾT THÚC THAY THẾ -----

	user, err := h.store.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "E4041", "resource.not_found", gin.H{"resource": "user"})
		return
	}

	// Xây dựng claims dựa trên scope
	claims := buildUserInfoClaims(user, scopes)

	response.Success(c, http.StatusOK, "userinfo.success", claims, nil)
}

// buildUserInfoClaims lọc và xây dựng các claims cho UserInfo dựa trên scope được cấp phép.
func buildUserInfoClaims(user *User, scopes []string) map[string]any {
	claims := make(map[string]any)
	claims["sub"] = user.ID // Subject claim luôn có theo chuẩn OpenID

	scopeMap := make(map[string]bool)
	for _, s := range scopes {
		scopeMap[s] = true
	}

	if scopeMap[string(domain.ScopeProfile)] {
		claims["name"] = user.Name
		claims["picture"] = user.Picture
	}

	if scopeMap[string(domain.ScopeEmail)] {
		claims["email"] = user.Email
		claims["email_verified"] = true // Giả lập
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
			"openid", "profile", "email", "offline_access",
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
