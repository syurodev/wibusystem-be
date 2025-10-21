package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"wibusystem/internal/infrastructure/database"
	catalogHttp "wibusystem/internal/modules/catalog/handler/http"
	catalogRepo "wibusystem/internal/modules/catalog/repository/postgres"
	catalogService "wibusystem/internal/modules/catalog/service"
	identityHttp "wibusystem/internal/modules/identity/handler/http"
	identityRepo "wibusystem/internal/modules/identity/repository/postgres"
	identityService "wibusystem/internal/modules/identity/service"
	"wibusystem/internal/platform/config"
	"wibusystem/internal/platform/i18n"

	"github.com/gofiber/fiber/v2"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize i18n Translator
	i18nConfig := i18n.Config{
		BundlePath:         cfg.Localization.BundlePath,
		DefaultLanguage:    cfg.Localization.DefaultLanguage,
		SupportedLanguages: cfg.Localization.SupportedLanguages,
		QueryParam:         cfg.Localization.QueryParam,
		HeaderName:         cfg.Localization.HeaderName,
		CookieName:         cfg.Localization.CookieName,
	}
	translator, err := i18n.NewTranslator(i18nConfig)
	if err != nil {
		log.Fatalf("Failed to initialize i18n translator: %v", err)
	}
	log.Println("✅ i18n Translator initialized")

	// Initialize database
	db, err := database.New(context.Background(), &cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("✅ Database connected successfully")

	// Run migrations
	if err := db.RunMigrations(cfg.Database.MigrationsPath); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	log.Println("✅ Migrations completed successfully")

	// --- Initialize Identity Module ---
	log.Println("Initializing Identity Module...")
	userRepo := identityRepo.NewUserRepository(db.Pool(), cfg.Database.IdentitySchema)
	tenantRepo := identityRepo.NewTenantRepository(db.Pool(), cfg.Database.IdentitySchema)
	tenantMemberRepo := identityRepo.NewTenantMemberRepository(db.Pool(), cfg.Database.IdentitySchema)
	sessionRepo := identityRepo.NewSessionRepository(db.Pool(), cfg.Database.IdentitySchema)

	authService := identityService.NewAuthService(userRepo, sessionRepo)
	userService := identityService.NewUserService(userRepo)
	tenantService := identityService.NewTenantService(tenantRepo, tenantMemberRepo, userRepo)
	log.Println("✅ Identity Module initialized")

	// --- Initialize Catalog Module ---
	log.Println("Initializing Catalog Module...")
	novelRepo := catalogRepo.NewNovelRepositoryPG(db.Pool())
	volumeRepo := catalogRepo.NewVolumeRepositoryPG(db.Pool())
	chapterRepo := catalogRepo.NewChapterRepositoryPG(db.Pool())
	creatorRepo := catalogRepo.NewCreatorRepositoryPG(db.Pool())
	genreRepo := catalogRepo.NewGenreRepositoryPG(db.Pool())
	characterRepo := catalogRepo.NewCharacterRepositoryPG(db.Pool())

	novelSvc := catalogService.NewNovelService(novelRepo)
	volumeSvc := catalogService.NewVolumeService(volumeRepo, novelRepo)
	chapterSvc := catalogService.NewChapterService(chapterRepo, volumeRepo)
	creatorSvc := catalogService.NewCreatorService(creatorRepo)
	genreSvc := catalogService.NewGenreService(genreRepo)
	characterSvc := catalogService.NewCharacterService(characterRepo, novelRepo)
	log.Println("✅ Catalog Module initialized")

	// Create Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			log.Printf("Error: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Internal Server Error",
			})
		},
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	})

	// Add i18n middleware
	app.Use(i18n.LocaleMiddleware(translator))

	// Setup routes
	identityHttp.SetupRouter(app, identityHttp.RouterConfig{
		AuthService:    authService,
		UserService:    userService,
		TenantService:  tenantService,
		Environment:    cfg.Server.Environment,
		AllowedOrigins: cfg.Security.CORS.AllowOrigins,
	})

	catalogHttp.SetupCatalogRoutes(app, novelSvc, volumeSvc, chapterSvc, creatorSvc, genreSvc, characterSvc)

	log.Println("✅ Routes configured")

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		portStr := strconv.Itoa(cfg.Server.Port)
		log.Printf("🚀 Server starting on http://localhost:%s", portStr)

		if err := app.Listen(":" + portStr); err != nil {
			log.Printf("Server error: %v", err)
		}
	}()

	// Wait for shutdown signal
	<-sigChan
	log.Println("Shutting down gracefully...")

	// Create shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	// Shutdown Fiber server
	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		log.Printf("Error during server shutdown: %v", err)
	}

	log.Println("✅ Shutdown completed")
	// Database will be closed by defer db.Close()
}
