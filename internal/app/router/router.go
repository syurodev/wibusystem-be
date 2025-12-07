package router

import (
	"html/template"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"system/configs"
	v1 "system/internal/app/handler/v1"
	"system/internal/app/handler/v1/analytics"
	"system/internal/app/handler/v1/artist"
	"system/internal/app/handler/v1/auth"
	"system/internal/app/handler/v1/author"
	"system/internal/app/handler/v1/creator"
	"system/internal/app/handler/v1/genre"
	"system/internal/app/handler/v1/novel"
	novel_volume "system/internal/app/handler/v1/novel/volume"
	volume_chapter "system/internal/app/handler/v1/novel/volume/chapter"
	oauth2_handler "system/internal/app/handler/v1/oauth2"
	"system/internal/app/handler/v1/oauth2_admin"
	"system/internal/app/handler/v1/public"
	"system/internal/app/handler/v1/user"
	"system/internal/app/middleware"
	"system/internal/app/worker"
	"system/internal/oauth2"
	fosite_storage "system/internal/oauth2/storage"
	txdb "system/internal/pkg/db"
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
func NewRouter(cfg *configs.Config, i18nInstance *i18n.I18n, zapLogger *zap.Logger, db *database.PostgresDB, rdb *database.RedisClient, ch *database.ClickHouseClient) *gin.Engine {
	// Thiết lập Gin mode
	if cfg.Server.IsProd {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	router := gin.New()

	// Load HTML templates recursively from web/templates
	// Build list of all template files
	var templateFiles []string
	err := filepath.WalkDir("web/templates", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && filepath.Ext(path) == ".html" {
			templateFiles = append(templateFiles, path)
		}
		return nil
	})

	if err != nil {
		zapLogger.Fatal("Failed to walk templates directory", zap.Error(err))
	}

	// Parse all templates into a single template set with custom names
	if len(templateFiles) > 0 {
		funcMap := template.FuncMap{
			"contains": strings.Contains,
			"dict": func(values ...any) map[string]any {
				if len(values)%2 != 0 {
					return nil
				}
				dict := make(map[string]any, len(values)/2)
				for i := 0; i < len(values); i += 2 {
					key, ok := values[i].(string)
					if !ok {
						return nil
					}
					dict[key] = values[i+1]
				}
				return dict
			},
		}

		tmpl := template.New("").Funcs(funcMap)

		// Parse each template file with a custom name (relative to web/templates/)
		for _, filePath := range templateFiles {
			// Remove "web/templates/" prefix to get template name
			templateName := strings.TrimPrefix(filePath, "web/templates/")

			// Read file content
			content, err := os.ReadFile(filePath)
			if err != nil {
				zapLogger.Fatal("Failed to read template file",
					zap.String("file", filePath),
					zap.Error(err),
				)
			}

			// Parse the template content with custom name
			_, err = tmpl.New(templateName).Parse(string(content))
			if err != nil {
				zapLogger.Fatal("Failed to parse template",
					zap.String("file", filePath),
					zap.String("name", templateName),
					zap.Error(err),
				)
			}
		}

		router.SetHTMLTemplate(tmpl)
	} else {
		zapLogger.Warn("No templates found")
	}

	// Serve static files
	router.Static("/images", "./web/images")
	router.Static("/static", "./web/static")

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
	txManager := txdb.NewTransactionManager(db.Pool)
	webauthnCredentialRepo := repository.NewWebAuthnCredentialRepository(db.Pool)
	webauthnSessionRepo := repository.NewWebAuthnSessionRepository(db.Pool)
	viewAnalyticsRepo := repository.NewViewAnalyticsClickHouseRepository(ch)
	viewTrackingRepo := repository.NewViewTrackingRedisRepository(rdb)
	roleRepo := repository.NewRoleRepository(db.Pool)
	creatorRepo := repository.NewCreatorRepository(db.Pool)

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
		roleRepo,
	)

	emailService := service.NewEmailService(&cfg.Email, zapLogger)

	oauth2AdminService := service.NewOAuth2AdminService(oauth2ClientRepo)

	genreService := service.NewGenreService(genreRepo)
	authorService := service.NewAuthorService(authorRepo)
	artistService := service.NewArtistService(artistRepo)
	novelService := service.NewNovelService(novelRepo, volumeRepo, genreRepo, authorRepo, artistRepo, creatorRepo, txManager)
	volumeService := service.NewVolumeService(volumeRepo, volumeHistoryRepo)
	chapterService := service.NewChapterService(chapterRepo, volumeRepo, chapterHistoryRepo, creatorRepo)
	viewTrackingService := service.NewViewTrackingService(viewTrackingRepo, viewAnalyticsRepo, chapterRepo, novelRepo, genreRepo, zapLogger, &cfg.ViewTracking)
	analyticsService := service.NewAnalyticsService(viewAnalyticsRepo, novelRepo, zapLogger)
	creatorService := service.NewCreatorService(creatorRepo, viewAnalyticsRepo, novelRepo, zapLogger)

	// Cache and Public services
	cacheService := service.NewCacheService(rdb, zapLogger)
	publicService := service.NewPublicService(analyticsService, creatorService, genreService, cacheService, zapLogger)

	// Start View Tracking Workers
	worker.StartViewTrackingWorkers(&cfg.ViewTracking, viewTrackingService, zapLogger)

	webauthnService, err := service.NewWebAuthnService(
		cfg.WebAuthn,
		webauthnCredentialRepo,
		webauthnSessionRepo,
		userRepo,
		zapLogger,
	)
	if err != nil {
		zapLogger.Fatal("Failed to initialize WebAuthn service", zap.Error(err))
	}

	// Handlers
	oauth2Handler := oauth2_handler.NewHandler(
		&cfg.OAuth2,
		oauth2Provider,
		oauth2Service,
		authRequestRepo,
		oauth2ClientRepo,
		webauthnService,
		zapLogger,
	)

	authHandler := auth.NewHandler(authService, emailService, webauthnService, oauth2Service)

	genreHandler := genre.NewHandler(genreService, zapLogger)
	authorHandler := author.NewHandler(authorService, zapLogger)
	artistHandler := artist.NewHandler(artistService)
	novelHandler := novel.NewHandler(novelService)
	volumeHandler := novel_volume.NewHandler(volumeService, novelService)

	chapterHandler := volume_chapter.NewHandler(chapterService, volumeService, viewTrackingService, novelService)
	userHandler := user.NewHandler(userRepo, sessionRepo)
	analyticsHandler := analytics.NewHandler(analyticsService)
	creatorHandler := creator.NewHandler(creatorService)
	publicHandler := public.NewHandler(publicService)

	// --- Đăng ký Routes ---
	apiV1 := router.Group("/api/v1")
	{
		v1.RegisterRoutes(apiV1)

		// Genre routes
		genreGroup := apiV1.Group("/genres")
		{
			genreHandler.RegisterRoutes(genreGroup, middleware.RequireAuth(oauth2Provider, zapLogger))
		}

		// Author routes
		authorGroup := apiV1.Group("/authors")
		{
			authorHandler.RegisterRoutes(authorGroup, middleware.RequireAuth(oauth2Provider, zapLogger))
		}

		// Artist routes
		artistHandler.RegisterRoutes(apiV1, middleware.RequireAuth(oauth2Provider, zapLogger))

		// Novel routes
		novelGroup := apiV1.Group("/novels")
		{
			novelHandler.RegisterRoutes(novelGroup, middleware.RequireAuth(oauth2Provider, zapLogger))

			// Nested: Volumes by novel
			novelGroup.GET("/:id/volumes", volumeHandler.ListVolumesByNovel)
			novelGroup.POST("/:id/volumes", middleware.RequireAuth(oauth2Provider, zapLogger), volumeHandler.CreateVolume)
		}

		// Novel Volume routes (Namespace: /novels/volumes)
		novelVolumeGroup := apiV1.Group("/novels/volumes")
		{
			novelVolumeGroup.GET("/:id", volumeHandler.GetVolume)
			
			// Protected volume operations
			protectedVolumeGroup := novelVolumeGroup.Group("")
			protectedVolumeGroup.Use(middleware.RequireAuth(oauth2Provider, zapLogger))
			{
				protectedVolumeGroup.PUT("/:id", volumeHandler.UpdateVolume)
				protectedVolumeGroup.DELETE("/:id", volumeHandler.DeleteVolume)
				protectedVolumeGroup.PUT("/:id/display-order", volumeHandler.UpdateDisplayOrder)
				protectedVolumeGroup.POST("/:id/publish", volumeHandler.PublishVolume)
				protectedVolumeGroup.POST("/:id/unpublish", volumeHandler.UnpublishVolume)
			}

			// Nested: Chapters by volume
			novelVolumeGroup.GET("/:id/chapters", chapterHandler.ListChaptersByVolume)
			novelVolumeGroup.POST("/:id/chapters", middleware.RequireAuth(oauth2Provider, zapLogger), chapterHandler.CreateChapter)
		}

		// Novel Chapter routes (Namespace: /novels/chapters)
		novelChapterGroup := apiV1.Group("/novels/chapters")
		{
			novelChapterGroup.GET("/:id", chapterHandler.GetChapter)
			
			// Protected chapter operations
			protectedChapterGroup := novelChapterGroup.Group("")
			protectedChapterGroup.Use(middleware.RequireAuth(oauth2Provider, zapLogger))
			{
				protectedChapterGroup.PUT("/:id", chapterHandler.UpdateChapter)
				protectedChapterGroup.DELETE("/:id", chapterHandler.DeleteChapter)
				protectedChapterGroup.POST("/:id/publish", chapterHandler.PublishChapter)
				protectedChapterGroup.POST("/:id/schedule", chapterHandler.ScheduleChapter)
				protectedChapterGroup.PUT("/:id/statistics", chapterHandler.UpdateStatistics)
			}
			
			// Public chapter operations
			novelChapterGroup.POST("/:id/view", chapterHandler.IncrementViewCount)
		}


		userGroup := apiV1.Group("/users/me")
		userGroup.Use(middleware.RequireAuth(oauth2Provider, zapLogger))
		{
			userHandler.RegisterRoutes(userGroup)
		}

		// Analytics routes
		analyticsGroup := apiV1.Group("/analytics")
		{
			analyticsGroup.GET("/trending", analyticsHandler.GetTrending)
		}

		// Creator routes (public)
		apiV1.GET("/creators", creatorHandler.ListCreators)

		// Public routes (home page data)
		publicGroup := apiV1.Group("/public")
		{
			publicHandler.RegisterRoutes(publicGroup)
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

	// Passkey HTML pages
	authHTMLGroup := router.Group("/auth")
	{
		authHTMLGroup.GET("/passkey/setup", authHandler.PasskeySetupPage)
		authHTMLGroup.GET("/passkey/register", authHandler.PasskeyRegisterPage)
		authHTMLGroup.GET("/passkey/manage", authHandler.PasskeyManagePage)
	}

	// Account management pages with htmx fragments (requires authentication)
	accountGroup := router.Group("/account", middleware.RequireSessionAuth(oauth2Service, zapLogger))
	{
		// Passkey management htmx fragments
		accountGroup.GET("/passkeys", authHandler.PasskeyListFragment)
		accountGroup.DELETE("/passkeys/:id", authHandler.PasskeyDeleteFragment)
		accountGroup.PUT("/passkeys/:id/name", authHandler.PasskeyUpdateNameFragment)
	}

	// Auth API endpoints
	authAPIGroup := apiV1.Group("/auth")
	{
		authAPIGroup.POST("/register", authHandler.Register)
		authAPIGroup.GET("/verify-email", authHandler.VerifyEmail)
		authAPIGroup.POST("/verify-email", authHandler.VerifyEmail)
		authAPIGroup.POST("/forgot-password", authHandler.ForgotPassword)
		authAPIGroup.POST("/reset-password", authHandler.ResetPassword)

		// WebAuthn/Passkey endpoints - Registration (requires authentication)
		authAPIGroup.POST("/passkey/register/begin", middleware.RequireSessionAuth(oauth2Service, zapLogger), authHandler.PasskeyRegisterBegin)
		authAPIGroup.POST("/passkey/register/finish", middleware.RequireSessionAuth(oauth2Service, zapLogger), authHandler.PasskeyRegisterFinish)

		// WebAuthn/Passkey endpoints - Authentication (public)
		authAPIGroup.POST("/passkey/authenticate/begin", authHandler.PasskeyAuthenticateBegin)
		authAPIGroup.POST("/passkey/authenticate/finish", authHandler.PasskeyAuthenticateFinish)

		// WebAuthn/Passkey endpoints - Credential management (requires authentication)
		authAPIGroup.GET("/passkey/credentials", middleware.RequireSessionAuth(oauth2Service, zapLogger), authHandler.PasskeyListCredentials)
		authAPIGroup.DELETE("/passkey/credentials", middleware.RequireSessionAuth(oauth2Service, zapLogger), authHandler.PasskeyDeleteCredential)
		authAPIGroup.PUT("/passkey/credentials/name", middleware.RequireSessionAuth(oauth2Service, zapLogger), authHandler.PasskeyUpdateCredentialName)
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
