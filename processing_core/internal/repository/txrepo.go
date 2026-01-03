package repository

import (
	"context"
	"errors"
	"fmt"
	"processing_core/internal/utils"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// RollBackUnlessCommitted - роллбэк, если транзакция не закоммичена
func (r *pgDb) RollBackUnlessCommitted(ctx context.Context, tx pgx.Tx) {
	if tx == nil {
		return
	}

	err := tx.Rollback(ctx)
	if errors.Is(err, pgx.ErrTxClosed) {
		return
	}

	// if err != nil {
		// logger.Errorf(ctx, "can't rollback transaction: %v", err)
	// }
}

// BeginTx - возвращает открытую транзакцию
func (r *pgDb) BeginTx(ctx context.Context) (pgx.Tx, error) {
	tx, err := r.conn.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("can't begin transaction, error: %w", err)
	}

	return tx, nil
}

// CommitTx - коммитит транзакцию
func (r *pgDb) CommitTx(ctx context.Context, tx pgx.Tx) error {
	return tx.Commit(ctx)
}

// Transactional – открывает транзакцию и выполняет функцию в ней.
func (r *pgDb) Transactional(
	ctx context.Context,
	f func(tx pgx.Tx) error,
) error {
	tx, err := r.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer r.RollBackUnlessCommitted(ctx, tx)
	err = f(tx)
	if err == nil {
		return r.CommitTx(ctx, tx)
	}
	return err
}

// LockClient - делает блокировку для клиента, чтобы работать с ним в одном потоке
func (r *pgDb) LockClient(
	ctx context.Context,
	tx pgx.Tx,
	clientID uuid.UUID,
) error {
	query := `SELECT pg_advisory_xact_lock($1)`

	idHash, err := utils.Hash(clientID[:])
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, query, idHash)
	if err != nil {
		return fmt.Errorf("can't lock client with id: %v, hash: %v, error: %w", clientID, idHash, err)
	}

	return nil
}
