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

func TestInternalUsecase_Process(t *testing.T) {
	t.Parallel()

	type args struct {
		domainRequest *model.InternalDomainRequest
	}
	type fields struct {
		setupMocks func(commonRepo *mocks.MockCommonRepo, clientRepo *mocks.MockInternalClientRepo, antifraud *mocks.MockAntifraudInternalCheck, args *args)
	}
	tests := []struct {
		name          string
		fields        fields
		args          args
		expectedResp  *desc.InternalResponse
		expectedError string
	}{
		{
			name: "success - process internal transfer",
			fields: fields{
				setupMocks: func(commonRepo *mocks.MockCommonRepo, clientRepo *mocks.MockInternalClientRepo, antifraud *mocks.MockAntifraudInternalCheck, args *args) {
					antifraud.EXPECT().
						InternalCheck(gomock.Any(), args.domainRequest).
						Return(nil)
					commonRepo.EXPECT().
						LockClient(gomock.Any(), gomock.Any(), args.domainRequest.Transaction.SenderID).
						Return(nil)
					commonRepo.EXPECT().
						LockClient(gomock.Any(), gomock.Any(), args.domainRequest.ReceiverId).
						Return(nil)
					clientRepo.EXPECT().
						GetCurrentBalanceTx(gomock.Any(), gomock.Any(), args.domainRequest.Transaction.SenderID).
						Return(int64(2000), nil)
					senderNewBalance := int64(2000) - args.domainRequest.Transaction.Amount
					clientRepo.EXPECT().
						UpdateBalance(gomock.Any(), gomock.Any(), args.domainRequest.Transaction.SenderID, senderNewBalance).
						Return(nil)
					clientRepo.EXPECT().
						GetCurrentBalanceTx(gomock.Any(), gomock.Any(), args.domainRequest.ReceiverId).
						Return(int64(1000), nil)
					receiverNewBalance := int64(1000) + args.domainRequest.Transaction.Amount
					clientRepo.EXPECT().
						UpdateBalance(gomock.Any(), gomock.Any(), args.domainRequest.ReceiverId, receiverNewBalance).
						Return(nil)
					clientRepo.EXPECT().
						UpsertTransaction(gomock.Any(), gomock.Any(), repository.MapInternalDomainToTransaction(*args.domainRequest)).
						Return(nil)
					commonRepo.EXPECT().
						Transactional(gomock.Any(), gomock.Any()).
						DoAndReturn(func(ctx context.Context, f func(tx pgx.Tx) error) error {
							return f(nil)
						})
				},
			},
			args: args{
				domainRequest: &model.InternalDomainRequest{
					ID: uuid.New(),
					Transaction: &model.Transaction{
						ID:       1,
						Amount:   500,
						SenderID: uuid.New(),
					},
					ReceiverId: uuid.New(),
				},
			},
			expectedResp: &desc.InternalResponse{
				NewStatus: desc.OperationStatus_Approved,
			},
			expectedError: "",
		},
		{
			name: "error - antifraud check fails",
			fields: fields{
				setupMocks: func(commonRepo *mocks.MockCommonRepo, clientRepo *mocks.MockInternalClientRepo, antifraud *mocks.MockAntifraudInternalCheck, args *args) {
					antifraudErr := errors.New("antifraud check failed")
					antifraud.EXPECT().
						InternalCheck(gomock.Any(), args.domainRequest).
						Return(antifraudErr)
				},
			},
			args: args{
				domainRequest: &model.InternalDomainRequest{
					ID: uuid.New(),
					Transaction: &model.Transaction{
						ID:       1,
						Amount:   500,
						SenderID: uuid.New(),
					},
					ReceiverId: uuid.New(),
				},
			},
			expectedResp: &desc.InternalResponse{
				NewStatus: desc.OperationStatus_Declined,
			},
			expectedError: "antifraud check failed",
		},
		{
			name: "error - LockClient sender fails",
			fields: fields{
				setupMocks: func(commonRepo *mocks.MockCommonRepo, clientRepo *mocks.MockInternalClientRepo, antifraud *mocks.MockAntifraudInternalCheck, args *args) {
					antifraud.EXPECT().
						InternalCheck(gomock.Any(), args.domainRequest).
						Return(nil)
					lockErr := errors.New("failed to lock sender")
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
				domainRequest: &model.InternalDomainRequest{
					ID: uuid.New(),
					Transaction: &model.Transaction{
						ID:       1,
						Amount:   500,
						SenderID: uuid.New(),
					},
					ReceiverId: uuid.New(),
				},
			},
			expectedResp: &desc.InternalResponse{
				NewStatus: desc.OperationStatus_Declined,
			},
			expectedError: "failed to lock sender",
		},
		{
			name: "error - LockClient receiver fails",
			fields: fields{
				setupMocks: func(commonRepo *mocks.MockCommonRepo, clientRepo *mocks.MockInternalClientRepo, antifraud *mocks.MockAntifraudInternalCheck, args *args) {
					antifraud.EXPECT().
						InternalCheck(gomock.Any(), args.domainRequest).
						Return(nil)
					commonRepo.EXPECT().
						LockClient(gomock.Any(), gomock.Any(), args.domainRequest.Transaction.SenderID).
						Return(nil)
					lockErr := errors.New("failed to lock receiver")
					commonRepo.EXPECT().
						LockClient(gomock.Any(), gomock.Any(), args.domainRequest.ReceiverId).
						Return(lockErr)
					commonRepo.EXPECT().
						Transactional(gomock.Any(), gomock.Any()).
						DoAndReturn(func(ctx context.Context, f func(tx pgx.Tx) error) error {
							return f(nil)
						})
				},
			},
			args: args{
				domainRequest: &model.InternalDomainRequest{
					ID: uuid.New(),
					Transaction: &model.Transaction{
						ID:       1,
						Amount:   500,
						SenderID: uuid.New(),
					},
					ReceiverId: uuid.New(),
				},
			},
			expectedResp: &desc.InternalResponse{
				NewStatus: desc.OperationStatus_Declined,
			},
			expectedError: "failed to lock receiver",
		},
		{
			name: "error - GetCurrentBalanceTx sender fails",
			fields: fields{
				setupMocks: func(commonRepo *mocks.MockCommonRepo, clientRepo *mocks.MockInternalClientRepo, antifraud *mocks.MockAntifraudInternalCheck, args *args) {
					antifraud.EXPECT().
						InternalCheck(gomock.Any(), args.domainRequest).
						Return(nil)
					commonRepo.EXPECT().
						LockClient(gomock.Any(), gomock.Any(), args.domainRequest.Transaction.SenderID).
						Return(nil)
					commonRepo.EXPECT().
						LockClient(gomock.Any(), gomock.Any(), args.domainRequest.ReceiverId).
						Return(nil)
					balanceErr := errors.New("failed to get sender balance")
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
				domainRequest: &model.InternalDomainRequest{
					ID: uuid.New(),
					Transaction: &model.Transaction{
						ID:       1,
						Amount:   500,
						SenderID: uuid.New(),
					},
					ReceiverId: uuid.New(),
				},
			},
			expectedResp: &desc.InternalResponse{
				NewStatus: desc.OperationStatus_Declined,
			},
			expectedError: "failed to get sender balance",
		},
		{
			name: "error - sender balance insufficient",
			fields: fields{
				setupMocks: func(commonRepo *mocks.MockCommonRepo, clientRepo *mocks.MockInternalClientRepo, antifraud *mocks.MockAntifraudInternalCheck, args *args) {
					antifraud.EXPECT().
						InternalCheck(gomock.Any(), args.domainRequest).
						Return(nil)
					commonRepo.EXPECT().
						LockClient(gomock.Any(), gomock.Any(), args.domainRequest.Transaction.SenderID).
						Return(nil)
					commonRepo.EXPECT().
						LockClient(gomock.Any(), gomock.Any(), args.domainRequest.ReceiverId).
						Return(nil)
					// баланс меньше суммы перевода
					clientRepo.EXPECT().
						GetCurrentBalanceTx(gomock.Any(), gomock.Any(), args.domainRequest.Transaction.SenderID).
						Return(int64(300), nil)
					commonRepo.EXPECT().
						Transactional(gomock.Any(), gomock.Any()).
						DoAndReturn(func(ctx context.Context, f func(tx pgx.Tx) error) error {
							return f(nil)
						})
				},
			},
			args: args{
				domainRequest: &model.InternalDomainRequest{
					ID: uuid.New(),
					Transaction: &model.Transaction{
						ID:       1,
						Amount:   500,
						SenderID: uuid.New(),
					},
					ReceiverId: uuid.New(),
				},
			},
			expectedResp: &desc.InternalResponse{
				NewStatus: desc.OperationStatus_Declined,
			},
			expectedError: "LIMIT OVERFLOW",
		},
		{
			name: "error - UpdateBalance sender fails",
			fields: fields{
				setupMocks: func(commonRepo *mocks.MockCommonRepo, clientRepo *mocks.MockInternalClientRepo, antifraud *mocks.MockAntifraudInternalCheck, args *args) {
					antifraud.EXPECT().
						InternalCheck(gomock.Any(), args.domainRequest).
						Return(nil)
					commonRepo.EXPECT().
						LockClient(gomock.Any(), gomock.Any(), args.domainRequest.Transaction.SenderID).
						Return(nil)
					commonRepo.EXPECT().
						LockClient(gomock.Any(), gomock.Any(), args.domainRequest.ReceiverId).
						Return(nil)
					clientRepo.EXPECT().
						GetCurrentBalanceTx(gomock.Any(), gomock.Any(), args.domainRequest.Transaction.SenderID).
						Return(int64(2000), nil)
					updateErr := errors.New("failed to update sender balance")
					senderNewBalance := int64(2000) - args.domainRequest.Transaction.Amount
					clientRepo.EXPECT().
						UpdateBalance(gomock.Any(), gomock.Any(), args.domainRequest.Transaction.SenderID, senderNewBalance).
						Return(updateErr)
					commonRepo.EXPECT().
						Transactional(gomock.Any(), gomock.Any()).
						DoAndReturn(func(ctx context.Context, f func(tx pgx.Tx) error) error {
							return f(nil)
						})
				},
			},
			args: args{
				domainRequest: &model.InternalDomainRequest{
					ID: uuid.New(),
					Transaction: &model.Transaction{
						ID:       1,
						Amount:   500,
						SenderID: uuid.New(),
					},
					ReceiverId: uuid.New(),
				},
			},
			expectedResp: &desc.InternalResponse{
				NewStatus: desc.OperationStatus_Declined,
			},
			expectedError: "failed to update sender balance",
		},
		{
			name: "error - GetCurrentBalanceTx receiver fails",
			fields: fields{
				setupMocks: func(commonRepo *mocks.MockCommonRepo, clientRepo *mocks.MockInternalClientRepo, antifraud *mocks.MockAntifraudInternalCheck, args *args) {
					antifraud.EXPECT().
						InternalCheck(gomock.Any(), args.domainRequest).
						Return(nil)
					commonRepo.EXPECT().
						LockClient(gomock.Any(), gomock.Any(), args.domainRequest.Transaction.SenderID).
						Return(nil)
					commonRepo.EXPECT().
						LockClient(gomock.Any(), gomock.Any(), args.domainRequest.ReceiverId).
						Return(nil)
					clientRepo.EXPECT().
						GetCurrentBalanceTx(gomock.Any(), gomock.Any(), args.domainRequest.Transaction.SenderID).
						Return(int64(2000), nil)
					senderNewBalance := int64(2000) - args.domainRequest.Transaction.Amount
					clientRepo.EXPECT().
						UpdateBalance(gomock.Any(), gomock.Any(), args.domainRequest.Transaction.SenderID, senderNewBalance).
						Return(nil)
					balanceErr := errors.New("failed to get receiver balance")
					clientRepo.EXPECT().
						GetCurrentBalanceTx(gomock.Any(), gomock.Any(), args.domainRequest.ReceiverId).
						Return(int64(0), balanceErr)
					commonRepo.EXPECT().
						Transactional(gomock.Any(), gomock.Any()).
						DoAndReturn(func(ctx context.Context, f func(tx pgx.Tx) error) error {
							return f(nil)
						})
				},
			},
			args: args{
				domainRequest: &model.InternalDomainRequest{
					ID: uuid.New(),
					Transaction: &model.Transaction{
						ID:       1,
						Amount:   500,
						SenderID: uuid.New(),
					},
					ReceiverId: uuid.New(),
				},
			},
			expectedResp: &desc.InternalResponse{
				NewStatus: desc.OperationStatus_Declined,
			},
			expectedError: "failed to get receiver balance",
		},
		{
			name: "error - UpdateBalance receiver fails",
			fields: fields{
				setupMocks: func(commonRepo *mocks.MockCommonRepo, clientRepo *mocks.MockInternalClientRepo, antifraud *mocks.MockAntifraudInternalCheck, args *args) {
					antifraud.EXPECT().
						InternalCheck(gomock.Any(), args.domainRequest).
						Return(nil)
					commonRepo.EXPECT().
						LockClient(gomock.Any(), gomock.Any(), args.domainRequest.Transaction.SenderID).
						Return(nil)
					commonRepo.EXPECT().
						LockClient(gomock.Any(), gomock.Any(), args.domainRequest.ReceiverId).
						Return(nil)
					clientRepo.EXPECT().
						GetCurrentBalanceTx(gomock.Any(), gomock.Any(), args.domainRequest.Transaction.SenderID).
						Return(int64(2000), nil)
					senderNewBalance := int64(2000) - args.domainRequest.Transaction.Amount
					clientRepo.EXPECT().
						UpdateBalance(gomock.Any(), gomock.Any(), args.domainRequest.Transaction.SenderID, senderNewBalance).
						Return(nil)
					clientRepo.EXPECT().
						GetCurrentBalanceTx(gomock.Any(), gomock.Any(), args.domainRequest.ReceiverId).
						Return(int64(1000), nil)
					updateErr := errors.New("failed to update receiver balance")
					receiverNewBalance := int64(1000) + args.domainRequest.Transaction.Amount
					clientRepo.EXPECT().
						UpdateBalance(gomock.Any(), gomock.Any(), args.domainRequest.ReceiverId, receiverNewBalance).
						Return(updateErr)
					commonRepo.EXPECT().
						Transactional(gomock.Any(), gomock.Any()).
						DoAndReturn(func(ctx context.Context, f func(tx pgx.Tx) error) error {
							return f(nil)
						})
				},
			},
			args: args{
				domainRequest: &model.InternalDomainRequest{
					ID: uuid.New(),
					Transaction: &model.Transaction{
						ID:       1,
						Amount:   500,
						SenderID: uuid.New(),
					},
					ReceiverId: uuid.New(),
				},
			},
			expectedResp: &desc.InternalResponse{
				NewStatus: desc.OperationStatus_Declined,
			},
			expectedError: "failed to update receiver balance",
		},
		{
			name: "error - AddOperationToHistory fails",
			fields: fields{
				setupMocks: func(commonRepo *mocks.MockCommonRepo, clientRepo *mocks.MockInternalClientRepo, antifraud *mocks.MockAntifraudInternalCheck, args *args) {
					antifraud.EXPECT().
						InternalCheck(gomock.Any(), args.domainRequest).
						Return(nil)
					commonRepo.EXPECT().
						LockClient(gomock.Any(), gomock.Any(), args.domainRequest.Transaction.SenderID).
						Return(nil)
					commonRepo.EXPECT().
						LockClient(gomock.Any(), gomock.Any(), args.domainRequest.ReceiverId).
						Return(nil)
					clientRepo.EXPECT().
						GetCurrentBalanceTx(gomock.Any(), gomock.Any(), args.domainRequest.Transaction.SenderID).
						Return(int64(2000), nil)
					senderNewBalance := int64(2000) - args.domainRequest.Transaction.Amount
					clientRepo.EXPECT().
						UpdateBalance(gomock.Any(), gomock.Any(), args.domainRequest.Transaction.SenderID, senderNewBalance).
						Return(nil)
					clientRepo.EXPECT().
						GetCurrentBalanceTx(gomock.Any(), gomock.Any(), args.domainRequest.ReceiverId).
						Return(int64(1000), nil)
					receiverNewBalance := int64(1000) + args.domainRequest.Transaction.Amount
					clientRepo.EXPECT().
						UpdateBalance(gomock.Any(), gomock.Any(), args.domainRequest.ReceiverId, receiverNewBalance).
						Return(nil)
					historyErr := errors.New("failed to add operation to history")
					clientRepo.EXPECT().
						UpsertTransaction(gomock.Any(), gomock.Any(), repository.MapInternalDomainToTransaction(*args.domainRequest)).
						Return(historyErr)
					commonRepo.EXPECT().
						Transactional(gomock.Any(), gomock.Any()).
						DoAndReturn(func(ctx context.Context, f func(tx pgx.Tx) error) error {
							return f(nil)
						})
				},
			},
			args: args{
				domainRequest: &model.InternalDomainRequest{
					ID: uuid.New(),
					Transaction: &model.Transaction{
						ID:       1,
						Amount:   500,
						SenderID: uuid.New(),
					},
					ReceiverId: uuid.New(),
				},
			},
			expectedResp: &desc.InternalResponse{
				NewStatus: desc.OperationStatus_Declined,
			},
			expectedError: "failed to add operation to history",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			commonRepo := mocks.NewMockCommonRepo(ctrl)
			clientRepo := mocks.NewMockInternalClientRepo(ctrl)
			antifraud := mocks.NewMockAntifraudInternalCheck(ctrl)

			tt.fields.setupMocks(commonRepo, clientRepo, antifraud, &tt.args)

			usecase := &internalUsecase{
				commonRepo: commonRepo,
				clientRepo: clientRepo,
				antifraud:  antifraud,
			}

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
