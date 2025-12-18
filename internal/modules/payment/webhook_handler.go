package payment

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	paymentdto "system/internal/dto/payment"
	pkgerrors "system/pkg/errors"
	"system/pkg/util/response"
)

// WebhookHandler handles SePay webhook requests
type WebhookHandler struct {
	topupUseCase  TopupUseCase
	configUseCase ConfigUseCase
	logger        *zap.Logger
}

// NewWebhookHandler creates a new webhook handler
func NewWebhookHandler(topupUseCase TopupUseCase, configUseCase ConfigUseCase, logger *zap.Logger) *WebhookHandler {
	return &WebhookHandler{
		topupUseCase:  topupUseCase,
		configUseCase: configUseCase,
		logger:        logger,
	}
}

// HandleSepayWebhook processes SePay webhook callbacks
// @Summary Handle SePay webhook
// @Tags Webhook
// @Accept json
// @Produce json
// @Param body body paymentdto.SepayWebhookRequest true "SePay webhook payload"
// @Success 200
// @Router /api/webhook/sepay [post]
func (h *WebhookHandler) HandleSepayWebhook(c *gin.Context) {
	ctx := c.Request.Context()

	var req paymentdto.SepayWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("invalid webhook payload", zap.Error(err))
		response.Error(c, http.StatusBadRequest, "INVALID_PAYLOAD", I18nTopupInvalidWebhook, nil)
		return
	}

	// Log incoming webhook
	h.logger.Info("received SePay webhook",
		zap.String("transaction_id", req.TransactionID),
		zap.String("content", req.Content),
		zap.Int64("amount", req.TransferAmount),
		zap.String("gateway", req.Gateway),
	)

	// Only process incoming transfers
	if req.TransferType != "in" {
		h.logger.Debug("ignoring outgoing transfer")
		c.JSON(http.StatusOK, gin.H{"success": true, "message": "ignored"})
		return
	}

	// TODO: Validate webhook signature if SePay provides one
	// apiKey, _ := h.configUseCase.GetString(ctx, "sepay.api_key")
	// if !validateSignature(req, apiKey) { ... }

	// Process the webhook
	if err := h.topupUseCase.ProcessWebhook(ctx, req.TransactionID, req.TransferAmount, req.Content); err != nil {
		if appErr, ok := pkgerrors.AsAppError(err); ok {
			// Business errors - still return 200 to SePay but log them
			h.logger.Warn("webhook processing failed (business error)",
				zap.String("transaction_id", req.TransactionID),
				zap.Error(err),
			)
			c.JSON(http.StatusOK, gin.H{"success": false, "code": appErr.ErrCode, "message": appErr.Message})
			return
		}

		// System errors
		h.logger.Error("webhook processing failed (system error)",
			zap.String("transaction_id", req.TransactionID),
			zap.Error(err),
		)
		response.Error(c, http.StatusInternalServerError, "PROCESSING_ERROR", I18nTopupCreateFailed, nil)
		return
	}

	h.logger.Info("webhook processed successfully",
		zap.String("transaction_id", req.TransactionID),
	)

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "processed"})
}
