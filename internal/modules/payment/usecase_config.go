package payment

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"

	"system/internal/domain"
	pkgerrors "system/pkg/errors"
)

// configUseCase implements ConfigUseCase (replaces ConfigService)
type configUseCase struct {
	repo domain.PaymentConfigurationRepository
}

// NewConfigUseCase creates a new configuration use case
func NewConfigUseCase(repo domain.PaymentConfigurationRepository) ConfigUseCase {
	return &configUseCase{repo: repo}
}

// GetAll retrieves all configurations (masks sensitive values)
func (uc *configUseCase) GetAll(ctx context.Context) ([]*domain.PaymentConfiguration, error) {
	configs, err := uc.repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	// Mask sensitive values
	for _, cfg := range configs {
		if cfg.IsSensitive {
			cfg.Value = cfg.MaskedValue()
		}
	}

	return configs, nil
}

// GetByKey retrieves a configuration by key (masks sensitive values)
func (uc *configUseCase) GetByKey(ctx context.Context, key string) (*domain.PaymentConfiguration, error) {
	config, err := uc.repo.GetByKey(ctx, key)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, pkgerrors.NotFound(I18nConfigNotFound, "configuration not found")
		}
		return nil, err
	}

	// Mask sensitive value
	if config.IsSensitive {
		config.Value = config.MaskedValue()
	}

	return config, nil
}

// GetByPrefix retrieves configurations by key prefix
func (uc *configUseCase) GetByPrefix(ctx context.Context, prefix string) ([]*domain.PaymentConfiguration, error) {
	configs, err := uc.repo.GetByPrefix(ctx, prefix)
	if err != nil {
		return nil, err
	}

	// Mask sensitive values
	for _, cfg := range configs {
		if cfg.IsSensitive {
			cfg.Value = cfg.MaskedValue()
		}
	}

	return configs, nil
}

// Update updates a configuration value
func (uc *configUseCase) Update(ctx context.Context, key string, value string, updatedBy uuid.UUID) (*domain.PaymentConfiguration, error) {
	// Check if config exists
	existing, err := uc.repo.GetByKey(ctx, key)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, pkgerrors.NotFound(I18nConfigNotFound, "configuration not found")
		}
		return nil, err
	}

	// Validate value type
	if err := uc.validateValueType(value, existing.ValueType); err != nil {
		return nil, pkgerrors.BadRequest(I18nConfigInvalidValue, "invalid value for type "+string(existing.ValueType))
	}

	// Update
	if err := uc.repo.Update(ctx, key, value, updatedBy); err != nil {
		return nil, err
	}

	// Return updated config (with masked value if sensitive)
	return uc.GetByKey(ctx, key)
}

// Create creates a new configuration
func (uc *configUseCase) Create(ctx context.Context, config *domain.PaymentConfiguration) error {
	// Check if key already exists
	_, err := uc.repo.GetByKey(ctx, config.Key)
	if err == nil {
		return pkgerrors.Conflict(I18nConfigAlreadyExists, "configuration key already exists")
	}
	if err != pgx.ErrNoRows {
		return err
	}

	// Validate value type
	if err := uc.validateValueType(config.Value, config.ValueType); err != nil {
		return pkgerrors.BadRequest(I18nConfigInvalidValue, "invalid value for type "+string(config.ValueType))
	}

	return uc.repo.Create(ctx, config)
}

// Delete deletes a configuration
func (uc *configUseCase) Delete(ctx context.Context, key string) error {
	err := uc.repo.Delete(ctx, key)
	if err == pgx.ErrNoRows {
		return pkgerrors.NotFound(I18nConfigNotFound, "configuration not found")
	}
	return err
}

// UpsertMany updates or creates multiple configurations
func (uc *configUseCase) UpsertMany(ctx context.Context, configs []*domain.PaymentConfiguration) error {
	// Validate each config
	for _, config := range configs {
		if err := uc.validateValueType(config.Value, config.ValueType); err != nil {
			return pkgerrors.BadRequest(I18nConfigInvalidValue, "invalid value for type "+string(config.ValueType)+" (key: "+config.Key+")")
		}
	}
	return uc.repo.UpsertMany(ctx, configs)
}

// GetString retrieves a string configuration value (raw, no masking)
func (uc *configUseCase) GetString(ctx context.Context, key string) (string, error) {
	config, err := uc.repo.GetByKey(ctx, key)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", pkgerrors.NotFound(I18nConfigNotFound, "configuration not found")
		}
		return "", err
	}
	return config.Value, nil
}

// GetNumber retrieves a numeric configuration value
func (uc *configUseCase) GetNumber(ctx context.Context, key string) (float64, error) {
	value, err := uc.GetString(ctx, key)
	if err != nil {
		return 0, err
	}

	num, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, pkgerrors.BadRequest(I18nConfigInvalidValue, "configuration value is not a number")
	}

	return num, nil
}

// GetBool retrieves a boolean configuration value
func (uc *configUseCase) GetBool(ctx context.Context, key string) (bool, error) {
	value, err := uc.GetString(ctx, key)
	if err != nil {
		return false, err
	}

	b, err := strconv.ParseBool(value)
	if err != nil {
		return false, pkgerrors.BadRequest(I18nConfigInvalidValue, "configuration value is not a boolean")
	}

	return b, nil
}

// GetJSON retrieves a JSON configuration value
func (uc *configUseCase) GetJSON(ctx context.Context, key string, target interface{}) error {
	value, err := uc.GetString(ctx, key)
	if err != nil {
		return err
	}

	if err := json.Unmarshal([]byte(value), target); err != nil {
		return pkgerrors.BadRequest(I18nConfigInvalidValue, "configuration value is not valid JSON")
	}

	return nil
}

// validateValueType validates that the value matches the expected type
func (uc *configUseCase) validateValueType(value string, valueType domain.PaymentConfigValueType) error {
	switch valueType {
	case domain.PaymentConfigTypeNumber:
		if _, err := strconv.ParseFloat(value, 64); err != nil {
			return err
		}
	case domain.PaymentConfigTypeBoolean:
		if _, err := strconv.ParseBool(value); err != nil {
			return err
		}
	case domain.PaymentConfigTypeJSON:
		var js interface{}
		if err := json.Unmarshal([]byte(value), &js); err != nil {
			return err
		}
	}
	// string type accepts any value
	return nil
}
