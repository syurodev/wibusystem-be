package payment

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid/v5"
	"go.uber.org/zap"

	"system/internal/app/middleware"
	"system/internal/domain"
	paymentdto "system/internal/dto/payment"
	pkgerrors "system/pkg/errors"
	"system/pkg/util/response"
)

// WalletHandler handles wallet HTTP requests
type WalletHandler struct {
	walletUseCase      WalletUseCase
	topupUseCase       TopupUseCase
	transactionUseCase TransactionUseCase
	logger             *zap.Logger
}

// NewWalletHandler creates a new wallet handler
func NewWalletHandler(
	walletUseCase WalletUseCase,
	topupUseCase TopupUseCase,
	transactionUseCase TransactionUseCase,
	logger *zap.Logger,
) *WalletHandler {
	return &WalletHandler{
		walletUseCase:      walletUseCase,
		topupUseCase:       topupUseCase,
		transactionUseCase: transactionUseCase,
		logger:             logger,
	}
}

// GetWallet returns current user's wallet information
// @Summary Get wallet information
// @Tags Wallet
// @Security BearerAuth
// @Produce json
// @Success 200 {object} paymentdto.WalletResponse
// @Router /api/v1/wallet [get]
func (h *WalletHandler) GetWallet(c *gin.Context) {
	ctx := c.Request.Context()
	userID, err := h.getUserID(c)
	if err != nil {
		return
	}

	wallet, err := h.walletUseCase.GetWallet(ctx, userID)
	if err != nil {
		h.logger.Error("failed to get wallet", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", I18nWalletGetFailed, nil)
		return
	}

	resp := paymentdto.WalletResponse{
		CoinBalance:            paymentdto.DecimalToString(wallet.CoinBalance),
		TotalDeposited:         paymentdto.DecimalToString(wallet.TotalDeposited),
		TotalSpent:             paymentdto.DecimalToString(wallet.TotalSpent),
		TotalSubscriptionSpent: paymentdto.DecimalToString(wallet.TotalSubscriptionSpent),
		UpdatedAt:              wallet.UpdatedAt.Format(time.RFC3339),
	}

	response.Success(c, http.StatusOK, I18nWalletGetSuccess, resp, nil)
}

// ListPackages returns all active coin packages
// @Summary List coin packages
// @Tags Wallet
// @Produce json
// @Success 200 {object} paymentdto.CoinPackageListResponse
// @Router /api/v1/payment/packages [get]
func (h *WalletHandler) ListPackages(c *gin.Context) {
	ctx := c.Request.Context()

	packages, err := h.topupUseCase.ListPackages(ctx)
	if err != nil {
		h.logger.Error("failed to list packages", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", I18nPackageListFailed, nil)
		return
	}

	resp := paymentdto.CoinPackageListResponse{
		Packages: make([]paymentdto.CoinPackageResponse, len(packages)),
	}

	for i, pkg := range packages {
		resp.Packages[i] = paymentdto.CoinPackageResponse{
			ID:           pkg.ID.String(),
			Name:         pkg.Name,
			Slug:         pkg.Slug,
			CoinAmount:   paymentdto.DecimalToString(pkg.CoinAmount),
			BonusPercent: pkg.BonusPercent,
			TotalCoins:   paymentdto.DecimalToString(pkg.TotalCoins()),
			PriceVND:     paymentdto.DecimalToString(pkg.PriceVND),
			IsPopular:    pkg.IsPopular,
		}
	}

	response.Success(c, http.StatusOK, I18nPackageListSuccess, resp, nil)
}

// CreateTopup creates a new top-up order
// @Summary Create top-up order
// @Tags Wallet
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body paymentdto.CreateTopupRequest true "Package ID"
// @Success 201 {object} paymentdto.CreateTopupResponse
// @Router /api/v1/payment/topup [post]
func (h *WalletHandler) CreateTopup(c *gin.Context) {
	ctx := c.Request.Context()
	userID, err := h.getUserID(c)
	if err != nil {
		return
	}

	var req paymentdto.CreateTopupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", I18nValidationFailed, err.Error())
		return
	}

	order, err := h.topupUseCase.CreateTopupOrder(ctx, userID, req.PackageID)
	if err != nil {
		if appErr, ok := pkgerrors.AsAppError(err); ok {
			response.Error(c, appErr.StatusCode, appErr.ErrCode, appErr.I18nKey, nil)
			return
		}
		h.logger.Error("failed to create topup", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", I18nTopupCreateFailed, nil)
		return
	}

	// Calculate expires in seconds
	expiresIn := int(time.Until(order.ExpiredAt).Seconds())
	if expiresIn < 0 {
		expiresIn = 0
	}

	orderResp := toTopupOrderResponse(order)
	transferInfo := paymentdto.TransferInfoResponse{
		BankName:      safeDeref(order.BankName),
		BankAccount:   safeDeref(order.BankAccount),
		AccountName:   safeDeref(order.AccountName),
		Amount:        paymentdto.DecimalToString(order.VNDAmount),
		Content:       order.OrderCode,
		QRCodeURL:     generateVietQRURL(safeDeref(order.BankName), safeDeref(order.BankAccount), safeDeref(order.AccountName), paymentdto.DecimalToString(order.VNDAmount), order.OrderCode),
		ExpiresInSecs: expiresIn,
	}

	resp := paymentdto.CreateTopupResponse{
		Order:        orderResp,
		TransferInfo: transferInfo,
	}

	response.Success(c, http.StatusCreated, I18nTopupCreateSuccess, resp, nil)
}

// GetTopup returns a top-up order by ID
// @Summary Get top-up order
// @Tags Wallet
// @Security BearerAuth
// @Produce json
// @Param id path string true "Order ID"
// @Success 200 {object} paymentdto.TopupOrderResponse
// @Router /api/v1/payment/topup/{id} [get]
func (h *WalletHandler) GetTopup(c *gin.Context) {
	ctx := c.Request.Context()
	userID, err := h.getUserID(c)
	if err != nil {
		return
	}

	orderID, err := uuid.FromString(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", I18nValidationFailed, nil)
		return
	}

	order, err := h.topupUseCase.GetTopupOrder(ctx, orderID)
	if err != nil {
		if appErr, ok := pkgerrors.AsAppError(err); ok {
			response.Error(c, appErr.StatusCode, appErr.ErrCode, appErr.I18nKey, nil)
			return
		}
		h.logger.Error("failed to get topup", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", I18nTopupNotFound, nil)
		return
	}

	// Check ownership
	if order.UserID != userID {
		response.Error(c, http.StatusNotFound, "NOT_FOUND", I18nTopupNotFound, nil)
		return
	}

	response.Success(c, http.StatusOK, I18nTopupGetSuccess, toTopupOrderResponse(order), nil)
}

// ListTopups returns top-up orders for current user
// @Summary List top-up orders
// @Tags Wallet
// @Security BearerAuth
// @Produce json
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Success 200 {object} paymentdto.TopupOrderListResponse
// @Router /api/v1/payment/topup [get]
func (h *WalletHandler) ListTopups(c *gin.Context) {
	ctx := c.Request.Context()
	userID, err := h.getUserID(c)
	if err != nil {
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	orders, total, err := h.topupUseCase.ListTopupOrders(ctx, userID, limit, offset)
	if err != nil {
		h.logger.Error("failed to list topups", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", I18nTopupNotFound, nil)
		return
	}

	resp := paymentdto.TopupOrderListResponse{
		Orders:     make([]paymentdto.TopupOrderResponse, len(orders)),
		TotalItems: total,
	}

	for i, order := range orders {
		resp.Orders[i] = toTopupOrderResponse(order)
	}

	response.Success(c, http.StatusOK, I18nTopupListSuccess, resp, nil)
}

// CancelTopup cancels a pending top-up order
// @Summary Cancel top-up order
// @Tags Wallet
// @Security BearerAuth
// @Param id path string true "Order ID"
// @Success 200
// @Router /api/v1/payment/topup/{id}/cancel [post]
func (h *WalletHandler) CancelTopup(c *gin.Context) {
	ctx := c.Request.Context()
	userID, err := h.getUserID(c)
	if err != nil {
		return
	}

	orderID, err := uuid.FromString(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", I18nValidationFailed, nil)
		return
	}

	if err := h.topupUseCase.CancelTopupOrder(ctx, userID, orderID); err != nil {
		if appErr, ok := pkgerrors.AsAppError(err); ok {
			response.Error(c, appErr.StatusCode, appErr.ErrCode, appErr.I18nKey, nil)
			return
		}
		h.logger.Error("failed to cancel topup", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", I18nTopupCreateFailed, nil)
		return
	}

	response.Success(c, http.StatusOK, I18nTopupCancelSuccess, nil, nil)
}

// ListTransactions returns transaction history for current user
// @Summary List transactions
// @Tags Wallet
// @Security BearerAuth
// @Produce json
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Success 200 {object} paymentdto.TransactionListResponse
// @Router /api/v1/wallet/transactions [get]
func (h *WalletHandler) ListTransactions(c *gin.Context) {
	ctx := c.Request.Context()
	userID, err := h.getUserID(c)
	if err != nil {
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	txs, total, err := h.transactionUseCase.ListTransactions(ctx, userID, limit, offset)
	if err != nil {
		h.logger.Error("failed to list transactions", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", I18nTransactionListFailed, nil)
		return
	}

	resp := paymentdto.TransactionListResponse{
		Transactions: make([]paymentdto.TransactionResponse, len(txs)),
		TotalItems:   total,
	}

	for i, tx := range txs {
		resp.Transactions[i] = toTransactionResponse(tx)
	}

	response.Success(c, http.StatusOK, I18nTransactionListSuccess, resp, nil)
}

// Helper functions

func (h *WalletHandler) getUserID(c *gin.Context) (uuid.UUID, error) {
	userIDStr, exists := middleware.GetUserID(c)
	if !exists {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", I18nAuthUnauthorized, nil)
		return uuid.UUID{}, pkgerrors.Unauthorized(I18nAuthUnauthorized, "unauthorized")
	}

	userID, err := uuid.FromString(userIDStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_USER_ID", I18nAuthInvalidUserID, nil)
		return uuid.UUID{}, err
	}

	return userID, nil
}

func toTopupOrderResponse(order *domain.TopupOrder) paymentdto.TopupOrderResponse {
	resp := paymentdto.TopupOrderResponse{
		ID:              order.ID.String(),
		OrderCode:       order.OrderCode,
		CoinAmount:      paymentdto.DecimalToString(order.CoinAmount),
		BaseCoinAmount:  paymentdto.DecimalToString(order.BaseCoinAmount),
		BonusCoinAmount: paymentdto.DecimalToString(order.BonusCoinAmount),
		VNDAmount:       paymentdto.DecimalToString(order.VNDAmount),
		Status:          string(order.Status),
		BankName:        order.BankName,
		BankAccount:     order.BankAccount,
		AccountName:     order.AccountName,
		ExpiredAt:       order.ExpiredAt.Format(time.RFC3339),
		CreatedAt:       order.CreatedAt.Format(time.RFC3339),
	}

	if order.CompletedAt != nil {
		completedAt := order.CompletedAt.Format(time.RFC3339)
		resp.CompletedAt = &completedAt
	}

	return resp
}

func toTransactionResponse(tx *domain.Transaction) paymentdto.TransactionResponse {
	resp := paymentdto.TransactionResponse{
		ID:           tx.ID.String(),
		Type:         string(tx.Type),
		CoinAmount:   paymentdto.DecimalToString(tx.CoinAmount),
		BalanceAfter: paymentdto.DecimalToString(tx.BalanceAfter),
		Description:  tx.Description,
		CreatedAt:    tx.CreatedAt.Format(time.RFC3339),
	}

	if tx.VNDAmount != nil {
		vnd := paymentdto.DecimalToString(*tx.VNDAmount)
		resp.VNDAmount = &vnd
	}

	if tx.ReferenceType != nil {
		resp.ReferenceType = tx.ReferenceType
	}

	if tx.ReferenceID != nil {
		refID := tx.ReferenceID.String()
		resp.ReferenceID = &refID
	}

	return resp
}

func safeDeref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func generateVietQRURL(bank, account, name, amount, content string) string {
	v := url.Values{}
	v.Set("amount", amount)
	v.Set("addInfo", content)
	v.Set("accountName", name)
	return fmt.Sprintf("https://img.vietqr.io/image/%s-%s-compact2.png?%s",
		url.PathEscape(bank),
		url.PathEscape(account),
		v.Encode(),
	)
}
