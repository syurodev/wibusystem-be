package payment

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofrs/uuid/v5"
	"go.uber.org/zap"

	"system/internal/domain"
	paymentdto "system/internal/dto/payment"
	ent "system/internal/ent/generated"
	pkgerrors "system/pkg/errors"
)

type topupUseCase struct {
	entClient     *ent.Client
	walletRepo    domain.WalletRepository
	packageRepo   domain.CoinPackageRepository
	topupRepo     domain.TopupOrderRepository
	txRepo        domain.TransactionRepository
	configUseCase ConfigUseCase
	notifier      TopupNotifier
	logger        *zap.Logger
}

func NewTopupUseCase(
	entClient *ent.Client,
	walletRepo domain.WalletRepository,
	packageRepo domain.CoinPackageRepository,
	topupRepo domain.TopupOrderRepository,
	txRepo domain.TransactionRepository,
	configUseCase ConfigUseCase,
	notifier TopupNotifier,
	logger *zap.Logger,
) TopupUseCase {
	return &topupUseCase{
		entClient:     entClient,
		walletRepo:    walletRepo,
		packageRepo:   packageRepo,
		topupRepo:     topupRepo,
		txRepo:        txRepo,
		configUseCase: configUseCase,
		notifier:      notifier,
		logger:        logger,
	}
}

func (uc *topupUseCase) ListPackages(ctx context.Context) ([]*domain.CoinPackage, error) {
	return uc.packageRepo.ListActive(ctx)
}

func (uc *topupUseCase) GetPackage(ctx context.Context, packageID uuid.UUID) (*domain.CoinPackage, error) {
	pkg, err := uc.packageRepo.GetByID(ctx, packageID)
	if err != nil {
		if isNotFoundError(err) {
			return nil, pkgerrors.NotFound(I18nPackageNotFound, "coin package not found")
		}
		return nil, err
	}
	if !pkg.IsActive {
		return nil, pkgerrors.BadRequest(I18nPackageInactive, "coin package is inactive")
	}
	return pkg, nil
}

func (uc *topupUseCase) CreateTopupOrder(ctx context.Context, userID uuid.UUID, packageID uuid.UUID) (*domain.TopupOrder, error) {
	// Get package
	pkg, err := uc.GetPackage(ctx, packageID)
	if err != nil {
		return nil, err
	}

	// Ensure wallet exists
	_, err = uc.walletRepo.GetOrCreateByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Get config values
	expiryMinutes, err := uc.configUseCase.GetNumber(ctx, "topup.expiry_minutes")
	if err != nil {
		expiryMinutes = 15 // default
	}

	orderPrefix, err := uc.configUseCase.GetString(ctx, "topup.order_prefix")
	if err != nil {
		orderPrefix = "NAP" // default
	}

	bankName, _ := uc.configUseCase.GetString(ctx, "sepay.bank_name")
	bankAccount, _ := uc.configUseCase.GetString(ctx, "sepay.bank_account")
	accountName, _ := uc.configUseCase.GetString(ctx, "sepay.account_name")

	// Generate order code
	orderCode := generateOrderCode(orderPrefix)

	// Calculate coins
	baseCoins := pkg.CoinAmount
	bonusCoins := pkg.BonusCoins()
	totalCoins := pkg.TotalCoins()

	// Create order
	order := &domain.TopupOrder{
		UserID:          userID,
		PackageID:       packageID,
		OrderCode:       orderCode,
		CoinAmount:      totalCoins,
		BaseCoinAmount:  baseCoins,
		BonusCoinAmount: bonusCoins,
		VNDAmount:       pkg.PriceVND,
		Status:          domain.TopupStatusPending,
		BankName:        &bankName,
		BankAccount:     &bankAccount,
		AccountName:     &accountName,
		ExpiredAt:       time.Now().Add(time.Duration(expiryMinutes) * time.Minute),
	}

	if err := uc.topupRepo.Create(ctx, order); err != nil {
		return nil, err
	}

	order.Package = pkg
	return order, nil
}

func (uc *topupUseCase) GetTopupOrder(ctx context.Context, orderID uuid.UUID) (*domain.TopupOrder, error) {
	order, err := uc.topupRepo.GetByID(ctx, orderID)
	if err != nil {
		if isNotFoundError(err) {
			return nil, pkgerrors.NotFound(I18nTopupNotFound, "top-up order not found")
		}
		return nil, err
	}
	return order, nil
}

func (uc *topupUseCase) GetTopupOrderByCode(ctx context.Context, orderCode string) (*domain.TopupOrder, error) {
	order, err := uc.topupRepo.GetByOrderCode(ctx, orderCode)
	if err != nil {
		if isNotFoundError(err) {
			return nil, pkgerrors.NotFound(I18nTopupNotFound, "top-up order not found")
		}
		return nil, err
	}
	return order, nil
}

func (uc *topupUseCase) ListTopupOrders(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.TopupOrder, int, error) {
	return uc.topupRepo.ListByUserID(ctx, userID, limit, offset)
}

