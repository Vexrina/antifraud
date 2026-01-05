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

func TestService_SbpOutgoing(t *testing.T) {
	t.Parallel()
	type args struct {
		ctx context.Context
		req *desc.SbpOutgoingRequest
	}
	type fields struct {
		sbpOut func(ctrl *gomock.Controller, a *args) SbpOutgoingOperations
	}
	now := time.Now()
	tests := []struct {
		name        string
		fields      fields
		args        args
		want        *desc.SbpOutgoingResponse
		expectedErr string
	}{
		{
			name: "success",
			fields: fields{
				sbpOut: func(ctrl *gomock.Controller, a *args) SbpOutgoingOperations {
					m := mocks.NewMockSbpOutgoingOperations(ctrl)
					m.EXPECT().
						Process(gomock.Any(), gomock.Any()).
						DoAndReturn(
							func(_ context.Context, req *model.SbpOutgoingDomainRequest) (*desc.SbpOutgoingResponse, error) {
								mod := model.SbpOutgoingDomainRequest{
									ID: uuid.MustParse(a.req.Id),
									Transaction: &model.Transaction{
										ID:            123,
										TransactionID: uuid.MustParse(a.req.Transaction.TransactionId),
										CreatedAt:     &now,
										Amount:        300_00,
										Currency:      123,
										Merchant:      "213",
										Country:       "321",
										SenderID:      uuid.MustParse(a.req.Transaction.SenderId),
									},
									ReceiverId: uuid.MustParse(a.req.ReceiverId),
									Bic:        a.req.Bic,
								}
								if !cmp.Equal(&mod, req) {
									t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(&mod, req))
								}
								return &desc.SbpOutgoingResponse{
									NewStatus: desc.OperationStatus_Approved,
								}, nil
							},
						)
					return m
				},
			},
			args: args{
				ctx: context.Background(),
				req: &desc.SbpOutgoingRequest{
					Id: uuid.NewString(),
					Transaction: &desc.Transaction{
						Id:            123,
						TransactionId: uuid.NewString(),
						CreatedAt:     timestamppb.New(now),
						Amount:        300_00,
						Currency:      123,
						Merchant:      "213",
						Country:       "321",
						SenderId:      uuid.NewString(),
					},
					ReceiverId: uuid.NewString(),
					Bic:        "044525225",
				},
			},
			want: &desc.SbpOutgoingResponse{
				NewStatus: desc.OperationStatus_Approved,
			},
			expectedErr: "",
		},
		{
			name: "error in usecase",
			fields: fields{
				sbpOut: func(ctrl *gomock.Controller, a *args) SbpOutgoingOperations {
					m := mocks.NewMockSbpOutgoingOperations(ctrl)
					m.EXPECT().
						Process(gomock.Any(), gomock.Any()).
						DoAndReturn(
							func(_ context.Context, _ *model.SbpOutgoingDomainRequest) (*desc.SbpOutgoingResponse, error) {
								return &desc.SbpOutgoingResponse{
									NewStatus: desc.OperationStatus_Approved,
								}, errors.New("usecase")
							},
						)
					return m
				},
			},
			args: args{
				ctx: context.Background(),
				req: &desc.SbpOutgoingRequest{
					Id: uuid.NewString(),
					Transaction: &desc.Transaction{
						Id:            123,
						TransactionId: uuid.NewString(),
						CreatedAt:     timestamppb.New(now),
						Amount:        300_00,
						Currency:      123,
						Merchant:      "213",
						Country:       "321",
						SenderId:      uuid.NewString(),
					},
					ReceiverId: uuid.NewString(),
					Bic:        "044525225",
				},
			},
			want:        nil,
			expectedErr: "rpc error: code = Internal desc = usecase",
		},
		{
			name: "error in validation (UUID id)",
			fields: fields{
				sbpOut: func(ctrl *gomock.Controller, a *args) SbpOutgoingOperations {
					m := mocks.NewMockSbpOutgoingOperations(ctrl)
					return m
				},
			},
			args: args{
				ctx: context.Background(),
				req: &desc.SbpOutgoingRequest{
					Id: "123", // неверный UUID
					Transaction: &desc.Transaction{
						Id:            123,
						TransactionId: uuid.NewString(),
						CreatedAt:     timestamppb.New(now),
						Amount:        300_00,
						Currency:      123,
						Merchant:      "213",
						Country:       "321",
						SenderId:      uuid.NewString(),
					},
					ReceiverId: uuid.NewString(),
					Bic:        "044525225",
				},
			},
			want:        nil,
			expectedErr: "rpc error: code = InvalidArgument desc = id: must be a valid UUID.",
		},
		{
			name: "error in validation (transaction)",
			fields: fields{
				sbpOut: func(ctrl *gomock.Controller, a *args) SbpOutgoingOperations {
					m := mocks.NewMockSbpOutgoingOperations(ctrl)
					return m
				},
			},
			args: args{
				ctx: context.Background(),
				req: &desc.SbpOutgoingRequest{
					Id: uuid.NewString(),
					Transaction: &desc.Transaction{
						Id:            0, // пропущено required поле
						TransactionId: "invalid-uuid",
						CreatedAt:     nil,
						Amount:        0,
						Currency:      0,
						Merchant:      "",
						Country:       "",
						SenderId:      "invalid-uuid",
					},
					ReceiverId: uuid.NewString(),
					Bic:        "044525225",
				},
			},
			want:        nil,
			expectedErr: "rpc error: code = InvalidArgument desc = amount: cannot be blank; country: cannot be blank; created_at: cannot be blank; currency: cannot be blank; id: cannot be blank; merchant: cannot be blank; sender_id: must be a valid UUID; transaction_id: must be a valid UUID.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			s := &Service{
				sbpOut: tt.fields.sbpOut(ctrl, &tt.args),
			}
			got, err := s.SbpOutgoing(tt.args.ctx, tt.args.req)
			if tt.expectedErr != "" {
				require.Error(t, err)
				assert.Equal(t, tt.expectedErr, err.Error())
				return
			}
			require.NoError(t, err)
			if !cmp.Equal(got, tt.want, protocmp.Transform()) {
				t.Errorf("SbpOutgoing() diffrents answers: %s", cmp.Diff(got, tt.want, protocmp.Transform()))
			}
		})
	}
}
