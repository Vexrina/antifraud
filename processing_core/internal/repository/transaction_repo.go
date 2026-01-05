package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/go-jet/jet/v2/postgres"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"processing_core/generated/proc_core_db/public/model"
	"processing_core/generated/proc_core_db/public/table"
)

// TransactionFilter - фильтр для поиска транзакций
type TransactionFilter struct {
	ID              *int64
	TransactionID   *uuid.UUID
	SenderID        *uuid.UUID
	ReceiverID      *uuid.UUID
	TransactionType *model.TransactionType
	Merchant        *string
	Country         *string
	AmountMin       *int64
	AmountMax       *int64
	CreatedAtFrom   *time.Time
	CreatedAtTo     *time.Time
	Limit           *int
	Offset          *int
}

// AppendTransaction - добавляет новую транзакцию в историю
func (db *pgDb) AppendTransaction(ctx context.Context, tx pgx.Tx, transaction model.TransactionsHistory) error {
	t := table.TransactionsHistory
	stmt := t.INSERT(
		t.AllColumns.Except(
			t.Revision,
		),
	).MODEL(transaction)

	sql, args := stmt.Sql()
	_, err := tx.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("can't append transaction: %w", err)
	}
	return nil
}

// UpsertTransaction - обновляет транзакцию, если она существует, или создает новую
func (db *pgDb) UpsertTransaction(ctx context.Context, tx pgx.Tx, transaction model.TransactionsHistory) error {
	t := table.TransactionsHistory
	stmt := t.INSERT(
		t.AllColumns.Except(
			t.Revision,
		),
	).MODEL(transaction).ON_CONFLICT(
		t.TransactionID,
	).DO_UPDATE(postgres.SET(
		t.Amount.SET(t.EXCLUDED.Amount),
		t.Currency.SET(t.EXCLUDED.Currency),
		t.Merchant.SET(t.EXCLUDED.Merchant),
		t.Country.SET(t.EXCLUDED.Country),
		t.SenderID.SET(t.EXCLUDED.SenderID),
		t.ReceiverID.SET(t.EXCLUDED.ReceiverID),
		t.ReceiverBic.SET(t.EXCLUDED.ReceiverBic),
		t.AtmID.SET(t.EXCLUDED.AtmID),
		t.TransactionType.SET(t.EXCLUDED.TransactionType),
		t.CreatedAt.SET(t.EXCLUDED.CreatedAt),
	))

	sql, args := stmt.Sql()
	_, err := tx.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("can't upsert transaction: %w", err)
	}
	return nil
}

// GetTransactions - получает транзакции по составному фильтру
func (db *pgDb) GetTransactions(ctx context.Context, tx pgx.Tx, filter TransactionFilter) ([]model.TransactionsHistory, error) {
	t := table.TransactionsHistory
	stmt := t.SELECT(t.AllColumns.As(""))

	// Применяем фильтры
	conditions := make([]postgres.BoolExpression, 0)

	if filter.ID != nil {
		conditions = append(conditions, t.ID.EQ(postgres.Int(*filter.ID)))
	}
	if filter.TransactionID != nil {
		conditions = append(conditions, t.TransactionID.EQ(postgres.UUID(*filter.TransactionID)))
	}
	if filter.SenderID != nil {
		conditions = append(conditions, t.SenderID.EQ(postgres.UUID(*filter.SenderID)))
	}
	if filter.ReceiverID != nil {
		conditions = append(conditions, t.ReceiverID.EQ(postgres.UUID(*filter.ReceiverID)))
	}
	if filter.TransactionType != nil {
		conditions = append(conditions,
			t.TransactionType.EQ(
				postgres.RawString(
					"val::public.transaction_type",
					map[string]interface{}{
						"val": string(*filter.TransactionType),
					},
				),
			),
		)
	}
	if filter.Merchant != nil {
		conditions = append(conditions, t.Merchant.EQ(postgres.String(*filter.Merchant)))
	}
	if filter.Country != nil {
		conditions = append(conditions, t.Country.EQ(postgres.String(*filter.Country)))
	}
	if filter.AmountMin != nil {
		conditions = append(conditions, t.Amount.GT_EQ(postgres.Int64(*filter.AmountMin)))
	}
	if filter.AmountMax != nil {
		conditions = append(conditions, t.Amount.LT_EQ(postgres.Int64(*filter.AmountMax)))
	}
	if filter.CreatedAtFrom != nil {
		conditions = append(conditions, t.CreatedAt.GT_EQ(postgres.TimestampzT(*filter.CreatedAtFrom)))
	}
	if filter.CreatedAtTo != nil {
		conditions = append(conditions, t.CreatedAt.LT_EQ(postgres.TimestampzT(*filter.CreatedAtTo)))
	}

	// Объединяем все условия через AND1
	if len(conditions) > 0 {
		whereClause := conditions[0]
		for i := 1; i < len(conditions); i++ {
			whereClause = whereClause.AND(conditions[i])
		}
		stmt = stmt.WHERE(whereClause)
	}

	// Сортировка по дате создания (новые сначала)
	stmt = stmt.ORDER_BY(t.CreatedAt.DESC())

	// Применяем LIMIT и OFFSET
	if filter.Limit != nil {
		stmt = stmt.LIMIT(int64(*filter.Limit))
	}
	if filter.Offset != nil {
		stmt = stmt.OFFSET(int64(*filter.Offset))
	}

	sql, args := stmt.Sql()

	rows, err := tx.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("can't get transactions: %w", err)
	}

	transactions, err := pgx.CollectRows(rows, pgx.RowToStructByName[model.TransactionsHistory])
	if err != nil {
		return nil, fmt.Errorf("can't collect transactions: %w", err)
	}

	return transactions, nil
}
