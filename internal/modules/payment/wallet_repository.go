package payment

import (
	"context"
	_ "embed"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"system/internal/domain"
)

//go:embed queries/wallet/get_by_user_id.sql
var getWalletByUserIDQuery string

//go:embed queries/wallet/get_or_create.sql
var getOrCreateWalletQuery string

//go:embed queries/wallet/lock_for_update.sql
var lockWalletForUpdateQuery string

//go:embed queries/wallet/add_coins.sql
var addCoinsQuery string

//go:embed queries/wallet/deduct_coins.sql
var deductCoinsQuery string

// walletRepository implements WalletRepository
type walletRepository struct {
	pool *pgxpool.Pool
}

// NewWalletRepository creates a new wallet repository
func NewWalletRepository(pool *pgxpool.Pool) domain.WalletRepository {
	return &walletRepository{pool: pool}
}

// GetByUserID retrieves a wallet by user ID
func (r *walletRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.UserWallet, error) {
	row := r.pool.QueryRow(ctx, getWalletByUserIDQuery, userID)
	return r.scanWallet(row)
}

// GetOrCreateByUserID gets or creates a wallet for a user
func (r *walletRepository) GetOrCreateByUserID(ctx context.Context, userID uuid.UUID) (*domain.UserWallet, error) {
	row := r.pool.QueryRow(ctx, getOrCreateWalletQuery, userID)
	return r.scanWallet(row)
}

// LockForUpdate locks the wallet row for update
func (r *walletRepository) LockForUpdate(ctx context.Context, userID uuid.UUID) (*domain.UserWallet, error) {
	row := r.pool.QueryRow(ctx, lockWalletForUpdateQuery, userID)
	return r.scanWallet(row)
}

// AddCoins adds coins to a user's wallet
func (r *walletRepository) AddCoins(ctx context.Context, userID uuid.UUID, amount decimal.Decimal) (*domain.UserWallet, error) {
	row := r.pool.QueryRow(ctx, addCoinsQuery, amount, amount, userID)
	return r.scanWallet(row)
}

// DeductCoins deducts coins from a user's wallet
func (r *walletRepository) DeductCoins(ctx context.Context, userID uuid.UUID, amount decimal.Decimal) (*domain.UserWallet, error) {
	row := r.pool.QueryRow(ctx, deductCoinsQuery, amount, amount, userID)
	return r.scanWallet(row)
}

