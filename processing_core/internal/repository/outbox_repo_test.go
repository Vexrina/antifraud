package repository

import (
	"context"
	"encoding/json"
	"processing_core/generated/proc_core_db/public/model"
	"processing_core/generated/proc_core_db/public/table"
	"processing_core/internal/outbox"
	"processing_core/internal/utils"
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
					return nil
				},
				check: func(ctx context.Context, db *pgDb, a *args) {
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

func Test_mapTransactionTypeDbToProto(t *testing.T) {
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
			assert.Equalf(t, tt.want, mapTransactionTypeDbToProto(tt.args.transactionType), "mapTransactionTypeDbToProto(%v)", tt.args.transactionType)
		})
	}
}

func TestPgDb_GetUnpublishedMessages(t *testing.T) {
	ctx := context.Background()
	db := NewTestDB(ctx, t)
	now := time.Now()

	msgs := []model.Outbox{
		{
			ID:          1,
			AggregateID: uuid.New(),
			Published:   lo.ToPtr(false),
			EventType:   outbox.EventTypeTransaction.String(),
			CreatedAt:   lo.ToPtr(now.Add(-2 * time.Hour)),
			Payload:     `{"x":"y"}`,
		},
		{
			ID:          2,
			AggregateID: uuid.New(),
			Published:   lo.ToPtr(false),
			CreatedAt:   lo.ToPtr(now.Add(-1 * time.Hour)),
			EventType:   outbox.EventTypeTransaction.String(),
			Payload:     `{"x":"y"}`,
		},
		{
			ID:          3,
			AggregateID: uuid.New(),
			Published:   lo.ToPtr(true),
			CreatedAt:   lo.ToPtr(now),
			EventType:   outbox.EventTypeTransaction.String(),
			Payload:     `{"x":"y"}`,
		},
		{
			ID:          4,
			AggregateID: uuid.New(),
			Published:   lo.ToPtr(false),
			CreatedAt:   lo.ToPtr(now.Add(1 * time.Hour)),
			EventType:   outbox.EventTypeTransaction.String(),
			Payload:     `{"x":"y"}`,
		},
	}
	tests := []struct {
		name           string
		prepare        func(ctx context.Context, db *pgDb)
		expectedCount  int
		validateResult func(t *testing.T, results []model.Outbox)
	}{
		{
			name:          "success - get all unpublished messages",
			expectedCount: 3,
			prepare: func(ctx context.Context, db *pgDb) {
				err := db.Transactional(ctx, func(tx pgx.Tx) error {
					for _, msg := range msgs {
						payloadJSON, err := json.Marshal(msg.Payload)
						if err != nil {
							return err
						}

						_, err = tx.Exec(ctx,
							`INSERT INTO public.outbox (
                id,
                aggregate_id,
                published,
                created_at,
                event_type,
                payload
            ) VALUES ($1, $2, $3, $4, $5, $6)`,
							msg.ID,
							msg.AggregateID,
							*msg.Published,
							*msg.CreatedAt,
							msg.EventType,
							payloadJSON,
						)
						if err != nil {
							return err
						}
					}
					return nil
				})
				require.NoError(t, err)
			},
			validateResult: func(t *testing.T, results []model.Outbox) {
				for _, msg := range results {
					assert.False(t, *msg.Published)
				}
				assert.True(t, results[0].CreatedAt.Before(*results[1].CreatedAt) || results[0].CreatedAt.Equal(*results[1].CreatedAt))
				assert.True(t, results[1].CreatedAt.Before(*results[2].CreatedAt) || results[1].CreatedAt.Equal(*results[2].CreatedAt))
			},
		},
		{
			name:          "success - no unpublished messages",
			expectedCount: 0,
			prepare:       func(ctx context.Context, db *pgDb) {},
			validateResult: func(t *testing.T, results []model.Outbox) {
				assert.Empty(t, results)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Cleanup(func() {
				tl := table.Outbox
				var ids []int64
				for _, msg := range msgs {
					ids = append(ids, msg.ID)
				}
				stmt := tl.DELETE().WHERE(tl.ID.IN(utils.IntegerArray(ids)...))
				sql, args := stmt.Sql()
				_, err := db.conn.Exec(ctx, sql, args...)
				if err != nil {
					t.Fatal(err)
				}
			})
			tt.prepare(ctx, db)
			var results []model.Outbox
			err := db.Transactional(ctx, func(tx pgx.Tx) error {
				var txErr error
				results, txErr = db.GetUnpublishedMessages(ctx, tx)
				return txErr
			})

			require.NoError(t, err)
			assert.Len(t, results, tt.expectedCount)

			if tt.validateResult != nil {
				tt.validateResult(t, results)
			}
		})
	}
}

