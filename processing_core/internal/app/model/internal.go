package model

import (
	"github.com/google/uuid"

	desc "processing_core/pkg/core"
)

type InternalDomainRequest struct {
	// uuid проводки
	ID uuid.UUID
	// транзакция
	Transaction *Transaction
	// uuid получателя
	ReceiverId uuid.UUID
}

func MapProtoInternalToDomain(req *desc.InternalRequest) *InternalDomainRequest {
	return &InternalDomainRequest{
		ID:          uuid.MustParse(req.Id),
		Transaction: MapProtoTransactionToDomain(req.Transaction),
		ReceiverId:  uuid.MustParse(req.ReceiverId),
	}
}
