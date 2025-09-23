package handlers

import (
	"wibusystem/pkg/i18n"
	"wibusystem/services/identify/config"
	"wibusystem/services/identify/oauth2"
	"wibusystem/services/identify/repositories"
	"wibusystem/services/identify/services"
	"wibusystem/services/identify/services/interfaces"
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
func NewHandlers(repos *repositories.Repositories, provider *oauth2.Provider, sess *session.Manager, translator *i18n.Translator, cfg *config.Config, devMode bool, regSecret string, loginPageURL string) *Handlers {
	// Initialize services
	credentialService := services.NewCredentialService(repos, cfg.Security.BCryptCost)
	userService := services.NewUserService(repos)
	tenantService := services.NewTenantService(repos)
	authService := services.NewAuthService(repos, userService, credentialService, tenantService, sess)

	return &Handlers{
		Auth:   NewAuthHandler(authService, provider, translator),
		User:   NewUserHandler(userService, translator),
		Tenant: NewTenantHandler(tenantService, translator),
		OAuth2: oauth2.NewHandlers(provider, repos, sess, translator, devMode, regSecret, loginPageURL),
	}
}

// NewHandlersWithServices creates handlers with pre-initialized services
func NewHandlersWithServices(
	authService interfaces.AuthServiceInterface,
	userService interfaces.UserServiceInterface,
	tenantService interfaces.TenantServiceInterface,
	provider *oauth2.Provider,
	repos *repositories.Repositories,
	sess *session.Manager,
	translator *i18n.Translator,
	devMode bool,
	regSecret string,
	loginPageURL string,
) *Handlers {
	return &Handlers{
		Auth:   NewAuthHandler(authService, provider, translator),
		User:   NewUserHandler(userService, translator),
		Tenant: NewTenantHandler(tenantService, translator),
		OAuth2: oauth2.NewHandlers(provider, repos, sess, translator, devMode, regSecret, loginPageURL),
	}
}
