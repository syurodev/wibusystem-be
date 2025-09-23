// Package routes manages all HTTP routing configuration for the Identity Service.
// It provides a clean separation between routing logic and main application bootstrapping.
package routes

import (
	"fmt"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"wibusystem/pkg/database/factory"
	"wibusystem/pkg/database/providers/postgres"
	"wibusystem/pkg/i18n"
	"wibusystem/services/identify/config"
	"wibusystem/services/identify/handlers"
	"wibusystem/services/identify/middleware"
	"wibusystem/services/identify/oauth2"
	"wibusystem/services/identify/repositories"
	v1 "wibusystem/services/identify/routes/api/v1"
	"wibusystem/services/identify/services"
	"wibusystem/services/identify/services/interfaces"
	"wibusystem/services/identify/session"
)

// Dependencies holds all the dependencies needed for setting up routes
type Dependencies struct {
	DBManager   *factory.DatabaseManager
	Config      *config.Config
	Translator  *i18n.Translator
	Handlers    *handlers.Handlers
	Middleware  *middleware.Manager
	Provider    *oauth2.Provider
	UserService interfaces.UserServiceInterface
}

// SetupRouter initializes and configures the main Gin router with all routes and middleware
func SetupRouter(deps *Dependencies) *gin.Engine {
	router := gin.New()

	// Add basic middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// CORS middleware
	router.Use(cors.New(cors.Config{
		AllowOrigins:     deps.Config.Security.CORS.AllowOrigins,
		AllowMethods:     deps.Config.Security.CORS.AllowMethods,
		AllowHeaders:     deps.Config.Security.CORS.AllowHeaders,
		ExposeHeaders:    deps.Config.Security.CORS.ExposeHeaders,
		AllowCredentials: deps.Config.Security.CORS.AllowCredentials,
		MaxAge:           time.Duration(deps.Config.Security.CORS.MaxAge) * time.Second,
	}))

	// Setup common middleware
	deps.Middleware.SetupCommonMiddleware(router)

	// Setup public routes (non-versioned)
	setupPublicRoutes(router, deps)

	// Setup API versioned routes
	v1.SetupV1Routes(router, deps.DBManager, deps.Config, deps.Translator, deps.Handlers, deps.Middleware)

	return router
}

// NewDependencies creates and initializes all dependencies needed for routing
func NewDependencies(dbManager *factory.DatabaseManager, cfg *config.Config, translator *i18n.Translator) (*Dependencies, error) {
	// Get the primary database
	primary := dbManager.GetPrimary()
	if primary == nil {
		return nil, fmt.Errorf("no primary database available")
	}

	// For now, we need to extract the underlying pool to work with existing repositories
	// This is a temporary solution until we update the repositories
	pgProvider, ok := primary.(*postgres.PostgresProvider)
	if !ok {
		return nil, fmt.Errorf("primary database is not a PostgreSQL provider")
	}

	pool := pgProvider.GetPool()
	if pool == nil {
		return nil, fmt.Errorf("cannot get connection pool from PostgreSQL provider")
	}

	// Initialize repositories (they still expect pgxpool.Pool)
	repos := repositories.NewRepositories(pool)

	// Initialize OAuth2 provider (still needs pgxpool.Pool)
	provider, err := oauth2.NewProvider(cfg, pool)
	if err != nil {
		return nil, err
	}

	// Seed default OAuth2 clients in non-production environments
	if cfg.Server.Environment != "production" {
		oauth2.SeedDefaultClients(pool, cfg.OAuth2.Issuer)
		oauth2.EnsureNextAuthRedirects(pool)
	}

	// Initialize session manager
	sess := session.New(cfg.Security.JWTSecret, "idsess", cfg.Security.SessionDuration, cfg.Server.Environment == "production")

	// Initialize services (extract them so we can use them in gRPC as well)
	credentialService := services.NewCredentialService(repos, cfg.Security.BCryptCost)
	userService := services.NewUserService(repos)
	tenantService := services.NewTenantService(repos)
	authService := services.NewAuthService(repos, userService, credentialService, tenantService, sess)

	// Initialize handlers with services
	h := handlers.NewHandlersWithServices(authService, userService, tenantService, provider, repos, sess, translator, cfg.Server.Environment != "production", cfg.Security.Registration.RegistrationAccessTokenSecret, cfg.Security.LoginPageURL)

	// Initialize middleware manager
	m := middleware.NewManager(cfg, provider, repos, translator)

	return &Dependencies{
		DBManager:   dbManager,
		Config:      cfg,
		Translator:  translator,
		Handlers:    h,
		Middleware:  m,
		Provider:    provider,
		UserService: userService,
	}, nil
}
