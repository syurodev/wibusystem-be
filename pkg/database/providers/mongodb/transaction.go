package mongodb

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
)

// MongoTransaction implements the Transaction interface for MongoDB
type MongoTransaction struct {
	session mongo.Session
}

// Commit commits the MongoDB transaction
func (t *MongoTransaction) Commit(ctx context.Context) error {
	if t.session == nil {
		return fmt.Errorf("mongodb session is nil")
	}

	err := t.session.CommitTransaction(ctx)
	if err != nil {
		return fmt.Errorf("failed to commit mongodb transaction: %w", err)
	}

	t.session.EndSession(ctx)
	return nil
}

// Rollback rolls back the MongoDB transaction
func (t *MongoTransaction) Rollback(ctx context.Context) error {
	if t.session == nil {
		return fmt.Errorf("mongodb session is nil")
	}

	err := t.session.AbortTransaction(ctx)
	if err != nil {
		return fmt.Errorf("failed to rollback mongodb transaction: %w", err)
	}

	t.session.EndSession(ctx)
	return nil
}

// GetSession returns the underlying MongoDB session for transaction operations
func (t *MongoTransaction) GetSession() mongo.Session {
	return t.session
}
