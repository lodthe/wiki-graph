package taskqueue

import (
	"encoding/json"

	"github.com/google/uuid"
	zlog "github.com/rs/zerolog/log"
	"github.com/wagslane/go-rabbitmq"
)

type Consumer struct {
	consumer rabbitmq.Consumer

	queueName  string
	routingKey string
}

func NewConsumer(consumer rabbitmq.Consumer, queueName, routingKey string) *Consumer {
	return &Consumer{
		consumer:   consumer,
		queueName:  queueName,
		routingKey: routingKey,
	}
}

// StartConsuming consumes tasks and calls handler for each of them.
// If a message cannot be unmarshalled, it's discarded.
// If handler fails, the message is requeued.
// Otherwise, the message is acked.
func (c *Consumer) StartConsuming(handler func(taskID uuid.UUID) error) error {
	return c.consumer.StartConsuming(
		func(d rabbitmq.Delivery) rabbitmq.Action {
			var task Task
			err := json.Unmarshal(d.Body, &task)
			if err != nil {
				zlog.Error().Err(err).Interface("message", d).Msg("failed to unmarshal received message")
				return rabbitmq.NackDiscard
			}

			zlog.Info().Str("id", task.ID.String()).Msg("received a new task")

			err = handler(task.ID)
			if err != nil {
				zlog.Error().Err(err).Str("id", task.ID.String()).Msg("failed to process the task")
				return rabbitmq.NackRequeue
			}

			zlog.Info().Str("id", task.ID.String()).Msg("task handled")

			return rabbitmq.Ack
		},
		c.queueName,
		[]string{c.routingKey},
		rabbitmq.WithConsumeOptionsConcurrency(1),
		rabbitmq.WithConsumeOptionsQueueDurable,
	)
}
