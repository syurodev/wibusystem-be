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
	"system/internal/app/handler/v1/oauth2_admin"
	"system/internal/app/middleware"
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

	// Load HTML templates
	loadTemplates(router, zapLogger)

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
	deps, err := NewDependencies(cfg, db, rdb, ch, zapLogger)
	if err != nil {
		zapLogger.Fatal("Failed to initialize dependencies", zap.Error(err))
	}

	// --- Đăng ký Routes ---
	registerAPIRoutes(router, deps, zapLogger)
	registerOAuth2Routes(router, deps)
	registerAuthRoutes(router, deps.Handlers, deps.Services, zapLogger)

	return router
}

// loadTemplates tải tất cả HTML templates từ web/templates
func loadTemplates(router *gin.Engine, zapLogger *zap.Logger) {
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

	if len(templateFiles) == 0 {
		zapLogger.Warn("No templates found")
		return
	}

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

	for _, filePath := range templateFiles {
		templateName := strings.TrimPrefix(filePath, "web/templates/")
		content, err := os.ReadFile(filePath)
		if err != nil {
			zapLogger.Fatal("Failed to read template file", zap.String("file", filePath), zap.Error(err))
		}
		_, err = tmpl.New(templateName).Parse(string(content))
		if err != nil {
			zapLogger.Fatal("Failed to parse template", zap.String("file", filePath), zap.Error(err))
		}
	}

	router.SetHTMLTemplate(tmpl)
}

// registerAPIRoutes đăng ký tất cả API v1 routes
func registerAPIRoutes(router *gin.Engine, deps *Dependencies, zapLogger *zap.Logger) {
	apiV1 := router.Group("/api/v1")
	authMiddleware := middleware.RequireAuth(deps.OAuth2Provider, zapLogger)

	v1.RegisterRoutes(apiV1)

	// Genre routes
	genreGroup := apiV1.Group("/genres")
	deps.Handlers.Genre.RegisterRoutes(genreGroup, authMiddleware)

	// Author routes
	authorGroup := apiV1.Group("/authors")
	deps.Handlers.Author.RegisterRoutes(authorGroup, authMiddleware)

	// Artist routes
	deps.Handlers.Artist.RegisterRoutes(apiV1, authMiddleware)

	// Novel routes
	registerNovelRoutes(apiV1, deps, authMiddleware)

	// User routes
	userGroup := apiV1.Group("/users/me")
	userGroup.Use(authMiddleware)
	deps.Handlers.User.RegisterRoutes(userGroup)

	// Media routes
	mediaGroup := apiV1.Group("/media")
	mediaGroup.GET("/trending", deps.Handlers.Media.GetTrending)

	// Creator routes
	apiV1.GET("/creators", deps.Handlers.Creator.ListCreators)

	// Public routes
	publicGroup := apiV1.Group("/public")
	deps.Handlers.Public.RegisterRoutes(publicGroup)

	// Auth API routes
	registerAuthAPIRoutes(apiV1, deps)

	// OAuth2 Admin API
	oauth2_admin.RegisterRoutes(apiV1, deps.Handlers.OAuth2Admin)
}

// registerNovelRoutes đăng ký novel, volume, chapter routes
func registerNovelRoutes(apiV1 *gin.RouterGroup, deps *Dependencies, authMiddleware gin.HandlerFunc) {
	h := deps.Handlers

	// Novel routes
	novelGroup := apiV1.Group("/novels")
	h.Novel.RegisterRoutes(novelGroup, authMiddleware)
	novelGroup.GET("/:id/volumes", h.Volume.ListVolumesByNovel)
	novelGroup.POST("/:id/volumes", authMiddleware, h.Volume.CreateVolume)

	// Volume routes (Namespace: /novels/volumes)
	volumeGroup := apiV1.Group("/novels/volumes")
	volumeGroup.GET("/:id", h.Volume.GetVolume)
	
	protectedVolume := volumeGroup.Group("", authMiddleware)
	protectedVolume.PUT("/:id", h.Volume.UpdateVolume)
	protectedVolume.DELETE("/:id", h.Volume.DeleteVolume)
	protectedVolume.PUT("/:id/display-order", h.Volume.UpdateDisplayOrder)
	protectedVolume.POST("/:id/publish", h.Volume.PublishVolume)
	protectedVolume.POST("/:id/unpublish", h.Volume.UnpublishVolume)

	volumeGroup.GET("/:id/chapters", h.Chapter.ListChaptersByVolume)
	volumeGroup.POST("/:id/chapters", authMiddleware, h.Chapter.CreateChapter)

	// Chapter routes (Namespace: /novels/chapters)
	chapterGroup := apiV1.Group("/novels/chapters")
	chapterGroup.GET("/:id", h.Chapter.GetChapter)
	chapterGroup.POST("/:id/view", h.Chapter.IncrementViewCount)

	protectedChapter := chapterGroup.Group("", authMiddleware)
	protectedChapter.PUT("/:id", h.Chapter.UpdateChapter)
	protectedChapter.DELETE("/:id", h.Chapter.DeleteChapter)
	protectedChapter.POST("/:id/publish", h.Chapter.PublishChapter)
	protectedChapter.POST("/:id/schedule", h.Chapter.ScheduleChapter)
	protectedChapter.PUT("/:id/statistics", h.Chapter.UpdateStatistics)
}

