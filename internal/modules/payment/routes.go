package payment

import (
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers payment configuration routes (Admin only)
func (h *Handler) RegisterRoutes(adminRouter *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	// All payment config routes require authentication (Admin only)
	config := adminRouter.Group("/payment/config")
	if authMiddleware != nil {
		config.Use(authMiddleware)
	}

	config.GET("", h.ListConfigs)          // GET /api/v1/admin/payment/config
	config.PUT("", h.UpsertConfigs)        // PUT /api/v1/admin/payment/config (Bulk)
	config.GET("/:key", h.GetConfig)       // GET /api/v1/admin/payment/config/:key
	config.PUT("/:key", h.UpdateConfig)    // PUT /api/v1/admin/payment/config/:key
	config.POST("", h.CreateConfig)        // POST /api/v1/admin/payment/config
	config.DELETE("/:key", h.DeleteConfig) // DELETE /api/v1/admin/payment/config/:key
}

// RegisterWalletRoutes registers wallet and topup routes
func (h *WalletHandler) RegisterWalletRoutes(apiV1 *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	// Wallet routes (requires auth)
	wallet := apiV1.Group("/wallet")
	wallet.Use(authMiddleware)
	wallet.GET("", h.GetWallet)                     // GET /api/v1/wallet
	wallet.GET("/transactions", h.ListTransactions) // GET /api/v1/wallet/transactions

	// Payment routes
	payment := apiV1.Group("/payment")

	// Public: List packages
	payment.GET("/packages", h.ListPackages) // GET /api/v1/payment/packages

	// Protected: Topup operations
	topup := payment.Group("/topup")
	topup.Use(authMiddleware)
	topup.POST("", h.CreateTopup)            // POST /api/v1/payment/topup
	topup.GET("", h.ListTopups)              // GET /api/v1/payment/topup
	topup.GET("/:id", h.GetTopup)            // GET /api/v1/payment/topup/:id
	topup.POST("/:id/cancel", h.CancelTopup) // POST /api/v1/payment/topup/:id/cancel
}

// RegisterWebhookRoutes registers public webhook routes
func (h *WebhookHandler) RegisterWebhookRoutes(router *gin.Engine) {
	// Public webhook endpoint (no auth - SePay calls this)
	webhook := router.Group("/api/webhook")
	webhook.POST("/sepay", h.HandleSepayWebhook) // POST /api/webhook/sepay
}
