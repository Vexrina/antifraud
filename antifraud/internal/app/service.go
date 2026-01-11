package app

import (
	"context"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"

	"antifraud/internal/app/model"
	desc "antifraud/pkg/antifraud"
)

//go:generate mockgen -source=service.go -destination=./mocks/checker_mock.go -package=mocks
type (
	Checker interface {
		Check(ctx context.Context, transaction *model.DomainTransaction) error
	}
	Service struct {
		desc.UnimplementedOnlineCheckServer

		sbpOut   Checker
		internal Checker
		cashOut  Checker
	}
)

func NewService(
	sbpOut Checker,
	internal Checker,
	cashOut Checker,
) *Service {
	return &Service{
		sbpOut:   sbpOut,
		internal: internal,
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
