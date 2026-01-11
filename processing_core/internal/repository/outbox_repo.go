package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/go-jet/jet/v2/postgres"
	"github.com/jackc/pgx/v5"
	"github.com/samber/lo"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"

	"processing_core/generated/proc_core_db/public/model"
	"processing_core/generated/proc_core_db/public/table"
	"processing_core/internal/outbox"
	"processing_core/internal/utils"
	"processing_core/pkg/kafka_core"
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
		EventType:   outbox.EventTypeTransaction.String(),
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

func mapTransactionTypeDbToProto(transactionType *model.TransactionType) kafka_core.TransactionType {
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
		merchant    = ""
		country     = ""
		receiverID  = lo.ToPtr("")
		receiverBic = ""
		atmID       = lo.ToPtr("")
	)

	if transaction.Merchant != nil {
		merchant = *transaction.Merchant
	}
	if transaction.Country != nil {
		country = *transaction.Country
	}
	if transaction.ReceiverID != nil {
		receiverID = lo.ToPtr(transaction.ReceiverID.String())
	}
	if transaction.ReceiverBic != nil {
		receiverBic = *transaction.ReceiverBic
	}
	if transaction.AtmID != nil {
		atmID = lo.ToPtr(transaction.AtmID.String())
	}

	msg := &kafka_core.TransactionCore{
		Id:              uint64(transaction.ID),
		TransactionId:   transaction.TransactionID.String(),
		CreatedAt:       timestamppb.New(*transaction.CreatedAt),
		Amount:          *transaction.Amount,
		Currency:        fmt.Sprint(*transaction.Currency),
		Merchant:        merchant,
		Country:         country,
		SenderId:        transaction.SenderID.String(),
		ReceiverId:      receiverID,
		ReceiverBic:     &receiverBic,
		AtmId:           atmID,
		TransactionType: mapTransactionTypeDbToProto(transaction.TransactionType),
		Revision:        transaction.Revision,
	}

	return msg
}

func (db *pgDb) GetUnpublishedMessages(ctx context.Context, tx pgx.Tx) ([]model.Outbox, error) {
	t := table.Outbox
	stmt := table.Outbox.SELECT(
		t.AllColumns.As(""),
	).WHERE(
		t.Published.EQ(postgres.Bool(false)),
	).ORDER_BY(
		t.CreatedAt.ASC(),
	).LIMIT(500).FOR(
		postgres.UPDATE().SKIP_LOCKED(),
	)

	sql, args := stmt.Sql()
	rows, err := tx.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	msgs, err := pgx.CollectRows(rows, pgx.RowToStructByName[model.Outbox])
	if err != nil {
		return nil, err
	}
	return msgs, nil
}

func (db *pgDb) MarkMessagesAsProcessed(ctx context.Context, tx pgx.Tx, ids []int64) error {
	t := table.Outbox
	stmt := table.Outbox.UPDATE(
		t.Published,
	).SET(postgres.Bool(true)).WHERE(t.ID.IN(utils.IntegerArray(ids)...))

	sql, args := stmt.Sql()
	_, err := tx.Exec(ctx, sql, args...)
	return err
}
