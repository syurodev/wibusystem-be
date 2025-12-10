package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TxKey is the key for the transaction in the context
type TxKey struct{}

// DBTX is an interface that satisfies both *pgxpool.Pool and pgx.Tx
type DBTX interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

// GetDB returns the transaction from the context if present, otherwise the pool
func GetDB(ctx context.Context, pool DBTX) DBTX {
	if tx, ok := ctx.Value(TxKey{}).(pgx.Tx); ok {
		return tx
	}
	return pool
}

// TransactionManager handles database transactions
type TransactionManager interface {
	RunInTx(ctx context.Context, fn func(ctx context.Context) error) error
}

type txManager struct {
	pool *pgxpool.Pool
}

func NewTransactionManager(pool *pgxpool.Pool) TransactionManager {
	return &txManager{pool: pool}
}

func (m *txManager) RunInTx(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := m.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		} else if err != nil {
			_ = tx.Rollback(ctx)
		} else {
			err = tx.Commit(ctx)
		}
	}()

	// Inject transaction into context
	ctxWithTx := context.WithValue(ctx, TxKey{}, tx)
	err = fn(ctxWithTx)
	return err
}
