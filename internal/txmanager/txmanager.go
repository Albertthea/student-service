// Package txmanager provides utilities for managing database transactions.
package txmanager

import (
	"context"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type txKey struct{}

// ErrNoTransaction is returned when no transaction is found in context.
var ErrNoTransaction = errors.New("no transaction in context")

// Stores transaction in context
func contextWithTx(ctx context.Context, tx *sqlx.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

// GetTx returns the transaction stored in the context or an error if none is found.
func GetTx(ctx context.Context) (*sqlx.Tx, error) {
	tx, ok := ctx.Value(txKey{}).(*sqlx.Tx)
	if !ok || tx == nil {
		return nil, ErrNoTransaction
	}
	return tx, nil
}

// WithTransaction manages a transaction lifecycle, beginning a transaction if needed,
// running the function, and committing or rolling back depending on the result.
func WithTransaction(ctx context.Context, db *sqlx.DB, run func(ctx context.Context) error) error {
	if tx, _ := GetTx(ctx); tx != nil {
		return run(ctx)
	}

	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("transaction: %w", err)
	}

	ctxWithTx := contextWithTx(ctx, tx)

	err = run(ctxWithTx)
	if err != nil {
		rbErr := tx.Rollback()
		if rbErr != nil {
			return fmt.Errorf("transaction: %w", errors.Join(err, rbErr))
		}
		return fmt.Errorf("transaction: %w", err)
	}
	//nolint:gofmt //
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("transaction: %w", err)
	}

	return nil
}
