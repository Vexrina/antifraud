package model

import (
	"github.com/google/uuid"

	desc "processing_core/pkg/core"
)

type SbpOutgoingDomainRequest struct {
	// uuid проводки
	ID uuid.UUID
	// транзакция
	Transaction *Transaction
	// uuid получателя
	ReceiverId uuid.UUID
	// bic банка получателя
	Bic string
}

func MapProtoSbpOutToDomain(req *desc.SbpOutgoingRequest) *SbpOutgoingDomainRequest {
	return &SbpOutgoingDomainRequest{
		ID:          uuid.MustParse(req.Id),
		Transaction: MapProtoTransactionToDomain(req.Transaction),
		ReceiverId:  uuid.MustParse(req.ReceiverId),
		Bic:         req.Bic,
	}
}
