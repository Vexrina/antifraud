package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/samber/lo"

	desc "antifraud/pkg/antifraud"
)

type DomainTransaction struct {
	CreatedAt  *time.Time
	Amount     int64
	Currency   string
	Merchant   string
	Country    string
	SenderID   uuid.UUID
	ReceiverID *uuid.UUID
	BIC        *string
	AtmID      *uuid.UUID
}

func MapProtoToDomainTransaction(transaction *desc.Transaction) *DomainTransaction {
	var (
		receiverID *uuid.UUID
		bic        *string
		atmID      *uuid.UUID
	)

	if transaction.ReceiverId != nil {
		receiverID = lo.ToPtr(uuid.MustParse(*transaction.ReceiverId))
	}
	if transaction.AtmId != nil {
		atmID = lo.ToPtr(uuid.MustParse(*transaction.AtmId))
	}
	if transaction.Bic != nil {
		bic = lo.ToPtr(*transaction.Bic)
	}

	return &DomainTransaction{
		CreatedAt:  lo.ToPtr(transaction.CreatedAt.AsTime()),
		Amount:     transaction.Amount,
		Currency:   transaction.Currency,
		Merchant:   transaction.Merchant,
		Country:    transaction.Country,
		SenderID:   uuid.MustParse(transaction.SenderId),
		ReceiverID: receiverID,
		BIC:        bic,
		AtmID:      atmID,
	}
}
