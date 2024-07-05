package transaction

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Transaction struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Transaction {
	return &Transaction{
		db: db,
	}
}

// Exec запускает функцию f в новой транзакции, если не передана явно.
func (t *Transaction) Exec(
	ctx context.Context,
	f func(ctx context.Context, tx pgx.Tx) error,
) error {
	tx, err := t.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	err = f(ctx, tx)
	if err != nil {
		_ = tx.Rollback(ctx)

		return fmt.Errorf("failed to execute transaction: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
