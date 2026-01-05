package outbox

import (
	"context"
	"errors"
	"processing_core/generated/proc_core_db/public/model"
	"processing_core/internal/outbox/mocks"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/mock/gomock"
)

func TestKafkaOutboxPublisher_Run(t *testing.T) {
	t.Parallel()
	var (
		tx = &pgxpool.Tx{}
	)
	type args struct {
		ctx context.Context
	}
	type fields struct {
		db       func(ctrl *gomock.Controller, a *args) KafkaDb
		commonDb func(ctrl *gomock.Controller, a *args) CommonDb
		producer func(ctrl *gomock.Controller, a *args) KafkaProducer
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				db: func(ctrl *gomock.Controller, _ *args) KafkaDb {
					m := mocks.NewMockKafkaDb(ctrl)
					m.EXPECT().GetUnpublishedMessages(gomock.Any(), gomock.Any()).
						Return([]model.Outbox{
							{
								ID: 1,
							},
							{
								ID: 2,
							},
						}, nil)
					m.EXPECT().
						MarkMessagesAsProcessed(gomock.Any(), gomock.Any(), []int64{1, 2}).
						Return(nil)
					return m
				},
				commonDb: func(ctrl *gomock.Controller, a *args) CommonDb {
					m := mocks.NewMockCommonDb(ctrl)
					m.EXPECT().
						Transactional(gomock.Any(), gomock.Any()).
						DoAndReturn(func(_ context.Context, fn func(tx pgx.Tx) error) error {
							return fn(tx)
						}).AnyTimes()
					return m
				},
				producer: func(ctrl *gomock.Controller, a *args) KafkaProducer {
					m := mocks.NewMockKafkaProducer(ctrl)
					m.EXPECT().Produce(gomock.Any(), []model.Outbox{
						{
							ID: 1,
						},
						{
							ID: 2,
						},
					}).Return([]int64{1, 2}, nil)
					return m
				},
			},
			args: args{
				ctx: context.Background(),
			},
			wantErr: true,
		},
		{
			name: "mark messages as processed error",
			fields: fields{
				db: func(ctrl *gomock.Controller, _ *args) KafkaDb {
					m := mocks.NewMockKafkaDb(ctrl)
					m.EXPECT().GetUnpublishedMessages(gomock.Any(), gomock.Any()).
						Return([]model.Outbox{
							{
								ID: 1,
							},
							{
								ID: 2,
							},
						}, nil)
					m.EXPECT().
						MarkMessagesAsProcessed(gomock.Any(), gomock.Any(), []int64{1, 2}).
						Return(errors.New("someErr"))
					return m
				},
				commonDb: func(ctrl *gomock.Controller, a *args) CommonDb {
					m := mocks.NewMockCommonDb(ctrl)
					m.EXPECT().
						Transactional(gomock.Any(), gomock.Any()).
						DoAndReturn(func(_ context.Context, fn func(tx pgx.Tx) error) error {
							return fn(tx)
						}).Times(1)
					m.EXPECT().
						Transactional(gomock.Any(), gomock.Any()).
						DoAndReturn(func(_ context.Context, fn func(tx pgx.Tx) error) error {
							return fn(tx)
						}).Times(1)
					return m
				},
				producer: func(ctrl *gomock.Controller, a *args) KafkaProducer {
					m := mocks.NewMockKafkaProducer(ctrl)
					m.EXPECT().Produce(gomock.Any(), []model.Outbox{
						{
							ID: 1,
						},
						{
							ID: 2,
						},
					}).Return([]int64{1, 2}, nil)
					return m
				},
			},
			args: args{
				ctx: context.Background(),
			},
			wantErr: true,
		},
		{
			name: "transactional error 1",
			fields: fields{
				db: func(ctrl *gomock.Controller, _ *args) KafkaDb {
					m := mocks.NewMockKafkaDb(ctrl)
					m.EXPECT().GetUnpublishedMessages(gomock.Any(), gomock.Any()).
						Return([]model.Outbox{
							{
								ID: 1,
							},
							{
								ID: 2,
							},
						}, nil)
					return m
				},
				commonDb: func(ctrl *gomock.Controller, a *args) CommonDb {
					m := mocks.NewMockCommonDb(ctrl)
					m.EXPECT().
						Transactional(gomock.Any(), gomock.Any()).
						DoAndReturn(func(_ context.Context, fn func(tx pgx.Tx) error) error {
							return fn(tx)
						}).Times(1)
					m.EXPECT().
						Transactional(gomock.Any(), gomock.Any()).
						DoAndReturn(func(_ context.Context, _ func(tx pgx.Tx) error) error {
							return errors.New("someErr")
						}).Times(1)
					return m
				},
				producer: func(ctrl *gomock.Controller, a *args) KafkaProducer {
					m := mocks.NewMockKafkaProducer(ctrl)
					m.EXPECT().Produce(gomock.Any(), []model.Outbox{
						{
							ID: 1,
						},
						{
							ID: 2,
						},
					}).Return([]int64{1, 2}, nil)
					return m
				},
			},
			args: args{
				ctx: context.Background(),
			},
			wantErr: true,
		},
		{
			name: "produce error",
			fields: fields{
				db: func(ctrl *gomock.Controller, _ *args) KafkaDb {
					m := mocks.NewMockKafkaDb(ctrl)
					m.EXPECT().GetUnpublishedMessages(gomock.Any(), gomock.Any()).
						Return([]model.Outbox{
							{
								ID: 1,
							},
							{
								ID: 2,
							},
						}, nil)
					return m
				},
				commonDb: func(ctrl *gomock.Controller, a *args) CommonDb {
					m := mocks.NewMockCommonDb(ctrl)
					m.EXPECT().
						Transactional(gomock.Any(), gomock.Any()).
						DoAndReturn(func(_ context.Context, fn func(tx pgx.Tx) error) error {
							return fn(tx)
						}).Times(1)
					return m
				},
				producer: func(ctrl *gomock.Controller, a *args) KafkaProducer {
					m := mocks.NewMockKafkaProducer(ctrl)
					m.EXPECT().Produce(gomock.Any(), []model.Outbox{
						{
							ID: 1,
						},
						{
							ID: 2,
						},
					}).Return([]int64{1, 2}, errors.New("someErr"))
					return m
				},
			},
			args: args{
				ctx: context.Background(),
			},
			wantErr: true,
		},
		{
			name: "zero msgs",
			fields: fields{
				db: func(ctrl *gomock.Controller, _ *args) KafkaDb {
					m := mocks.NewMockKafkaDb(ctrl)
					m.EXPECT().GetUnpublishedMessages(gomock.Any(), gomock.Any()).
						Return([]model.Outbox{}, nil)
					return m
				},
				commonDb: func(ctrl *gomock.Controller, a *args) CommonDb {
					m := mocks.NewMockCommonDb(ctrl)
					m.EXPECT().
						Transactional(gomock.Any(), gomock.Any()).
						DoAndReturn(func(_ context.Context, fn func(tx pgx.Tx) error) error {
							return fn(tx)
						}).Times(1)
					return m
				},
				producer: func(ctrl *gomock.Controller, a *args) KafkaProducer {
					m := mocks.NewMockKafkaProducer(ctrl)
					return m
				},
			},
			args: args{
				ctx: context.Background(),
			},
			wantErr: true,
		},
		{
			name: "zero msgs",
			fields: fields{
				db: func(ctrl *gomock.Controller, _ *args) KafkaDb {
					m := mocks.NewMockKafkaDb(ctrl)
					m.EXPECT().GetUnpublishedMessages(gomock.Any(), gomock.Any()).
						Return([]model.Outbox{}, nil)
					return m
				},
				commonDb: func(ctrl *gomock.Controller, a *args) CommonDb {
					m := mocks.NewMockCommonDb(ctrl)
					m.EXPECT().
						Transactional(gomock.Any(), gomock.Any()).
						DoAndReturn(func(_ context.Context, fn func(tx pgx.Tx) error) error {
							return fn(tx)
						}).Times(1)
					return m
				},
				producer: func(ctrl *gomock.Controller, a *args) KafkaProducer {
					m := mocks.NewMockKafkaProducer(ctrl)
					return m
				},
			},
			args: args{
				ctx: context.Background(),
			},
			wantErr: true,
		},
		{
			name: "get msgs error",
			fields: fields{
				db: func(ctrl *gomock.Controller, _ *args) KafkaDb {
					m := mocks.NewMockKafkaDb(ctrl)
					m.EXPECT().GetUnpublishedMessages(gomock.Any(), gomock.Any()).
						Return([]model.Outbox{}, errors.New("someErr"))
					return m
				},
				commonDb: func(ctrl *gomock.Controller, a *args) CommonDb {
					m := mocks.NewMockCommonDb(ctrl)
					m.EXPECT().
						Transactional(gomock.Any(), gomock.Any()).
						DoAndReturn(func(_ context.Context, fn func(tx pgx.Tx) error) error {
							return fn(tx)
						}).Times(1)
					return m
				},
				producer: func(ctrl *gomock.Controller, a *args) KafkaProducer {
					m := mocks.NewMockKafkaProducer(ctrl)
					return m
				},
			},
			args: args{
				ctx: context.Background(),
			},
			wantErr: true,
		},
		{
			name: "transactional error 2",
			fields: fields{
				db: func(ctrl *gomock.Controller, _ *args) KafkaDb {
					m := mocks.NewMockKafkaDb(ctrl)
					return m
				},
				commonDb: func(ctrl *gomock.Controller, a *args) CommonDb {
					m := mocks.NewMockCommonDb(ctrl)
					m.EXPECT().
						Transactional(gomock.Any(), gomock.Any()).
						DoAndReturn(func(_ context.Context, _ func(tx pgx.Tx) error) error {
							return errors.New("someErr")
						}).Times(1)
					return m
				},
				producer: func(ctrl *gomock.Controller, a *args) KafkaProducer {
					m := mocks.NewMockKafkaProducer(ctrl)
					return m
				},
			},
			args: args{
				ctx: context.Background(),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			ctx, cancel := context.WithTimeout(tt.args.ctx, 995*time.Millisecond)
			p := &KafkaOutboxPublisher{
				db:       tt.fields.db(ctrl, &tt.args),
				commonDb: tt.fields.commonDb(ctrl, &tt.args),
				producer: tt.fields.producer(ctrl, &tt.args),
			}
			if err := p.Run(ctx); (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
			}
			cancel()
		})
	}
}
