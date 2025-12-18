package domain

import (
	"context"
	"time"

	"github.com/gofrs/uuid/v5"
)

// PaymentConfigValueType represents the type of configuration value
type PaymentConfigValueType string

const (
	PaymentConfigTypeString  PaymentConfigValueType = "string"
	PaymentConfigTypeNumber  PaymentConfigValueType = "number"
	PaymentConfigTypeBoolean PaymentConfigValueType = "boolean"
	PaymentConfigTypeJSON    PaymentConfigValueType = "json"
)

// PaymentConfiguration represents a payment configuration entry
type PaymentConfiguration struct {
	ID          uuid.UUID
	Key         string
	Value       string
	ValueType   PaymentConfigValueType
	Description *string
	IsSensitive bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
	UpdatedBy   *uuid.UUID
}

// MaskedValue returns the value with sensitive data masked
func (c *PaymentConfiguration) MaskedValue() string {
	if c.IsSensitive && c.Value != "" {
		return "***"
	}
	return c.Value
}

// PaymentConfigurationRepository defines the interface for payment configuration data access
type PaymentConfigurationRepository interface {
	// GetByKey retrieves a configuration by its key
	GetByKey(ctx context.Context, key string) (*PaymentConfiguration, error)

	// GetAll retrieves all configurations
	GetAll(ctx context.Context) ([]*PaymentConfiguration, error)

	// GetByPrefix retrieves configurations with keys starting with prefix
	GetByPrefix(ctx context.Context, prefix string) ([]*PaymentConfiguration, error)

	// Update updates a configuration value
	Update(ctx context.Context, key string, value string, updatedBy uuid.UUID) error

	// Create creates a new configuration
	Create(ctx context.Context, config *PaymentConfiguration) error

	// Delete deletes a configuration
	Delete(ctx context.Context, key string) error

	// UpsertMany updates or creates multiple configurations
	UpsertMany(ctx context.Context, configs []*PaymentConfiguration) error
}
