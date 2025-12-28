package payment

import "github.com/gofrs/uuid/v5"

// === Configuration Requests ===

// UpdateConfigRequest là request để cập nhật config
type UpdateConfigRequest struct {
	Value string `json:"value" binding:"required"`
}

// CreateConfigRequest là request để tạo config mới
type CreateConfigRequest struct {
	Key         string  `json:"key" binding:"required,min=1,max=100"`
	Value       string  `json:"value" binding:"required"`
	ValueType   string  `json:"value_type" binding:"required,oneof=string number boolean json"`
	Description *string `json:"description"`
	IsSensitive bool    `json:"is_sensitive"`
}

// UpsertConfigsRequest là request để cập nhật nhiều config
type UpsertConfigsItem struct {
	Key         string  `json:"key" binding:"required,min=1,max=100"`
	Value       string  `json:"value" binding:"required"`
	ValueType   string  `json:"value_type" binding:"required,oneof=string number boolean json"`
	Description *string `json:"description"`
	IsSensitive bool    `json:"is_sensitive"`
}

type UpsertConfigsRequest struct {
	Configs []UpsertConfigsItem `json:"configs" binding:"required,dive"`
}

// === Wallet Requests ===

// No specific request for wallet (uses auth user ID)

// === Coin Package Requests ===

// No specific request for listing packages (public API)

// === Top-up Requests ===

// CreateTopupRequest là request để tạo đơn nạp tiền
type CreateTopupRequest struct {
	PackageID uuid.UUID `json:"package_id" binding:"required"`
}

// CancelTopupRequest là request để hủy đơn nạp
type CancelTopupRequest struct {
	OrderID uuid.UUID `json:"order_id" binding:"required"`
}

// === Transaction Requests ===

// ListTransactionsRequest là request để lấy lịch sử giao dịch
type ListTransactionsRequest struct {
	Page  int    `form:"page" binding:"omitempty,min=1"`
	Limit int    `form:"limit" binding:"omitempty,min=1,max=100"`
	Type  string `form:"type" binding:"omitempty,oneof=topup purchase_chapter purchase_series rental subscription refund"`
}

// ListTopupOrdersRequest là request để lấy danh sách đơn nạp
type ListTopupOrdersRequest struct {
	Page   int    `form:"page" binding:"omitempty,min=1"`
	Limit  int    `form:"limit" binding:"omitempty,min=1,max=100"`
	Status string `form:"status" binding:"omitempty,oneof=pending success expired cancelled failed"`
}

// === SePay Webhook Request ===

// SepayWebhookRequest là payload từ SePay webhook
type SepayWebhookRequest struct {
	Gateway          string `json:"gateway"`          // "MB", "VCB", etc.
	TransactionID    string `json:"id"`               // ID giao dịch từ SePay
	TransferType     string `json:"transferType"`     // "in" hoặc "out"
	TransferAmount   int64  `json:"transferAmount"`   // Số tiền VND
	Content          string `json:"content"`          // Nội dung chuyển khoản
	TransactionDate  string `json:"transactionDate"`  // Format: "YYYY-MM-DD HH:mm:ss"
	ReferenceNumber  string `json:"referenceNumber"`  // Mã tham chiếu từ ngân hàng
	AccountNumber    string `json:"accountNumber"`    // Số tài khoản
	SubAccount       string `json:"subAccount"`       // Sub account (nếu có)
	Description      string `json:"description"`      // Mô tả
	BankBrandName    string `json:"bankBrandName"`    // Tên ngân hàng
	SepayAccountName string `json:"sepayAccountName"` // Tên account SePay
}
