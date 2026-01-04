package repository

import (
	"context"
	"fmt"
	"processing_core/generated/proc_core_db/public/model"
	"processing_core/generated/proc_core_db/public/table"
	"processing_core/internal/outbox"
	"processing_core/pkg/kafka_core"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/samber/lo"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// AppendOutbox аппендит запись в аутбокс, чтобы кафка его сожрала
func (db *pgDb) AppendOutbox(ctx context.Context, tx pgx.Tx, transaction model.TransactionsHistory) error {
	msg := mapTransactionDbModelToProto(transaction)
	payloadJSON, err := protojson.Marshal(msg)
	if err != nil {
		return err
	}

	insertedModel := model.Outbox{
		AggregateID: *transaction.TransactionID,
		EventType:   fmt.Sprint(outbox.OutboxEventType_Transaction),
		Payload:     string(payloadJSON),
		CreatedAt:   lo.ToPtr(time.Now()),
		Version:     lo.ToPtr(transaction.Revision),
	}

	stmt := table.Outbox.INSERT(
		table.Outbox.AllColumns.Except(
			table.Outbox.ID,
			table.Outbox.CreatedAt,
			table.Outbox.Published,
		),
	).MODEL(insertedModel)

	sql, args := stmt.Sql()
	_, err = tx.Exec(ctx, sql, args...)
	if err != nil {
		return err
	}

	return nil
}

func mapDbTransactionTypeToProto(transactionType *model.TransactionType) kafka_core.TransactionType {
	if transactionType == nil {
		return kafka_core.TransactionType_Unknown
	}
	switch *transactionType {
	case model.TransactionType_Internal:
		return kafka_core.TransactionType_Internal
	case model.TransactionType_CashIn:
		return kafka_core.TransactionType_CashIn
	case model.TransactionType_CashOut:
		return kafka_core.TransactionType_CashOut
	case model.TransactionType_SbpOutgoing:
		return kafka_core.TransactionType_SbpOutgoing
	default:
		return kafka_core.TransactionType_Unknown
	}
}

func mapTransactionDbModelToProto(transaction model.TransactionsHistory) *kafka_core.TransactionCore {
	var (
		receiverID  = lo.ToPtr("")
		atmID       = lo.ToPtr("")
		merchant    = ""
		country     = ""
		receiverBic = ""
	)

	if transaction.ReceiverID != nil {
		receiverID = lo.ToPtr(transaction.ReceiverID.String())
	}

	if transaction.AtmID != nil {
		atmID = lo.ToPtr(transaction.AtmID.String())
	}

	if transaction.Merchant != nil {
		merchant = *transaction.Merchant
	}

	if transaction.Country != nil {
		country = *transaction.Country
	}

	if transaction.ReceiverBic != nil {
		country = *transaction.ReceiverBic
	}

	msg := &kafka_core.TransactionCore{
		Id:              uint64(transaction.ID),
		TransactionId:   transaction.TransactionID.String(),
		CreatedAt:       timestamppb.New(*transaction.CreatedAt),
		Amount:          *transaction.Amount,
		Currency:        fmt.Sprint(transaction.Currency),
		Merchant:        merchant,
		Country:         country,
		SenderId:        transaction.SenderID.String(),
		ReceiverId:      receiverID,
		ReceiverBic:     &receiverBic,
		AtmId:           atmID,
		TransactionType: mapDbTransactionTypeToProto(transaction.TransactionType),
		Revision:        transaction.Revision,
	}

	return msg
}
