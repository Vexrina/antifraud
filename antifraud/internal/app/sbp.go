package app

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"antifraud/internal/app/model"
	desc "antifraud/pkg/antifraud"
)

func (s *Service) SbpOutgoing(ctx context.Context, req *desc.Transaction) (*desc.CheckResult, error) {
	err := validateTransaction(req)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	domainTx := model.MapProtoToDomainTransaction(req)
	err = s.sbpOut.Check(ctx, domainTx)
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}
	return &desc.CheckResult{NewStatus: desc.OperationStatus_Approved}, nil
}