// scanWallet scans a row into a UserWallet
func (r *walletRepository) scanWallet(row pgx.Row) (*domain.UserWallet, error) {
	var w domain.UserWallet
	err := row.Scan(
		&w.ID,
		&w.UserID,
		&w.CoinBalance,
		&w.TotalDeposited,
		&w.TotalSpent,
		&w.TotalSubscriptionSpent,
		&w.CreatedAt,
		&w.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &w, nil
}

//go:embed queries/package/get_by_id.sql
var getPackageByIDQuery string

//go:embed queries/package/get_by_slug.sql
var getPackageBySlugQuery string

//go:embed queries/package/list_active.sql
var listActivePackagesQuery string

// coinPackageRepository implements CoinPackageRepository
type coinPackageRepository struct {
	pool *pgxpool.Pool
}

// NewCoinPackageRepository creates a new coin package repository
func NewCoinPackageRepository(pool *pgxpool.Pool) domain.CoinPackageRepository {
	return &coinPackageRepository{pool: pool}
}

// GetByID retrieves a package by ID
func (r *coinPackageRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.CoinPackage, error) {
	row := r.pool.QueryRow(ctx, getPackageByIDQuery, id)
	return r.scanPackage(row)
}

// GetBySlug retrieves a package by slug
func (r *coinPackageRepository) GetBySlug(ctx context.Context, slug string) (*domain.CoinPackage, error) {
	row := r.pool.QueryRow(ctx, getPackageBySlugQuery, slug)
	return r.scanPackage(row)
}

// ListActive retrieves all active packages
func (r *coinPackageRepository) ListActive(ctx context.Context) ([]*domain.CoinPackage, error) {
	rows, err := r.pool.Query(ctx, listActivePackagesQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var packages []*domain.CoinPackage
	for rows.Next() {
		p, err := r.scanPackageFromRows(rows)
		if err != nil {
			return nil, err
		}
		packages = append(packages, p)
	}
	return packages, rows.Err()
}

// scanPackage scans a row into a CoinPackage
func (r *coinPackageRepository) scanPackage(row pgx.Row) (*domain.CoinPackage, error) {
	var p domain.CoinPackage
	err := row.Scan(
		&p.ID,
		&p.Name,
		&p.Slug,
		&p.CoinAmount,
		&p.PriceVND,
		&p.BonusPercent,
		&p.IsPopular,
		&p.IsActive,
		&p.DisplayOrder,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// scanPackageFromRows scans rows into a CoinPackage
func (r *coinPackageRepository) scanPackageFromRows(rows pgx.Rows) (*domain.CoinPackage, error) {
	var p domain.CoinPackage
	err := rows.Scan(
		&p.ID,
		&p.Name,
		&p.Slug,
		&p.CoinAmount,
		&p.PriceVND,
		&p.BonusPercent,
		&p.IsPopular,
		&p.IsActive,
		&p.DisplayOrder,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

//go:embed queries/topup/create.sql
var createTopupOrderQuery string

//go:embed queries/topup/get_by_id.sql
var getTopupByIDQuery string

//go:embed queries/topup/get_by_order_code.sql
var getTopupByOrderCodeQuery string

//go:embed queries/topup/get_by_sepay_tx_id.sql
var getTopupBySepayTxIDQuery string

//go:embed queries/topup/update_status.sql
var updateTopupStatusQuery string

//go:embed queries/topup/list_by_user_id.sql
var listTopupByUserIDQuery string

//go:embed queries/topup/expire_pending.sql
var expirePendingTopupsQuery string

// topupOrderRepository implements TopupOrderRepository
type topupOrderRepository struct {
	pool *pgxpool.Pool
}

// NewTopupOrderRepository creates a new topup order repository
func NewTopupOrderRepository(pool *pgxpool.Pool) domain.TopupOrderRepository {
	return &topupOrderRepository{pool: pool}
}

// Create creates a new topup order
func (r *topupOrderRepository) Create(ctx context.Context, order *domain.TopupOrder) error {
	return r.pool.QueryRow(ctx, createTopupOrderQuery,
		order.UserID,
		order.PackageID,
		order.OrderCode,
		order.CoinAmount,
		order.BaseCoinAmount,
		order.BonusCoinAmount,
		order.VNDAmount,
		order.BankName,
		order.BankAccount,
		order.AccountName,
		order.ExpiredAt,
	).Scan(&order.ID, &order.CreatedAt, &order.UpdatedAt)
}

// GetByID retrieves an order by ID
func (r *topupOrderRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.TopupOrder, error) {
	row := r.pool.QueryRow(ctx, getTopupByIDQuery, id)
	return r.scanTopupOrder(row)
}

// GetByOrderCode retrieves an order by order code
func (r *topupOrderRepository) GetByOrderCode(ctx context.Context, orderCode string) (*domain.TopupOrder, error) {
	row := r.pool.QueryRow(ctx, getTopupByOrderCodeQuery, orderCode)
	return r.scanTopupOrder(row)
}

// GetBySepayTransactionID retrieves an order by SePay transaction ID
func (r *topupOrderRepository) GetBySepayTransactionID(ctx context.Context, txID string) (*domain.TopupOrder, error) {
	row := r.pool.QueryRow(ctx, getTopupBySepayTxIDQuery, txID)
	return r.scanTopupOrder(row)
}

// UpdateStatus updates the order status
func (r *topupOrderRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.TopupStatus, sepayTxID *string, sepayContent *string) error {
	var completedAt *time.Time
	if status == domain.TopupStatusSuccess {
		now := time.Now()
		completedAt = &now
	}

	_, err := r.pool.Exec(ctx, updateTopupStatusQuery,
		status,
		sepayTxID,
		sepayContent,
		completedAt,
		id,
	)
	return err
}

// ListByUserID retrieves orders by user ID
func (r *topupOrderRepository) ListByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.TopupOrder, int, error) {
	rows, err := r.pool.Query(ctx, listTopupByUserIDQuery, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var orders []*domain.TopupOrder
	var total int
	for rows.Next() {
		o, t, err := r.scanTopupOrderWithCount(rows)
		if err != nil {
			return nil, 0, err
		}
		orders = append(orders, o)
		total = t
	}
	return orders, total, rows.Err()
}

// ExpirePendingOrders marks expired pending orders as expired
func (r *topupOrderRepository) ExpirePendingOrders(ctx context.Context) (int64, error) {
	result, err := r.pool.Exec(ctx, expirePendingTopupsQuery)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

// scanTopupOrder scans a row into a TopupOrder
func (r *topupOrderRepository) scanTopupOrder(row pgx.Row) (*domain.TopupOrder, error) {
	var o domain.TopupOrder
	err := row.Scan(
		&o.ID,
		&o.UserID,
		&o.PackageID,
		&o.OrderCode,
		&o.CoinAmount,
		&o.BaseCoinAmount,
		&o.BonusCoinAmount,
		&o.VNDAmount,
		&o.Status,
		&o.SepayTransactionID,
		&o.SepayContent,
		&o.BankName,
		&o.BankAccount,
		&o.AccountName,
		&o.CompletedAt,
		&o.ExpiredAt,
		&o.CreatedAt,
		&o.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &o, nil
}

// scanTopupOrderWithCount scans rows into a TopupOrder with total count
func (r *topupOrderRepository) scanTopupOrderWithCount(rows pgx.Rows) (*domain.TopupOrder, int, error) {
	var o domain.TopupOrder
	var total int
	err := rows.Scan(
		&o.ID,
		&o.UserID,
		&o.PackageID,
		&o.OrderCode,
		&o.CoinAmount,
		&o.BaseCoinAmount,
		&o.BonusCoinAmount,
		&o.VNDAmount,
		&o.Status,
		&o.SepayTransactionID,
		&o.SepayContent,
		&o.BankName,
		&o.BankAccount,
		&o.AccountName,
		&o.CompletedAt,
		&o.ExpiredAt,
		&o.CreatedAt,
		&o.UpdatedAt,
		&total,
	)
	if err != nil {
		return nil, 0, err
	}
	return &o, total, nil
}

//go:embed queries/transaction/create.sql
var createTransactionQuery string

//go:embed queries/transaction/list_by_user_id.sql
var listTransactionsByUserIDQuery string

//go:embed queries/transaction/get_by_reference.sql
var getTransactionByReferenceQuery string

// transactionRepository implements TransactionRepository
type transactionRepository struct {
	pool *pgxpool.Pool
}

// NewTransactionRepository creates a new transaction repository
func NewTransactionRepository(pool *pgxpool.Pool) domain.TransactionRepository {
	return &transactionRepository{pool: pool}
}

// Create creates a new transaction
func (r *transactionRepository) Create(ctx context.Context, tx *domain.Transaction) error {
	return r.pool.QueryRow(ctx, createTransactionQuery,
		tx.UserID,
		tx.Type,
		tx.CoinAmount,
		tx.VNDAmount,
		tx.BalanceAfter,
		tx.ReferenceType,
		tx.ReferenceID,
		tx.CreatorID,
		tx.CreatorRevenueVND,
		tx.PlatformRevenueVND,
		tx.Description,
		tx.Metadata,
	).Scan(&tx.ID, &tx.CreatedAt)
}

// ListByUserID retrieves transactions by user ID
func (r *transactionRepository) ListByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.Transaction, int, error) {
	rows, err := r.pool.Query(ctx, listTransactionsByUserIDQuery, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var txs []*domain.Transaction
	var total int
	for rows.Next() {
		t, tt, err := r.scanTransactionWithCount(rows)
		if err != nil {
			return nil, 0, err
		}
		txs = append(txs, t)
		total = tt
	}
	return txs, total, rows.Err()
}

// GetByReferenceID retrieves a transaction by reference ID
func (r *transactionRepository) GetByReferenceID(ctx context.Context, refType string, refID uuid.UUID) (*domain.Transaction, error) {
	row := r.pool.QueryRow(ctx, getTransactionByReferenceQuery, refType, refID)
	return r.scanTransaction(row)
}

// scanTransaction scans a row into a Transaction
func (r *transactionRepository) scanTransaction(row pgx.Row) (*domain.Transaction, error) {
	var t domain.Transaction
	err := row.Scan(
		&t.ID,
		&t.UserID,
		&t.Type,
		&t.CoinAmount,
		&t.VNDAmount,
		&t.BalanceAfter,
		&t.ReferenceType,
		&t.ReferenceID,
		&t.CreatorID,
		&t.CreatorRevenueVND,
		&t.PlatformRevenueVND,
		&t.Description,
		&t.Metadata,
		&t.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// scanTransactionWithCount scans rows into a Transaction with total count
func (r *transactionRepository) scanTransactionWithCount(rows pgx.Rows) (*domain.Transaction, int, error) {
	var t domain.Transaction
	var total int
	err := rows.Scan(
		&t.ID,
		&t.UserID,
		&t.Type,
		&t.CoinAmount,
		&t.VNDAmount,
		&t.BalanceAfter,
		&t.ReferenceType,
		&t.ReferenceID,
		&t.CreatorID,
		&t.CreatorRevenueVND,
		&t.PlatformRevenueVND,
		&t.Description,
		&t.Metadata,
		&t.CreatedAt,
		&total,
	)
	if err != nil {
		return nil, 0, err
	}
	return &t, total, nil
}
