package handlers

import (
	"wibusystem/pkg/i18n"
	"wibusystem/services/identify/oauth2"
	"wibusystem/services/identify/repositories"
	"wibusystem/services/identify/session"
)

// Handlers holds all handler instances
type Handlers struct {
	Auth   *AuthHandler
	User   *UserHandler
	Tenant *TenantHandler
	OAuth2 *oauth2.Handlers
}

// NewHandlers creates a new handlers manager
func NewHandlers(repos *repositories.Repositories, provider *oauth2.Provider, sess *session.Manager, translator *i18n.Translator, devMode bool, regSecret string, loginPageURL string) *Handlers {
	return &Handlers{
		Auth:   NewAuthHandler(repos, provider, sess, translator),
		User:   NewUserHandler(repos, translator),
		Tenant: NewTenantHandler(repos, translator),
		OAuth2: oauth2.NewHandlers(provider, repos, sess, translator, devMode, regSecret, loginPageURL),
	}
}
