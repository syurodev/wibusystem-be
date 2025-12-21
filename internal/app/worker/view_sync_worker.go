/*
View Sync Worker - Background Sync Flow
========================================

Worker này chạy trong background goroutine và định kỳ đồng bộ data
từ Redis sang PostgreSQL và ClickHouse.

WORKER LIFECYCLE:
-----------------
                                    ┌─────────────────┐
                                    │   Application   │
                                    │     Start       │
                                    └────────┬────────┘
                                             │
                                             ▼
                              ┌──────────────────────────────┐
                              │     Check WorkerEnabled      │
                              │        (from config)         │
                              └──────────────┬───────────────┘
                                             │
                          ┌──────────────────┼──────────────────┐
                          ▼                                     ▼
                    [Disabled]                            [Enabled]
                          │                                     │
                          ▼                                     ▼
                   Close doneChan                     Go goroutine: run()
                   Return immediately                           │
                                                                │
                                    ┌───────────────────────────┘
                                    ▼
                         ┌──────────────────────┐
                         │  Immediate Sync      │ ←── Sync pending data từ lần trước
                         │  (on startup)        │
                         └──────────┬───────────┘
                                    │
                                    ▼
                    ┌─────────────────────────────────┐
                    │         MAIN LOOP               │
                    │                                 │
                    │   select {                      │
                    │     case <-ticker.C:            │ ←── Định kỳ (SyncIntervalMinutes)
                    │       sync()                    │
                    │                                 │
                    │     case <-stopChan:            │ ←── Graceful shutdown
                    │       sync()                    │     (final sync before exit)
                    │       return                    │
                    │                                 │
                    │     case <-ctx.Done():          │ ←── Context cancelled
                    │       return                    │
                    │   }                             │
                    └─────────────────────────────────┘

SYNC CYCLE (sync method):
-------------------------
┌─────────────────────────────────────────────────────────────────────────────────┐
│ 1. SyncBuffersToPostgreSQL                                                      │
│    Redis Buffers → Batch Update PostgreSQL (novels, chapters, genres)          │
└─────────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│ 2. SyncEventsToClickHouse (up to 10 iterations)                                │
│    Redis Event Queue → Batch Insert ClickHouse                                 │
└─────────────────────────────────────────────────────────────────────────────────┘

GRACEFUL SHUTDOWN:
------------------
Stop() → Signal stopChan → Worker performs final sync → Close doneChan → Return
*/

package worker

import (
	"context"
	"time"

	"go.uber.org/zap"

	"system/configs"
	analytics_module "system/internal/modules/analytics"
)

// ViewSyncWorker handles background synchronization of view counts.
// Worker này chạy trong goroutine và định kỳ sync data từ:
// - Redis buffers → PostgreSQL (view_count updates)
// - Redis event queue → ClickHouse (analytics events)
type ViewSyncWorker struct {
	viewTrackingService *analytics_module.ViewTrackingService
	logger              *zap.Logger
	config              *configs.ViewTrackingConfig
	stopChan            chan struct{}
	doneChan            chan struct{}
}

// NewViewSyncWorker creates a new background worker instance.
//
// Parameters:
//   - viewTrackingService: Service để handle sync logic
//   - logger: Zap logger instance
//   - config: ViewTracking configuration
//
// Returns:
//   - *ViewSyncWorker: Worker instance
func NewViewSyncWorker(
	viewTrackingService *analytics_module.ViewTrackingService,
	logger *zap.Logger,
	config *configs.ViewTrackingConfig,
) *ViewSyncWorker {
	return &ViewSyncWorker{
		viewTrackingService: viewTrackingService,
		logger:              logger,
		config:              config,
		stopChan:            make(chan struct{}),
		doneChan:            make(chan struct{}),
	}
}

