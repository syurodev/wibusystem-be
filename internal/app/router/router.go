package router

import (
	"net/http"
	"system/configs"
	v1 "system/internal/app/handler/v1"
	oauth2_handler "system/internal/app/handler/v1/oauth2"
	"system/internal/oauth2"
	fosite_storage "system/internal/oauth2/storage"
	"system/internal/pkg/repository"
	"system/internal/platform/database"
	"system/internal/platform/i18n"
	"system/internal/platform/logger"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// NewRouter khởi tạo và cấu hình Gin router.
func NewRouter(cfg *configs.Config, i18nInstance *i18n.I18n, zapLogger *zap.Logger, db *database.PostgresDB, rdb *database.RedisClient) *gin.Engine {
	// Thiết lập Gin mode
	if cfg.Server.IsProd {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	router := gin.New()

	// Middleware
	router.Use(logger.GinZap(zapLogger, cfg.Server.IsProd))
	router.Use(gin.Recovery())
	router.Use(cors.New(buildCorsConfig(cfg)))
	router.Use(i18n.GinI18n(i18nInstance))

	// Health Check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "UP"})
	})

	// --- Khởi tạo Dependencies ---

	// Repositories
	oauth2ClientRepo := repository.NewOAuth2ClientRepository(db.Pool)
	oauth2SessionRepo := repository.NewOAuth2SessionRepository(db.Pool)

	// Fosite Storage
	sqlStore := fosite_storage.NewSQLStore(oauth2ClientRepo, oauth2SessionRepo)
	redisStore := fosite_storage.NewRedisStore(rdb)
	_ = fosite_storage.NewHybridStore(sqlStore, redisStore) // Gán cho _ cho đến khi dùng

	// Fosite Provider
	hybridStore := fosite_storage.NewHybridStore(sqlStore, redisStore)
	oauth2Provider := oauth2.NewOAuth2Provider(hybridStore, &cfg.OAuth2)

	// Handlers
	mockStore := oauth2_handler.NewMockStore() // Vẫn dùng mock store cho UserInfo handler
	oauth2Handler := oauth2_handler.NewHandler(&cfg.OAuth2, oauth2Provider, mockStore)

	// --- Đăng ký Routes ---
	apiV1 := router.Group("/api/v1")
	{
		v1.RegisterRoutes(apiV1)
	}

	wellKnownGroup := router.Group("/.well-known")
	{
		oauth2Handler.RegisterWellKnownRoutes(wellKnownGroup)
	}

	oauth2Group := router.Group("/oauth2")
	{
		oauth2Handler.RegisterRoutes(oauth2Group)
	}

	return router
}

// buildCorsConfig xây dựng cấu hình CORS từ config của ứng dụng.
func buildCorsConfig(cfg *configs.Config) cors.Config {
	return cors.Config{
		AllowOrigins:     cfg.CORS.AllowOrigins,
		AllowMethods:     cfg.CORS.AllowMethods,
		AllowHeaders:     cfg.CORS.AllowHeaders,
		ExposeHeaders:    cfg.CORS.ExposeHeaders,
		AllowCredentials: cfg.CORS.AllowCredentials,
		MaxAge:           time.Duration(cfg.CORS.MaxAge) * time.Second,
	}
}
