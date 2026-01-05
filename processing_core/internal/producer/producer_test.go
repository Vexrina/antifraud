package producer

import (
	"context"
	"processing_core/generated/proc_core_db/public/model"
	"testing"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_kafkaProducer_Produce(t *testing.T) {
	topic := "test-topic"
	pr, err := kafka.NewProducer(
		&kafka.ConfigMap{
			"bootstrap.servers": "localhost:19092",
		},
	)
	require.NoError(t, err)
	p := &kafkaProducer{
		producer: pr,
		topic:    &topic,
	}

	outbox := []model.Outbox{
		{
			ID:          1,
			AggregateID: uuid.New(),
			Payload:     "hello",
		},
		{
			ID:          2,
			AggregateID: uuid.New(),
			Payload:     "world",
		},
	}
	ctx := context.Background()
	ids, err := p.Produce(ctx, outbox)
	require.NoError(t, err)
	require.Len(t, ids, 2)
	assert.ElementsMatch(t, []int64{1, 2}, ids)
}
