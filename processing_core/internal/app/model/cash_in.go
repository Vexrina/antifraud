package model

import (
	desc "processing_core/pkg/core"

	"github.com/google/uuid"
)

type CashInDomainRequest struct {
	ID          uuid.UUID
	Transaction Transaction
	AtmID       uuid.UUID
}

func MapProtoCashInToDomain(req *desc.CashInRequest) *CashInDomainRequest {
	return &CashInDomainRequest{
		ID:          uuid.MustParse(req.Id),
		Transaction: *MapProtoTransactionToDomain(req.Transaction),
		AtmID:       uuid.MustParse(req.AtmId),
	}
}
