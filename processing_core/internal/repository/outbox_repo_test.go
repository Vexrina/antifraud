package repository

import (
	"context"
	"processing_core/generated/proc_core_db/public/model"
	"processing_core/generated/proc_core_db/public/table"
	"processing_core/pkg/kafka_core"
	"testing"
	"time"

	"github.com/go-jet/jet/v2/postgres"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/timestamppb"
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

func Test_mapTransactionDbModelToProto(t *testing.T) {
	t.Parallel()
	type args struct {
		transaction model.TransactionsHistory
	}

	TransactionID := uuid.New()
	CreatedAt := time.Now()
	SenderID := uuid.New()
	ReceiverID := uuid.New()
	AtmID := uuid.New()

	tests := []struct {
		name string
		args args
		want *kafka_core.TransactionCore
	}{
		{
			name: "all nillable fields is nil",
			args: args{
				transaction: model.TransactionsHistory{
					ID:              0,
					TransactionID:   lo.ToPtr(TransactionID),
					CreatedAt:       lo.ToPtr(CreatedAt),
					Amount:          lo.ToPtr(int64(100_00)),
					Currency:        lo.ToPtr(int32(23)),
					Merchant:        nil,
					Country:         nil,
					SenderID:        lo.ToPtr(SenderID),
					ReceiverID:      nil,
					ReceiverBic:     nil,
					AtmID:           nil,
					TransactionType: lo.ToPtr(model.TransactionType_Internal),
					Revision:        0,
				},
			},
			want: &kafka_core.TransactionCore{
				Id:              0,
				TransactionId:   TransactionID.String(),
				CreatedAt:       timestamppb.New(CreatedAt),
				Amount:          int64(100_00),
				Currency:        "23",
				Merchant:        "",
				Country:         "",
				SenderId:        SenderID.String(),
				ReceiverId:      lo.ToPtr(""),
				ReceiverBic:     lo.ToPtr(""),
				AtmId:           lo.ToPtr(""),
				TransactionType: kafka_core.TransactionType_Internal,
				Revision:        0,
			},
		},
		{
			name: "all nillable fields is NOT nil",
			args: args{
				transaction: model.TransactionsHistory{
					ID:              0,
					TransactionID:   lo.ToPtr(TransactionID),
					CreatedAt:       lo.ToPtr(CreatedAt),
					Amount:          lo.ToPtr(int64(100_00)),
					Currency:        lo.ToPtr(int32(23)),
					Merchant:        lo.ToPtr("123"),
					Country:         lo.ToPtr("132"),
					SenderID:        lo.ToPtr(SenderID),
					ReceiverID:      lo.ToPtr(ReceiverID),
					ReceiverBic:     lo.ToPtr("213"),
					AtmID:           lo.ToPtr(AtmID),
					TransactionType: lo.ToPtr(model.TransactionType_Internal),
					Revision:        0,
				},
			},
			want: &kafka_core.TransactionCore{
				Id:              0,
				TransactionId:   TransactionID.String(),
				CreatedAt:       timestamppb.New(CreatedAt),
				Amount:          int64(100_00),
				Currency:        "23",
				Merchant:        "123",
				Country:         "132",
				SenderId:        SenderID.String(),
				ReceiverId:      lo.ToPtr(ReceiverID.String()),
				ReceiverBic:     lo.ToPtr("213"),
				AtmId:           lo.ToPtr(AtmID.String()),
				TransactionType: kafka_core.TransactionType_Internal,
				Revision:        0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := mapTransactionDbModelToProto(tt.args.transaction)
			if !cmp.Equal(tt.want, got, cmpopts.IgnoreUnexported(), protocmp.Transform()) {
				t.Errorf(
					"mapTransactionDbModelToProto() is not equal to wanted: %s",
					cmp.Diff(tt.want, got, cmpopts.IgnoreUnexported(), protocmp.Transform()),
				)
			}
		})
	}
}

func Test_mapDbTransactionTypeToProto(t *testing.T) {
	t.Parallel()
	type args struct {
		transactionType *model.TransactionType
	}
	tests := []struct {
		name string
		args args
		want kafka_core.TransactionType
	}{
		{
			name: "internal",
			args: args{
				transactionType: lo.ToPtr(model.TransactionType_Internal),
			},
			want: kafka_core.TransactionType_Internal,
		},
		{
			name: "CashIn",
			args: args{
				transactionType: lo.ToPtr(model.TransactionType_CashIn),
			},
			want: kafka_core.TransactionType_CashIn,
		},
		{
			name: "SbpOutgoing",
			args: args{
				transactionType: lo.ToPtr(model.TransactionType_SbpOutgoing),
			},
			want: kafka_core.TransactionType_SbpOutgoing,
		},
		{
			name: "CashOut",
			args: args{
				transactionType: lo.ToPtr(model.TransactionType_CashOut),
			},
			want: kafka_core.TransactionType_CashOut,
		},
		{
			name: "Unknown",
			args: args{
				transactionType: lo.ToPtr(model.TransactionType("unknown")),
			},
			want: kafka_core.TransactionType_Unknown,
		},
		{
			name: "nil",
			args: args{
				transactionType: nil,
			},
			want: kafka_core.TransactionType_Unknown,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equalf(t, tt.want, mapDbTransactionTypeToProto(tt.args.transactionType), "mapDbTransactionTypeToProto(%v)", tt.args.transactionType)
		})
	}
}
