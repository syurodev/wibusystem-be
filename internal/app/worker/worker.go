package worker

import (
	"context"
	"time"

	"go.uber.org/zap"

	"system/configs"
	analytics_module "system/internal/modules/analytics"
)

// StartViewTrackingWorkers khởi động các background worker cho view tracking service.
func StartViewTrackingWorkers(cfg *configs.ViewTrackingConfig, viewTrackingService *analytics_module.ViewTrackingService, zapLogger *zap.Logger) {
	if !cfg.WorkerEnabled {
		return
	}

	// Worker 1: Sync buffers from Redis to Postgres
	go func() {
		defer func() {
			if r := recover(); r != nil {
				zapLogger.Error("Panic in PostgreSQL sync worker", zap.Any("panic", r))
			}
		}()

		ticker := time.NewTicker(time.Duration(cfg.SyncIntervalMinutes) * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			func() {
				defer func() {
					if r := recover(); r != nil {
						zapLogger.Error("Panic during PostgreSQL sync", zap.Any("panic", r))
					}
				}()
				if err := viewTrackingService.SyncBuffersToPostgreSQL(context.Background()); err != nil {
					zapLogger.Error("Failed to sync view buffers to PostgreSQL", zap.Error(err))
				}
			}()
		}
	}()

	// Worker 2: Sync events from Redis to ClickHouse
	go func() {
		defer func() {
			if r := recover(); r != nil {
				zapLogger.Error("Panic in ClickHouse sync worker", zap.Any("panic", r))
			}
		}()

		// Sync frequently for near real-time analytics (e.g., every 10 seconds)
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			func() {
				defer func() {
					if r := recover(); r != nil {
						zapLogger.Error("Panic during ClickHouse event sync", zap.Any("panic", r))
					}
				}()
				if err := viewTrackingService.SyncEventsToClickHouse(context.Background()); err != nil {
					zapLogger.Error("Failed to sync view events to ClickHouse", zap.Error(err))
				}
			}()
		}
	}()

	// Worker 3: Sync active readers (PostgreSQL <-> ClickHouse)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				zapLogger.Error("Panic in active readers sync worker", zap.Any("panic", r))
			}
		}()

		// Sync hourly for active readers (heavy query)
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			func() {
				defer func() {
					if r := recover(); r != nil {
						zapLogger.Error("Panic during active readers sync", zap.Any("panic", r))
					}
				}()
				if err := viewTrackingService.SyncActiveReaders(context.Background()); err != nil {
					zapLogger.Error("Failed to sync active readers", zap.Error(err))
				}
			}()
		}
	}()

	// Worker 4: Sync content activities from Redis to ClickHouse
	go func() {
		defer func() {
			if r := recover(); r != nil {
				zapLogger.Error("Panic in activities sync worker", zap.Any("panic", r))
			}
		}()

		// Sync frequently (e.g., every 30 seconds)
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			func() {
				defer func() {
					if r := recover(); r != nil {
						zapLogger.Error("Panic during activities sync", zap.Any("panic", r))
					}
				}()
				if err := viewTrackingService.SyncActivitiesToClickHouse(context.Background()); err != nil {
					zapLogger.Error("Failed to sync content activities to ClickHouse", zap.Error(err))
				}
			}()
		}
	}()
}
