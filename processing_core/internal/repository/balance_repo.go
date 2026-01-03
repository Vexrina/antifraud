package repository

import (
	"context"
	"fmt"

	"github.com/go-jet/jet/v2/postgres"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"processing_core/generated/proc_core_db/public/model"
	"processing_core/generated/proc_core_db/public/table"
)

func (db *pgDb) GetCurrentBalanceTx(ctx context.Context, tx pgx.Tx, clientID uuid.UUID) (int64, error) {
	t := table.UserBalance
	stmt := t.SELECT(
		t.AllColumns.As(""),
	).WHERE(
		t.ClientID.EQ(postgres.UUID(clientID)),
	)
	sql, args := stmt.Sql()

	rows, err := tx.Query(ctx, sql, args...)
	if err != nil {
		return 0, err
	}
	res, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[model.UserBalance])

	if err != nil {
		return 0, err
	}

	if res.Balance == nil {
		return 0, fmt.Errorf("got nullable balance")
	}
	return *res.Balance, nil
}

func (db *pgDb) UpdateBalance(ctx context.Context, tx pgx.Tx, clientID uuid.UUID, newBalance int64) error {
	t := table.UserBalance
	balanceModel := model.UserBalance{
		ClientID: &clientID,
		Balance:  &newBalance,
	}
	stmt := t.INSERT(
		t.AllColumns.Except(
			t.Revision,
			t.CreatedAt,
			t.UpdatedAt,
		),
	).MODEL(balanceModel).ON_CONFLICT(
		t.ClientID,
	).DO_UPDATE(postgres.SET(
		t.Balance.SET(t.EXCLUDED.Balance),
		t.UpdatedAt.SET(postgres.RawTimestampz("now()")),
		t.Revision.SET(postgres.RawInt("nextval('user_balance_revision_seq'::regclass)")),
	))

	sql, args := stmt.Sql()
	_, err := tx.Exec(ctx, sql, args...)
	if err != nil {
		return err
	}
	return nil
}
