// ============================================================================
// Payment Configuration Repository (Ent Implementation)
// ============================================================================

package payment

import (
	"context"
	"system/internal/platform/database"

	"github.com/gofrs/uuid/v5"

	"system/internal/domain"
	ent "system/internal/ent/generated"
	"system/internal/ent/generated/paymentconfiguration"
	"system/internal/ent/generated/predicate"
)

// paymentConfigEntRepository triển khai PaymentConfigurationRepository sử dụng Ent
type paymentConfigEntRepository struct {
	client *ent.Client
}

// NewPaymentConfigEntRepository tạo instance mới
func NewPaymentConfigEntRepository(client *ent.Client) domain.PaymentConfigurationRepository {
	return &paymentConfigEntRepository{client: client}
}

// GetByKey retrieves a configuration by its key
func (r *paymentConfigEntRepository) GetByKey(ctx context.Context, key string) (*domain.PaymentConfiguration, error) {
	cfg, err := database.GetClientFromContext(ctx, r.client).PaymentConfiguration.Query().
		Where(paymentconfiguration.KeyEQ(key)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return entPaymentConfigToDomain(cfg), nil
}

// GetAll retrieves all configurations
func (r *paymentConfigEntRepository) GetAll(ctx context.Context) ([]*domain.PaymentConfiguration, error) {
	configs, err := database.GetClientFromContext(ctx, r.client).PaymentConfiguration.Query().
		Order(ent.Asc(paymentconfiguration.FieldKey)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return entPaymentConfigsToDomain(configs), nil
}

// GetByPrefix retrieves configurations with keys starting with prefix
func (r *paymentConfigEntRepository) GetByPrefix(ctx context.Context, prefix string) ([]*domain.PaymentConfiguration, error) {
	configs, err := database.GetClientFromContext(ctx, r.client).PaymentConfiguration.Query().
		Where(paymentconfiguration.KeyHasPrefix(prefix)).
		Order(ent.Asc(paymentconfiguration.FieldKey)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	return entPaymentConfigsToDomain(configs), nil
}

// Update updates a configuration value
func (r *paymentConfigEntRepository) Update(ctx context.Context, key string, value string, updatedBy uuid.UUID) error {
	_, err := database.GetClientFromContext(ctx, r.client).PaymentConfiguration.Update().
		Where(paymentconfiguration.KeyEQ(key)).
		SetValue(value).
		SetUpdatedBy(updatedBy).
		Save(ctx)
	return err
}

// Create creates a new configuration
func (r *paymentConfigEntRepository) Create(ctx context.Context, config *domain.PaymentConfiguration) error {
	builder := database.GetClientFromContext(ctx, r.client).PaymentConfiguration.Create().
		SetKey(config.Key).
		SetValue(config.Value).
		SetValueType(paymentconfiguration.ValueType(config.ValueType)).
		SetIsSensitive(config.IsSensitive)

	if config.Description != nil {
		builder.SetDescription(*config.Description)
	}
	if config.UpdatedBy != nil {
		builder.SetUpdatedBy(*config.UpdatedBy)
	}

	created, err := builder.Save(ctx)
	if err != nil {
		return err
	}
	config.ID = created.ID
	config.CreatedAt = created.CreatedAt
	config.UpdatedAt = created.UpdatedAt
	return nil
}

// Delete deletes a configuration
func (r *paymentConfigEntRepository) Delete(ctx context.Context, key string) error {
	_, err := database.GetClientFromContext(ctx, r.client).PaymentConfiguration.Delete().
		Where(paymentconfiguration.KeyEQ(key)).
		Exec(ctx)
	return err
}

// UpsertMany updates or creates multiple configurations
func (r *paymentConfigEntRepository) UpsertMany(ctx context.Context, configs []*domain.PaymentConfiguration) error {
	for _, config := range configs {
		// Check if exists
		exists, err := database.GetClientFromContext(ctx, r.client).PaymentConfiguration.Query().
			Where(paymentconfiguration.KeyEQ(config.Key)).
			Exist(ctx)
		if err != nil {
			return err
		}

		if exists {
			// Update
			builder := database.GetClientFromContext(ctx, r.client).PaymentConfiguration.Update().
				Where(paymentconfiguration.KeyEQ(config.Key)).
				SetValue(config.Value).
				SetValueType(paymentconfiguration.ValueType(config.ValueType)).
				SetIsSensitive(config.IsSensitive)

			if config.Description != nil {
				builder.SetDescription(*config.Description)
			}
			if config.UpdatedBy != nil {
				builder.SetUpdatedBy(*config.UpdatedBy)
			}

			_, err = builder.Save(ctx)
		} else {
			// Create
			builder := database.GetClientFromContext(ctx, r.client).PaymentConfiguration.Create().
				SetKey(config.Key).
				SetValue(config.Value).
				SetValueType(paymentconfiguration.ValueType(config.ValueType)).
				SetIsSensitive(config.IsSensitive)

			if config.Description != nil {
				builder.SetDescription(*config.Description)
			}
			if config.UpdatedBy != nil {
				builder.SetUpdatedBy(*config.UpdatedBy)
			}

			_, err = builder.Save(ctx)
		}

		if err != nil {
			return err
		}
	}
	return nil
}

func entPaymentConfigToDomain(cfg *ent.PaymentConfiguration) *domain.PaymentConfiguration {
	return &domain.PaymentConfiguration{
		ID:          cfg.ID,
		Key:         cfg.Key,
		Value:       cfg.Value,
		ValueType:   domain.PaymentConfigValueType(cfg.ValueType),
		Description: cfg.Description,
		IsSensitive: cfg.IsSensitive,
		UpdatedBy:   cfg.UpdatedBy,
		CreatedAt:   cfg.CreatedAt,
		UpdatedAt:   cfg.UpdatedAt,
	}
}

func entPaymentConfigsToDomain(configs []*ent.PaymentConfiguration) []*domain.PaymentConfiguration {
	results := make([]*domain.PaymentConfiguration, len(configs))
	for i, cfg := range configs {
		results[i] = entPaymentConfigToDomain(cfg)
	}
	return results
}

// Ensure predicate is used (to avoid import error)
var _ predicate.PaymentConfiguration = nil
