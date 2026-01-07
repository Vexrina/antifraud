package outgoing

import (
	"context"
	"errors"
	"testing"

	"github.com/gocql/gocql"
	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	"antifraud/internal/app/model"
	"antifraud/internal/constants"
	"antifraud/internal/outgoing/mocks"
)

func Test_bigSpent_Check(t *testing.T) {
	type args struct {
		ctx         context.Context
		transaction *model.DomainTransaction
	}
	type fields struct {
		largeThs int64
		fs       func(ctrl *gomock.Controller, a *args) FeatureStore
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
				largeThs: 1_000_00,
				fs: func(ctrl *gomock.Controller, a *args) FeatureStore {
					m := mocks.NewMockFeatureStore(ctrl)

					userUUID, _ := gocql.ParseUUID(a.transaction.SenderID.String())

					m.EXPECT().GetFeatureInteger(
						gomock.Any(), constants.FeatureSpent3H, constants.FeatureFilter{
							UserID: userUUID,
							Limit:  1,
						},
					).Return([]int64{1_000_00}, nil)
					return m
				},
			},
			args: args{
				ctx: context.Background(),
				transaction: &model.DomainTransaction{
					SenderID: uuid.New(),
					Amount:   500_00,
				},
			},
			wantErr: true,
		},
		{
			name: "not enough spents",
			fields: fields{
				largeThs: 1_000_00,
				fs: func(ctrl *gomock.Controller, a *args) FeatureStore {
					m := mocks.NewMockFeatureStore(ctrl)

					userUUID, _ := gocql.ParseUUID(a.transaction.SenderID.String())

					m.EXPECT().GetFeatureInteger(
						gomock.Any(), constants.FeatureSpent3H, constants.FeatureFilter{
							UserID: userUUID,
							Limit:  1,
						},
					).Return([]int64{0}, nil)
					return m
				},
			},
			args: args{
				ctx: context.Background(),
				transaction: &model.DomainTransaction{
					SenderID: uuid.New(),
					Amount:   500_00,
				},
			},
		},
		{
			name: "nil feature",
			fields: fields{
				largeThs: 1_000_00,
				fs: func(ctrl *gomock.Controller, a *args) FeatureStore {
					m := mocks.NewMockFeatureStore(ctrl)

					userUUID, _ := gocql.ParseUUID(a.transaction.SenderID.String())

					m.EXPECT().GetFeatureInteger(
						gomock.Any(), constants.FeatureSpent3H, constants.FeatureFilter{
							UserID: userUUID,
							Limit:  1,
						},
					).Return(nil, nil)
					return m
				},
			},
			args: args{
				ctx: context.Background(),
				transaction: &model.DomainTransaction{
					SenderID: uuid.New(),
					Amount:   500_00,
				},
			},
		},
		{
			name: "error feature store",
			fields: fields{
				largeThs: 1_000_00,
				fs: func(ctrl *gomock.Controller, a *args) FeatureStore {
					m := mocks.NewMockFeatureStore(ctrl)

					userUUID, _ := gocql.ParseUUID(a.transaction.SenderID.String())

					m.EXPECT().GetFeatureInteger(
						gomock.Any(), constants.FeatureSpent3H, constants.FeatureFilter{
							UserID: userUUID,
							Limit:  1,
						},
					).Return(nil, errors.New("some error"))
					return m
				},
			},
			args: args{
				ctx: context.Background(),
				transaction: &model.DomainTransaction{
					SenderID: uuid.New(),
					Amount:   500_00,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			l := &bigSpent{
				largeThs: tt.fields.largeThs,
				fs:       tt.fields.fs(ctrl, &tt.args),
			}

			if err := l.Check(tt.args.ctx, tt.args.transaction); (err != nil) != tt.wantErr {
				t.Errorf("Check() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_bigSpent_ShouldRun(t *testing.T) {
	type fields struct {
		isOn      bool
		amountThs int64
	}
	type args struct {
		transaction *model.DomainTransaction
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "run",
			fields: fields{
				isOn:      true,
				amountThs: 10_00,
			},
			args: args{
				transaction: &model.DomainTransaction{
					Amount: 11_00,
				},
			},
			want: true,
		},
		{
			name: "less money",
			fields: fields{
				isOn:      true,
				amountThs: 10_00,
			},
			args: args{
				transaction: &model.DomainTransaction{
					Amount: 1_00,
				},
			},
		},
		{
			name: "is off",
			fields: fields{
				isOn:      false,
				amountThs: 10_00,
			},
			args: args{
				transaction: &model.DomainTransaction{
					Amount: 1_00,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &bigSpent{
				isOn:      tt.fields.isOn,
				amountThs: tt.fields.amountThs,
			}
			if got := l.ShouldRun(context.Background(), tt.args.transaction); got != tt.want {
				t.Errorf("ShouldRun() = %v, want %v", got, tt.want)
			}
		})
	}
}
