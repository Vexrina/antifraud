package repository

import (
	"context"
	"processing_core/generated/proc_core_db/public/model"
	"processing_core/generated/proc_core_db/public/table"
	"processing_core/pkg/kafka_core"
	"testing"
	"time"

	"github.com/go-jet/jet/v2/postgres"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestPgDb_AppendOutbox(t *testing.T) {
	t.Parallel()

	type args struct {
		transaction model.TransactionsHistory
	}
	type fields struct {
		prepare func(ctx context.Context, db *pgDb, a *args) error
		check   func(ctx context.Context, db *pgDb, a *args)
	}
	tests := []struct {
		name          string
		fields        fields
		args          args
		expectedError string
	}{
		{
			name: "insert new transaction",
			fields: fields{
				prepare: func(ctx context.Context, db *pgDb, a *args) error {
					// ничего не создаем заранее
					return nil
				},
				check: func(ctx context.Context, db *pgDb, a *args) {
					// Проверяем, что запись реально вставилась
					var payload string
					var eventType string
					stmt := table.Outbox.SELECT(
						table.Outbox.Payload,
						table.Outbox.EventType,
					).WHERE(table.Outbox.AggregateID.EQ(postgres.UUID(a.transaction.TransactionID)))
					sql, args := stmt.Sql()

					rows := db.conn.QueryRow(ctx, sql, args...)
					err := rows.Scan(&payload, &eventType)

					require.NoError(t, err)

					// Декодируем protojson обратно и проверяем ключевые поля
					var decoded kafka_core.TransactionCore
					err = protojson.Unmarshal([]byte(payload), &decoded)
					require.NoError(t, err)
					assert.Equal(t, a.transaction.TransactionID.String(), decoded.TransactionId)
					assert.Equal(t, *a.transaction.Amount, decoded.Amount)
				},
			},
			args: args{
				transaction: model.TransactionsHistory{
					ID:              1,
					TransactionID:   lo.ToPtr(uuid.New()),
					CreatedAt:       lo.ToPtr(time.Now()),
					Amount:          lo.ToPtr(int64(1000)),
					Currency:        lo.ToPtr(int32(23)),
					Merchant:        lo.ToPtr("TestMerchant"),
					Country:         lo.ToPtr("US"),
					SenderID:        lo.ToPtr(uuid.New()),
					ReceiverID:      nil,
					ReceiverBic:     nil,
					AtmID:           nil,
					TransactionType: lo.ToPtr(model.TransactionType_CashIn),
					Revision:        1,
				},
			},
			expectedError: "",
		},
		{
			name: "insert transaction with nil fields",
			fields: fields{
				prepare: func(ctx context.Context, db *pgDb, a *args) error {
					return nil
				},
				check: func(ctx context.Context, db *pgDb, a *args) {
					// Проверяем, что запись реально вставилась
					var payload string
					var eventType string
					stmt := table.Outbox.SELECT(
						table.Outbox.Payload,
						table.Outbox.EventType,
					).WHERE(table.Outbox.AggregateID.EQ(postgres.UUID(a.transaction.TransactionID)))
					sql, args := stmt.Sql()

					rows := db.conn.QueryRow(ctx, sql, args...)
					err := rows.Scan(&payload, &eventType)

					require.NoError(t, err)

					// Декодируем protojson обратно и проверяем ключевые поля
					var decoded kafka_core.TransactionCore
					err = protojson.Unmarshal([]byte(payload), &decoded)
					require.NoError(t, err)
					assert.Equal(t, a.transaction.TransactionID.String(), decoded.TransactionId)
					assert.Equal(t, *a.transaction.Amount, decoded.Amount)
				},
			},
			args: args{
				transaction: model.TransactionsHistory{
					ID:              2,
					TransactionID:   lo.ToPtr(uuid.New()),
					CreatedAt:       lo.ToPtr(time.Now()),
					Amount:          lo.ToPtr(int64(500)),
					Currency:        lo.ToPtr(int32(23)),
					Merchant:        nil,
					Country:         nil,
					SenderID:        lo.ToPtr(uuid.New()),
					ReceiverID:      nil,
					ReceiverBic:     nil,
					AtmID:           nil,
					TransactionType: nil,
					Revision:        1,
				},
			},
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			db := NewTestDB(ctx, t)

			t.Cleanup(func() {
				tl := table.Outbox
				stmt := tl.DELETE().WHERE(
					tl.AggregateID.EQ(postgres.UUID(tt.args.transaction.TransactionID)),
				)
				sql, args := stmt.Sql()
				_, err := db.conn.Exec(ctx, sql, args...)
				if err != nil {
					t.Fatal(err)
				}
			})

			err := tt.fields.prepare(ctx, db, &tt.args)
			require.NoError(t, err)

			err = db.Transactional(ctx, func(tx pgx.Tx) error {
				return db.AppendOutbox(ctx, tx, tt.args.transaction)
			})

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Equal(t, tt.expectedError, err.Error())
				return
			}
			require.NoError(t, err)

			tt.fields.check(ctx, db, &tt.args)
		})
	}
}
