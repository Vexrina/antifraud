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
	"processing_core/internal/usecases/mocks"
	desc "processing_core/pkg/core"
)

func TestSbpOutgoingUsecase_Process(t *testing.T) {
	t.Parallel()

	type args struct {
		domainRequest *model.SbpOutgoingDomainRequest
	}
	type fields struct {
		setupMocks func(commonRepo *mocks.MockCommonRepo, clientRepo *mocks.MockSbpOutgoingClientRepo, antifraud *mocks.MockAntifraudSbpOutgoingCheck, sbpIntegration *mocks.MockSbpIntegrationInterface, args *args)
	}
	tests := []struct {
		name          string
		fields        fields
		args          args
		expectedResp  *desc.SbpOutgoingResponse
		expectedError string
	}{
		{
			name: "success - process sbp outgoing",
			fields: fields{
				setupMocks: func(commonRepo *mocks.MockCommonRepo, clientRepo *mocks.MockSbpOutgoingClientRepo, antifraud *mocks.MockAntifraudSbpOutgoingCheck, sbpIntegration *mocks.MockSbpIntegrationInterface, args *args) {
					antifraud.EXPECT().
						SbpOutgoingCheck(gomock.Any(), *args.domainRequest.Transaction).
						Return(nil)
					commonRepo.EXPECT().
						LockClient(gomock.Any(), gomock.Any(), args.domainRequest.Transaction.SenderID).
						Return(nil)
					clientRepo.EXPECT().
						GetCurrentBalanceTx(gomock.Any(), gomock.Any(), args.domainRequest.Transaction.SenderID).
						Return(int64(2000), nil)
					senderNewBalance := int64(2000) - args.domainRequest.Transaction.Amount
					clientRepo.EXPECT().
						UpdateBalance(gomock.Any(), gomock.Any(), args.domainRequest.Transaction.SenderID, senderNewBalance).
						Return(nil)
					clientRepo.EXPECT().
						AddOperationToHistory(gomock.Any(), gomock.Any(), uuid.Nil, *args.domainRequest.Transaction).
						Return(nil)
					commonRepo.EXPECT().
						Transactional(gomock.Any(), gomock.Any()).
						DoAndReturn(func(ctx context.Context, f func(tx pgx.Tx) error) error {
							return f(nil)
						})
					sbpIntegration.EXPECT().
						ToAnotherBank(gomock.Any(), *args.domainRequest.Transaction)
				},
			},
			args: args{
				domainRequest: &model.SbpOutgoingDomainRequest{
					ID: uuid.New(),
					Transaction: &model.Transaction{
						ID:       1,
						Amount:   500,
						SenderID: uuid.New(),
					},
					ReceiverId: uuid.New(),
					Bic:        "044525225",
				},
			},
			expectedResp: &desc.SbpOutgoingResponse{
				NewStatus: desc.OperationStatus_Approved,
			},
			expectedError: "",
		},
		{
			name: "error - antifraud check fails",
			fields: fields{
				setupMocks: func(commonRepo *mocks.MockCommonRepo, clientRepo *mocks.MockSbpOutgoingClientRepo, antifraud *mocks.MockAntifraudSbpOutgoingCheck, sbpIntegration *mocks.MockSbpIntegrationInterface, args *args) {
					antifraudErr := errors.New("antifraud check failed")
					antifraud.EXPECT().
						SbpOutgoingCheck(gomock.Any(), *args.domainRequest.Transaction).
						Return(antifraudErr)
				},
			},
			args: args{
				domainRequest: &model.SbpOutgoingDomainRequest{
					ID: uuid.New(),
					Transaction: &model.Transaction{
						ID:       1,
						Amount:   500,
						SenderID: uuid.New(),
					},
					ReceiverId: uuid.New(),
					Bic:        "044525225",
				},
			},
			expectedResp: &desc.SbpOutgoingResponse{
				NewStatus: desc.OperationStatus_Declined,
			},
			expectedError: "antifraud check failed",
		},
		{
			name: "error - LockClient fails",
			fields: fields{
				setupMocks: func(commonRepo *mocks.MockCommonRepo, clientRepo *mocks.MockSbpOutgoingClientRepo, antifraud *mocks.MockAntifraudSbpOutgoingCheck, sbpIntegration *mocks.MockSbpIntegrationInterface, args *args) {
					antifraud.EXPECT().
						SbpOutgoingCheck(gomock.Any(), *args.domainRequest.Transaction).
						Return(nil)
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
				domainRequest: &model.SbpOutgoingDomainRequest{
					ID: uuid.New(),
					Transaction: &model.Transaction{
						ID:       1,
						Amount:   500,
						SenderID: uuid.New(),
					},
					ReceiverId: uuid.New(),
					Bic:        "044525225",
				},
			},
			expectedResp: &desc.SbpOutgoingResponse{
				NewStatus: desc.OperationStatus_Declined,
			},
			expectedError: "failed to lock client",
		},
		{
			name: "error - GetCurrentBalanceTx fails",
			fields: fields{
				setupMocks: func(commonRepo *mocks.MockCommonRepo, clientRepo *mocks.MockSbpOutgoingClientRepo, antifraud *mocks.MockAntifraudSbpOutgoingCheck, sbpIntegration *mocks.MockSbpIntegrationInterface, args *args) {
					antifraud.EXPECT().
						SbpOutgoingCheck(gomock.Any(), *args.domainRequest.Transaction).
						Return(nil)
					commonRepo.EXPECT().
						LockClient(gomock.Any(), gomock.Any(), args.domainRequest.Transaction.SenderID).
						Return(nil)
					balanceErr := errors.New("failed to get balance")
					clientRepo.EXPECT().
						GetCurrentBalanceTx(gomock.Any(), gomock.Any(), args.domainRequest.Transaction.SenderID).
						Return(int64(0), balanceErr)
					commonRepo.EXPECT().
						Transactional(gomock.Any(), gomock.Any()).
						DoAndReturn(func(ctx context.Context, f func(tx pgx.Tx) error) error {
							return f(nil)
						})
					// ToAnotherBank не должен вызываться при ошибке
				},
			},
			args: args{
				domainRequest: &model.SbpOutgoingDomainRequest{
					ID: uuid.New(),
					Transaction: &model.Transaction{
						ID:       1,
						Amount:   500,
						SenderID: uuid.New(),
					},
					ReceiverId: uuid.New(),
					Bic:        "044525225",
				},
			},
			expectedResp: &desc.SbpOutgoingResponse{
				NewStatus: desc.OperationStatus_Declined,
			},
			expectedError: "failed to get balance",
		},
		{
			name: "error - sender balance insufficient",
			fields: fields{
				setupMocks: func(commonRepo *mocks.MockCommonRepo, clientRepo *mocks.MockSbpOutgoingClientRepo, antifraud *mocks.MockAntifraudSbpOutgoingCheck, sbpIntegration *mocks.MockSbpIntegrationInterface, args *args) {
					antifraud.EXPECT().
						SbpOutgoingCheck(gomock.Any(), *args.domainRequest.Transaction).
						Return(nil)
					commonRepo.EXPECT().
						LockClient(gomock.Any(), gomock.Any(), args.domainRequest.Transaction.SenderID).
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
				domainRequest: &model.SbpOutgoingDomainRequest{
					ID: uuid.New(),
					Transaction: &model.Transaction{
						ID:       1,
						Amount:   500,
						SenderID: uuid.New(),
					},
					ReceiverId: uuid.New(),
					Bic:        "044525225",
				},
			},
			expectedResp: &desc.SbpOutgoingResponse{
				NewStatus: desc.OperationStatus_Declined,
			},
			expectedError: "LIMIT OVERFLOW",
		},
		{
			name: "error - UpdateBalance fails",
			fields: fields{
				setupMocks: func(commonRepo *mocks.MockCommonRepo, clientRepo *mocks.MockSbpOutgoingClientRepo, antifraud *mocks.MockAntifraudSbpOutgoingCheck, sbpIntegration *mocks.MockSbpIntegrationInterface, args *args) {
					antifraud.EXPECT().
						SbpOutgoingCheck(gomock.Any(), *args.domainRequest.Transaction).
						Return(nil)
					commonRepo.EXPECT().
						LockClient(gomock.Any(), gomock.Any(), args.domainRequest.Transaction.SenderID).
						Return(nil)
					clientRepo.EXPECT().
						GetCurrentBalanceTx(gomock.Any(), gomock.Any(), args.domainRequest.Transaction.SenderID).
						Return(int64(2000), nil)
					updateErr := errors.New("failed to update balance")
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
				domainRequest: &model.SbpOutgoingDomainRequest{
					ID: uuid.New(),
					Transaction: &model.Transaction{
						ID:       1,
						Amount:   500,
						SenderID: uuid.New(),
					},
					ReceiverId: uuid.New(),
					Bic:        "044525225",
				},
			},
			expectedResp: &desc.SbpOutgoingResponse{
				NewStatus: desc.OperationStatus_Declined,
			},
			expectedError: "failed to update balance",
		},
		{
			name: "error - AddOperationToHistory fails",
			fields: fields{
				setupMocks: func(commonRepo *mocks.MockCommonRepo, clientRepo *mocks.MockSbpOutgoingClientRepo, antifraud *mocks.MockAntifraudSbpOutgoingCheck, sbpIntegration *mocks.MockSbpIntegrationInterface, args *args) {
					antifraud.EXPECT().
						SbpOutgoingCheck(gomock.Any(), *args.domainRequest.Transaction).
						Return(nil)
					commonRepo.EXPECT().
						LockClient(gomock.Any(), gomock.Any(), args.domainRequest.Transaction.SenderID).
						Return(nil)
					clientRepo.EXPECT().
						GetCurrentBalanceTx(gomock.Any(), gomock.Any(), args.domainRequest.Transaction.SenderID).
						Return(int64(2000), nil)
					senderNewBalance := int64(2000) - args.domainRequest.Transaction.Amount
					clientRepo.EXPECT().
						UpdateBalance(gomock.Any(), gomock.Any(), args.domainRequest.Transaction.SenderID, senderNewBalance).
						Return(nil)
					historyErr := errors.New("failed to add operation to history")
					clientRepo.EXPECT().
						AddOperationToHistory(gomock.Any(), gomock.Any(), uuid.Nil, *args.domainRequest.Transaction).
						Return(historyErr)
					commonRepo.EXPECT().
						Transactional(gomock.Any(), gomock.Any()).
						DoAndReturn(func(ctx context.Context, f func(tx pgx.Tx) error) error {
							return f(nil)
						})
				},
			},
			args: args{
				domainRequest: &model.SbpOutgoingDomainRequest{
					ID: uuid.New(),
					Transaction: &model.Transaction{
						ID:       1,
						Amount:   500,
						SenderID: uuid.New(),
					},
					ReceiverId: uuid.New(),
					Bic:        "044525225",
				},
			},
			expectedResp: &desc.SbpOutgoingResponse{
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
			clientRepo := mocks.NewMockSbpOutgoingClientRepo(ctrl)
			antifraud := mocks.NewMockAntifraudSbpOutgoingCheck(ctrl)
			sbpIntegration := mocks.NewMockSbpIntegrationInterface(ctrl)

			tt.fields.setupMocks(commonRepo, clientRepo, antifraud, sbpIntegration, &tt.args)

			usecase := &sbpOutgoingUsecase{
				commonRepo:     commonRepo,
				clientRepo:     clientRepo,
				antifraud:      antifraud,
				sbpIntegration: sbpIntegration,
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
