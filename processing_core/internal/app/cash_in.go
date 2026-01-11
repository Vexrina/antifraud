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

func (s *Service) CashIn(ctx context.Context, req *desc.CashInRequest) (*desc.CashInResponse, error) {
	if err := validateCashInRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	domainRequest := model.MapProtoCashInToDomain(req)
	resp, err := s.cashIn.Process(ctx, domainRequest)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return resp, nil
}

func validateCashInRequest(req *desc.CashInRequest) error {
	transactionErr := validateTransaction(req.Transaction)
	if transactionErr != nil {
		return transactionErr
	}
	return validation.ValidateStruct(req,
		validation.Field(&req.Id, validation.Required, is.UUID),
		validation.Field(&req.AtmId, validation.Required, is.UUID),
	)
}
