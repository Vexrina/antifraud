package repository

import (
	"context"
	"testing"

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
		shouldCommit    bool
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
			shouldCommit:    true,
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
			shouldCommit:    true,
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
			shouldCommit:    true,
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
			shouldCommit:    true,
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
			shouldCommit:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			db := NewTestDB(ctx, t)
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

func TestPgDb_UpdateBalance_MultipleClients(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		clients []struct {
			clientID uuid.UUID
			balance  int64
		}
		shouldCommit bool
	}{
		{
			name: "two clients with different balances",
			clients: []struct {
				clientID uuid.UUID
				balance  int64
			}{
				{uuid.New(), 1000},
				{uuid.New(), 2000},
			},
			shouldCommit: true,
		},
		{
			name: "three clients with same balance",
			clients: []struct {
				clientID uuid.UUID
				balance  int64
			}{
				{uuid.New(), 500},
				{uuid.New(), 500},
				{uuid.New(), 500},
			},
			shouldCommit: true,
		},
		{
			name: "multiple clients with various balances",
			clients: []struct {
				clientID uuid.UUID
				balance  int64
			}{
				{uuid.New(), 0},
				{uuid.New(), -100},
				{uuid.New(), 10000},
			},
			shouldCommit: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			db := NewTestDB(ctx, t)

			tx, err := db.BeginTx(ctx)
			if err != nil {
				t.Fatalf("can't begin transaction: %v", err)
			}
			defer db.RollBackUnlessCommitted(ctx, tx)

			// создаем балансы для всех клиентов
			for _, client := range tt.clients {
				err = db.UpdateBalance(ctx, tx, client.clientID, client.balance)
				if err != nil {
					t.Fatalf("can't create balance for client %v: %v", client.clientID, err)
				}
			}

			// проверяем балансы всех клиентов
			for _, client := range tt.clients {
				gotBalance, err := db.GetCurrentBalanceTx(ctx, tx, client.clientID)
				if err != nil {
					t.Fatalf("can't get balance for client %v: %v", client.clientID, err)
				}
				assert.Equal(t, client.balance, gotBalance)
			}

			if tt.shouldCommit {
				err = db.CommitTx(ctx, tx)
				if err != nil {
					t.Fatalf("can't commit transaction: %v", err)
				}
			} else {
				db.RollBackUnlessCommitted(ctx, tx)
			}
		})
	}
}
