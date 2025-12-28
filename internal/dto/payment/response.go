package payment

import "github.com/shopspring/decimal"

// === Configuration Responses ===

// ConfigResponse là response cho một config
type ConfigResponse struct {
	Key         string  `json:"key"`
	Value       string  `json:"value"`
	ValueType   string  `json:"value_type"`
	Description *string `json:"description,omitempty"`
	IsSensitive bool    `json:"is_sensitive"`
	UpdatedAt   string  `json:"updated_at"`
}

// ConfigListResponse là response cho danh sách configs
type ConfigListResponse struct {
	Configs []ConfigResponse `json:"configs"`
}

// === Wallet Responses ===

// WalletResponse là response cho thông tin ví
type WalletResponse struct {
	CoinBalance            string `json:"coin_balance"`
	TotalDeposited         string `json:"total_deposited"`
	TotalSpent             string `json:"total_spent"`
	TotalSubscriptionSpent string `json:"total_subscription_spent"`
	UpdatedAt              string `json:"updated_at"`
}

// === Coin Package Responses ===

// CoinPackageResponse là response cho một gói nạp
type CoinPackageResponse struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Slug         string `json:"slug"`
	CoinAmount   string `json:"coin_amount"`   // Base coins
	BonusPercent int    `json:"bonus_percent"` // % bonus
	TotalCoins   string `json:"total_coins"`   // Base + bonus
	PriceVND     string `json:"price_vnd"`     // Giá VND
	IsPopular    bool   `json:"is_popular"`
}

// CoinPackageListResponse là response cho danh sách gói nạp
type CoinPackageListResponse struct {
	Packages []CoinPackageResponse `json:"packages"`
}

// === Top-up Responses ===

// TopupOrderResponse là response cho đơn nạp
type TopupOrderResponse struct {
	ID              string  `json:"id"`
	OrderCode       string  `json:"order_code"`  // Nội dung CK: "NAP" + code
	CoinAmount      string  `json:"coin_amount"` // Total coins (base + bonus)
	BaseCoinAmount  string  `json:"base_coin_amount"`
	BonusCoinAmount string  `json:"bonus_coin_amount"`
	VNDAmount       string  `json:"vnd_amount"`
	Status          string  `json:"status"`
	BankName        *string `json:"bank_name,omitempty"`
	BankAccount     *string `json:"bank_account,omitempty"`
	AccountName     *string `json:"account_name,omitempty"`
	QRCodeURL       *string `json:"qr_code_url,omitempty"` // URL to QR image
	ExpiredAt       string  `json:"expired_at"`
	CreatedAt       string  `json:"created_at"`
	CompletedAt     *string `json:"completed_at,omitempty"`
}

// CreateTopupResponse là response khi tạo đơn nạp
type CreateTopupResponse struct {
	Order        TopupOrderResponse   `json:"order"`
	TransferInfo TransferInfoResponse `json:"transfer_info"`
}

// TransferInfoResponse là thông tin chuyển khoản
type TransferInfoResponse struct {
	BankName      string `json:"bank_name"`
	BankAccount   string `json:"bank_account"`
	AccountName   string `json:"account_name"`
	Amount        string `json:"amount"`          // VND
	Content       string `json:"content"`         // Nội dung CK
	QRCodeURL     string `json:"qr_code_url"`     // URL to QR image (if available)
	ExpiresInSecs int    `json:"expires_in_secs"` // Thời gian còn lại
}

// TopupOrderListResponse là response cho danh sách đơn nạp
type TopupOrderListResponse struct {
	Orders     []TopupOrderResponse `json:"orders"`
	TotalItems int                  `json:"total_items"`
}

// === Transaction Responses ===

// TransactionResponse là response cho một giao dịch
type TransactionResponse struct {
	ID            string  `json:"id"`
	Type          string  `json:"type"`
	CoinAmount    string  `json:"coin_amount"` // + nạp, - tiêu
	VNDAmount     *string `json:"vnd_amount,omitempty"`
	BalanceAfter  string  `json:"balance_after"`
	Description   *string `json:"description,omitempty"`
	ReferenceType *string `json:"reference_type,omitempty"`
	ReferenceID   *string `json:"reference_id,omitempty"`
	CreatedAt     string  `json:"created_at"`
}

// TransactionListResponse là response cho lịch sử giao dịch
type TransactionListResponse struct {
	Transactions []TransactionResponse `json:"transactions"`
	TotalItems   int                   `json:"total_items"`
}

// === WebSocket Notification ===

// TopupNotification là thông báo realtime khi topup hoàn thành
type TopupNotification struct {
	Type          string            `json:"type"`
	OrderID       string            `json:"order_id"`
	OrderCode     string            `json:"order_code"`
	CoinAmount    string            `json:"coin_amount"`
	NewBalance    string            `json:"new_balance"`
	Message       string            `json:"message"`
	MessageKey    string            `json:"message_key"`
	MessageParams map[string]string `json:"message_params,omitempty"`
}

// === Helper functions ===

// DecimalToString converts decimal to string for JSON
func DecimalToString(d decimal.Decimal) string {
	return d.StringFixed(2)
}
