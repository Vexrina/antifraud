package app

import (
	"context"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"

	"processing_core/internal/app/model"
	desc "processing_core/pkg/core"
)

//go:generate mockgen -source=service.go -destination=./mocks/service_mock.go -package=mocks
type (
	SbpOutgoingOperations interface {
		Process(ctx context.Context, domainRequest *desc.SbpOutgoingRequest) (*desc.SbpOutgoingResponse, error)
	}
	InternalOperations interface {
		Process(ctx context.Context, domainRequest *desc.InternalRequest) (*desc.InternalResponse, error)
	}
	CashInOperations interface {
		Process(ctx context.Context, domainRequest *model.CashInDomainRequest) (*desc.CashInResponse, error)
	}
	CashOutOperations interface {
		Process(ctx context.Context, domainRequest *desc.CashOutRequest) (*desc.CashOutResponse, error)
	}
)

type Service struct {
	desc.UnimplementedCoreServer

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

func validateTransaction(transaction *desc.Transaction) error {
	return validation.ValidateStruct(transaction,
		validation.Field(&transaction.Id, validation.Required),
		validation.Field(&transaction.TransactionId, validation.Required, is.UUID),
		validation.Field(&transaction.CreatedAt, validation.Required),
		validation.Field(&transaction.Amount, validation.Required),
		validation.Field(&transaction.Currency, validation.Required),
		validation.Field(&transaction.Merchant, validation.Required),
		validation.Field(&transaction.Country, validation.Required),
		validation.Field(&transaction.SenderId, validation.Required, is.UUID),
	)
}
