// Package postgres implements transaction wrapper for PostgreSQL
package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
)

// PostgresTransaction wraps pgx.Tx to implement interfaces.Transaction
type PostgresTransaction struct {
	tx pgx.Tx
}

func (t *PostgresTransaction) Commit(ctx context.Context) error {
	return t.tx.Commit(ctx)
}

func (t *PostgresTransaction) Rollback(ctx context.Context) error {
	return t.tx.Rollback(ctx)
}
