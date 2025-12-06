package worker

import (
	"context"
	"time"

	"go.uber.org/zap"

	"system/configs"
	"system/internal/pkg/service"
)

// StartViewTrackingWorkers khởi động các background worker cho view tracking service.
func StartViewTrackingWorkers(cfg *configs.ViewTrackingConfig, viewTrackingService *service.ViewTrackingService, zapLogger *zap.Logger) {
	if !cfg.WorkerEnabled {
		return
	}

	// Worker 1: Sync buffers from Redis to Postgres
	go func() {
		ticker := time.NewTicker(time.Duration(cfg.SyncIntervalMinutes) * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			if err := viewTrackingService.SyncBuffersToPostgreSQL(context.Background()); err != nil {
				zapLogger.Error("Failed to sync view buffers to PostgreSQL", zap.Error(err))
			}
		}
	}()

	// Worker 2: Sync events from Redis to ClickHouse
	go func() {
		// Sync frequently for near real-time analytics (e.g., every 10 seconds)
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			if err := viewTrackingService.SyncEventsToClickHouse(context.Background()); err != nil {
				zapLogger.Error("Failed to sync view events to ClickHouse", zap.Error(err))
			}
		}
	}()

	// Worker 3: Sync active readers (PostgreSQL <-> ClickHouse)
	go func() {
		// Sync hourly for active readers (heavy query)
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			if err := viewTrackingService.SyncActiveReaders(context.Background()); err != nil {
				zapLogger.Error("Failed to sync active readers", zap.Error(err))
			}
		}
	}()
}
