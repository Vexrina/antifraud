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

func (s *Service) Internal(ctx context.Context, req *desc.InternalRequest) (*desc.InternalResponse, error) {
	if err := validateInternalRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	domainRequest := model.MapProtoInternalToDomain(req)
	resp, err := s.internal.Process(ctx, domainRequest)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return resp, nil
}

func validateInternalRequest(req *desc.InternalRequest) error {
	transactionErr := validateTransaction(req.Transaction)
	if transactionErr != nil {
		return transactionErr
	}
	return validation.ValidateStruct(req,
		validation.Field(&req.Id, validation.Required, is.UUID),
		validation.Field(&req.ReceiverId, validation.Required, is.UUID),
	)
}
