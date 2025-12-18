package payment

import (
	"context"
	_ "embed"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"system/internal/domain"
)

//go:embed queries/get_config_by_key.sql
var getConfigByKeyQuery string

//go:embed queries/get_all_configs.sql
var getAllConfigsQuery string

//go:embed queries/get_configs_by_prefix.sql
var getConfigsByPrefixQuery string

//go:embed queries/update_config.sql
var updateConfigQuery string

//go:embed queries/create_config.sql
var createConfigQuery string

//go:embed queries/delete_config.sql
var deleteConfigQuery string

//go:embed queries/upsert_config.sql
var upsertConfigQuery string

// configurationRepository implements PaymentConfigurationRepository
type configurationRepository struct {
	pool *pgxpool.Pool
}

// NewConfigurationRepository creates a new configuration repository
func NewConfigurationRepository(pool *pgxpool.Pool) domain.PaymentConfigurationRepository {
	return &configurationRepository{pool: pool}
}

// GetByKey retrieves a configuration by its key
func (r *configurationRepository) GetByKey(ctx context.Context, key string) (*domain.PaymentConfiguration, error) {
	row := r.pool.QueryRow(ctx, getConfigByKeyQuery, key)
	return r.scanConfig(row)
}

// GetAll retrieves all configurations
func (r *configurationRepository) GetAll(ctx context.Context) ([]*domain.PaymentConfiguration, error) {
	rows, err := r.pool.Query(ctx, getAllConfigsQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanConfigs(rows)
}

// GetByPrefix retrieves configurations with keys starting with prefix
func (r *configurationRepository) GetByPrefix(ctx context.Context, prefix string) ([]*domain.PaymentConfiguration, error) {
	rows, err := r.pool.Query(ctx, getConfigsByPrefixQuery, prefix+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanConfigs(rows)
}

// Update updates a configuration value
func (r *configurationRepository) Update(ctx context.Context, key string, value string, updatedBy uuid.UUID) error {
	result, err := r.pool.Exec(ctx, updateConfigQuery, value, updatedBy, key)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// Create creates a new configuration
func (r *configurationRepository) Create(ctx context.Context, config *domain.PaymentConfiguration) error {
	_, err := r.pool.Exec(ctx, createConfigQuery,
		config.Key,
		config.Value,
		config.ValueType,
		config.Description,
		config.IsSensitive,
	)
	return err
}

// Delete deletes a configuration
func (r *configurationRepository) Delete(ctx context.Context, key string) error {
	result, err := r.pool.Exec(ctx, deleteConfigQuery, key)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// UpsertMany updates or creates multiple configurations
func (r *configurationRepository) UpsertMany(ctx context.Context, configs []*domain.PaymentConfiguration) error {
	batch := &pgx.Batch{}

	for _, cfg := range configs {
		batch.Queue(upsertConfigQuery,
			cfg.Key,
			cfg.Value,
			cfg.ValueType,
			cfg.Description,
			cfg.IsSensitive,
			cfg.UpdatedBy,
		)
	}

	br := r.pool.SendBatch(ctx, batch)
	defer br.Close()

	for range configs {
		_, err := br.Exec()
		if err != nil {
			return err
		}
	}

	return nil
}

// scanConfig scans a single row into a PaymentConfiguration
func (r *configurationRepository) scanConfig(row pgx.Row) (*domain.PaymentConfiguration, error) {
	var config domain.PaymentConfiguration
	err := row.Scan(
		&config.ID,
		&config.Key,
		&config.Value,
		&config.ValueType,
		&config.Description,
		&config.IsSensitive,
		&config.CreatedAt,
		&config.UpdatedAt,
		&config.UpdatedBy,
	)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// scanConfigs scans multiple rows into PaymentConfiguration slice
func (r *configurationRepository) scanConfigs(rows pgx.Rows) ([]*domain.PaymentConfiguration, error) {
	var configs []*domain.PaymentConfiguration
	for rows.Next() {
		var config domain.PaymentConfiguration
		err := rows.Scan(
			&config.ID,
			&config.Key,
			&config.Value,
			&config.ValueType,
			&config.Description,
			&config.IsSensitive,
			&config.CreatedAt,
			&config.UpdatedAt,
			&config.UpdatedBy,
		)
		if err != nil {
			return nil, err
		}
		configs = append(configs, &config)
	}
	return configs, rows.Err()
}
