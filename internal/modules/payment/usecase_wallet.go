package payment

import (
	"context"

	"github.com/gofrs/uuid/v5"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"

	"system/internal/domain"
)

type walletUseCase struct {
	walletRepo domain.WalletRepository
	logger     *zap.Logger
}

func NewWalletUseCase(walletRepo domain.WalletRepository, logger *zap.Logger) WalletUseCase {
	return &walletUseCase{
		walletRepo: walletRepo,
		logger:     logger,
	}
}

func (uc *walletUseCase) GetWallet(ctx context.Context, userID uuid.UUID) (*domain.UserWallet, error) {
	return uc.walletRepo.GetOrCreateByUserID(ctx, userID)
}

func (uc *walletUseCase) GetBalance(ctx context.Context, userID uuid.UUID) (decimal.Decimal, error) {
	wallet, err := uc.walletRepo.GetOrCreateByUserID(ctx, userID)
	if err != nil {
		return decimal.Zero, err
	}
	return wallet.CoinBalance, nil
}
