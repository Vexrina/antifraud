package repository

import (
	"context"
	"testing"
	"time"

	"github.com/go-jet/jet/v2/postgres"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"processing_core/generated/proc_core_db/public/model"
	"processing_core/generated/proc_core_db/public/table"
	"processing_core/internal/utils"
)

func TestPgDb_AppendTransaction(t *testing.T) {
	type args struct {
		transaction model.TransactionsHistory
	}
	tests := []struct {
		name          string
		args          args
		expectedError string
		prepare       func(t *testing.T, db *pgDb, a *args)
	}{
		{
			name: "success - append new transaction",
			args: args{
				transaction: model.TransactionsHistory{
					ID:              42,
					TransactionID:   lo.ToPtr(uuid.New()),
					CreatedAt:       lo.ToPtr(time.Now()),
					Amount:          lo.ToPtr(int64(1000)),
					Currency:        lo.ToPtr(int32(643)),
					Merchant:        lo.ToPtr("test_merchant"),
					Country:         lo.ToPtr("RU"),
					SenderID:        lo.ToPtr(uuid.New()),
					ReceiverID:      lo.ToPtr(uuid.New()),
					TransactionType: lo.ToPtr(model.TransactionType_CashIn),
				},
			},
			expectedError: "",
			prepare:       func(t *testing.T, db *pgDb, a *args) {},
		},
		{
			name: "success - append transaction with all fields",
			args: args{
				transaction: model.TransactionsHistory{
					ID:              45,
					TransactionID:   lo.ToPtr(uuid.New()),
					CreatedAt:       lo.ToPtr(time.Now()),
					Amount:          lo.ToPtr(int64(5000)),
					Currency:        lo.ToPtr(int32(840)),
					Merchant:        lo.ToPtr("merchant_2"),
					Country:         lo.ToPtr("US"),
					SenderID:        lo.ToPtr(uuid.New()),
					ReceiverID:      lo.ToPtr(uuid.New()),
					ReceiverBic:     lo.ToPtr("BIC123"),
					AtmID:           lo.ToPtr(uuid.New()),
					TransactionType: lo.ToPtr(model.TransactionType_CashOut),
				},
			},
			expectedError: "",

			prepare: func(t *testing.T, db *pgDb, a *args) {},
		},
		{
			name: "error - duplicate transaction_id",
			args: args{
				transaction: model.TransactionsHistory{
					ID:              42,
					TransactionID:   lo.ToPtr(uuid.New()),
					CreatedAt:       lo.ToPtr(time.Now()),
					Amount:          lo.ToPtr(int64(1000)),
					Currency:        lo.ToPtr(int32(643)),
					Merchant:        lo.ToPtr("test_merchant"),
					Country:         lo.ToPtr("RU"),
					SenderID:        lo.ToPtr(uuid.New()),
					ReceiverID:      lo.ToPtr(uuid.New()),
					TransactionType: lo.ToPtr(model.TransactionType_CashIn),
				},
			},
			expectedError: "can't append transaction",

			prepare: func(t *testing.T, db *pgDb, a *args) {
				err := db.Transactional(context.Background(), func(tx pgx.Tx) error {
					return db.AppendTransaction(context.Background(), tx, a.transaction)
				})
				if err != nil {
					t.Fatal(err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			db := NewTestDB(ctx, t)

			t.Cleanup(func() {
				tl := table.TransactionsHistory
				stmt := tl.DELETE().WHERE(
					tl.TransactionID.EQ(postgres.UUID(tt.args.transaction.TransactionID)),
				)
				sql, args := stmt.Sql()
				_, err := db.conn.Exec(ctx, sql, args...)
				if err != nil {
					t.Fatal(err)
				}
			})

			tt.prepare(t, db, &tt.args)
			err := db.Transactional(ctx, func(tx pgx.Tx) error {
				return db.AppendTransaction(ctx, tx, tt.args.transaction)
			})

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestPgDb_UpsertTransaction(t *testing.T) {
	t.Parallel()

	type args struct {
		transaction model.TransactionsHistory
	}
	type fields struct {
		prepare func(ctx context.Context, db *pgDb, a *args) error
	}
	tests := []struct {
		name          string
		fields        fields
		args          args
		expectedError string
		shouldCommit  bool
	}{
		{
			name: "success - insert new transaction",
			fields: fields{
				prepare: func(ctx context.Context, db *pgDb, a *args) error {
					return nil
				},
			},
			args: args{
				transaction: model.TransactionsHistory{
					ID:              42,
					TransactionID:   lo.ToPtr(uuid.New()),
					CreatedAt:       lo.ToPtr(time.Now()),
					Amount:          lo.ToPtr(int64(1000)),
					Currency:        lo.ToPtr(int32(643)),
					Merchant:        lo.ToPtr("test_merchant"),
					Country:         lo.ToPtr("RU"),
					SenderID:        lo.ToPtr(uuid.New()),
					ReceiverID:      lo.ToPtr(uuid.New()),
					TransactionType: lo.ToPtr(model.TransactionType_CashIn),
				},
			},
			expectedError: "",
			shouldCommit:  true,
		},
		{
			name: "success - update existing transaction",
			fields: fields{
				prepare: func(ctx context.Context, db *pgDb, a *args) error {
					return db.Transactional(ctx, func(tx pgx.Tx) error {
						return db.AppendTransaction(ctx, tx, a.transaction)
					})
				},
			},
			args: args{
				transaction: model.TransactionsHistory{
					ID:              42,
					TransactionID:   lo.ToPtr(uuid.New()),
					CreatedAt:       lo.ToPtr(time.Now()),
					Amount:          lo.ToPtr(int64(1000)),
					Currency:        lo.ToPtr(int32(643)),
					Merchant:        lo.ToPtr("test_merchant"),
					Country:         lo.ToPtr("RU"),
					SenderID:        lo.ToPtr(uuid.New()),
					ReceiverID:      lo.ToPtr(uuid.New()),
					TransactionType: lo.ToPtr(model.TransactionType_CashIn),
				},
			},
			expectedError: "",
			shouldCommit:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			db := NewTestDB(ctx, t)

			t.Cleanup(func() {
				tl := table.TransactionsHistory
				stmt := tl.DELETE().WHERE(
					tl.TransactionID.EQ(postgres.UUID(tt.args.transaction.TransactionID)),
				)
				sql, args := stmt.Sql()
				_, err := db.conn.Exec(ctx, sql, args...)
				if err != nil {
					t.Fatal(err)
				}
			})

			err := tt.fields.prepare(ctx, db, &tt.args)
			require.NoError(t, err)

			// Обновляем транзакцию для теста обновления
			if tt.name == "success - update existing transaction" {
				tt.args.transaction.Amount = lo.ToPtr(int64(2000))
				tt.args.transaction.Merchant = lo.ToPtr("updated_merchant")
			}

			err = db.Transactional(ctx, func(tx pgx.Tx) error {
				return db.UpsertTransaction(ctx, tx, tt.args.transaction)
			})

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				return
			}
			require.NoError(t, err)

			// Проверяем, что транзакция была сохранена/обновлена
			var foundTransactions []model.TransactionsHistory
			err = db.Transactional(ctx, func(tx pgx.Tx) error {
				var txErr error
				foundTransactions, txErr = db.GetTransactions(ctx, tx, TransactionFilter{
					TransactionID: tt.args.transaction.TransactionID,
				})
				return txErr
			})
			require.NoError(t, err)
			require.Len(t, foundTransactions, 1)
			assert.Equal(t, *tt.args.transaction.Amount, *foundTransactions[0].Amount)
		})
	}
}

func TestPgDb_GetTransactions(t *testing.T) {
	// Подготовка данных
	senderID1 := uuid.New()
	senderID2 := uuid.New()
	receiverID1 := uuid.New()
	receiverID2 := uuid.New()
	merchant1 := "merchant_1"
	merchant2 := "merchant_2"
	country1 := "RU"
	country2 := "US"
	now := time.Now()

	ctx := context.Background()
	db := NewTestDB(ctx, t)
	transactions := []model.TransactionsHistory{
		{
			ID:              42,
			TransactionID:   lo.ToPtr(uuid.New()),
			CreatedAt:       lo.ToPtr(now.Add(-2 * time.Hour)),
			Amount:          lo.ToPtr(int64(1000)),
			Currency:        lo.ToPtr(int32(643)),
			Merchant:        &merchant1,
			Country:         &country1,
			SenderID:        &senderID1,
			ReceiverID:      &receiverID1,
			TransactionType: lo.ToPtr(model.TransactionType_CashIn),
		},
		{
			ID:              42,
			TransactionID:   lo.ToPtr(uuid.New()),
			CreatedAt:       lo.ToPtr(now.Add(-1 * time.Hour)),
			Amount:          lo.ToPtr(int64(2000)),
			Currency:        lo.ToPtr(int32(643)),
			Merchant:        &merchant2,
			Country:         &country1,
			SenderID:        &senderID1,
			ReceiverID:      &receiverID2,
			TransactionType: lo.ToPtr(model.TransactionType_CashOut),
		},
		{
			ID:              42,
			TransactionID:   lo.ToPtr(uuid.New()),
			CreatedAt:       lo.ToPtr(now),
			Amount:          lo.ToPtr(int64(3000)),
			Currency:        lo.ToPtr(int32(840)),
			Merchant:        &merchant1,
			Country:         &country2,
			SenderID:        &senderID2,
			ReceiverID:      &receiverID1,
			TransactionType: lo.ToPtr(model.TransactionType_Internal),
		},
		{
			ID:              42,
			TransactionID:   lo.ToPtr(uuid.New()),
			CreatedAt:       lo.ToPtr(now.Add(1 * time.Hour)),
			Amount:          lo.ToPtr(int64(5000)),
			Currency:        lo.ToPtr(int32(840)),
			Merchant:        &merchant2,
			Country:         &country2,
			SenderID:        &senderID2,
			ReceiverID:      &receiverID2,
			TransactionType: lo.ToPtr(model.TransactionType_SbpOutgoing),
		},
	}

	// Добавляем все транзакции
	err := db.Transactional(ctx, func(tx pgx.Tx) error {
		for _, tr := range transactions {
			if err := db.AppendTransaction(ctx, tx, tr); err != nil {
				return err
			}
		}
		return nil
	})
	require.NoError(t, err)

	tests := []struct {
		name           string
		filter         TransactionFilter
		expectedCount  int
		expectedError  string
		validateResult func(t *testing.T, results []model.TransactionsHistory)
	}{
		{
			name:          "success - get all transactions",
			filter:        TransactionFilter{},
			expectedCount: 4,
			expectedError: "",
		},
		{
			name: "success - filter by sender_id",
			filter: TransactionFilter{
				SenderID: &senderID1,
			},
			expectedCount: 2,
			expectedError: "",
			validateResult: func(t *testing.T, results []model.TransactionsHistory) {
				for _, tr := range results {
					assert.Equal(t, senderID1, *tr.SenderID)
				}
			},
		},
		{
			name: "success - filter by receiver_id",
			filter: TransactionFilter{
				ReceiverID: &receiverID1,
			},
			expectedCount: 2,
			expectedError: "",
			validateResult: func(t *testing.T, results []model.TransactionsHistory) {
				for _, tr := range results {
					assert.Equal(t, receiverID1, *tr.ReceiverID)
				}
			},
		},
		{
			name: "success - filter by transaction_type",
			filter: TransactionFilter{
				TransactionType: lo.ToPtr(model.TransactionType_CashIn),
			},
			expectedCount: 1,
			expectedError: "",
			validateResult: func(t *testing.T, results []model.TransactionsHistory) {
				assert.Equal(t, model.TransactionType_CashIn, *results[0].TransactionType)
			},
		},
		{
			name: "success - filter by merchant",
			filter: TransactionFilter{
				Merchant: &merchant1,
			},
			expectedCount: 2,
			expectedError: "",
			validateResult: func(t *testing.T, results []model.TransactionsHistory) {
				for _, tr := range results {
					assert.Equal(t, merchant1, *tr.Merchant)
				}
			},
		},
		{
			name: "success - filter by country",
			filter: TransactionFilter{
				Country: &country1,
			},
			expectedCount: 2,
			expectedError: "",
			validateResult: func(t *testing.T, results []model.TransactionsHistory) {
				for _, tr := range results {
					assert.Equal(t, country1, *tr.Country)
				}
			},
		},
		{
			name: "success - filter by amount range",
			filter: TransactionFilter{
				AmountMin: lo.ToPtr(int64(2000)),
				AmountMax: lo.ToPtr(int64(4000)),
			},
			expectedCount: 2,
			expectedError: "",
			validateResult: func(t *testing.T, results []model.TransactionsHistory) {
				for _, tr := range results {
					assert.GreaterOrEqual(t, *tr.Amount, int64(2000))
					assert.LessOrEqual(t, *tr.Amount, int64(4000))
				}
			},
		},
		{
			name: "success - filter by created_at range",
			filter: TransactionFilter{
				CreatedAtFrom: lo.ToPtr(now.Add(-90 * time.Minute)),
				CreatedAtTo:   lo.ToPtr(now.Add(30 * time.Minute)),
			},
			expectedCount: 2,
			expectedError: "",
		},
		{
			name: "success - filter by transaction_id",
			filter: TransactionFilter{
				TransactionID: transactions[0].TransactionID,
			},
			expectedCount: 1,
			expectedError: "",
			validateResult: func(t *testing.T, results []model.TransactionsHistory) {
				assert.Equal(t, *transactions[0].TransactionID, *results[0].TransactionID)
			},
		},
		{
			name: "success - filter by id",
			filter: TransactionFilter{
				ID: lo.ToPtr(int64(transactions[0].ID)),
			},
			expectedCount: 1,
			expectedError: "",
			validateResult: func(t *testing.T, results []model.TransactionsHistory) {
				assert.Equal(t, transactions[0].ID, results[0].ID)
			},
		},
		{
			name: "success - composite filter",
			filter: TransactionFilter{
				SenderID:        &senderID1,
				Country:         &country1,
				TransactionType: lo.ToPtr(model.TransactionType_CashOut),
			},
			expectedCount: 1,
			expectedError: "",
			validateResult: func(t *testing.T, results []model.TransactionsHistory) {
				assert.Equal(t, senderID1, *results[0].SenderID)
				assert.Equal(t, country1, *results[0].Country)
				assert.Equal(t, model.TransactionType_CashOut, *results[0].TransactionType)
			},
		},
		{
			name: "success - filter with limit",
			filter: TransactionFilter{
				Limit: lo.ToPtr(2),
			},
			expectedCount: 2,
			expectedError: "",
		},
		{
			name: "success - filter with limit and offset",
			filter: TransactionFilter{
				Limit:  lo.ToPtr(2),
				Offset: lo.ToPtr(1),
			},
			expectedCount: 2,
			expectedError: "",
		},
		{
			name: "success - no results",
			filter: TransactionFilter{
				SenderID: lo.ToPtr(uuid.New()),
			},
			expectedCount: 0,
			expectedError: "",
		},
	}
	t.Cleanup(func() {
		tl := table.TransactionsHistory
		var transactionsIDs []uuid.UUID
		for _, transaction := range transactions {
			transactionsIDs = append(transactionsIDs, *transaction.TransactionID)
		}
		stmt := tl.DELETE().WHERE(
			tl.TransactionID.IN(utils.UUIDArray(transactionsIDs)...),
		)
		sql, args := stmt.Sql()
		_, err = db.conn.Exec(ctx, sql, args...)
		if err != nil {
			t.Fatal(err)
		}
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var results []model.TransactionsHistory
			err := db.Transactional(ctx, func(tx pgx.Tx) error {
				var txErr error
				results, txErr = db.GetTransactions(ctx, tx, tt.filter)
				return txErr
			})

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				return
			}
			require.NoError(t, err)
			assert.Len(t, results, tt.expectedCount)

			if tt.validateResult != nil {
				tt.validateResult(t, results)
			}
		})
	}
}
