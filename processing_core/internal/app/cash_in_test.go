package app

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/timestamppb"

	"processing_core/internal/app/mocks"
	"processing_core/internal/app/model"
	desc "processing_core/pkg/core"
)

func TestService_CashIn(t *testing.T) {
	t.Parallel()
	type args struct {
		ctx context.Context
		req *desc.CashInRequest
	}
	type fields struct {
		cashIn func(ctrl *gomock.Controller, a *args) CashInOperations
	}
	now := time.Now()
	tests := []struct {
		name        string
		fields      fields
		args        args
		want        *desc.CashInResponse
		expectedErr string
	}{
		{
			name: "success",
			fields: fields{
				cashIn: func(ctrl *gomock.Controller, a *args) CashInOperations {
					m := mocks.NewMockCashInOperations(ctrl)
					m.EXPECT().
						Process(gomock.Any(), gomock.Any()).
						DoAndReturn(
							func(_ context.Context, req *model.CashInDomainRequest) (*desc.CashInResponse, error) {
								mod := model.CashInDomainRequest{
									ID: uuid.MustParse(a.req.Id),
									Transaction: model.Transaction{
										ID:            123,
										TransactionID: uuid.MustParse(a.req.Transaction.TransactionId),
										CreatedAt:     &now,
										Amount:        100_00,
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
								return &desc.CashInResponse{
									NewStatus: desc.OperationStatus_Approved,
								}, nil
							},
						)
					return m
				},
			},
			args: args{
				ctx: context.Background(),
				req: &desc.CashInRequest{
					Id: uuid.NewString(),
					Transaction: &desc.Transaction{
						Id:            123,
						TransactionId: uuid.NewString(),
						CreatedAt:     timestamppb.New(now),
						Amount:        100_00,
						Currency:      123,
						Merchant:      "213",
						Country:       "321",
						SenderId:      uuid.NewString(),
					},
					AtmId: uuid.NewString(),
				},
			},
			want: &desc.CashInResponse{
				NewStatus: desc.OperationStatus_Approved,
			},
			expectedErr: "",
		},
		{
			name: "error in usecase",
			fields: fields{
				cashIn: func(ctrl *gomock.Controller, a *args) CashInOperations {
					m := mocks.NewMockCashInOperations(ctrl)
					m.EXPECT().
						Process(gomock.Any(), gomock.Any()).
						DoAndReturn(
							func(_ context.Context, _ *model.CashInDomainRequest) (*desc.CashInResponse, error) {
								return &desc.CashInResponse{
									NewStatus: desc.OperationStatus_Approved,
								}, errors.New("usecase")
							},
						)
					return m
				},
			},
			args: args{
				ctx: context.Background(),
				req: &desc.CashInRequest{
					Id: uuid.NewString(),
					Transaction: &desc.Transaction{
						Id:            123,
						TransactionId: uuid.NewString(),
						CreatedAt:     timestamppb.New(now),
						Amount:        100_00,
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
				cashIn: func(ctrl *gomock.Controller, a *args) CashInOperations {
					m := mocks.NewMockCashInOperations(ctrl)
					return m
				},
			},
			args: args{
				ctx: context.Background(),
				req: &desc.CashInRequest{
					Id: "123",
					Transaction: &desc.Transaction{
						Id:            123,
						TransactionId: uuid.NewString(),
						CreatedAt:     timestamppb.New(now),
						Amount:        100_00,
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
				cashIn: tt.fields.cashIn(ctrl, &tt.args),
			}
			got, err := s.CashIn(tt.args.ctx, tt.args.req)
			if tt.expectedErr != "" {
				require.Error(t, err)
				assert.Equal(t, tt.expectedErr, err.Error())
				return
			}
			require.NoError(t, err)
			if !cmp.Equal(got, tt.want, protocmp.Transform()) {
				t.Errorf("CashIn() diffrents answers: %s", cmp.Diff(got, tt.want, protocmp.Transform()))
			}
		})
	}
}
