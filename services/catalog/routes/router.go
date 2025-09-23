// Package routes manages HTTP routing for the Catalog service, echoing the
// conventions established in the Identify service for familiarity.
package routes

import (
	"fmt"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"wibusystem/pkg/database/factory"
	"wibusystem/pkg/database/providers/postgres"
	"wibusystem/pkg/i18n"
	"wibusystem/services/catalog/config"
	"wibusystem/services/catalog/handlers"
	"wibusystem/services/catalog/middleware"
	"wibusystem/services/catalog/repositories"
	"wibusystem/services/catalog/services"
	v1 "wibusystem/services/catalog/routes/api/v1"
)

// Dependencies groups runtime dependencies required to setup routes.
type Dependencies struct {
	DBManager    *factory.DatabaseManager
	Config       *config.Config
	Translator   *i18n.Translator
	Handlers     *handlers.Handlers
	Middleware   *middleware.Manager
	Repositories *repositories.Repositories
	Services     *services.Services
}

// SetupRouter initializes a Gin engine with middlewares and versioned routes.
func SetupRouter(deps *Dependencies) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())

	corsCfg := cors.Config{
		AllowMethods:     deps.Config.Security.CORS.AllowMethods,
		AllowHeaders:     deps.Config.Security.CORS.AllowHeaders,
		ExposeHeaders:    deps.Config.Security.CORS.ExposeHeaders,
		AllowCredentials: deps.Config.Security.CORS.AllowCredentials,
		MaxAge:           time.Duration(deps.Config.Security.CORS.MaxAge) * time.Second,
	}

	if len(deps.Config.Security.CORS.AllowOrigins) == 0 {
		corsCfg.AllowAllOrigins = true
	} else if len(deps.Config.Security.CORS.AllowOrigins) == 1 && deps.Config.Security.CORS.AllowOrigins[0] == "*" {
		corsCfg.AllowAllOrigins = true
	} else {
		corsCfg.AllowOrigins = deps.Config.Security.CORS.AllowOrigins
	}

	router.Use(cors.New(corsCfg))

	deps.Middleware.SetupCommonMiddleware(router)

	setupPublicRoutes(router, deps)

	v1.SetupRoutes(router, deps.Config, deps.Handlers, deps.Middleware)

	return router
}

// NewDependencies wires repositories, handlers, and middleware.
func NewDependencies(dbManager *factory.DatabaseManager, cfg *config.Config, translator *i18n.Translator) (*Dependencies, error) {
	primary := dbManager.GetPrimary()
	if primary == nil {
		return nil, fmt.Errorf("no primary database configured")
	}

	pgProvider, ok := primary.(*postgres.PostgresProvider)
	if !ok {
		return nil, fmt.Errorf("primary database is not PostgreSQL provider")
	}

	pool := pgProvider.GetPool()
	if pool == nil {
		return nil, fmt.Errorf("postgres connection pool is unavailable")
	}

	repos := repositories.NewRepositories(pool)
	services := services.NewServices(repos)
	h := handlers.NewHandlers(repos, services, translator)
	m := middleware.NewManager(cfg, translator)

	return &Dependencies{
		DBManager:    dbManager,
		Config:       cfg,
		Translator:   translator,
		Handlers:     h,
		Middleware:   m,
		Repositories: repos,
		Services:     services,
	}, nil
}

func setupPublicRoutes(router *gin.Engine, deps *Dependencies) {
	router.GET("/healthz", deps.Handlers.Health.Status)
}
