package app

import (
	"context"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"processing_core/internal/app/model"
	desc "processing_core/pkg/core"
)

func (s *Service) SbpOutgoing(ctx context.Context, req *desc.SbpOutgoingRequest) (*desc.SbpOutgoingResponse, error) {
	if err := validateSbpOutgoingRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	domainRequest := model.MapProtoSbpOutToDomain(req)
	resp, err := s.sbpOut.Process(ctx, domainRequest)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return resp, nil
}

func validateSbpOutgoingRequest(req *desc.SbpOutgoingRequest) error {
	transactionErr := validateTransaction(req.Transaction)
	if transactionErr != nil {
		return transactionErr
	}
	return validation.ValidateStruct(req,
		validation.Field(&req.Id, validation.Required, is.UUID),
		validation.Field(&req.ReceiverId, validation.Required, is.UUID),
		validation.Field(&req.Bic, validation.Required),
	)
}
