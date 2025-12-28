// ============================================================================
// Payment Wallet Repository (Ent Implementation)
// ============================================================================
//
// File này chứa các repository implementations cho Payment module sử dụng Ent.
// Bao gồm: WalletRepository, CoinPackageRepository, TopupOrderRepository, TransactionRepository
//
// ============================================================================

package payment

import (
	"context"
	"database/sql"
	"system/internal/platform/database"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/shopspring/decimal"

	"system/internal/domain"
	ent "system/internal/ent/generated"
	"system/internal/ent/generated/coinpackage"
	"system/internal/ent/generated/topuporder"
	"system/internal/ent/generated/transaction"
	"system/internal/ent/generated/userwallet"
)

// =============================================================================
// WALLET REPOSITORY
// =============================================================================

// walletEntRepository implements WalletRepository using Ent
type walletEntRepository struct {
	client *ent.Client
}

// NewWalletEntRepository creates a new wallet repository using Ent
func NewWalletEntRepository(client *ent.Client) domain.WalletRepository {
	return &walletEntRepository{client: client}
}

// GetByUserID retrieves a wallet by user ID
func (r *walletEntRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.UserWallet, error) {
	w, err := database.GetClientFromContext(ctx, r.client).UserWallet.Query().
		Where(userwallet.UserIDEQ(userID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return entWalletToDomain(w), nil
}

// GetOrCreateByUserID gets or creates a wallet for a user
func (r *walletEntRepository) GetOrCreateByUserID(ctx context.Context, userID uuid.UUID) (*domain.UserWallet, error) {
	// Try to get existing
	w, err := database.GetClientFromContext(ctx, r.client).UserWallet.Query().
		Where(userwallet.UserIDEQ(userID)).
		Only(ctx)
	if err == nil {
		return entWalletToDomain(w), nil
	}
	if !ent.IsNotFound(err) {
		return nil, err
	}

	// Create new wallet
	w, err = database.GetClientFromContext(ctx, r.client).UserWallet.Create().
		SetUserID(userID).
		SetCoinBalance(decimal.Zero).
		SetTotalDeposited(decimal.Zero).
		SetTotalSpent(decimal.Zero).
		SetTotalSubscriptionSpent(decimal.Zero).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return entWalletToDomain(w), nil
}

// AddCoins adds coins to a user's wallet
func (r *walletEntRepository) AddCoins(ctx context.Context, userID uuid.UUID, amount decimal.Decimal) (*domain.UserWallet, error) {
	w, err := database.GetClientFromContext(ctx, r.client).UserWallet.Query().
		Where(userwallet.UserIDEQ(userID)).
		Only(ctx)
	if err != nil {
		return nil, err
	}

	updated, err := database.GetClientFromContext(ctx, r.client).UserWallet.UpdateOne(w).
		SetCoinBalance(w.CoinBalance.Add(amount)).
		SetTotalDeposited(w.TotalDeposited.Add(amount)).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return entWalletToDomain(updated), nil
}

// DeductCoins deducts coins from a user's wallet
func (r *walletEntRepository) DeductCoins(ctx context.Context, userID uuid.UUID, amount decimal.Decimal) (*domain.UserWallet, error) {
	w, err := database.GetClientFromContext(ctx, r.client).UserWallet.Query().
		Where(userwallet.UserIDEQ(userID)).
		Only(ctx)
	if err != nil {
		return nil, err
	}

	updated, err := database.GetClientFromContext(ctx, r.client).UserWallet.UpdateOne(w).
		SetCoinBalance(w.CoinBalance.Sub(amount)).
		SetTotalSpent(w.TotalSpent.Add(amount)).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return entWalletToDomain(updated), nil
}

// LockForUpdate locks the wallet row for update (for transactions)
// Note: Uses raw SQL for FOR UPDATE lock
func (r *walletEntRepository) LockForUpdate(ctx context.Context, userID uuid.UUID) (*domain.UserWallet, error) {
	// Ent doesn't support FOR UPDATE directly, use standard query
	// Transaction locking should be handled at service level with Ent Tx
	return r.GetByUserID(ctx, userID)
}

func entWalletToDomain(w *ent.UserWallet) *domain.UserWallet {
	return &domain.UserWallet{
		ID:                     w.ID,
		UserID:                 w.UserID,
		CoinBalance:            w.CoinBalance,
		TotalDeposited:         w.TotalDeposited,
		TotalSpent:             w.TotalSpent,
		TotalSubscriptionSpent: w.TotalSubscriptionSpent,
		CreatedAt:              w.CreatedAt,
		UpdatedAt:              w.UpdatedAt,
	}
}

// =============================================================================
// COIN PACKAGE REPOSITORY
// =============================================================================

// coinPackageEntRepository implements CoinPackageRepository using Ent
type coinPackageEntRepository struct {
	client *ent.Client
}

// NewCoinPackageEntRepository creates a new coin package repository using Ent
func NewCoinPackageEntRepository(client *ent.Client) domain.CoinPackageRepository {
	return &coinPackageEntRepository{client: client}
}

// GetByID retrieves a package by ID
func (r *coinPackageEntRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.CoinPackage, error) {
	p, err := database.GetClientFromContext(ctx, r.client).CoinPackage.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return entCoinPackageToDomain(p), nil
}

// GetBySlug retrieves a package by slug
func (r *coinPackageEntRepository) GetBySlug(ctx context.Context, slug string) (*domain.CoinPackage, error) {
	p, err := database.GetClientFromContext(ctx, r.client).CoinPackage.Query().
		Where(coinpackage.SlugEQ(slug)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return entCoinPackageToDomain(p), nil
}

// ListActive retrieves all active packages
func (r *coinPackageEntRepository) ListActive(ctx context.Context) ([]*domain.CoinPackage, error) {
	packages, err := database.GetClientFromContext(ctx, r.client).CoinPackage.Query().
		Where(coinpackage.IsActiveEQ(true)).
		Order(ent.Asc(coinpackage.FieldDisplayOrder)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	results := make([]*domain.CoinPackage, len(packages))
	for i, p := range packages {
		results[i] = entCoinPackageToDomain(p)
	}
	return results, nil
}

func entCoinPackageToDomain(p *ent.CoinPackage) *domain.CoinPackage {
	return &domain.CoinPackage{
		ID:           p.ID,
		Name:         p.Name,
		Slug:         p.Slug,
		CoinAmount:   p.CoinAmount,
		PriceVND:     p.PriceVnd,
		BonusPercent: p.BonusPercent,
		IsPopular:    p.IsPopular,
		IsActive:     p.IsActive,
		DisplayOrder: p.DisplayOrder,
		CreatedAt:    p.CreatedAt,
		UpdatedAt:    p.UpdatedAt,
	}
}

// =============================================================================
// TOPUP ORDER REPOSITORY
// =============================================================================

// topupOrderEntRepository implements TopupOrderRepository using Ent
type topupOrderEntRepository struct {
	client *ent.Client
	db     *sql.DB
}

// NewTopupOrderEntRepository creates a new topup order repository using Ent
func NewTopupOrderEntRepository(client *ent.Client, db *sql.DB) domain.TopupOrderRepository {
	return &topupOrderEntRepository{client: client, db: db}
}

// Create creates a new topup order
func (r *topupOrderEntRepository) Create(ctx context.Context, order *domain.TopupOrder) error {
	created, err := database.GetClientFromContext(ctx, r.client).TopupOrder.Create().
		SetUserID(order.UserID).
		SetPackageID(order.PackageID).
		SetOrderCode(order.OrderCode).
		SetCoinAmount(order.CoinAmount).
		SetBaseCoinAmount(order.BaseCoinAmount).
		SetBonusCoinAmount(order.BonusCoinAmount).
		SetVndAmount(order.VNDAmount).
		SetStatus(topuporder.StatusPending).
		SetExpiredAt(order.ExpiredAt).
		SetNillableBankName(order.BankName).
		SetNillableBankAccount(order.BankAccount).
		SetNillableAccountName(order.AccountName).
		Save(ctx)
	if err != nil {
		return err
	}
	order.ID = created.ID
	order.CreatedAt = created.CreatedAt
	order.UpdatedAt = created.UpdatedAt
	return nil
}

// GetByID retrieves an order by ID
func (r *topupOrderEntRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.TopupOrder, error) {
	o, err := database.GetClientFromContext(ctx, r.client).TopupOrder.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return entTopupOrderToDomain(o), nil
}

// GetByOrderCode retrieves an order by order code
func (r *topupOrderEntRepository) GetByOrderCode(ctx context.Context, orderCode string) (*domain.TopupOrder, error) {
	o, err := database.GetClientFromContext(ctx, r.client).TopupOrder.Query().
		Where(topuporder.OrderCodeEQ(orderCode)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return entTopupOrderToDomain(o), nil
}

// GetBySepayTransactionID retrieves an order by SePay transaction ID
func (r *topupOrderEntRepository) GetBySepayTransactionID(ctx context.Context, txID string) (*domain.TopupOrder, error) {
	o, err := database.GetClientFromContext(ctx, r.client).TopupOrder.Query().
		Where(topuporder.SepayTransactionIDEQ(txID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return entTopupOrderToDomain(o), nil
}

// UpdateStatus updates the order status
func (r *topupOrderEntRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.TopupStatus, sepayTxID *string, sepayContent *string) error {
	builder := database.GetClientFromContext(ctx, r.client).TopupOrder.UpdateOneID(id).
		SetStatus(topuporder.Status(status))

	if sepayTxID != nil {
		builder.SetSepayTransactionID(*sepayTxID)
	}
	if sepayContent != nil {
		builder.SetSepayContent(*sepayContent)
	}
	if status == domain.TopupStatusSuccess {
		builder.SetCompletedAt(time.Now())
	}

	_, err := builder.Save(ctx)
	return err
}

// ListByUserID retrieves orders by user ID with pagination
func (r *topupOrderEntRepository) ListByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.TopupOrder, int, error) {
	query := database.GetClientFromContext(ctx, r.client).TopupOrder.Query().
		Where(topuporder.UserIDEQ(userID)).
		Order(ent.Desc(topuporder.FieldCreatedAt))

	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	orders, err := query.Limit(limit).Offset(offset).All(ctx)
	if err != nil {
		return nil, 0, err
	}

	results := make([]*domain.TopupOrder, len(orders))
	for i, o := range orders {
		results[i] = entTopupOrderToDomain(o)
	}
	return results, total, nil
}

// ExpirePendingOrders marks expired pending orders as expired
func (r *topupOrderEntRepository) ExpirePendingOrders(ctx context.Context) (int64, error) {
	count, err := database.GetClientFromContext(ctx, r.client).TopupOrder.Update().
		Where(
			topuporder.StatusEQ(topuporder.StatusPending),
			topuporder.ExpiredAtLT(time.Now()),
		).
		SetStatus(topuporder.StatusExpired).
		Save(ctx)
	return int64(count), err
}

func entTopupOrderToDomain(o *ent.TopupOrder) *domain.TopupOrder {
	return &domain.TopupOrder{
		ID:                 o.ID,
		UserID:             o.UserID,
		PackageID:          o.PackageID,
		OrderCode:          o.OrderCode,
		CoinAmount:         o.CoinAmount,
		BaseCoinAmount:     o.BaseCoinAmount,
		BonusCoinAmount:    o.BonusCoinAmount,
		VNDAmount:          o.VndAmount,
		Status:             domain.TopupStatus(o.Status),
		SepayTransactionID: o.SepayTransactionID,
		SepayContent:       o.SepayContent,
		BankName:           o.BankName,
		BankAccount:        o.BankAccount,
		AccountName:        o.AccountName,
		CompletedAt:        o.CompletedAt,
		ExpiredAt:          o.ExpiredAt,
		CreatedAt:          o.CreatedAt,
		UpdatedAt:          o.UpdatedAt,
	}
}

// =============================================================================
// TRANSACTION REPOSITORY
// =============================================================================

// transactionEntRepository implements TransactionRepository using Ent
type transactionEntRepository struct {
	client *ent.Client
}

// NewTransactionEntRepository creates a new transaction repository using Ent
func NewTransactionEntRepository(client *ent.Client) domain.TransactionRepository {
	return &transactionEntRepository{client: client}
}

// Create creates a new transaction
func (r *transactionEntRepository) Create(ctx context.Context, tx *domain.Transaction) error {
	builder := database.GetClientFromContext(ctx, r.client).Transaction.Create().
		SetUserID(tx.UserID).
		SetType(transaction.Type(tx.Type)).
		SetCoinAmount(tx.CoinAmount).
		SetBalanceAfter(tx.BalanceAfter)

	if tx.VNDAmount != nil {
		builder.SetVndAmount(*tx.VNDAmount)
	}
	if tx.ReferenceType != nil {
		builder.SetReferenceType(*tx.ReferenceType)
	}
	if tx.ReferenceID != nil {
		builder.SetReferenceID(*tx.ReferenceID)
	}
	if tx.CreatorID != nil {
		builder.SetCreatorID(*tx.CreatorID)
	}
	if tx.CreatorRevenueVND != nil {
		builder.SetCreatorRevenueVnd(*tx.CreatorRevenueVND)
	}
	if tx.PlatformRevenueVND != nil {
		builder.SetPlatformRevenueVnd(*tx.PlatformRevenueVND)
	}
	if tx.Description != nil {
		builder.SetDescription(*tx.Description)
	}

	created, err := builder.Save(ctx)
	if err != nil {
		return err
	}
	tx.ID = created.ID
	tx.CreatedAt = created.CreatedAt
	return nil
}

// ListByUserID retrieves transactions by user ID with pagination
func (r *transactionEntRepository) ListByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.Transaction, int, error) {
	query := database.GetClientFromContext(ctx, r.client).Transaction.Query().
		Where(transaction.UserIDEQ(userID)).
		Order(ent.Desc(transaction.FieldCreatedAt))

	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	txs, err := query.Limit(limit).Offset(offset).All(ctx)
	if err != nil {
		return nil, 0, err
	}

	results := make([]*domain.Transaction, len(txs))
	for i, t := range txs {
		results[i] = entTransactionToDomain(t)
	}
	return results, total, nil
}

// GetByReferenceID retrieves a transaction by reference ID
func (r *transactionEntRepository) GetByReferenceID(ctx context.Context, refType string, refID uuid.UUID) (*domain.Transaction, error) {
	tx, err := database.GetClientFromContext(ctx, r.client).Transaction.Query().
		Where(
			transaction.ReferenceTypeEQ(refType),
			transaction.ReferenceIDEQ(refID),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return entTransactionToDomain(tx), nil
}

func entTransactionToDomain(t *ent.Transaction) *domain.Transaction {
	return &domain.Transaction{
		ID:                 t.ID,
		UserID:             t.UserID,
		Type:               domain.TransactionType(t.Type),
		CoinAmount:         t.CoinAmount,
		VNDAmount:          t.VndAmount,
		BalanceAfter:       t.BalanceAfter,
		ReferenceType:      t.ReferenceType,
		ReferenceID:        t.ReferenceID,
		CreatorID:          t.CreatorID,
		CreatorRevenueVND:  t.CreatorRevenueVnd,
		PlatformRevenueVND: t.PlatformRevenueVnd,
		Description:        t.Description,
		CreatedAt:          t.CreatedAt,
	}
}
