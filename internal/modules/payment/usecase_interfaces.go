package payment

import (
	"context"

	"github.com/gofrs/uuid/v5"
	"github.com/shopspring/decimal"

	"system/internal/domain"
)

// ConfigUseCase defines use cases for payment configuration
type ConfigUseCase interface {
	GetAll(ctx context.Context) ([]*domain.PaymentConfiguration, error)
	GetByKey(ctx context.Context, key string) (*domain.PaymentConfiguration, error)
	GetByPrefix(ctx context.Context, prefix string) ([]*domain.PaymentConfiguration, error)
	Update(ctx context.Context, key string, value string, updatedBy uuid.UUID) (*domain.PaymentConfiguration, error)
	Create(ctx context.Context, config *domain.PaymentConfiguration) error
	Delete(ctx context.Context, key string) error
	UpsertMany(ctx context.Context, configs []*domain.PaymentConfiguration) error

	// Helpers
	GetString(ctx context.Context, key string) (string, error)
	GetNumber(ctx context.Context, key string) (float64, error)
	GetBool(ctx context.Context, key string) (bool, error)
	GetJSON(ctx context.Context, key string, target interface{}) error
}

// WalletUseCase defines use cases for user wallet
type WalletUseCase interface {
	GetWallet(ctx context.Context, userID uuid.UUID) (*domain.UserWallet, error)
	GetBalance(ctx context.Context, userID uuid.UUID) (decimal.Decimal, error)
}

// TopupUseCase defines use cases for coin top-up
type TopupUseCase interface {
	ListPackages(ctx context.Context) ([]*domain.CoinPackage, error)
	GetPackage(ctx context.Context, packageID uuid.UUID) (*domain.CoinPackage, error)
	CreateTopupOrder(ctx context.Context, userID uuid.UUID, packageID uuid.UUID) (*domain.TopupOrder, error)
	GetTopupOrder(ctx context.Context, orderID uuid.UUID) (*domain.TopupOrder, error)
	GetTopupOrderByCode(ctx context.Context, orderCode string) (*domain.TopupOrder, error)
	ListTopupOrders(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.TopupOrder, int, error)
	CancelTopupOrder(ctx context.Context, userID uuid.UUID, orderID uuid.UUID) error
	ProcessWebhook(ctx context.Context, transactionID string, amount int64, content string) error
	ExpirePendingOrders(ctx context.Context) (int64, error)
}

// TopupNotifier is interface for sending notifications
type TopupNotifier interface {
	SendToUser(userID uuid.UUID, message []byte)
}

// TransactionUseCase defines use cases for transaction history
type TransactionUseCase interface {
	ListTransactions(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.Transaction, int, error)
}
