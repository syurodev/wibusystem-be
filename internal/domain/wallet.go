package domain

import (
	"context"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/shopspring/decimal"
)

// UserWallet represents a user's coin wallet
type UserWallet struct {
	ID                    uuid.UUID
	UserID                uuid.UUID
	CoinBalance           decimal.Decimal
	TotalDeposited        decimal.Decimal
	TotalSpent            decimal.Decimal
	TotalSubscriptionSpent decimal.Decimal
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

// CoinPackage represents a coin top-up package
type CoinPackage struct {
	ID           uuid.UUID
	Name         string
	Slug         string
	CoinAmount   decimal.Decimal
	PriceVND     decimal.Decimal
	BonusPercent int
	IsPopular    bool
	IsActive     bool
	DisplayOrder int
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// TotalCoins returns the total coins including bonus
func (p *CoinPackage) TotalCoins() decimal.Decimal {
	bonus := p.CoinAmount.Mul(decimal.NewFromInt(int64(p.BonusPercent))).Div(decimal.NewFromInt(100))
	return p.CoinAmount.Add(bonus)
}

// BonusCoins returns only the bonus coins
func (p *CoinPackage) BonusCoins() decimal.Decimal {
	return p.CoinAmount.Mul(decimal.NewFromInt(int64(p.BonusPercent))).Div(decimal.NewFromInt(100))
}

// TopupStatus represents the status of a top-up order
type TopupStatus string

const (
	TopupStatusPending   TopupStatus = "pending"
	TopupStatusSuccess   TopupStatus = "success"
	TopupStatusExpired   TopupStatus = "expired"
	TopupStatusCancelled TopupStatus = "cancelled"
	TopupStatusFailed    TopupStatus = "failed"
)

// IsValid kểm tra xem status có hợp lệ không
func (s TopupStatus) IsValid() bool {
	switch s {
	case TopupStatusPending, TopupStatusSuccess, TopupStatusExpired, TopupStatusCancelled, TopupStatusFailed:
		return true
	default:
		return false
	}
}

// TopupOrder represents a coin top-up order
type TopupOrder struct {
	ID                  uuid.UUID
	UserID              uuid.UUID
	PackageID           uuid.UUID
	OrderCode           string
	CoinAmount          decimal.Decimal // Total coins (base + bonus)
	BaseCoinAmount      decimal.Decimal // Base coins from package
	BonusCoinAmount     decimal.Decimal // Bonus coins
	VNDAmount           decimal.Decimal
	Status              TopupStatus
	SepayTransactionID  *string
	SepayContent        *string
	BankName            *string
	BankAccount         *string
	AccountName         *string
	CompletedAt         *time.Time
	ExpiredAt           time.Time
	CreatedAt           time.Time
	UpdatedAt           time.Time

	// Loaded via JOIN
	Package *CoinPackage `db:"-"`
}

// IsExpired checks if the order has expired
func (o *TopupOrder) IsExpired() bool {
	return time.Now().After(o.ExpiredAt)
}

// TransactionType represents the type of transaction
type TransactionType string

const (
	TransactionTypeTopup           TransactionType = "topup"
	TransactionTypePurchaseChapter TransactionType = "purchase_chapter"
	TransactionTypePurchaseSeries  TransactionType = "purchase_series"
	TransactionTypeRental          TransactionType = "rental"
	TransactionTypeSubscription    TransactionType = "subscription"
	TransactionTypeRefund          TransactionType = "refund"
	TransactionTypeAdminAdjustment TransactionType = "admin_adjustment"
)

// IsValid kiểm tra xem transaction type có hợp lệ không
func (t TransactionType) IsValid() bool {
	switch t {
	case TransactionTypeTopup, TransactionTypePurchaseChapter, TransactionTypePurchaseSeries,
		TransactionTypeRental, TransactionTypeSubscription, TransactionTypeRefund, TransactionTypeAdminAdjustment:
		return true
	default:
		return false
	}
}

// NotificationType represents the type of WebSocket notification
type NotificationType string

const (
	NotificationTypeTopupSuccess NotificationType = "TOPUP_SUCCESS"
	NotificationTypeTopupFailed  NotificationType = "TOPUP_FAILED"
)

// Transaction represents a coin transaction
type Transaction struct {
	ID                 uuid.UUID
	UserID             uuid.UUID
	Type               TransactionType
	CoinAmount         decimal.Decimal // Positive for deposits, negative for spending
	VNDAmount          *decimal.Decimal
	BalanceAfter       decimal.Decimal
	ReferenceType      *string
	ReferenceID        *uuid.UUID
	CreatorID          *uuid.UUID
	CreatorRevenueVND  *decimal.Decimal
	PlatformRevenueVND *decimal.Decimal
	Description        *string
	Metadata           map[string]interface{}
	CreatedAt          time.Time
}

// WalletRepository defines the interface for wallet data access
type WalletRepository interface {
	// GetByUserID retrieves a wallet by user ID
	GetByUserID(ctx context.Context, userID uuid.UUID) (*UserWallet, error)

	// GetOrCreateByUserID gets or creates a wallet for a user
	GetOrCreateByUserID(ctx context.Context, userID uuid.UUID) (*UserWallet, error)

	// AddCoins adds coins to a user's wallet (with transaction)
	AddCoins(ctx context.Context, userID uuid.UUID, amount decimal.Decimal) (*UserWallet, error)

	// DeductCoins deducts coins from a user's wallet (with transaction)
	DeductCoins(ctx context.Context, userID uuid.UUID, amount decimal.Decimal) (*UserWallet, error)

	// LockForUpdate locks the wallet row for update (use within transaction)
	LockForUpdate(ctx context.Context, userID uuid.UUID) (*UserWallet, error)
}

// CoinPackageRepository defines the interface for coin package data access
type CoinPackageRepository interface {
	// GetByID retrieves a package by ID
	GetByID(ctx context.Context, id uuid.UUID) (*CoinPackage, error)

	// GetBySlug retrieves a package by slug
	GetBySlug(ctx context.Context, slug string) (*CoinPackage, error)

	// ListActive retrieves all active packages
	ListActive(ctx context.Context) ([]*CoinPackage, error)
}

// TopupOrderRepository defines the interface for topup order data access
type TopupOrderRepository interface {
	// Create creates a new topup order
	Create(ctx context.Context, order *TopupOrder) error

	// GetByID retrieves an order by ID
	GetByID(ctx context.Context, id uuid.UUID) (*TopupOrder, error)

	// GetByOrderCode retrieves an order by order code
	GetByOrderCode(ctx context.Context, orderCode string) (*TopupOrder, error)

	// GetBySepayTransactionID retrieves an order by SePay transaction ID
	GetBySepayTransactionID(ctx context.Context, txID string) (*TopupOrder, error)

	// UpdateStatus updates the order status
	UpdateStatus(ctx context.Context, id uuid.UUID, status TopupStatus, sepayTxID *string, sepayContent *string) error

	// ListByUserID retrieves orders by user ID
	ListByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*TopupOrder, int, error)

	// ExpirePendingOrders marks expired pending orders as expired
	ExpirePendingOrders(ctx context.Context) (int64, error)
}

// TransactionRepository defines the interface for transaction data access
type TransactionRepository interface {
	// Create creates a new transaction
	Create(ctx context.Context, tx *Transaction) error

	// ListByUserID retrieves transactions by user ID
	ListByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*Transaction, int, error)

	// GetByReferenceID retrieves a transaction by reference ID
	GetByReferenceID(ctx context.Context, refType string, refID uuid.UUID) (*Transaction, error)
}