func TestPgDb_MarkMessagesAsProcessed(t *testing.T) {
	ctx := context.Background()
	db := NewTestDB(ctx, t)
	now := time.Now()

	msgs := []model.Outbox{
		{
			ID:          1,
			AggregateID: uuid.New(),
			Published:   lo.ToPtr(false),
			EventType:   outbox.EventTypeTransaction.String(),
			CreatedAt:   lo.ToPtr(now.Add(-2 * time.Hour)),
			Payload:     `{"x":"y"}`,
		},
		{
			ID:          2,
			AggregateID: uuid.New(),
			Published:   lo.ToPtr(false),
			CreatedAt:   lo.ToPtr(now.Add(-1 * time.Hour)),
			EventType:   outbox.EventTypeTransaction.String(),
			Payload:     `{"x":"y"}`,
		},
		{
			ID:          3,
			AggregateID: uuid.New(),
			Published:   lo.ToPtr(false),
			CreatedAt:   lo.ToPtr(now),
			EventType:   outbox.EventTypeTransaction.String(),
			Payload:     `{"x":"y"}`,
		},
	}

	tests := []struct {
		name          string
		idsToMark     []int64
		expectedState map[int64]bool
		prepare       func(ctx context.Context, db *pgDb)
	}{
		{
			name:      "success - mark 1 and 2 as processed",
			idsToMark: []int64{1, 2},
			expectedState: map[int64]bool{
				1: true,
				2: true,
				3: false,
			},
			prepare: func(ctx context.Context, db *pgDb) {
				err := db.Transactional(ctx, func(tx pgx.Tx) error {
					for _, msg := range msgs {
						payloadJSON, err := json.Marshal(msg.Payload)
						if err != nil {
							return err
						}
						_, err = tx.Exec(ctx,
							`INSERT INTO public.outbox (
								id, aggregate_id, published, created_at, event_type, payload
							) VALUES ($1,$2,$3,$4,$5,$6)`,
							msg.ID, msg.AggregateID, *msg.Published, *msg.CreatedAt, msg.EventType, payloadJSON,
						)
						if err != nil {
							return err
						}
					}
					return nil
				})
				require.NoError(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Cleanup(func() {
				tl := table.Outbox
				var ids []int64
				for _, msg := range msgs {
					ids = append(ids, msg.ID)
				}
				stmt := tl.DELETE().WHERE(tl.ID.IN(utils.IntegerArray(ids)...))
				sql, args := stmt.Sql()
				_, err := db.conn.Exec(ctx, sql, args...)
				if err != nil {
					t.Fatal(err)
				}
			})

			tt.prepare(ctx, db)

			err := db.Transactional(ctx, func(tx pgx.Tx) error {
				return db.MarkMessagesAsProcessed(ctx, tx, tt.idsToMark)
			})
			require.NoError(t, err)

			err = db.Transactional(ctx, func(tx pgx.Tx) error {
				tl := table.Outbox
				stmt := tl.SELECT(tl.ID, tl.Published).WHERE(tl.ID.IN(utils.IntegerArray([]int64{1, 2, 3})...))
				sql, args := stmt.Sql()
				rows, err := tx.Query(ctx, sql, args...)
				if err != nil {
					return err
				}
				results, err := pgx.CollectRows(rows, func(r pgx.CollectableRow) (struct {
					ID        int64
					Published bool
				}, error) {
					var id int64
					var published bool
					err := r.Scan(&id, &published)
					return struct {
						ID        int64
						Published bool
					}{id, published}, err
				})
				if err != nil {
					return err
				}

				for _, res := range results {
					expected, ok := tt.expectedState[res.ID]
					require.True(t, ok)
					assert.Equal(t, expected, res.Published)
				}
				return nil
			})
			require.NoError(t, err)
		})
	}
}
