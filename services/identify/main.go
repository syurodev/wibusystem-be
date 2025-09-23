// Package main boots the Identity Service HTTP server, wires up
// configuration, database, OAuth2 provider, routing, and middleware.
//
// This file intentionally keeps logic focused on composition rather than
// business rules. Handlers, middleware, repositories, and OAuth2 provider
// implementations live in their respective packages.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"runtime"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"wibusystem/pkg/database/factory"
	"wibusystem/pkg/i18n"
	"wibusystem/services/identify/config"
	identitygrpc "wibusystem/services/identify/grpc"
	"wibusystem/services/identify/routes"
)

// main is the entry point for the Identity Service.
// It loads configuration, initializes dependencies, registers routes, and
// starts the Gin HTTP server with graceful shutdown.
func main() {
	log.Println("Identity Service starting...")

	// Load environment from .env files first (service-local), then OS env
	loadEnvFiles()

	// Load configuration
	cfg := config.Load()

	serviceDir := getServiceDir()

	// Point to pkg/i18n/locales instead of service-local locales
	bundlePath := filepath.Join(serviceDir, "../../pkg/i18n/locales")

	// Initialize localization bundle with new i18n config
	i18nConfig := i18n.Config{
		BundlePath:             bundlePath,
		DefaultLanguage:        cfg.Localization.DefaultLanguage,
		SupportedLanguages:     cfg.Localization.SupportedLanguages,
		FallbackToDefault:      true,
		ServiceNamespace:       "identify",
		LoadCommonTranslations: true,
		QueryParam:             cfg.Localization.QueryParam,
		HeaderName:             "Accept-Language",
		CookieName:             "locale",
	}
	translator, err := i18n.NewTranslator(i18nConfig)
	if err != nil {
		log.Fatalf("Failed to initialize localization: %v", err)
	}

	// Set Gin mode based on the environment
	if cfg.Server.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize database manager
	dbManager, err := factory.NewDatabaseManager(cfg.Database.DatabaseConfig)
	if err != nil {
		log.Fatalf("Failed to create database manager: %v", err)
	}

	// Connect to databases
	ctx := context.Background()
	if err := dbManager.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect to databases: %v", err)
	}
	defer dbManager.Close()

	// Setup and run database migrations
	if err := setupAndRunMigrations(ctx, cfg, dbManager); err != nil {
		log.Printf("Migration warning: %v", err)
	}

	// Initialize routing dependencies
	deps, err := routes.NewDependencies(dbManager, cfg, translator)
	if err != nil {
		log.Fatalf("Failed to initialize routing dependencies: %v", err)
	}

	// Setup router with all routes and middleware
	router := routes.SetupRouter(deps)

	// Setup gRPC server with both token validation and user services
	grpcServer, err := identitygrpc.SetupGRPCServer(deps.Provider, deps.UserService, cfg.GRPC.ServerConfig)
	if err != nil {
		log.Fatalf("Failed to setup gRPC server: %v", err)
	}

	// Start gRPC server (contains both services)
	identitygrpc.StartGRPCServer(grpcServer)

	// Create an HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start HTTP server in a goroutine
	go func() {
		log.Printf("Identity Service HTTP server listening on %s:%d", cfg.Server.Host, cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP Server failed to start: %v", err)
		}
	}()

	// Wait for the interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Create a context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown HTTP server
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("HTTP Server forced to shutdown: %v", err)
	}

	// Stop gRPC server
	grpcServer.Stop()

	log.Println("Servers gracefully stopped")
}

// loadEnvFiles loads environment variables from .env files located next to the
// service code, allowing local overrides without polluting OS environment.
// Order: .env.local then .env (later entries do not override earlier ones).
func loadEnvFiles() {
	// Determine the directory of this source file at runtime
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		return
	}
	serviceDir := filepath.Dir(thisFile)

	// Try to load .env.local first, then .env
	// Ignore errors if files are missing
	_ = godotenv.Load(filepath.Join(serviceDir, ".env.local"))
	_ = godotenv.Overload(filepath.Join(serviceDir, ".env"))
}

func getServiceDir() string {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		return "."
	}
	return filepath.Dir(thisFile)
}

func resolveBundlePath(configuredPath, serviceDir string) string {
	if configuredPath == "" {
		configuredPath = "locales"
	}

	if filepath.IsAbs(configuredPath) {
		return configuredPath
	}

	if _, err := os.Stat(configuredPath); err == nil {
		abs, absErr := filepath.Abs(configuredPath)
		if absErr == nil {
			return abs
		}
		return configuredPath
	}

	candidate := filepath.Join(serviceDir, configuredPath)
	if _, err := os.Stat(candidate); err == nil {
		return candidate
	}

	return configuredPath
}

// setupAndRunMigrations configures and runs database migrations using the new migration system.
// It is safe to call at startup; errors are returned to allow soft-fail
// handling (e.g., logging a warning) by the caller.
func setupAndRunMigrations(ctx context.Context, cfg *config.Config, dbManager *factory.DatabaseManager) error {
	// Get the migrations base path from config
	migrationsBasePath := cfg.Database.MigrationsPath
	if migrationsBasePath == "" {
		log.Printf("No migrations path configured, skipping migrations")
		return nil
	}

	// For file-based migrations, ensure the path exists
	if !filepath.IsAbs(migrationsBasePath) {
		// Make relative to service directory
		serviceDir := getServiceDir()
		migrationsBasePath = filepath.Join(serviceDir, migrationsBasePath)
	}

	// Setup migrations for all configured databases
	if err := dbManager.SetupMigrations(ctx, migrationsBasePath); err != nil {
		return fmt.Errorf("failed to setup migrations: %w", err)
	}

	// Run migrations
	if err := dbManager.RunMigrations(ctx); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Get and log migration status
	status, err := dbManager.GetMigrationStatus(ctx)
	if err != nil {
		log.Printf("Warning: failed to get migration status: %v", err)
	} else {
		for dbType, migrationStatus := range status {
			if migrationStatus.Error != nil {
				log.Printf("Migration status for %s: Error - %v", dbType, migrationStatus.Error)
			} else {
				log.Printf("Migration status for %s: Version %d, Dirty: %t", dbType, migrationStatus.Version, migrationStatus.Dirty)
			}
		}
	}

	return nil
}
