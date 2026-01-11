package app

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"antifraud/internal/app/mocks"
	desc "antifraud/pkg/antifraud"
)

func TestService_CashOut(t *testing.T) {
	t.Parallel()
	type fields struct {
		cashOut *mocks.MockChecker
	}
	type args struct {
		ctx context.Context
		req *desc.Transaction
	}

	validReq := &desc.Transaction{
		Id:            1,
		TransactionId: uuid.NewString(),
		CreatedAt:     timestamppb.Now(),
		Amount:        500_00,
		Currency:      "USD",
		Merchant:      "merchant",
		Country:       "US",
		SenderId:      uuid.NewString(),
	}

	tests := []struct {
		name            string
		fields          func(ctrl *gomock.Controller) *fields
		args            args
		wantErr         bool
		errCode         codes.Code
		wantedNewStatus desc.OperationStatus
	}{
		{
			name: "validation error",
			fields: func(ctrl *gomock.Controller) *fields {
				return &fields{}
			},
			args: args{
				ctx: context.Background(),
				req: &desc.Transaction{
					TransactionId: "",
				},
			},
			wantErr:         true,
			errCode:         codes.InvalidArgument,
			wantedNewStatus: desc.OperationStatus_Approved,
		},
		{
			name: "cashOut check fails",
			fields: func(ctrl *gomock.Controller) *fields {
				mockCheck := mocks.NewMockChecker(ctrl)
				mockCheck.EXPECT().Check(gomock.Any(), gomock.Any()).Return(errors.New("someerr"))
				return &fields{cashOut: mockCheck}
			},
			args: args{
				ctx: context.Background(),
				req: validReq,
			},
			wantErr:         true,
			errCode:         codes.FailedPrecondition,
			wantedNewStatus: desc.OperationStatus_Approved,
		},
		{
			name: "cashOut check declined",
			fields: func(ctrl *gomock.Controller) *fields {
				mockCheck := mocks.NewMockChecker(ctrl)
				mockCheck.EXPECT().Check(gomock.Any(), gomock.Any()).Return(errors.New("declined"))
				return &fields{cashOut: mockCheck}
			},
			args: args{
				ctx: context.Background(),
				req: validReq,
			},
			wantErr:         false,
			wantedNewStatus: desc.OperationStatus_Declined,
		},
		{
			name: "success",
			fields: func(ctrl *gomock.Controller) *fields {
				mockCheck := mocks.NewMockChecker(ctrl)
				mockCheck.EXPECT().Check(gomock.Any(), gomock.Any()).Return(nil)
				return &fields{cashOut: mockCheck}
			},
			args: args{
				ctx: context.Background(),
				req: validReq,
			},
			wantErr:         false,
			wantedNewStatus: desc.OperationStatus_Approved,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			f := tt.fields(ctrl)
			s := &Service{
				cashOut: f.cashOut,
			}
			got, err := s.CashOut(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Fatalf("CashOut() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				st, ok := status.FromError(err)
				if !ok {
					t.Fatalf("expected gRPC status error, got %v", err)
				}
				if st.Code() != tt.errCode {
					t.Errorf("CashOut() error code = %v, want %v", st.Code(), tt.errCode)
				}
				return
			}
			if got.NewStatus != tt.wantedNewStatus {
				t.Errorf("CashOut() = %v, want %v", got.NewStatus, tt.wantedNewStatus)
			}
		})
	}
}
