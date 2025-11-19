package router

import (
	"net/http"
	"system/configs"
	v1 "system/internal/app/handler/v1"
	"system/internal/app/handler/v1/artist"
	"system/internal/app/handler/v1/auth"
	"system/internal/app/handler/v1/author"
	"system/internal/app/handler/v1/chapter"
	"system/internal/app/handler/v1/genre"
	"system/internal/app/handler/v1/novel"
	oauth2_handler "system/internal/app/handler/v1/oauth2"
	"system/internal/app/handler/v1/oauth2_admin"
	"system/internal/app/handler/v1/volume"
	"system/internal/oauth2"
	fosite_storage "system/internal/oauth2/storage"
	"system/internal/pkg/repository"
	"system/internal/pkg/service"
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

	// Load HTML templates
	router.LoadHTMLGlob("web/templates/**/*")

	// Serve static files
	router.Static("/images", "./web/images")

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
	userRepo := repository.NewUserRepository(db.Pool)
	sessionRepo := repository.NewSessionRepository(rdb)
	authRequestRepo := repository.NewAuthRequestRepository(rdb, oauth2ClientRepo)
	consentRepo := repository.NewConsentRepository(db.Pool)
	emailVerificationRepo := repository.NewEmailVerificationRepository(db.Pool)
	passwordResetRepo := repository.NewPasswordResetRepository(db.Pool)
	genreRepo := repository.NewGenreRepository(db.Pool)
	authorRepo := repository.NewAuthorRepository(db.Pool)
	artistRepo := repository.NewArtistRepository(db.Pool)
	novelRepo := repository.NewNovelRepository(db.Pool)
	volumeRepo := repository.NewVolumeRepository(db.Pool)
	volumeHistoryRepo := repository.NewVolumeHistoryRepository(db.Pool)
	chapterRepo := repository.NewChapterRepository(db.Pool)
	chapterHistoryRepo := repository.NewChapterHistoryRepository(db.Pool)

	// Fosite Storage
	sqlStore := fosite_storage.NewSQLStore(oauth2ClientRepo, oauth2SessionRepo)
	redisStore := fosite_storage.NewRedisStore(rdb, oauth2ClientRepo, zapLogger)

	// Fosite Provider
	hybridStore := fosite_storage.NewHybridStore(sqlStore, redisStore)
	oauth2Provider := oauth2.NewOAuth2Provider(hybridStore, &cfg.OAuth2)

	// Services
	oauth2Service := service.NewOAuth2Service(
		userRepo,
		sessionRepo,
		authRequestRepo,
		consentRepo,
		oauth2SessionRepo,
		oauth2ClientRepo,
	)

	authService := service.NewAuthService(
		userRepo,
		emailVerificationRepo,
		passwordResetRepo,
	)

	emailService := service.NewEmailService(&cfg.Email, zapLogger)

	oauth2AdminService := service.NewOAuth2AdminService(oauth2ClientRepo)

	genreService := service.NewGenreService(genreRepo)
	authorService := service.NewAuthorService(authorRepo)
	artistService := service.NewArtistService(artistRepo)
	novelService := service.NewNovelService(novelRepo)
	volumeService := service.NewVolumeService(volumeRepo, volumeHistoryRepo)
	chapterService := service.NewChapterService(chapterRepo, chapterHistoryRepo)

	// Handlers
	oauth2Handler := oauth2_handler.NewHandler(
		&cfg.OAuth2,
		oauth2Provider,
		oauth2Service,
		authRequestRepo,
		oauth2ClientRepo,
		zapLogger,
	)

	authHandler := auth.NewHandler(authService, emailService)

	genreHandler := genre.NewHandler(genreService, zapLogger)
	authorHandler := author.NewHandler(authorService)
	artistHandler := artist.NewHandler(artistService)
	novelHandler := novel.NewHandler(novelService)
	volumeHandler := volume.NewHandler(volumeService)
	chapterHandler := chapter.NewHandler(chapterService)

	// --- Đăng ký Routes ---
	apiV1 := router.Group("/api/v1")
	{
		v1.RegisterRoutes(apiV1)

		// Genre routes
		genreGroup := apiV1.Group("/genres")
		{
			genreHandler.RegisterRoutes(genreGroup)
		}

		// Author routes
		authorGroup := apiV1.Group("/authors")
		{
			authorHandler.RegisterRoutes(authorGroup)
		}

		// Artist routes
		artistGroup := apiV1.Group("/artists")
		{
			artistHandler.RegisterRoutes(artistGroup)
		}

		// Novel routes
		novelGroup := apiV1.Group("/novels")
		{
			novelHandler.RegisterRoutes(novelGroup)

			// Nested: Volumes by novel (use :id to match existing routes)
			novelGroup.GET("/:id/volumes", volumeHandler.ListVolumesByNovel)

			// Nested: Chapters by novel (use :id to match existing routes)
			novelChaptersGroup := novelGroup.Group("/:id/chapters")
			{
				chapterHandler.RegisterNovelChaptersRoutes(novelChaptersGroup)
			}
		}

		// Volume routes
		volumeGroup := apiV1.Group("/volumes")
		{
			volumeHandler.RegisterRoutes(volumeGroup)

			// Nested: Chapters by volume (use :id to match existing routes)
			volumeChaptersGroup := volumeGroup.Group("/:id/chapters")
			{
				chapterHandler.RegisterVolumeChaptersRoutes(volumeChaptersGroup)
			}
		}

		// Chapter routes
		chapterGroup := apiV1.Group("/chapters")
		{
			chapterHandler.RegisterRoutes(chapterGroup)
		}
	}

	wellKnownGroup := router.Group("/.well-known")
	{
		oauth2Handler.RegisterWellKnownRoutes(wellKnownGroup)
	}

	oauth2Group := router.Group("/oauth2")
	{
		oauth2Handler.RegisterRoutes(oauth2Group)

		// Auth pages under /oauth2 for consistency with login page
		oauth2Group.GET("/register", authHandler.RegisterPage)
		oauth2Group.GET("/verify-email", authHandler.VerifyEmailPage)
		oauth2Group.GET("/forgot-password", authHandler.ForgotPasswordPage)
		oauth2Group.GET("/reset-password", authHandler.ResetPasswordPage)
	}

	// Auth API endpoints
	authAPIGroup := apiV1.Group("/auth")
	{
		authAPIGroup.POST("/register", authHandler.Register)
		authAPIGroup.GET("/verify-email", authHandler.VerifyEmail)
		authAPIGroup.POST("/verify-email", authHandler.VerifyEmail)
		authAPIGroup.POST("/forgot-password", authHandler.ForgotPassword)
		authAPIGroup.POST("/reset-password", authHandler.ResetPassword)
	}

	// OAuth2 Admin API
	oauth2AdminHandler := oauth2_admin.NewHandler(oauth2AdminService)
	oauth2_admin.RegisterRoutes(apiV1, oauth2AdminHandler)

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
