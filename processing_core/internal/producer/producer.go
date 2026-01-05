package producer

import (
	"context"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"

	"processing_core/generated/proc_core_db/public/model"
)

func NewKafkaProducer(p *kafka.Producer, topic *string) *kafkaProducer {
	return &kafkaProducer{
		producer: p,
		topic:    topic,
	}
}

type kafkaProducer struct {
	producer *kafka.Producer
	topic    *string
}

func (k *kafkaProducer) Produce(
	ctx context.Context,
	outbox []model.Outbox,
) ([]int64, error) {

	deliveryChan := make(chan kafka.Event, len(outbox))

	for _, out := range outbox {
		err := k.producer.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{
				Topic:     k.topic,
				Partition: kafka.PartitionAny,
			},
			Value: []byte(out.Payload),
			Key:   []byte(out.AggregateID.String()),
		}, deliveryChan)

		if err != nil {
			return nil, err
		}
	}

	var delivered int
	for delivered < len(outbox) {
		select {
		case e := <-deliveryChan:
			m := e.(*kafka.Message)
			if m.TopicPartition.Error != nil {
				return nil, m.TopicPartition.Error
			}
			delivered++
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	k.producer.Flush(1_000)

	ids := make([]int64, 0, len(outbox))
	for _, o := range outbox {
		ids = append(ids, o.ID)
	}

	return ids, nil
}
