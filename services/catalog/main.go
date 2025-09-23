// Package main bootstraps the Catalog service HTTP server. The structure mirrors
// the Identify service so engineers experience a consistent setup across
// services: load configuration, initialise shared dependencies, configure
// routes, and start the HTTP server with graceful shutdown.
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
	"runtime"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"wibusystem/pkg/database/factory"
	"wibusystem/pkg/i18n"
	"wibusystem/services/catalog/config"
	"wibusystem/services/catalog/routes"
)

func main() {
	log.Println("Catalog Service starting...")

	loadEnvFiles()

	cfg := config.Load()

	serviceDir := getServiceDir()

	bundlePath := resolveBundlePath(cfg.Localization.BundlePath, serviceDir)

	i18nConfig := i18n.Config{
		BundlePath:             bundlePath,
		DefaultLanguage:        cfg.Localization.DefaultLanguage,
		SupportedLanguages:     cfg.Localization.SupportedLanguages,
		FallbackToDefault:      true,
		ServiceNamespace:       "catalog",
		LoadCommonTranslations: true,
		QueryParam:             cfg.Localization.QueryParam,
		HeaderName:             cfg.Localization.HeaderName,
		CookieName:             cfg.Localization.CookieName,
	}

	translator, err := i18n.NewTranslator(i18nConfig)
	if err != nil {
		log.Fatalf("Failed to initialise localization: %v", err)
	}

	if cfg.Server.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	dbManager, err := factory.NewDatabaseManager(cfg.Database.DatabaseConfig)
	if err != nil {
		log.Fatalf("Failed to create database manager: %v", err)
	}

	ctx := context.Background()
	if err := dbManager.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect to databases: %v", err)
	}
	defer func() {
		if closeErr := dbManager.Close(); closeErr != nil {
			log.Printf("Error closing database manager: %v", closeErr)
		}
	}()

	if err := setupAndRunMigrations(ctx, cfg, serviceDir, dbManager); err != nil {
		log.Printf("Migration warning: %v", err)
	}

	deps, err := routes.NewDependencies(dbManager, cfg, translator)
	if err != nil {
		log.Fatalf("Failed to initialise routing dependencies: %v", err)
	}

	router := routes.SetupRouter(deps)

	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	go func() {
		log.Printf("Catalog Service HTTP server listening on %s:%d", cfg.Server.Host, cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	log.Println("Catalog Service shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server forced to shutdown: %v", err)
	}

	log.Println("Catalog Service stopped")
}

func loadEnvFiles() {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		return
	}
	serviceDir := filepath.Dir(thisFile)

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

func setupAndRunMigrations(ctx context.Context, cfg *config.Config, serviceDir string, dbManager *factory.DatabaseManager) error {
	migrationsPath := cfg.Database.MigrationsPath
	if migrationsPath == "" {
		log.Printf("No migrations path configured, skipping migrations")
		return nil
	}

	if !filepath.IsAbs(migrationsPath) {
		migrationsPath = filepath.Join(serviceDir, migrationsPath)
	}

	if err := dbManager.SetupMigrations(ctx, migrationsPath); err != nil {
		return fmt.Errorf("failed to setup migrations: %w", err)
	}

	if err := dbManager.RunMigrations(ctx); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	status, err := dbManager.GetMigrationStatus(ctx)
	if err != nil {
		log.Printf("Warning: unable to fetch migration status: %v", err)
		return nil
	}

	for dbType, st := range status {
		if st.Error != nil {
			log.Printf("Migration status for %s: error %v", dbType, st.Error)
			continue
		}
		log.Printf("Migration status for %s: version %d, dirty=%t", dbType, st.Version, st.Dirty)
	}

	return nil
}