func (uc *topupUseCase) CancelTopupOrder(ctx context.Context, userID uuid.UUID, orderID uuid.UUID) error {
	order, err := uc.topupRepo.GetByID(ctx, orderID)
	if err != nil {
		if isNotFoundError(err) {
			return pkgerrors.NotFound(I18nTopupNotFound, "top-up order not found")
		}
		return err
	}

	// Check ownership
	if order.UserID != userID {
		return pkgerrors.Forbidden(I18nAuthUnauthorized, "not authorized")
	}

	// Check status
	if order.Status != domain.TopupStatusPending {
		return pkgerrors.BadRequest(I18nTopupAlreadyCompleted, "order already completed or cancelled")
	}

	return uc.topupRepo.UpdateStatus(ctx, orderID, domain.TopupStatusCancelled, nil, nil)
}

func (uc *topupUseCase) ProcessWebhook(ctx context.Context, transactionID string, amount int64, content string) error {
	// Check idempotency
	existing, _ := uc.topupRepo.GetBySepayTransactionID(ctx, transactionID)
	if existing != nil {
		uc.logger.Info("webhook already processed", zap.String("transaction_id", transactionID))
		return nil // Idempotent - already processed
	}

	// Extract order code from content
	orderCode := extractOrderCode(content)
	if orderCode == "" {
		return pkgerrors.BadRequest(I18nTopupInvalidWebhook, "cannot extract order code from content")
	}

	// Get order
	order, err := uc.topupRepo.GetByOrderCode(ctx, orderCode)
	if err != nil {
		return pkgerrors.NotFound(I18nTopupNotFound, "order not found")
	}

	// Check status
	if order.Status != domain.TopupStatusPending {
		uc.logger.Warn("order not pending",
			zap.String("order_code", orderCode),
			zap.String("status", string(order.Status)),
		)
		return nil // Ignore
	}

	// Check amount
	expectedAmount := order.VNDAmount.IntPart()
	if amount < expectedAmount {
		uc.logger.Warn("amount mismatch",
			zap.String("order_code", orderCode),
			zap.Int64("expected", expectedAmount),
			zap.Int64("received", amount),
		)
		return pkgerrors.BadRequest(I18nTopupAmountMismatch, "amount mismatch")
	}

	// Process using Ent transaction
	tx, err := uc.entClient.Tx(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Update order status
	if err = uc.topupRepo.UpdateStatus(ctx, order.ID, domain.TopupStatusSuccess, &transactionID, &content); err != nil {
		return err
	}

	// Add coins to wallet
	wallet, err := uc.walletRepo.AddCoins(ctx, order.UserID, order.CoinAmount)
	if err != nil {
		return err
	}

	// Create transaction record
	refType := "topup_order"
	txRecord := &domain.Transaction{
		UserID:        order.UserID,
		Type:          domain.TransactionTypeTopup,
		CoinAmount:    order.CoinAmount,
		VNDAmount:     &order.VNDAmount,
		BalanceAfter:  wallet.CoinBalance,
		ReferenceType: &refType,
		ReferenceID:   &order.ID,
		Description:   stringPtr(fmt.Sprintf("Nạp %s Coin", order.CoinAmount.StringFixed(2))),
	}

	if err = uc.txRepo.Create(ctx, txRecord); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	uc.logger.Info("topup completed",
		zap.String("order_code", orderCode),
		zap.String("user_id", order.UserID.String()),
		zap.String("coins", order.CoinAmount.StringFixed(2)),
	)

	// Send notification using WebSocket
	if uc.notifier != nil {
		notification := paymentdto.TopupNotification{
			Type:       string(domain.NotificationTypeTopupSuccess),
			OrderID:    order.ID.String(),
			OrderCode:  order.OrderCode,
			CoinAmount: order.CoinAmount.StringFixed(2),
			NewBalance: wallet.CoinBalance.StringFixed(2),
			Message:    "Nạp coin thành công",
			MessageKey: I18nTopupSuccessNotification,
			MessageParams: map[string]string{
				"amount": order.CoinAmount.StringFixed(2),
			},
		}

		msgBytes, _ := json.Marshal(notification)
		uc.notifier.SendToUser(order.UserID, msgBytes)
	}

	return nil
}

func (uc *topupUseCase) ExpirePendingOrders(ctx context.Context) (int64, error) {
	return uc.topupRepo.ExpirePendingOrders(ctx)
}

// Helpers

func generateOrderCode(prefix string) string {
	bytes := make([]byte, 4)
	rand.Read(bytes)
	return fmt.Sprintf("%s%s", prefix, hex.EncodeToString(bytes))
}

func extractOrderCode(content string) string {
	prefixes := []string{"NAP"}

	for _, prefix := range prefixes {
		idx := findPrefixIndex(content, prefix)
		if idx >= 0 && idx+len(prefix)+8 <= len(content) {
			return content[idx : idx+len(prefix)+8]
		}
	}
	return ""
}

func findPrefixIndex(s, prefix string) int {
	for i := 0; i <= len(s)-len(prefix); i++ {
		if s[i:i+len(prefix)] == prefix {
			return i
		}
	}
	return -1
}

func stringPtr(s string) *string {
	return &s
}

// isNotFoundError checks if the error is a "not found" type error
func isNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	// Check for Ent's IsNotFound
	if ent.IsNotFound(err) {
		return true
	}
	// Check for common error messages
	errStr := err.Error()
	return errStr == "no rows in result set" || errStr == "sql: no rows in result set"
}
