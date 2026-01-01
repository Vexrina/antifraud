package app

import (
	"context"

	"processing_core/pkg/core"
)

//go:generate mockgen -source=service.go -destination=./mocks/sbpOutMock.go -package=mocks SbpOutgoingOperations
//go:generate mockgen -source=service.go -destination=./mocks/internalMock.go -package=mocks InternalOperations
//go:generate mockgen -source=service.go -destination=./mocks/cashInMock.go -package=mocks CashInOperations
//go:generate mockgen -source=service.go -destination=./mocks/casgOutMock.go -package=mocks CashOutOperations
type (
	SbpOutgoingOperations interface {
		Process(ctx context.Context, domainRequest any) error
	}
	InternalOperations interface {
		Process(ctx context.Context, domainRequest any) error
	}
	CashInOperations interface {
		Process(ctx context.Context, domainRequest any) error
	}
	CashOutOperations interface {
		Process(ctx context.Context, domainRequest any) error
	}
)

type Service struct {
	core.UnimplementedCoreServer

	sbpOut   SbpOutgoingOperations
	internal InternalOperations
	cashIn   CashInOperations
	cashOut  CashOutOperations
}

func NewService(
	sbpOut SbpOutgoingOperations,
	internal InternalOperations,
	cashIn CashInOperations,
	cashOut CashOutOperations,
) *Service {
	return &Service{
		sbpOut:   sbpOut,
		internal: internal,
		cashIn:   cashIn,
		cashOut:  cashOut,
	}
}
