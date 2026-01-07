package outgoing

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	"antifraud/internal/app/model"
	"antifraud/internal/constants"
	"antifraud/internal/outgoing/mocks"
)

func Test_manyPartners_ShouldRun(t *testing.T) {
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
			want: false,
		},
		{
			name: "is off",
			fields: fields{
				isOn:      false,
				amountThs: 10_00,
			},
			args: args{
				transaction: &model.DomainTransaction{
					Amount: 11_00,
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &manyPartners{
				isOn:      tt.fields.isOn,
				amountThs: tt.fields.amountThs,
			}
			if got := l.ShouldRun(context.Background(), tt.args.transaction); got != tt.want {
				t.Errorf("ShouldRun() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_manyPartners_Check(t *testing.T) {
	type args struct {
		ctx         context.Context
		transaction *model.DomainTransaction
	}
	type fields struct {
		partners int
		mode     int
		fs       func(ctrl *gomock.Controller, a *args) FeatureStore
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "success internal mode",
			fields: fields{
				partners: 2,
				mode:     0,
				fs: func(ctrl *gomock.Controller, a *args) FeatureStore {
					m := mocks.NewMockFeatureStore(ctrl)
					m.EXPECT().GetFeatureString(
						gomock.Any(), constants.FeatureInternalPartners30M, gomock.Any(),
					).Return([]string{"p1", "p2", "p2"}, nil)
					return m
				},
			},
			args: args{
				ctx: context.Background(),
				transaction: &model.DomainTransaction{
					SenderID: uuid.New(),
				},
			},
			wantErr: true,
		},
		{
			name: "not enough partners sbp mode",
			fields: fields{
				partners: 3,
				mode:     1,
				fs: func(ctrl *gomock.Controller, a *args) FeatureStore {
					m := mocks.NewMockFeatureStore(ctrl)
					m.EXPECT().GetFeatureString(
						gomock.Any(), constants.FeatureSbpPartners30M, gomock.Any(),
					).Return([]string{"p1", "p2"}, nil)
					return m
				},
			},
			args: args{
				ctx: context.Background(),
				transaction: &model.DomainTransaction{
					SenderID: uuid.New(),
				},
			},
			wantErr: false, // уникальных партнеров 2 < 3 => ok
		},
		{
			name: "nil feature",
			fields: fields{
				partners: 3,
				mode:     0,
				fs: func(ctrl *gomock.Controller, a *args) FeatureStore {
					m := mocks.NewMockFeatureStore(ctrl)
					m.EXPECT().GetFeatureString(
						gomock.Any(), constants.FeatureInternalPartners30M, gomock.Any(),
					).Return(nil, nil)
					return m
				},
			},
			args: args{
				ctx: context.Background(),
				transaction: &model.DomainTransaction{
					SenderID: uuid.New(),
				},
			},
			wantErr: false,
		},
		{
			name: "feature store error",
			fields: fields{
				partners: 3,
				mode:     1,
				fs: func(ctrl *gomock.Controller, a *args) FeatureStore {
					m := mocks.NewMockFeatureStore(ctrl)
					m.EXPECT().GetFeatureString(
						gomock.Any(), constants.FeatureSbpPartners30M, gomock.Any(),
					).Return(nil, errors.New("some error"))
					return m
				},
			},
			args: args{
				ctx: context.Background(),
				transaction: &model.DomainTransaction{
					SenderID: uuid.New(),
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			l := &manyPartners{
				partners: tt.fields.partners,
				mode:     tt.fields.mode,
				fs:       tt.fields.fs(ctrl, &tt.args),
			}

			if err := l.Check(tt.args.ctx, tt.args.transaction); (err != nil) != tt.wantErr {
				t.Errorf("Check() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
