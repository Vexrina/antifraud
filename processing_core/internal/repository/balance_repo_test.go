package repository

import (
	"context"
	"processing_core/generated/proc_core_db/public/table"
	"testing"

	"github.com/go-jet/jet/v2/postgres"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPgDb_GetCurrentBalanceTx(t *testing.T) {
	t.Parallel()

	type args struct {
		clientID uuid.UUID
	}
	type fields struct {
		prepare func(ctx context.Context, db *pgDb, a *args) error
	}
	tests := []struct {
		name            string
		fields          fields
		args            args
		expectedBalance int64
		expectedError   string
		shouldCommit    bool
	}{
		{
			name: "success - get existing balance",
			fields: fields{
				prepare: func(ctx context.Context, db *pgDb, a *args) error {
					return db.Transactional(ctx, func(tx pgx.Tx) error {
						err := db.UpdateBalance(ctx, tx, a.clientID, int64(1000))
						if err != nil {
							t.Fatal(err.Error())
							return err
						}
						return nil
					})
				},
			},
			args: args{
				clientID: uuid.New(),
			},
			expectedBalance: 1000,
			expectedError:   "",
			shouldCommit:    true,
		},
		{
			name: "not found - balance does not exist",
			fields: fields{
				prepare: func(ctx context.Context, db *pgDb, a *args) error {
					// не создаем баланс, чтобы получить ошибку
					return nil
				},
			},
			args: args{
				clientID: uuid.New(),
			},
			expectedBalance: 0,
			expectedError:   "no rows in result set",
			shouldCommit:    false,
		},
		{
			name: "success - get zero balance",
			fields: fields{
				prepare: func(ctx context.Context, db *pgDb, a *args) error {
					return db.Transactional(ctx, func(tx pgx.Tx) error {
						err := db.UpdateBalance(ctx, tx, a.clientID, int64(0))
						if err != nil {
							t.Fatal(err.Error())
							return err
						}
						return nil
					})
				},
			},
			args: args{
				clientID: uuid.New(),
			},
			expectedBalance: 0,
			expectedError:   "",
			shouldCommit:    true,
		},
		{
			name: "success - get negative balance",
			fields: fields{
				prepare: func(ctx context.Context, db *pgDb, a *args) error {
					return db.Transactional(ctx, func(tx pgx.Tx) error {
						err := db.UpdateBalance(ctx, tx, a.clientID, int64(-500))
						if err != nil {
							t.Fatal(err.Error())
							return err
						}
						return nil
					})
				},
			},
			args: args{
				clientID: uuid.New(),
			},
			expectedBalance: -500,
			expectedError:   "",
			shouldCommit:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			db := NewTestDB(ctx, t)

			t.Cleanup(func() {
				tl := table.UserBalance
				stmt := tl.DELETE().WHERE(
					tl.ClientID.EQ(postgres.UUID(tt.args.clientID)),
				)
				sql, args := stmt.Sql()
				_, err := db.conn.Exec(ctx, sql, args...)
				if err != nil {
					t.Fatal(err)
				}
			})

			err := tt.fields.prepare(ctx, db, &tt.args)
			require.NoError(t, err)

			var resBalance int64
			err = db.Transactional(ctx, func(tx pgx.Tx) error {
				res, txErr := db.GetCurrentBalanceTx(ctx, tx, tt.args.clientID)
				resBalance = res
				return txErr
			})

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Equal(t, tt.expectedError, err.Error())
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expectedBalance, resBalance)

		})
	}
}

func TestPgDb_UpdateBalance(t *testing.T) {
	t.Parallel()

	type args struct {
		clientID      uuid.UUID
		updateBalance int64
	}
	type fields struct {
		prepare func(ctx context.Context, db *pgDb, a *args) error
	}
	tests := []struct {
		name            string
		fields          fields
		args            args
		expectedBalance int64
		expectedError   string
	}{
		{
			name: "create new balance",
			fields: fields{
				prepare: func(ctx context.Context, db *pgDb, a *args) error {
					// не создаем начальный баланс
					return nil
				},
			},
			args: args{
				clientID:      uuid.New(),
				updateBalance: 500,
			},
			expectedBalance: 500,
			expectedError:   "",
		},
		{
			name: "update existing balance",
			fields: fields{
				prepare: func(ctx context.Context, db *pgDb, a *args) error {
					return db.Transactional(ctx, func(tx pgx.Tx) error {
						err := db.UpdateBalance(ctx, tx, a.clientID, int64(500))
						if err != nil {
							t.Fatal(err.Error())
							return err
						}
						return nil
					})
				},
			},
			args: args{
				clientID:      uuid.New(),
				updateBalance: 1500,
			},
			expectedBalance: 1500,
			expectedError:   "",
		},
		{
			name: "update to zero balance",
			fields: fields{
				prepare: func(ctx context.Context, db *pgDb, a *args) error {
					return db.Transactional(ctx, func(tx pgx.Tx) error {
						err := db.UpdateBalance(ctx, tx, a.clientID, int64(1000))
						if err != nil {
							t.Fatal(err.Error())
							return err
						}
						return nil
					})
				},
			},
			args: args{
				clientID:      uuid.New(),
				updateBalance: 0,
			},
			expectedBalance: 0,
			expectedError:   "",
		},
		{
			name: "update to negative balance",
			fields: fields{
				prepare: func(ctx context.Context, db *pgDb, a *args) error {
					return db.Transactional(ctx, func(tx pgx.Tx) error {
						err := db.UpdateBalance(ctx, tx, a.clientID, int64(1000))
						if err != nil {
							t.Fatal(err.Error())
							return err
						}
						return nil
					})
				},
			},
			args: args{
				clientID:      uuid.New(),
				updateBalance: -200,
			},
			expectedBalance: -200,
			expectedError:   "",
		},
		{
			name: "multiple updates in same transaction",
			fields: fields{
				prepare: func(ctx context.Context, db *pgDb, a *args) error {
					// не создаем начальный баланс
					return nil
				},
			},
			args: args{
				clientID:      uuid.New(),
				updateBalance: 3000,
			},
			expectedBalance: 3000,
			expectedError:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			db := NewTestDB(ctx, t)

			t.Cleanup(func() {
				tl := table.UserBalance
				stmt := tl.DELETE().WHERE(
					tl.ClientID.EQ(postgres.UUID(tt.args.clientID)),
				)
				sql, args := stmt.Sql()
				_, err := db.conn.Exec(ctx, sql, args...)
				if err != nil {
					t.Fatal(err)
				}
			})

			err := tt.fields.prepare(ctx, db, &tt.args)
			require.NoError(t, err)

			var resBalance int64
			err = db.Transactional(ctx, func(tx pgx.Tx) error {
				updateErr := db.UpdateBalance(ctx, tx, tt.args.clientID, tt.args.updateBalance)
				if updateErr != nil {
					return updateErr
				}
				res, txErr := db.GetCurrentBalanceTx(ctx, tx, tt.args.clientID)
				resBalance = res
				return txErr
			})

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Equal(t, tt.expectedError, err.Error())
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expectedBalance, resBalance)
		})
	}
}
