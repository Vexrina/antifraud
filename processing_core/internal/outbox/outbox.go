package outbox

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"

	"processing_core/generated/proc_core_db/public/model"
)

const (
	KafkaOutboxTopic_Unknown     = "unknown"
	KafkaOutboxTopic_Transaction = "best_bank_transactions"
)

type (
	OutboxPublisher interface {
		Run(ctx context.Context) error
		Publish(ctx context.Context) error
	}
	KafkaProducer interface {
		Produce(ctx context.Context, outbox []model.Outbox) ([]int64, error)
	}
	CommonDb interface {
		Transactional(ctx context.Context, f func(tx pgx.Tx) error) error
	}
	OutboxDb interface {
		GetUnpublisedMessages(ctx context.Context) ([]model.Outbox, error)
		MarkMessagesAsProcessed(ctx context.Context, id []int64) error
	}
	KafkaOutboxPublisher struct {
		db       OutboxDb
		commonDb CommonDb
		producer KafkaProducer
	}
)

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
	msgs, err := p.db.GetUnpublisedMessages(ctx)
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
		txErr := p.db.MarkMessagesAsProcessed(ctx, processedIDs)
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
