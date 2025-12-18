package payment

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid/v5"
	"go.uber.org/zap"

	"system/internal/app/middleware"
	"system/internal/domain"
	paymentdto "system/internal/dto/payment"
	pkgerrors "system/pkg/errors"
	"system/pkg/util/response"
)

// Handler handles payment configuration HTTP requests
type Handler struct {
	configUseCase ConfigUseCase
	logger        *zap.Logger
}

// NewHandler creates a new payment handler
func NewHandler(configUseCase ConfigUseCase, logger *zap.Logger) *Handler {
	return &Handler{
		configUseCase: configUseCase,
		logger:        logger,
	}
}

// ListConfigs returns all configurations
// @Summary List all payment configurations
// @Tags Admin - Payment Config
// @Security BearerAuth
// @Produce json
// @Param prefix query string false "Filter by key prefix (e.g., 'sepay.', 'coin.')"
// @Success 200 {object} paymentdto.ConfigListResponse
// @Router /api/v1/admin/payment/config [get]
func (h *Handler) ListConfigs(c *gin.Context) {
	ctx := c.Request.Context()
	prefix := c.Query("prefix")

	var configs []*domain.PaymentConfiguration
	var err error

	if prefix != "" {
		configs, err = h.configUseCase.GetByPrefix(ctx, prefix)
	} else {
		configs, err = h.configUseCase.GetAll(ctx)
	}

	if err != nil {
		h.logger.Error("failed to list configs", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", I18nConfigListFailed, nil)
		return
	}

	resp := paymentdto.ConfigListResponse{
		Configs: make([]paymentdto.ConfigResponse, len(configs)),
	}

	for i, cfg := range configs {
		resp.Configs[i] = toConfigResponse(cfg)
	}

	response.Success(c, http.StatusOK, I18nConfigListSuccess, resp, nil)
}

// GetConfig returns a single configuration by key
// @Summary Get a payment configuration by key
// @Tags Admin - Payment Config
// @Security BearerAuth
// @Produce json
// @Param key path string true "Configuration key"
// @Success 200 {object} paymentdto.ConfigResponse
// @Router /api/v1/admin/payment/config/{key} [get]
func (h *Handler) GetConfig(c *gin.Context) {
	ctx := c.Request.Context()
	key := c.Param("key")

	config, err := h.configUseCase.GetByKey(ctx, key)
	if err != nil {
		if appErr, ok := pkgerrors.AsAppError(err); ok {
			response.Error(c, appErr.StatusCode, appErr.ErrCode, appErr.I18nKey, nil)
			return
		}
		h.logger.Error("failed to get config", zap.String("key", key), zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", I18nConfigGetFailed, nil)
		return
	}

	response.Success(c, http.StatusOK, I18nConfigGetSuccess, toConfigResponse(config), nil)
}

// UpdateConfig updates a configuration value
// @Summary Update a payment configuration
// @Tags Admin - Payment Config
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param key path string true "Configuration key"
// @Param body body paymentdto.UpdateConfigRequest true "New value"
// @Success 200 {object} paymentdto.ConfigResponse
// @Router /api/v1/admin/payment/config/{key} [put]
func (h *Handler) UpdateConfig(c *gin.Context) {
	ctx := c.Request.Context()
	key := c.Param("key")

	var req paymentdto.UpdateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", I18nValidationFailed, err.Error())
		return
	}

	// Get user ID from context
	userID, exists := middleware.GetUserID(c)
	if !exists {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", I18nAuthUnauthorized, nil)
		return
	}

	updatedBy, err := uuid.FromString(userID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_USER_ID", I18nAuthInvalidUserID, nil)
		return
	}

	config, err := h.configUseCase.Update(ctx, key, req.Value, updatedBy)
	if err != nil {
		if appErr, ok := pkgerrors.AsAppError(err); ok {
			response.Error(c, appErr.StatusCode, appErr.ErrCode, appErr.I18nKey, nil)
			return
		}
		h.logger.Error("failed to update config", zap.String("key", key), zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", I18nConfigUpdateFailed, nil)
		return
	}

	h.logger.Info("configuration updated",
		zap.String("key", key),
		zap.String("updated_by", userID),
	)

	response.Success(c, http.StatusOK, I18nConfigUpdateSuccess, toConfigResponse(config), nil)
}

// CreateConfig creates a new configuration
// @Summary Create a new payment configuration
// @Tags Admin - Payment Config
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body paymentdto.CreateConfigRequest true "Configuration data"
// @Success 201 {object} paymentdto.ConfigResponse
// @Router /api/v1/admin/payment/config [post]
func (h *Handler) CreateConfig(c *gin.Context) {
	ctx := c.Request.Context()

	var req paymentdto.CreateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", I18nValidationFailed, err.Error())
		return
	}

	config := &domain.PaymentConfiguration{
		Key:         req.Key,
		Value:       req.Value,
		ValueType:   domain.PaymentConfigValueType(req.ValueType),
		Description: req.Description,
		IsSensitive: req.IsSensitive,
	}

	if err := h.configUseCase.Create(ctx, config); err != nil {
		if appErr, ok := pkgerrors.AsAppError(err); ok {
			response.Error(c, appErr.StatusCode, appErr.ErrCode, appErr.I18nKey, nil)
			return
		}
		h.logger.Error("failed to create config", zap.String("key", req.Key), zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", I18nConfigCreateFailed, nil)
		return
	}

	h.logger.Info("configuration created", zap.String("key", req.Key))

	// Get created config
	created, _ := h.configUseCase.GetByKey(ctx, req.Key)
	response.Success(c, http.StatusCreated, I18nConfigCreateSuccess, toConfigResponse(created), nil)
}

// UpsertConfigs upserts multiple configurations
// @Summary Upsert multiple payment configurations
// @Tags Admin - Payment Config
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body paymentdto.UpsertConfigsRequest true "Configuration list"
// @Success 200 {object} map[string]string "Success"
// @Router /api/v1/admin/payment/config [put]
func (h *Handler) UpsertConfigs(c *gin.Context) {
	ctx := c.Request.Context()

	var req paymentdto.UpsertConfigsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", I18nValidationFailed, err.Error())
		return
	}

	// Get user ID
	userID, exists := middleware.GetUserID(c)
	if !exists {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", I18nAuthUnauthorized, nil)
		return
	}
	updatedBy, _ := uuid.FromString(userID)

	configs := make([]*domain.PaymentConfiguration, len(req.Configs))
	for i, item := range req.Configs {
		configs[i] = &domain.PaymentConfiguration{
			Key:         item.Key,
			Value:       item.Value,
			ValueType:   domain.PaymentConfigValueType(item.ValueType),
			Description: item.Description,
			IsSensitive: item.IsSensitive,
			UpdatedBy:   &updatedBy,
		}
	}

	if err := h.configUseCase.UpsertMany(ctx, configs); err != nil {
		h.logger.Error("failed to upsert configs", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", I18nConfigUpdateFailed, nil)
		return
	}

	h.logger.Info("configurations upserted", zap.Int("count", len(configs)), zap.String("updated_by", userID))
	response.Success(c, http.StatusOK, I18nConfigUpdateSuccess, gin.H{"message": "success"}, nil)
}

// DeleteConfig deletes a configuration
// @Summary Delete a payment configuration
// @Tags Admin - Payment Config
// @Security BearerAuth
// @Param key path string true "Configuration key"
// @Success 204 "No Content"
// @Router /api/v1/admin/payment/config/{key} [delete]
func (h *Handler) DeleteConfig(c *gin.Context) {
	ctx := c.Request.Context()
	key := c.Param("key")

	if err := h.configUseCase.Delete(ctx, key); err != nil {
		if appErr, ok := pkgerrors.AsAppError(err); ok {
			response.Error(c, appErr.StatusCode, appErr.ErrCode, appErr.I18nKey, nil)
			return
		}
		h.logger.Error("failed to delete config", zap.String("key", key), zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", I18nConfigDeleteFailed, nil)
		return
	}

	h.logger.Info("configuration deleted", zap.String("key", key))
	c.Status(http.StatusNoContent)
}

// toConfigResponse converts a domain.PaymentConfiguration to paymentdto.ConfigResponse
func toConfigResponse(cfg *domain.PaymentConfiguration) paymentdto.ConfigResponse {
	return paymentdto.ConfigResponse{
		Key:         cfg.Key,
		Value:       cfg.Value, // Already masked by service if sensitive
		ValueType:   string(cfg.ValueType),
		Description: cfg.Description,
		IsSensitive: cfg.IsSensitive,
		UpdatedAt:   cfg.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
