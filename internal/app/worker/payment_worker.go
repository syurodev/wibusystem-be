package worker

import (
	"context"
	"time"

	"go.uber.org/zap"

	payment_module "system/internal/modules/payment"
)

// StartPaymentWorkers starts background workers for payment module
func StartPaymentWorkers(topupUseCase payment_module.TopupUseCase, zapLogger *zap.Logger) {
	zapLogger.Info("Starting Payment workers...")

	// Worker 1: Expire pending orders
	go func() {
		// Run every minute
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			count, err := topupUseCase.ExpirePendingOrders(context.Background())
			if err != nil {
				zapLogger.Error("Failed to expire pending orders", zap.Error(err))
			} else if count > 0 {
				zapLogger.Info("Expired pending orders", zap.Int64("count", count))
			}
		}
	}()
}
