package oauth2

import (
	"context"
	"system/configs"
	"system/internal/domain"
	"system/internal/oauth2/storage"
	"time"

	"github.com/ory/fosite"
	"github.com/ory/fosite/compose"
)

// NewOAuth2Provider khởi tạo và cấu hình một Fosite OAuth2 provider hoàn chỉnh.
func NewOAuth2Provider(store *storage.HybridStore, cfg *configs.OAuthConfig) fosite.OAuth2Provider {

	// Cấu hình chung cho Fosite
	fositeConfig := &fosite.Config{
		AccessTokenLifespan:   time.Hour * 1,
		AuthorizeCodeLifespan: time.Minute * 10,
		RefreshTokenLifespan:  time.Hour * 24 * 30,
		IDTokenLifespan:       time.Hour * 1,
		IDTokenIssuer:         cfg.Issuer,
		AccessTokenIssuer:     cfg.Issuer,
		GlobalSecret:          []byte(cfg.HMACSecret),
		EnforcePKCE:           true,
		RefreshTokenScopes:    []string{string(domain.ScopeOfflineAccess), string(domain.ScopeOffline)},
		HashCost:              12,
	}

	// Tạo một strategy chung, kết hợp cả HMAC (cho access token) và OIDC (dùng private key cho id token)
	strategy := &compose.CommonStrategy{
		CoreStrategy: compose.NewOAuth2HMACStrategy(fositeConfig),
		OpenIDConnectTokenStrategy: compose.NewOpenIDConnectStrategy(
			func(ctx context.Context) (any, error) {
				// Trả về private key RSA/EC dùng ký ID Token
				return cfg.PrivateKey, nil
			},
			fositeConfig,
		),
		// Nếu bạn muốn access token là JWT + stateless introspection,
		// có thể bổ sung jwt.Signer / NewOAuth2JWTStrategy ở đây (không bắt buộc).
	}

	// Tự lắp ráp provider với các factory cần thiết, loại bỏ ROPC
	return compose.Compose(
		fositeConfig,
		store,
		strategy, // Tham số thứ 3 là strategy, không phải private key

		// Liệt kê các factory cần thiết
		compose.OAuth2AuthorizeExplicitFactory,
		compose.OAuth2RefreshTokenGrantFactory,
		compose.OAuth2ClientCredentialsGrantFactory,
		compose.OAuth2TokenIntrospectionFactory,
		compose.OAuth2TokenRevocationFactory,
		compose.OpenIDConnectExplicitFactory,
		compose.OAuth2PKCEFactory,
	)
}
