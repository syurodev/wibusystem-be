/*
Worker Manager - Unified Background Worker Orchestration
=========================================================

WorkerManager quản lý lifecycle của tất cả background workers:
- ViewSyncWorker: Sync view data từ Redis sang PostgreSQL/ClickHouse
- EmbeddingWorker: Xử lý embedding queue
- RankSnapshotWorker: Tạo rank snapshots định kỳ
- PaymentWorker: Xử lý payment tasks

LIFECYCLE:
----------
                    ┌─────────────────┐
                    │  Application    │
                    │     Start       │
                    └────────┬────────┘
                             │
                             ▼
                  ┌──────────────────────┐
                  │  WorkerManager.Start │
                  └──────────┬───────────┘
                             │
         ┌───────────────────┼───────────────────┐
         │                   │                   │
         ▼                   ▼                   ▼
   ViewSyncWorker    EmbeddingWorker    RankSnapshotWorker
      .Start()          .Start()            .Start()
         │                   │                   │
         └───────────────────┼───────────────────┘
                             │
                       [Running...]
                             │
                  ┌──────────────────────┐
                  │  WorkerManager.Stop  │   ← Graceful shutdown
                  └──────────┬───────────┘
                             │
         ┌───────────────────┼───────────────────┐
         │                   │                   │
         ▼                   ▼                   ▼
   ViewSyncWorker    EmbeddingWorker    RankSnapshotWorker
       .Stop()          .Stop()             .Stop()
         │                   │                   │
         └───────────────────┼───────────────────┘
                             │
                        [All Done]
*/

package worker

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"

	"system/configs"
	"system/internal/domain"
	"system/internal/modules/analytics"
	"system/internal/modules/embedding"
	"system/internal/modules/novel"
	payment_module "system/internal/modules/payment"
)

// WorkerManager manages all background workers lifecycle.
type WorkerManager struct {
	viewSyncWorker     *ViewSyncWorker
	embeddingWorker    *EmbeddingWorker
	rankSnapshotWorker *RankSnapshotWorker
	paymentWorker      *PaymentWorker
	logger             *zap.Logger
	wg                 sync.WaitGroup
}

// NewWorkerManager creates a new WorkerManager with all workers.
func NewWorkerManager(
	// ViewSync dependencies
	viewTrackingService *analytics.ViewTrackingService,
	viewTrackingConfig *configs.ViewTrackingConfig,
	// Embedding dependencies
	embeddingService *embedding.Service,
	novelService novel.NovelService,
	embeddingConfig *configs.EmbeddingConfig,
	// RankSnapshot dependencies
	analyticsRepo domain.ViewAnalyticsRepository,
	// Payment dependencies
	topupUseCase payment_module.TopupUseCase,
	// Common
	logger *zap.Logger,
) *WorkerManager {
	return &WorkerManager{
		viewSyncWorker: NewViewSyncWorker(
			viewTrackingService,
			logger,
			viewTrackingConfig,
		),
		embeddingWorker: NewEmbeddingWorker(
			embeddingService,
			novelService,
			logger,
			embeddingConfig,
		),
		rankSnapshotWorker: NewRankSnapshotWorker(
			analyticsRepo,
			logger,
		),
		paymentWorker: NewPaymentWorker(
			topupUseCase,
			logger,
		),
		logger: logger,
	}
}

// Start starts all background workers.
// This method is non-blocking - workers run in background goroutines.
func (m *WorkerManager) Start(ctx context.Context) {
	m.logger.Info("WorkerManager: Starting all background workers...")

	// Start ViewSyncWorker
	m.viewSyncWorker.Start(ctx)
	m.logger.Info("WorkerManager: ViewSyncWorker started")

	// Start EmbeddingWorker
	m.embeddingWorker.Start(ctx)
	m.logger.Info("WorkerManager: EmbeddingWorker started")

	// Start RankSnapshotWorker
	m.rankSnapshotWorker.Start()
	m.logger.Info("WorkerManager: RankSnapshotWorker started")

	// Start PaymentWorker
	m.paymentWorker.Start(ctx)
	m.logger.Info("WorkerManager: PaymentWorker started")

	m.logger.Info("WorkerManager: All background workers started successfully")
}

// Stop gracefully stops all workers with timeout.
// Workers are stopped in reverse order of starting.
func (m *WorkerManager) Stop(ctx context.Context) error {
	m.logger.Info("WorkerManager: Stopping all background workers...")

	// Use WaitGroup to wait for all workers to stop
	var stopErr error
	errChan := make(chan error, 4)

	// Stop in parallel for faster shutdown
	m.wg.Add(4)

	// Stop PaymentWorker
	go func() {
		defer m.wg.Done()
		m.paymentWorker.Stop()
		m.logger.Info("WorkerManager: PaymentWorker stopped")
	}()

	// Stop RankSnapshotWorker
	go func() {
		defer m.wg.Done()
		m.rankSnapshotWorker.Stop()
		m.logger.Info("WorkerManager: RankSnapshotWorker stopped")
	}()

	// Stop EmbeddingWorker
	go func() {
		defer m.wg.Done()
		if err := m.embeddingWorker.Stop(ctx); err != nil {
			errChan <- err
		}
		m.logger.Info("WorkerManager: EmbeddingWorker stopped")
	}()

	// Stop ViewSyncWorker
	go func() {
		defer m.wg.Done()
		if err := m.viewSyncWorker.Stop(ctx); err != nil {
			errChan <- err
		}
		m.logger.Info("WorkerManager: ViewSyncWorker stopped")
	}()

	// Wait for all workers to stop or context timeout
	done := make(chan struct{})
	go func() {
		m.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		close(errChan)
		for err := range errChan {
			if err != nil {
				stopErr = err
			}
		}
		m.logger.Info("WorkerManager: All workers stopped successfully")
	case <-ctx.Done():
		m.logger.Warn("WorkerManager: Stop timeout, some workers may not have stopped gracefully")
		return ctx.Err()
	}

	return stopErr
}

// PaymentWorker wraps payment background tasks.
type PaymentWorker struct {
	topupUseCase payment_module.TopupUseCase
	logger       *zap.Logger
	stopChan     chan struct{}
	doneChan     chan struct{}
}

// NewPaymentWorker creates a new PaymentWorker.
func NewPaymentWorker(
	topupUseCase payment_module.TopupUseCase,
	logger *zap.Logger,
) *PaymentWorker {
	return &PaymentWorker{
		topupUseCase: topupUseCase,
		logger:       logger,
		stopChan:     make(chan struct{}),
		doneChan:     make(chan struct{}),
	}
}

// Start starts the payment worker.
func (w *PaymentWorker) Start(ctx context.Context) {
	w.logger.Info("Starting Payment worker...")

	go func() {
		defer close(w.doneChan)

		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				count, err := w.topupUseCase.ExpirePendingOrders(context.Background())
				if err != nil {
					w.logger.Error("Failed to expire pending orders", zap.Error(err))
				} else if count > 0 {
					w.logger.Info("Expired pending orders", zap.Int64("count", count))
				}
			case <-w.stopChan:
				w.logger.Info("Payment worker received stop signal")
				return
			case <-ctx.Done():
				return
			}
		}
	}()
}

// Stop stops the payment worker.
func (w *PaymentWorker) Stop() {
	close(w.stopChan)
	<-w.doneChan
}
