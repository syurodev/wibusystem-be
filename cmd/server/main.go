package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"system/internal/app/router"
	"time"

	"go.uber.org/zap"

	"system/configs"
	"system/internal/platform/database"
	"system/internal/platform/i18n"
	"system/internal/platform/logger"
)

func main() {
	// 1. Tải cấu hình từ file .env
	cfg, err := configs.LoadConfig(".env")
	if err != nil {
		// Log.Fatal hoặc panic nếu cấu hình bị lỗi
		panic("Failed to load configuration: " + err.Error())
	}

	// 2. Khởi tạo Logger
	appLogger, err := logger.InitLogger(cfg.Server.IsProd)
	if err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}

	// Đảm bảo logger buffer được flush khi main kết thúc
	defer logger.SyncLogger()

	appLogger.Info("Application logger initialized successfully.")

	// 3. Khởi tạo I18n
	if err := i18n.InitI18n(appLogger); err != nil {
		appLogger.Fatal("Failed to initialize i18n bundle.", zap.Error(err))
	}

	appLogger.Info("I18n bundle initialized successfully.")

	// 4. Khởi tạo Database Connection
	// Tạo context với timeout cho việc khởi tạo database
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := database.NewPostgresDB(ctx, &cfg.DB, cfg.Log.DBLogQueries, appLogger)
	if err != nil {
		appLogger.Fatal("Failed to initialize database connection", zap.Error(err))
	}

	// Đảm bảo database connection được đóng khi application shutdown
	defer db.Close()

	appLogger.Info("Database connection initialized successfully.")

	// 5. Health check database
	healthInfo, err := db.Health(context.Background())
	if err != nil {
		appLogger.Warn("Database health check warning", zap.Error(err))
	} else {
		appLogger.Info("Database health check passed", zap.Any("health_info", healthInfo))
	}

	// 6. Khởi tạo Redis Connection
	redisCtx, redisCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer redisCancel()

	rdb, err := database.NewRedisClient(redisCtx, &cfg.Redis, appLogger)
	if err != nil {
		appLogger.Fatal("Failed to initialize Redis connection", zap.Error(err))
	}

	// Đảm bảo Redis connection được đóng khi application shutdown
	defer rdb.Close()

	appLogger.Info("Redis connection initialized successfully.")

	// 7. Health check Redis
	redisHealthInfo, err := rdb.Health(context.Background())
	if err != nil {
		appLogger.Warn("Redis health check warning", zap.Error(err))
	} else {
		appLogger.Info("Redis health check passed", zap.Any("health_info", redisHealthInfo))
	}

	// 8. Khởi tạo ClickHouse Connection
	chCtx, chCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer chCancel()

	ch, err := database.NewClickHouseClient(chCtx, &cfg.ClickHouse, appLogger)
	if err != nil {
		appLogger.Fatal("Failed to initialize ClickHouse connection", zap.Error(err))
	}
	defer ch.Close()

	appLogger.Info("ClickHouse connection initialized successfully.")

	// 9. Health check ClickHouse
	chHealthInfo, err := ch.Health(context.Background())
	if err != nil {
		appLogger.Warn("ClickHouse health check warning", zap.Error(err))
	} else {
		appLogger.Info("ClickHouse health check passed", zap.Any("health_info", chHealthInfo))
	}

	// 10. Khởi tạo Gin Router
	appRouter := router.NewRouter(cfg, i18n.GetInstance(), appLogger, db, rdb, ch)
	appLogger.Info("Gin router initialized successfully.")

	// 11. Khởi tạo và chạy HTTP Server
	srv := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: appRouter,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.Fatal("Failed to start HTTP server", zap.Error(err))
		}
	}()

	appLogger.Info("HTTP Server started", zap.String("port", cfg.Server.Port))

	// 12. Setup graceful shutdown
	// Tạo channel để listen shutdown signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	appLogger.Info("Application started successfully. Press Ctrl+C to shutdown.")

	// Đợi signal shutdown
	<-quit

	appLogger.Info("Shutting down application gracefully...")

	// Thực hiện graceful shutdown cho server
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		appLogger.Error("Server forced to shutdown", zap.Error(err))
	}

	appLogger.Info("Application shutdown completed.")
}
