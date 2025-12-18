package payment

import (
	"context"

	"github.com/gofrs/uuid/v5"
	"go.uber.org/zap"

	"system/internal/domain"
)

type transactionUseCase struct {
	txRepo domain.TransactionRepository
	logger *zap.Logger
}

func NewTransactionUseCase(txRepo domain.TransactionRepository, logger *zap.Logger) TransactionUseCase {
	return &transactionUseCase{
		txRepo: txRepo,
		logger: logger,
	}
}

func (uc *transactionUseCase) ListTransactions(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.Transaction, int, error) {
	return uc.txRepo.ListByUserID(ctx, userID, limit, offset)
}
