package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"system/configs"
	"system/internal/di"
	"system/internal/platform/database"
	"system/internal/platform/i18n"
	"system/internal/platform/logger"
)

func main() {
	// 1. Load configuration from .env
	cfg, err := configs.LoadConfig(".env")
	if err != nil {
		panic("Failed to load configuration: " + err.Error())
	}

	// 2. Initialize Logger
	appLogger, err := logger.InitLogger(cfg.Server.IsProd)
	if err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
	defer logger.SyncLogger()

	appLogger.Info("Application logger initialized successfully.")

	// 3. Initialize I18n
	if err := i18n.InitI18n(appLogger); err != nil {
		appLogger.Fatal("Failed to initialize i18n bundle.", zap.Error(err))
	}
	appLogger.Info("I18n bundle initialized successfully.")

	// 4. Initialize Redis Connection
	redisCtx, redisCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer redisCancel()

	rdb, err := database.NewRedisClient(redisCtx, &cfg.Redis, appLogger)
	if err != nil {
		appLogger.Fatal("Failed to initialize Redis connection", zap.Error(err))
	}
	defer rdb.Close()
	appLogger.Info("Redis connection initialized successfully.")

	// 5. Health check Redis
	redisHealthInfo, err := rdb.Health(context.Background())
	if err != nil {
		appLogger.Warn("Redis health check warning", zap.Error(err))
	} else {
		appLogger.Info("Redis health check passed", zap.Any("health_info", redisHealthInfo))
	}

	// 6. Initialize ClickHouse Connection
	chCtx, chCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer chCancel()

	ch, err := database.NewClickHouseClient(chCtx, &cfg.ClickHouse, appLogger)
	if err != nil {
		appLogger.Fatal("Failed to initialize ClickHouse connection", zap.Error(err))
	}
	defer ch.Close()
	appLogger.Info("ClickHouse connection initialized successfully.")

	// 7. Health check ClickHouse
	chHealthInfo, err := ch.Health(context.Background())
	if err != nil {
		appLogger.Warn("ClickHouse health check warning", zap.Error(err))
	} else {
		appLogger.Info("ClickHouse health check passed", zap.Any("health_info", chHealthInfo))
	}

	// 8. Initialize Application using Wire-generated injector
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	app, err := di.InitializeApplication(ctx, cfg, appLogger, i18n.GetInstance(), rdb, ch)
	if err != nil {
		appLogger.Fatal("Failed to initialize application via Wire", zap.Error(err))
	}
	appLogger.Info("Application initialized successfully via Wire DI.")

	// 11. Start HTTP Server
	go func() {
		if err := app.HTTPServer.ListenAndServe(); err != nil && err.Error() != "http: Server closed" {
			appLogger.Fatal("Failed to start HTTP server", zap.Error(err))
		}
	}()
	appLogger.Info("HTTP Server started", zap.String("port", cfg.Server.Port))

	// 12. Setup graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	appLogger.Info("Application started successfully. Press Ctrl+C to shutdown.")

	// Wait for shutdown signal
	<-quit

	appLogger.Info("Shutting down application gracefully...")

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := app.HTTPServer.Shutdown(shutdownCtx); err != nil {
		appLogger.Error("Server forced to shutdown", zap.Error(err))
	}

	appLogger.Info("Application shutdown completed.")
}