// registerAuthAPIRoutes đăng ký auth API routes
func registerAuthAPIRoutes(apiV1 *gin.RouterGroup, deps *Dependencies) {
	authGroup := apiV1.Group("/auth")
	sessionAuth := middleware.RequireSessionAuth(deps.Services.OAuth2, nil)
	h := deps.Handlers.Auth

	// Public auth endpoints
	authGroup.POST("/register", h.Register)
	authGroup.GET("/verify-email", h.VerifyEmail)
	authGroup.POST("/verify-email", h.VerifyEmail)
	authGroup.POST("/forgot-password", h.ForgotPassword)
	authGroup.POST("/reset-password", h.ResetPassword)

	// WebAuthn Registration (requires auth)
	authGroup.POST("/passkey/register/begin", sessionAuth, h.PasskeyRegisterBegin)
	authGroup.POST("/passkey/register/finish", sessionAuth, h.PasskeyRegisterFinish)

	// WebAuthn Authentication (public)
	authGroup.POST("/passkey/authenticate/begin", h.PasskeyAuthenticateBegin)
	authGroup.POST("/passkey/authenticate/finish", h.PasskeyAuthenticateFinish)

	// WebAuthn Credential management (requires auth)
	authGroup.GET("/passkey/credentials", sessionAuth, h.PasskeyListCredentials)
	authGroup.DELETE("/passkey/credentials", sessionAuth, h.PasskeyDeleteCredential)
	authGroup.PUT("/passkey/credentials/name", sessionAuth, h.PasskeyUpdateCredentialName)
}

// registerOAuth2Routes đăng ký OAuth2 routes
func registerOAuth2Routes(router *gin.Engine, deps *Dependencies) {
	// Well-known endpoints
	wellKnown := router.Group("/.well-known")
	deps.Handlers.OAuth2.RegisterWellKnownRoutes(wellKnown)

	// OAuth2 endpoints
	oauth2Group := router.Group("/oauth2")
	deps.Handlers.OAuth2.RegisterRoutes(oauth2Group)

	// Auth pages under /oauth2
	h := deps.Handlers.Auth
	oauth2Group.GET("/register", h.RegisterPage)
	oauth2Group.GET("/verify-email", h.VerifyEmailPage)
	oauth2Group.GET("/forgot-password", h.ForgotPasswordPage)
	oauth2Group.GET("/reset-password", h.ResetPasswordPage)
}

// registerAuthRoutes đăng ký auth HTML pages và account routes
func registerAuthRoutes(router *gin.Engine, handlers *Handlers, services *Services, zapLogger *zap.Logger) {
	h := handlers.Auth
	sessionAuth := middleware.RequireSessionAuth(services.OAuth2, zapLogger)

	// Passkey HTML pages
	authHTML := router.Group("/auth")
	authHTML.GET("/passkey/setup", h.PasskeySetupPage)
	authHTML.GET("/passkey/register", h.PasskeyRegisterPage)
	authHTML.GET("/passkey/manage", h.PasskeyManagePage)

	// Account management (requires session auth)
	account := router.Group("/account", sessionAuth)
	account.GET("/passkeys", h.PasskeyListFragment)
	account.DELETE("/passkeys/:id", h.PasskeyDeleteFragment)
	account.PUT("/passkeys/:id/name", h.PasskeyUpdateNameFragment)
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
