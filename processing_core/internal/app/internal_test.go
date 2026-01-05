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

func TestService_Internal(t *testing.T) {
	t.Parallel()
	type args struct {
		ctx context.Context
		req *desc.InternalRequest
	}
	type fields struct {
		internal func(ctrl *gomock.Controller, a *args) InternalOperations
	}
	now := time.Now()
	tests := []struct {
		name        string
		fields      fields
		args        args
		want        *desc.InternalResponse
		expectedErr string
	}{
		{
			name: "success",
			fields: fields{
				internal: func(ctrl *gomock.Controller, a *args) InternalOperations {
					m := mocks.NewMockInternalOperations(ctrl)
					m.EXPECT().
						Process(gomock.Any(), gomock.Any()).
						DoAndReturn(
							func(_ context.Context, req *model.InternalDomainRequest) (*desc.InternalResponse, error) {
								mod := model.InternalDomainRequest{
									ID: uuid.MustParse(a.req.Id),
									Transaction: &model.Transaction{
										ID:            123,
										TransactionID: uuid.MustParse(a.req.Transaction.TransactionId),
										CreatedAt:     &now,
										Amount:        200_00,
										Currency:      123,
										Merchant:      "213",
										Country:       "321",
										SenderID:      uuid.MustParse(a.req.Transaction.SenderId),
									},
									ReceiverId: uuid.MustParse(a.req.ReceiverId),
								}
								if !cmp.Equal(&mod, req) {
									t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(&mod, req))
								}
								return &desc.InternalResponse{
									NewStatus: desc.OperationStatus_Approved,
								}, nil
							},
						)
					return m
				},
			},
			args: args{
				ctx: context.Background(),
				req: &desc.InternalRequest{
					Id: uuid.NewString(),
					Transaction: &desc.Transaction{
						Id:            123,
						TransactionId: uuid.NewString(),
						CreatedAt:     timestamppb.New(now),
						Amount:        200_00,
						Currency:      123,
						Merchant:      "213",
						Country:       "321",
						SenderId:      uuid.NewString(),
					},
					ReceiverId: uuid.NewString(),
				},
			},
			want: &desc.InternalResponse{
				NewStatus: desc.OperationStatus_Approved,
			},
			expectedErr: "",
		},
		{
			name: "error in usecase",
			fields: fields{
				internal: func(ctrl *gomock.Controller, a *args) InternalOperations {
					m := mocks.NewMockInternalOperations(ctrl)
					m.EXPECT().
						Process(gomock.Any(), gomock.Any()).
						DoAndReturn(
							func(_ context.Context, _ *model.InternalDomainRequest) (*desc.InternalResponse, error) {
								return &desc.InternalResponse{
									NewStatus: desc.OperationStatus_Approved,
								}, errors.New("usecase")
							},
						)
					return m
				},
			},
			args: args{
				ctx: context.Background(),
				req: &desc.InternalRequest{
					Id: uuid.NewString(),
					Transaction: &desc.Transaction{
						Id:            123,
						TransactionId: uuid.NewString(),
						CreatedAt:     timestamppb.New(now),
						Amount:        200_00,
						Currency:      123,
						Merchant:      "213",
						Country:       "321",
						SenderId:      uuid.NewString(),
					},
					ReceiverId: uuid.NewString(),
				},
			},
			want:        nil,
			expectedErr: "rpc error: code = Internal desc = usecase",
		},
		{
			name: "error in validation",
			fields: fields{
				internal: func(ctrl *gomock.Controller, a *args) InternalOperations {
					m := mocks.NewMockInternalOperations(ctrl)
					return m
				},
			},
			args: args{
				ctx: context.Background(),
				req: &desc.InternalRequest{
					Id: "123", // неверный UUID
					Transaction: &desc.Transaction{
						Id:            123,
						TransactionId: uuid.NewString(),
						CreatedAt:     timestamppb.New(now),
						Amount:        200_00,
						Currency:      123,
						Merchant:      "213",
						Country:       "321",
						SenderId:      uuid.NewString(),
					},
					ReceiverId: uuid.NewString(),
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
				internal: tt.fields.internal(ctrl, &tt.args),
			}
			got, err := s.Internal(tt.args.ctx, tt.args.req)
			if tt.expectedErr != "" {
				require.Error(t, err)
				assert.Equal(t, tt.expectedErr, err.Error())
				return
			}
			require.NoError(t, err)
			if !cmp.Equal(got, tt.want, protocmp.Transform()) {
				t.Errorf("Internal() diffrents answers: %s", cmp.Diff(got, tt.want, protocmp.Transform()))
			}
		})
	}
}
