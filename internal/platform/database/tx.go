package database

import (
	"context"
	"fmt"

	ent "system/internal/ent/generated"
)

// TxKey is the key for the transaction in the context
type TxKey struct{}

// TransactionManager handles database transactions
type TransactionManager interface {
	RunInTx(ctx context.Context, fn func(ctx context.Context) error) error
}

type entTransactionManager struct {
	client *ent.Client
}

// NewTransactionManager creates a new Ent transaction manager
func NewTransactionManager(client *ent.Client) TransactionManager {
	return &entTransactionManager{client: client}
}

// RunInTx executes the function within an Ent transaction
func (m *entTransactionManager) RunInTx(ctx context.Context, fn func(ctx context.Context) error) error {
	// Start transaction
	tx, err := m.client.Tx(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if v := recover(); v != nil {
			_ = tx.Rollback()
			panic(v)
		}
	}()

	// Inject transactional client into context
	// tx.Client() returns a client that executes all queries within the transaction
	ctxWithTx := context.WithValue(ctx, TxKey{}, tx.Client())

	if err := fn(ctxWithTx); err != nil {
		if rerr := tx.Rollback(); rerr != nil {
			return fmt.Errorf("rolling back transaction: %w (original error: %v)", rerr, err)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}

	return nil
}

// GetClientFromContext retrieves the Ent client from context (if in transaction)
// or returns the fallback client
func GetClientFromContext(ctx context.Context, fallback *ent.Client) *ent.Client {
	if client, ok := ctx.Value(TxKey{}).(*ent.Client); ok {
		return client
	}
	return fallback
}
