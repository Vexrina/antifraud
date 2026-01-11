package helpers

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/types/known/timestamppb"

	"processing_core/internal/app/model"
	"processing_core/internal/helpers/mocks"
	"processing_core/pkg/antifraud"
)

func TestAfIntegration_CashOutCheck(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx       context.Context
		operation *model.CashOutDomainRequest
	}

	type fields struct {
		setupMocks func(af *mocks.MockOnlineCheckClient, args *args)
	}

	tests := []struct {
		name          string
		fields        fields
		args          args
		expectedError string
	}{
		{
			name: "success - approved",
			fields: fields{
				setupMocks: func(af *mocks.MockOnlineCheckClient, args *args) {
					af.
						EXPECT().
						CashOut(
							gomock.Any(),
							&antifraud.Transaction{
								Id:            args.operation.Transaction.ID,
								TransactionId: args.operation.Transaction.TransactionID.String(),
								CreatedAt:     timestamppb.New(*args.operation.Transaction.CreatedAt),
								Amount:        args.operation.Transaction.Amount,
								Currency:      fmt.Sprint(args.operation.Transaction.Currency),
								Merchant:      args.operation.Transaction.Merchant,
								Country:       args.operation.Transaction.Country,
								SenderId:      args.operation.Transaction.SenderID.String(),
								ReceiverId:    nil,
								Bic:           nil,
								AtmId:         lo.ToPtr(args.operation.AtmID.String()),
							},
						).
						Return(
							&antifraud.CheckResult{
								NewStatus: antifraud.OperationStatus_Approved,
							},
							nil,
						)
				},
			},
			args: args{
				ctx: context.Background(),
				operation: &model.CashOutDomainRequest{
					ID: uuid.New(),
					Transaction: &model.Transaction{
						ID:            1,
						TransactionID: uuid.New(),
						CreatedAt:     lo.ToPtr(time.Now()),
						Amount:        500,
						Currency:      0,
						Merchant:      "sad",
						Country:       "dsa",
						SenderID:      uuid.New(),
					},
					AtmID: uuid.New(),
				},
			},
			expectedError: "",
		},
		{
			name: "error during grpc - approved",
			fields: fields{
				setupMocks: func(af *mocks.MockOnlineCheckClient, args *args) {
					af.
						EXPECT().
						CashOut(
							gomock.Any(),
							&antifraud.Transaction{
								Id:            args.operation.Transaction.ID,
								TransactionId: args.operation.Transaction.TransactionID.String(),
								CreatedAt:     timestamppb.New(*args.operation.Transaction.CreatedAt),
								Amount:        args.operation.Transaction.Amount,
								Currency:      fmt.Sprint(args.operation.Transaction.Currency),
								Merchant:      args.operation.Transaction.Merchant,
								Country:       args.operation.Transaction.Country,
								SenderId:      args.operation.Transaction.SenderID.String(),
								ReceiverId:    nil,
								Bic:           nil,
								AtmId:         lo.ToPtr(args.operation.AtmID.String()),
							},
						).
						Return(
							&antifraud.CheckResult{
								NewStatus: antifraud.OperationStatus_Approved,
							},
							errors.New("someErr"),
						)
				},
			},
			args: args{
				ctx: context.Background(),
				operation: &model.CashOutDomainRequest{
					ID: uuid.New(),
					Transaction: &model.Transaction{
						ID:            1,
						TransactionID: uuid.New(),
						CreatedAt:     lo.ToPtr(time.Now()),
						Amount:        500,
						Currency:      0,
						Merchant:      "sad",
						Country:       "dsa",
						SenderID:      uuid.New(),
					},
					AtmID: uuid.New(),
				},
			},
			expectedError: "",
		},
		{
			name: "decline",
			fields: fields{
				setupMocks: func(af *mocks.MockOnlineCheckClient, args *args) {
					af.
						EXPECT().
						CashOut(
							gomock.Any(),
							&antifraud.Transaction{
								Id:            args.operation.Transaction.ID,
								TransactionId: args.operation.Transaction.TransactionID.String(),
								CreatedAt:     timestamppb.New(*args.operation.Transaction.CreatedAt),
								Amount:        args.operation.Transaction.Amount,
								Currency:      fmt.Sprint(args.operation.Transaction.Currency),
								Merchant:      args.operation.Transaction.Merchant,
								Country:       args.operation.Transaction.Country,
								SenderId:      args.operation.Transaction.SenderID.String(),
								ReceiverId:    nil,
								Bic:           nil,
								AtmId:         lo.ToPtr(args.operation.AtmID.String()),
							},
						).
						Return(
							&antifraud.CheckResult{
								NewStatus: antifraud.OperationStatus_Declined,
							},
							nil,
						)
				},
			},
			args: args{
				ctx: context.Background(),
				operation: &model.CashOutDomainRequest{
					ID: uuid.New(),
					Transaction: &model.Transaction{
						ID:            1,
						TransactionID: uuid.New(),
						CreatedAt:     lo.ToPtr(time.Now()),
						Amount:        500,
						Currency:      0,
						Merchant:      "sad",
						Country:       "dsa",
						SenderID:      uuid.New(),
					},
					AtmID: uuid.New(),
				},
			},
			expectedError: "operation declined",
		},
		{
			name: "timeout - ignored",
			fields: fields{
				setupMocks: func(af *mocks.MockOnlineCheckClient, args *args) {
					af.
						EXPECT().
						CashOut(
							gomock.Any(),
							gomock.Any(),
						).
						Return(
							nil,
							context.DeadlineExceeded,
						)
				},
			},
			args: args{
				ctx: context.Background(),
				operation: &model.CashOutDomainRequest{
					Transaction: &model.Transaction{
						ID:            1,
						TransactionID: uuid.New(),
						CreatedAt:     lo.ToPtr(time.Now()),
						Amount:        1500,
						Currency:      0,
						Merchant:      "merchant",
						Country:       "RU",
						SenderID:      uuid.New(),
					},
					AtmID: uuid.New(),
				},
			},
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			afClient := mocks.NewMockOnlineCheckClient(ctrl)
			tt.fields.setupMocks(afClient, &tt.args)

			integration := NewAfIntegration(afClient)

			err := integration.CashOutCheck(tt.args.ctx, tt.args.operation)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Equal(t, tt.expectedError, err.Error())
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestAfIntegration_InternalCheck(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx       context.Context
		operation *model.InternalDomainRequest
	}

	type fields struct {
		setupMocks func(af *mocks.MockOnlineCheckClient, args *args)
	}

	tests := []struct {
		name          string
		fields        fields
		args          args
		expectedError string
	}{
		{
			name: "success - approved",
			fields: fields{
				setupMocks: func(af *mocks.MockOnlineCheckClient, args *args) {
					af.
						EXPECT().
						Internal(
							gomock.Any(),
							&antifraud.Transaction{
								Id:            args.operation.Transaction.ID,
								TransactionId: args.operation.Transaction.TransactionID.String(),
								CreatedAt:     timestamppb.New(*args.operation.Transaction.CreatedAt),
								Amount:        args.operation.Transaction.Amount,
								Currency:      fmt.Sprint(args.operation.Transaction.Currency),
								Merchant:      args.operation.Transaction.Merchant,
								Country:       args.operation.Transaction.Country,
								SenderId:      args.operation.Transaction.SenderID.String(),
								ReceiverId:    lo.ToPtr(args.operation.ReceiverId.String()),
								Bic:           nil,
								AtmId:         nil,
							},
						).
						Return(
							&antifraud.CheckResult{
								NewStatus: antifraud.OperationStatus_Approved,
							},
							nil,
						)
				},
			},
			args: args{
				ctx: context.Background(),
				operation: &model.InternalDomainRequest{
					Transaction: &model.Transaction{
						ID:            1,
						TransactionID: uuid.New(),
						CreatedAt:     lo.ToPtr(time.Now()),
						Amount:        1000,
						Currency:      0,
						Merchant:      "merchant",
						Country:       "RU",
						SenderID:      uuid.New(),
					},
					ReceiverId: uuid.New(),
				},
			},
			expectedError: "",
		},
		{
			name: "error during grpc",
			fields: fields{
				setupMocks: func(af *mocks.MockOnlineCheckClient, args *args) {
					af.
						EXPECT().
						Internal(
							gomock.Any(),
							gomock.Any(),
						).
						Return(
							nil,
							errors.New("someErr"),
						)
				},
			},
			args: args{
				ctx: context.Background(),
				operation: &model.InternalDomainRequest{
					Transaction: &model.Transaction{
						ID:            1,
						TransactionID: uuid.New(),
						CreatedAt:     lo.ToPtr(time.Now()),
						Amount:        1000,
						Currency:      0,
						Merchant:      "merchant",
						Country:       "RU",
						SenderID:      uuid.New(),
					},
					ReceiverId: uuid.New(),
				},
			},
			expectedError: "",
		},
		{
			name: "declined",
			fields: fields{
				setupMocks: func(af *mocks.MockOnlineCheckClient, args *args) {
					af.
						EXPECT().
						Internal(
							gomock.Any(),
							gomock.Any(),
						).
						Return(
							&antifraud.CheckResult{
								NewStatus: antifraud.OperationStatus_Declined,
							},
							nil,
						)
				},
			},
			args: args{
				ctx: context.Background(),
				operation: &model.InternalDomainRequest{
					Transaction: &model.Transaction{
						ID:            1,
						TransactionID: uuid.New(),
						CreatedAt:     lo.ToPtr(time.Now()),
						Amount:        1000,
						Currency:      0,
						Merchant:      "merchant",
						Country:       "RU",
						SenderID:      uuid.New(),
					},
					ReceiverId: uuid.New(),
				},
			},
			expectedError: "operation declined",
		},
		{
			name: "timeout - ignored",
			fields: fields{
				setupMocks: func(af *mocks.MockOnlineCheckClient, args *args) {
					af.
						EXPECT().
						Internal(
							gomock.Any(),
							gomock.Any(),
						).
						Return(
							nil,
							context.DeadlineExceeded,
						)
				},
			},
			args: args{
				ctx: context.Background(),
				operation: &model.InternalDomainRequest{
					Transaction: &model.Transaction{
						ID:            1,
						TransactionID: uuid.New(),
						CreatedAt:     lo.ToPtr(time.Now()),
						Amount:        1500,
						Currency:      0,
						Merchant:      "merchant",
						Country:       "RU",
						SenderID:      uuid.New(),
					},
					ReceiverId: uuid.New(),
				},
			},
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			afClient := mocks.NewMockOnlineCheckClient(ctrl)
			tt.fields.setupMocks(afClient, &tt.args)

			integration := NewAfIntegration(afClient)

			err := integration.InternalCheck(tt.args.ctx, tt.args.operation)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Equal(t, tt.expectedError, err.Error())
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestAfIntegration_SbpOutgoingCheck(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx       context.Context
		operation *model.SbpOutgoingDomainRequest
	}

	type fields struct {
		setupMocks func(af *mocks.MockOnlineCheckClient, args *args)
	}

	tests := []struct {
		name          string
		fields        fields
		args          args
		expectedError string
	}{
		{
			name: "success - approved",
			fields: fields{
				setupMocks: func(af *mocks.MockOnlineCheckClient, args *args) {
					af.
						EXPECT().
						SbpOutgoing(
							gomock.Any(),
							&antifraud.Transaction{
								Id:            args.operation.Transaction.ID,
								TransactionId: args.operation.Transaction.TransactionID.String(),
								CreatedAt:     timestamppb.New(*args.operation.Transaction.CreatedAt),
								Amount:        args.operation.Transaction.Amount,
								Currency:      fmt.Sprint(args.operation.Transaction.Currency),
								Merchant:      args.operation.Transaction.Merchant,
								Country:       args.operation.Transaction.Country,
								SenderId:      args.operation.Transaction.SenderID.String(),
								ReceiverId:    lo.ToPtr(args.operation.ReceiverId.String()),
								Bic:           lo.ToPtr(args.operation.ReceiverId.String()),
								AtmId:         nil,
							},
						).
						Return(
							&antifraud.CheckResult{
								NewStatus: antifraud.OperationStatus_Approved,
							},
							nil,
						)
				},
			},
			args: args{
				ctx: context.Background(),
				operation: &model.SbpOutgoingDomainRequest{
					Transaction: &model.Transaction{
						ID:            1,
						TransactionID: uuid.New(),
						CreatedAt:     lo.ToPtr(time.Now()),
						Amount:        1500,
						Currency:      0,
						Merchant:      "merchant",
						Country:       "RU",
						SenderID:      uuid.New(),
					},
					ReceiverId: uuid.New(),
				},
			},
			expectedError: "",
		},
		{
			name: "error during grpc",
			fields: fields{
				setupMocks: func(af *mocks.MockOnlineCheckClient, args *args) {
					af.
						EXPECT().
						SbpOutgoing(
							gomock.Any(),
							gomock.Any(),
						).
						Return(
							nil,
							errors.New("someErr"),
						)
				},
			},
			args: args{
				ctx: context.Background(),
				operation: &model.SbpOutgoingDomainRequest{
					Transaction: &model.Transaction{
						ID:            1,
						TransactionID: uuid.New(),
						CreatedAt:     lo.ToPtr(time.Now()),
						Amount:        1500,
						Currency:      0,
						Merchant:      "merchant",
						Country:       "RU",
						SenderID:      uuid.New(),
					},
					ReceiverId: uuid.New(),
				},
			},
			expectedError: "",
		},
		{
			name: "declined",
			fields: fields{
				setupMocks: func(af *mocks.MockOnlineCheckClient, args *args) {
					af.
						EXPECT().
						SbpOutgoing(
							gomock.Any(),
							gomock.Any(),
						).
						Return(
							&antifraud.CheckResult{
								NewStatus: antifraud.OperationStatus_Declined,
							},
							nil,
						)
				},
			},
			args: args{
				ctx: context.Background(),
				operation: &model.SbpOutgoingDomainRequest{
					Transaction: &model.Transaction{
						ID:            1,
						TransactionID: uuid.New(),
						CreatedAt:     lo.ToPtr(time.Now()),
						Amount:        1500,
						Currency:      0,
						Merchant:      "merchant",
						Country:       "RU",
						SenderID:      uuid.New(),
					},
					ReceiverId: uuid.New(),
				},
			},
			expectedError: "operation declined",
		},
		{
			name: "timeout - ignored",
			fields: fields{
				setupMocks: func(af *mocks.MockOnlineCheckClient, args *args) {
					af.
						EXPECT().
						SbpOutgoing(
							gomock.Any(),
							gomock.Any(),
						).
						Return(
							nil,
							context.DeadlineExceeded,
						)
				},
			},
			args: args{
				ctx: context.Background(),
				operation: &model.SbpOutgoingDomainRequest{
					Transaction: &model.Transaction{
						ID:            1,
						TransactionID: uuid.New(),
						CreatedAt:     lo.ToPtr(time.Now()),
						Amount:        1500,
						Currency:      0,
						Merchant:      "merchant",
						Country:       "RU",
						SenderID:      uuid.New(),
					},
					ReceiverId: uuid.New(),
				},
			},
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			afClient := mocks.NewMockOnlineCheckClient(ctrl)
			tt.fields.setupMocks(afClient, &tt.args)

			integration := NewAfIntegration(afClient)

			err := integration.SbpOutgoingCheck(tt.args.ctx, tt.args.operation)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Equal(t, tt.expectedError, err.Error())
				return
			}

			require.NoError(t, err)
		})
	}
}
