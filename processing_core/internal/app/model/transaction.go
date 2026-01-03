package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/samber/lo"

	desc "processing_core/pkg/core"
)

type Transaction struct {
	ID            uint64
	TransactionID uuid.UUID
	CreatedAt     *time.Time
	Amount        int64
	Currency      string
	Merchant      string
	Country       string
	SenderID      uuid.UUID
}

func MapProtoTransactionToDomain(req *desc.Transaction) *Transaction {
	return &Transaction{
		ID:            req.Id,
		TransactionID: uuid.MustParse(req.TransactionId),
		CreatedAt:     lo.ToPtr(req.CreatedAt.AsTime()),
		Amount:        req.Amount,
		Currency:      req.Currency,
		Merchant:      req.Merchant,
		Country:       req.Country,
		SenderID:      uuid.MustParse(req.SenderId),
	}
}
