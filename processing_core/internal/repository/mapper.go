package repository

import (
	dbModel "processing_core/generated/proc_core_db/public/model"
	"processing_core/internal/app/model"

	"github.com/samber/lo"
)

func MapCashOutDomainToTransaction(domain model.CashOutDomainRequest) dbModel.TransactionsHistory {
	return dbModel.TransactionsHistory{
		ID:              int32(domain.Transaction.ID),
		TransactionID:   lo.ToPtr(domain.Transaction.TransactionID),
		CreatedAt:       domain.Transaction.CreatedAt,
		Amount:          lo.ToPtr(domain.Transaction.Amount),
		Currency:        lo.ToPtr(domain.Transaction.Currency),
		Merchant:        lo.ToPtr(domain.Transaction.Merchant),
		Country:         lo.ToPtr(domain.Transaction.Country),
		SenderID:        lo.ToPtr(domain.Transaction.SenderID),
		ReceiverID:      nil,
		ReceiverBic:     nil,
		AtmID:           lo.ToPtr(domain.AtmID),
		TransactionType: lo.ToPtr(dbModel.TransactionType_CashOut),
	}
}

func MapCashInDomainToTransaction(domain model.CashInDomainRequest) dbModel.TransactionsHistory {
	return dbModel.TransactionsHistory{
		ID:              int32(domain.Transaction.ID),
		TransactionID:   lo.ToPtr(domain.Transaction.TransactionID),
		CreatedAt:       domain.Transaction.CreatedAt,
		Amount:          lo.ToPtr(domain.Transaction.Amount),
		Currency:        lo.ToPtr(domain.Transaction.Currency),
		Merchant:        lo.ToPtr(domain.Transaction.Merchant),
		Country:         lo.ToPtr(domain.Transaction.Country),
		SenderID:        lo.ToPtr(domain.Transaction.SenderID),
		ReceiverID:      nil,
		ReceiverBic:     nil,
		AtmID:           lo.ToPtr(domain.AtmID),
		TransactionType: lo.ToPtr(dbModel.TransactionType_CashIn),
	}
}

func MapSbpOutgoingDomainToTransaction(domain model.SbpOutgoingDomainRequest) dbModel.TransactionsHistory {
	return dbModel.TransactionsHistory{
		ID:              int32(domain.Transaction.ID),
		TransactionID:   lo.ToPtr(domain.Transaction.TransactionID),
		CreatedAt:       domain.Transaction.CreatedAt,
		Amount:          lo.ToPtr(domain.Transaction.Amount),
		Currency:        lo.ToPtr(domain.Transaction.Currency),
		Merchant:        lo.ToPtr(domain.Transaction.Merchant),
		Country:         lo.ToPtr(domain.Transaction.Country),
		SenderID:        lo.ToPtr(domain.Transaction.SenderID),
		ReceiverID:      lo.ToPtr(domain.ReceiverId),
		ReceiverBic:     lo.ToPtr(domain.Bic),
		AtmID:           nil,
		TransactionType: lo.ToPtr(dbModel.TransactionType_SbpOutgoing),
	}
}

func MapInternalDomainToTransaction(domain model.InternalDomainRequest) dbModel.TransactionsHistory {
	return dbModel.TransactionsHistory{
		ID:              int32(domain.Transaction.ID),
		TransactionID:   lo.ToPtr(domain.Transaction.TransactionID),
		CreatedAt:       domain.Transaction.CreatedAt,
		Amount:          lo.ToPtr(domain.Transaction.Amount),
		Currency:        lo.ToPtr(domain.Transaction.Currency),
		Merchant:        lo.ToPtr(domain.Transaction.Merchant),
		Country:         lo.ToPtr(domain.Transaction.Country),
		SenderID:        lo.ToPtr(domain.Transaction.SenderID),
		ReceiverID:      lo.ToPtr(domain.ReceiverId),
		ReceiverBic:     nil,
		AtmID:           nil,
		TransactionType: lo.ToPtr(dbModel.TransactionType_Internal),
	}
}
