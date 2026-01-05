package outbox

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"

	"processing_core/generated/proc_core_db/public/model"
)

const TransactionTopic = "best_bank_transactions"

//go:generate mockgen -source=outbox.go -destination=./mocks/publisher.go -package=mocks
type (
	KafkaProducer interface {
		Produce(ctx context.Context, outbox []model.Outbox) ([]int64, error)
	}
	CommonDb interface {
		Transactional(ctx context.Context, f func(tx pgx.Tx) error) error
	}
	KafkaDb interface {
		GetUnpublishedMessages(ctx context.Context, tx pgx.Tx) ([]model.Outbox, error)
		MarkMessagesAsProcessed(ctx context.Context, tx pgx.Tx, ids []int64) error
	}
	KafkaOutboxPublisher struct {
		db       KafkaDb
		commonDb CommonDb
		producer KafkaProducer
	}
)

func NewKafkaOutboxPublisher(db KafkaDb, commonDb CommonDb, producer KafkaProducer) *KafkaOutboxPublisher {
	return &KafkaOutboxPublisher{
		db:       db,
		commonDb: commonDb,
		producer: producer,
	}
}

func (p *KafkaOutboxPublisher) Run(ctx context.Context) error {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			err := p.Publish(ctx)
			if err != nil {
				// todo log
				continue
			}
		}
	}
}

func (p *KafkaOutboxPublisher) Publish(ctx context.Context) error {
	var (
		msgs  []model.Outbox
		txErr error
		err   error
	)
	err = p.commonDb.Transactional(ctx, func(tx pgx.Tx) error {
		msgs, txErr = p.db.GetUnpublishedMessages(ctx, tx)
		return txErr
	})
	if err != nil {
		return err
	}
	if len(msgs) == 0 {
		return nil
	}

	processedIDs, err := p.producer.Produce(ctx, msgs)
	if err != nil {
		return err
	}

	err = p.commonDb.Transactional(ctx, func(tx pgx.Tx) error {
		txErr = p.db.MarkMessagesAsProcessed(ctx, tx, processedIDs)
		if txErr != nil {
			return txErr
		}
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