// Start begins the background worker goroutine.
// Worker sẽ check config.WorkerEnabled trước khi start.
// Nếu disabled, close doneChan ngay lập tức.
//
// Parameters:
//   - ctx: Context (parent context của application)
func (w *ViewSyncWorker) Start(ctx context.Context) {
	if !w.config.WorkerEnabled {
		w.logger.Info("View sync worker is disabled by configuration")
		close(w.doneChan)
		return
	}

	w.logger.Info("Starting view sync worker",
		zap.Int("interval_minutes", w.config.SyncIntervalMinutes))

	go w.run(ctx)
}

// Stop gracefully stops the worker.
// Method này wait cho worker finish hoặc context cancel.
//
// Parameters:
//   - ctx: Context với timeout cho graceful shutdown
//
// Returns:
//   - error: ctx.Err() nếu timeout, nil nếu stopped gracefully
func (w *ViewSyncWorker) Stop(ctx context.Context) error {
	w.logger.Info("Stopping view sync worker...")

	// Signal worker to stop
	close(w.stopChan)

	// Wait for worker to finish or context to cancel
	select {
	case <-w.doneChan:
		w.logger.Info("View sync worker stopped gracefully")
		return nil
	case <-ctx.Done():
		w.logger.Warn("View sync worker stop timed out")
		return ctx.Err()
	}
}

// run is the main worker loop.
// Loop này:
//  1. Run sync ngay lập tức khi start
//  2. Chạy ticker với interval từ config
//  3. Mỗi tick → run sync cycle
//  4. Khi receive stop signal → run final sync → return
func (w *ViewSyncWorker) run(ctx context.Context) {
	defer close(w.doneChan)

	ticker := time.NewTicker(time.Duration(w.config.SyncIntervalMinutes) * time.Minute)
	defer ticker.Stop()

	// Run immediately on start để sync pending data
	w.sync(ctx)

	for {
		select {
		case <-ticker.C:
			// Periodic sync
			w.sync(ctx)

		case <-w.stopChan:
			// Graceful shutdown: perform final sync before exit
			w.logger.Info("Worker received stop signal, performing final sync...")
			w.sync(ctx)
			return

		case <-ctx.Done():
			// Context cancelled (e.g., application shutdown)
			w.logger.Warn("Worker context cancelled")
			return
		}
	}
}

// sync performs one sync cycle.
// Method này gọi cả 2 sync operations:
//  1. Sync Redis buffers → PostgreSQL
//  2. Sync events → ClickHouse (có thể chạy multiple times để drain queue)
//
// Parameters:
//   - ctx: Context (thường là parent context, nhưng tạo timeout context riêng cho sync cycle)
func (w *ViewSyncWorker) sync(ctx context.Context) {
	w.logger.Debug("View sync cycle started")
	syncStart := time.Now()

	// Create timeout context cho sync cycle này (2 phút timeout)
	syncCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	// 1. Sync Redis buffers to PostgreSQL
	if err := w.viewTrackingService.SyncBuffersToPostgreSQL(syncCtx); err != nil {
		w.logger.Error("Failed to sync buffers to PostgreSQL", zap.Error(err))
		// Don't return - continue to ClickHouse sync
	}

	// 2. Sync events to ClickHouse
	// Run multiple times to drain the queue if it's large
	// Tối đa 10 iterations để tránh worker bị block quá lâu
	for i := 0; i < 10; i++ {
		if err := w.viewTrackingService.SyncEventsToClickHouse(syncCtx); err != nil {
			w.logger.Error("Failed to sync events to ClickHouse",
				zap.Error(err),
				zap.Int("iteration", i))
			break
		}

		// If we got here without error, check if there are more events
		// (the service returns nil when queue is empty)
		// For now, we'll just do one batch per cycle
		// Có thể enhance sau: continue loop nếu batch size == config.ClickHouseBatchSize
		break
	}

	syncDuration := time.Since(syncStart)
	w.logger.Info("View sync cycle completed",
		zap.Duration("duration", syncDuration))

	// TODO: Emit metrics (if you have Prometheus)
	// metrics.ViewSyncDuration.Observe(syncDuration.Seconds())
	// metrics.ViewSyncLastSuccess.SetToCurrentTime()
}
