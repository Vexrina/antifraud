package app

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/timestamppb"

	"processing_core/internal/app/mocks"
	"processing_core/internal/app/model"
	desc "processing_core/pkg/core"
)

func TestService_CashOut(t *testing.T) {
	t.Parallel()
	type args struct {
		ctx context.Context
		req *desc.CashOutRequest
	}
	type fields struct {
		cashOut func(ctrl *gomock.Controller, a *args) CashOutOperations
	}
	now := time.Now()
	tests := []struct {
		name        string
		fields      fields
		args        args
		want        *desc.CashOutResponse
		expectedErr string
	}{
		{
			name: "success",
			fields: fields{
				cashOut: func(ctrl *gomock.Controller, a *args) CashOutOperations {
					m := mocks.NewMockCashOutOperations(ctrl)
					m.EXPECT().
						Process(gomock.Any(), gomock.Any()).
						DoAndReturn(
							func(_ context.Context, req *model.CashOutDomainRequest) (*desc.CashOutResponse, error) {
								mod := model.CashOutDomainRequest{
									ID: uuid.MustParse(a.req.Id),
									Transaction: &model.Transaction{
										ID:            123,
										TransactionID: uuid.MustParse(a.req.Transaction.TransactionId),
										CreatedAt:     &now,
										Amount:        50_00,
										Currency:      123,
										Merchant:      "213",
										Country:       "321",
										SenderID:      uuid.MustParse(a.req.Transaction.SenderId),
									},
									AtmID: uuid.MustParse(a.req.AtmId),
								}
								if !cmp.Equal(&mod, req) {
									t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(&mod, req))
								}
								return &desc.CashOutResponse{
									NewStatus: desc.OperationStatus_Approved,
								}, nil
							},
						)
					return m
				},
			},
			args: args{
				ctx: context.Background(),
				req: &desc.CashOutRequest{
					Id: uuid.NewString(),
					Transaction: &desc.Transaction{
						Id:            123,
						TransactionId: uuid.NewString(),
						CreatedAt:     timestamppb.New(now),
						Amount:        50_00,
						Currency:      123,
						Merchant:      "213",
						Country:       "321",
						SenderId:      uuid.NewString(),
					},
					AtmId: uuid.NewString(),
				},
			},
			want: &desc.CashOutResponse{
				NewStatus: desc.OperationStatus_Approved,
			},
			expectedErr: "",
		},
		{
			name: "error in usecase",
			fields: fields{
				cashOut: func(ctrl *gomock.Controller, a *args) CashOutOperations {
					m := mocks.NewMockCashOutOperations(ctrl)
					m.EXPECT().
						Process(gomock.Any(), gomock.Any()).
						DoAndReturn(
							func(_ context.Context, _ *model.CashOutDomainRequest) (*desc.CashOutResponse, error) {
								return &desc.CashOutResponse{
									NewStatus: desc.OperationStatus_Approved,
								}, errors.New("usecase")
							},
						)
					return m
				},
			},
			args: args{
				ctx: context.Background(),
				req: &desc.CashOutRequest{
					Id: uuid.NewString(),
					Transaction: &desc.Transaction{
						Id:            123,
						TransactionId: uuid.NewString(),
						CreatedAt:     timestamppb.New(now),
						Amount:        50_00,
						Currency:      123,
						Merchant:      "213",
						Country:       "321",
						SenderId:      uuid.NewString(),
					},
					AtmId: uuid.NewString(),
				},
			},
			want:        nil,
			expectedErr: "rpc error: code = Internal desc = usecase",
		},
		{
			name: "error in validation",
			fields: fields{
				cashOut: func(ctrl *gomock.Controller, a *args) CashOutOperations {
					m := mocks.NewMockCashOutOperations(ctrl)
					return m
				},
			},
			args: args{
				ctx: context.Background(),
				req: &desc.CashOutRequest{
					Id: "123", // неверный UUID
					Transaction: &desc.Transaction{
						Id:            123,
						TransactionId: uuid.NewString(),
						CreatedAt:     timestamppb.New(now),
						Amount:        50_00,
						Currency:      123,
						Merchant:      "213",
						Country:       "321",
						SenderId:      uuid.NewString(),
					},
					AtmId: uuid.NewString(),
				},
			},
			want:        nil,
			expectedErr: "rpc error: code = InvalidArgument desc = id: must be a valid UUID.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			s := &Service{
				cashOut: tt.fields.cashOut(ctrl, &tt.args),
			}
			got, err := s.CashOut(tt.args.ctx, tt.args.req)
			if tt.expectedErr != "" {
				require.Error(t, err)
				assert.Equal(t, tt.expectedErr, err.Error())
				return
			}
			require.NoError(t, err)
			if !cmp.Equal(got, tt.want, protocmp.Transform()) {
				t.Errorf("CashOut() diffrents answers: %s", cmp.Diff(got, tt.want, protocmp.Transform()))
			}
		})
	}
}
