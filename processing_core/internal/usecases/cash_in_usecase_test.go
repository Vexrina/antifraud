package usecases

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"processing_core/internal/app/model"
	"processing_core/internal/repository"
	"processing_core/internal/usecases/mocks"
	desc "processing_core/pkg/core"
)

func TestCashInUsecase_Process(t *testing.T) {
	t.Parallel()

	type args struct {
		domainRequest *model.CashInDomainRequest
	}
	type fields struct {
		setupMocks func(commonRepo *mocks.MockCommonRepo, clientRepo *mocks.MockCashOutClientRepo, args *args)
	}
	tests := []struct {
		name          string
		fields        fields
		args          args
		expectedResp  *desc.CashInResponse
		expectedError string
	}{
		{
			name: "success - process cash in",
			fields: fields{
				setupMocks: func(commonRepo *mocks.MockCommonRepo, clientRepo *mocks.MockCashOutClientRepo, args *args) {
					// настраиваем ожидания для методов внутри транзакции до вызова Transactional
					commonRepo.EXPECT().
						LockClient(gomock.Any(), gomock.Any(), args.domainRequest.Transaction.SenderID).
						Return(nil)
					clientRepo.EXPECT().
						GetCurrentBalanceTx(gomock.Any(), gomock.Any(), args.domainRequest.Transaction.SenderID).
						Return(int64(1000), nil)
					expectedBalance := int64(1000) + args.domainRequest.Transaction.Amount
					clientRepo.EXPECT().
						UpdateBalance(gomock.Any(), gomock.Any(), args.domainRequest.Transaction.SenderID, expectedBalance).
						Return(nil)
					clientRepo.EXPECT().
						UpsertTransaction(gomock.Any(), gomock.Any(), repository.MapCashInDomainToTransaction(*args.domainRequest)).
						Return(nil)
					commonRepo.EXPECT().
						Transactional(gomock.Any(), gomock.Any()).
						DoAndReturn(func(ctx context.Context, f func(tx pgx.Tx) error) error {
							// вызываем функцию с nil транзакцией, так как мы мокируем все вызовы
							return f(nil)
						})
				},
			},
			args: args{
				domainRequest: &model.CashInDomainRequest{
					ID: uuid.New(),
					Transaction: model.Transaction{
						ID:       1,
						Amount:   500,
						SenderID: uuid.New(),
					},
					AtmID: uuid.New(),
				},
			},
			expectedResp: &desc.CashInResponse{
				NewStatus: desc.OperationStatus_Approved,
			},
			expectedError: "",
		},
		{
			name: "error - LockClient fails",
			fields: fields{
				setupMocks: func(commonRepo *mocks.MockCommonRepo, clientRepo *mocks.MockCashOutClientRepo, args *args) {
					lockErr := errors.New("failed to lock client")
					commonRepo.EXPECT().
						LockClient(gomock.Any(), gomock.Any(), args.domainRequest.Transaction.SenderID).
						Return(lockErr)
					commonRepo.EXPECT().
						Transactional(gomock.Any(), gomock.Any()).
						DoAndReturn(func(ctx context.Context, f func(tx pgx.Tx) error) error {
							return f(nil)
						})
				},
			},
			args: args{
				domainRequest: &model.CashInDomainRequest{
					ID: uuid.New(),
					Transaction: model.Transaction{
						ID:       1,
						Amount:   500,
						SenderID: uuid.New(),
					},
					AtmID: uuid.New(),
				},
			},
			expectedResp: &desc.CashInResponse{
				NewStatus: desc.OperationStatus_Declined,
			},
			expectedError: "failed to lock client",
		},
		{
			name: "error - GetCurrentBalanceTx fails",
			fields: fields{
				setupMocks: func(commonRepo *mocks.MockCommonRepo, clientRepo *mocks.MockCashOutClientRepo, args *args) {
					balanceErr := errors.New("failed to get balance")
					commonRepo.EXPECT().
						LockClient(gomock.Any(), gomock.Any(), args.domainRequest.Transaction.SenderID).
						Return(nil)
					clientRepo.EXPECT().
						GetCurrentBalanceTx(gomock.Any(), gomock.Any(), args.domainRequest.Transaction.SenderID).
						Return(int64(0), balanceErr)
					commonRepo.EXPECT().
						Transactional(gomock.Any(), gomock.Any()).
						DoAndReturn(func(ctx context.Context, f func(tx pgx.Tx) error) error {
							return f(nil)
						})
				},
			},
			args: args{
				domainRequest: &model.CashInDomainRequest{
					ID: uuid.New(),
					Transaction: model.Transaction{
						ID:       1,
						Amount:   500,
						SenderID: uuid.New(),
					},
					AtmID: uuid.New(),
				},
			},
			expectedResp: &desc.CashInResponse{
				NewStatus: desc.OperationStatus_Declined,
			},
			expectedError: "failed to get balance",
		},
		{
			name: "error - UpdateBalance fails",
			fields: fields{
				setupMocks: func(commonRepo *mocks.MockCommonRepo, clientRepo *mocks.MockCashOutClientRepo, args *args) {
					updateErr := errors.New("failed to update balance")
					commonRepo.EXPECT().
						LockClient(gomock.Any(), gomock.Any(), args.domainRequest.Transaction.SenderID).
						Return(nil)
					clientRepo.EXPECT().
						GetCurrentBalanceTx(gomock.Any(), gomock.Any(), args.domainRequest.Transaction.SenderID).
						Return(int64(1000), nil)
					expectedBalance := int64(1000) + args.domainRequest.Transaction.Amount
					clientRepo.EXPECT().
						UpdateBalance(gomock.Any(), gomock.Any(), args.domainRequest.Transaction.SenderID, expectedBalance).
						Return(updateErr)
					commonRepo.EXPECT().
						Transactional(gomock.Any(), gomock.Any()).
						DoAndReturn(func(ctx context.Context, f func(tx pgx.Tx) error) error {
							return f(nil)
						})
				},
			},
			args: args{
				domainRequest: &model.CashInDomainRequest{
					ID: uuid.New(),
					Transaction: model.Transaction{
						ID:       1,
						Amount:   500,
						SenderID: uuid.New(),
					},
					AtmID: uuid.New(),
				},
			},
			expectedResp: &desc.CashInResponse{
				NewStatus: desc.OperationStatus_Declined,
			},
			expectedError: "failed to update balance",
		},
		{
			name: "success - zero initial balance",
			fields: fields{
				setupMocks: func(commonRepo *mocks.MockCommonRepo, clientRepo *mocks.MockCashOutClientRepo, args *args) {
					commonRepo.EXPECT().
						LockClient(gomock.Any(), gomock.Any(), args.domainRequest.Transaction.SenderID).
						Return(nil)
					clientRepo.EXPECT().
						GetCurrentBalanceTx(gomock.Any(), gomock.Any(), args.domainRequest.Transaction.SenderID).
						Return(int64(0), nil)
					expectedBalance := int64(0) + args.domainRequest.Transaction.Amount
					clientRepo.EXPECT().
						UpdateBalance(gomock.Any(), gomock.Any(), args.domainRequest.Transaction.SenderID, expectedBalance).
						Return(nil)
					clientRepo.EXPECT().
						UpsertTransaction(gomock.Any(), gomock.Any(), repository.MapCashInDomainToTransaction(*args.domainRequest)).
						Return(nil)
					commonRepo.EXPECT().
						Transactional(gomock.Any(), gomock.Any()).
						DoAndReturn(func(ctx context.Context, f func(tx pgx.Tx) error) error {
							return f(nil)
						})
				},
			},
			args: args{
				domainRequest: &model.CashInDomainRequest{
					ID: uuid.New(),
					Transaction: model.Transaction{
						ID:       1,
						Amount:   1000,
						SenderID: uuid.New(),
					},
					AtmID: uuid.New(),
				},
			},
			expectedResp: &desc.CashInResponse{
				NewStatus: desc.OperationStatus_Approved,
			},
			expectedError: "",
		},
		{
			name: "success - negative initial balance",
			fields: fields{
				setupMocks: func(commonRepo *mocks.MockCommonRepo, clientRepo *mocks.MockCashOutClientRepo, args *args) {
					commonRepo.EXPECT().
						LockClient(gomock.Any(), gomock.Any(), args.domainRequest.Transaction.SenderID).
						Return(nil)
					clientRepo.EXPECT().
						GetCurrentBalanceTx(gomock.Any(), gomock.Any(), args.domainRequest.Transaction.SenderID).
						Return(int64(-500), nil)
					expectedBalance := int64(-500) + args.domainRequest.Transaction.Amount
					clientRepo.EXPECT().
						UpdateBalance(gomock.Any(), gomock.Any(), args.domainRequest.Transaction.SenderID, expectedBalance).
						Return(nil)
					clientRepo.EXPECT().
						UpsertTransaction(gomock.Any(), gomock.Any(), repository.MapCashInDomainToTransaction(*args.domainRequest)).
						Return(nil)
					commonRepo.EXPECT().
						Transactional(gomock.Any(), gomock.Any()).
						DoAndReturn(func(ctx context.Context, f func(tx pgx.Tx) error) error {
							return f(nil)
						})
				},
			},
			args: args{
				domainRequest: &model.CashInDomainRequest{
					ID: uuid.New(),
					Transaction: model.Transaction{
						ID:       1,
						Amount:   1000,
						SenderID: uuid.New(),
					},
					AtmID: uuid.New(),
				},
			},
			expectedResp: &desc.CashInResponse{
				NewStatus: desc.OperationStatus_Approved,
			},
			expectedError: "",
		},
		{
			name: "error - upsert fails",
			fields: fields{
				setupMocks: func(commonRepo *mocks.MockCommonRepo, clientRepo *mocks.MockCashOutClientRepo, args *args) {
					commonRepo.EXPECT().
						LockClient(gomock.Any(), gomock.Any(), args.domainRequest.Transaction.SenderID).
						Return(nil)
					clientRepo.EXPECT().
						GetCurrentBalanceTx(gomock.Any(), gomock.Any(), args.domainRequest.Transaction.SenderID).
						Return(int64(-500), nil)
					expectedBalance := int64(-500) + args.domainRequest.Transaction.Amount
					clientRepo.EXPECT().
						UpdateBalance(gomock.Any(), gomock.Any(), args.domainRequest.Transaction.SenderID, expectedBalance).
						Return(nil)
					clientRepo.EXPECT().
						UpsertTransaction(gomock.Any(), gomock.Any(), repository.MapCashInDomainToTransaction(*args.domainRequest)).
						Return(errors.New("upsert fail"))
					commonRepo.EXPECT().
						Transactional(gomock.Any(), gomock.Any()).
						DoAndReturn(func(ctx context.Context, f func(tx pgx.Tx) error) error {
							return f(nil)
						})
				},
			},
			args: args{
				domainRequest: &model.CashInDomainRequest{
					ID: uuid.New(),
					Transaction: model.Transaction{
						ID:       1,
						Amount:   1000,
						SenderID: uuid.New(),
					},
					AtmID: uuid.New(),
				},
			},
			expectedResp: &desc.CashInResponse{
				NewStatus: desc.OperationStatus_Declined,
			},
			expectedError: "upsert fail",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			commonRepo := mocks.NewMockCommonRepo(ctrl)
			clientRepo := mocks.NewMockCashOutClientRepo(ctrl)

			tt.fields.setupMocks(commonRepo, clientRepo, &tt.args)

			usecase := NewCashInUsecase(commonRepo, clientRepo)

			resp, err := usecase.Process(context.Background(), tt.args.domainRequest)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Equal(t, tt.expectedError, err.Error())
				assert.Equal(t, tt.expectedResp.NewStatus, resp.NewStatus)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedResp.NewStatus, resp.NewStatus)
		})
	}
}
