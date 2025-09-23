package repositories

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
)

// HealthRepository exposes health-check operations for backing stores.
type HealthRepository interface {
	Ping(ctx context.Context) error
}

type healthRepository struct {
	pool *pgxpool.Pool
}

// NewHealthRepository creates a HealthRepository backed by pgx.
func NewHealthRepository(pool *pgxpool.Pool) HealthRepository {
	return &healthRepository{pool: pool}
}

func (r *healthRepository) Ping(ctx context.Context) error {
	if r.pool == nil {
		return errors.New("postgres pool is nil")
	}
	return r.pool.Ping(ctx)
}
