package model

import (
	"github.com/google/uuid"

	desc "processing_core/pkg/core"
)

type CashOutDomainRequest struct {
	// uuid проводки
	ID uuid.UUID
	// транзакция
	Transaction *Transaction
	// uuid банкомата
	AtmID uuid.UUID
}

func MapProtoCashOutToDomain(req *desc.CashOutRequest) *CashOutDomainRequest {
	return &CashOutDomainRequest{
		ID:          uuid.MustParse(req.Id),
		Transaction: MapProtoTransactionToDomain(req.Transaction),
		AtmID:       uuid.MustParse(req.AtmId),
	}
}
