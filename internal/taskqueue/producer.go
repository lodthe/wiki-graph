package taskqueue

import (
	"encoding/json"

	"github.com/pkg/errors"
	zlog "github.com/rs/zerolog/log"
	"github.com/wagslane/go-rabbitmq"
)

type Producer struct {
	publisher *rabbitmq.Publisher

	exchangeName string
	routingKey   string
}

func NewProducer(publisher *rabbitmq.Publisher, exchangeName string, routingKey string) *Producer {
	return &Producer{
		publisher:    publisher,
		exchangeName: exchangeName,
		routingKey:   routingKey,
	}
}

func (p *Producer) Produce(task Task) error {
	data, err := json.Marshal(task)
	if err != nil {
		return errors.Wrap(err, "json marshalling failed")
	}

	err = p.publisher.Publish(
		data,
		[]string{p.routingKey},
		rabbitmq.WithPublishOptionsContentType("application/json"),
		rabbitmq.WithPublishOptionsPersistentDelivery,
		rabbitmq.WithPublishOptionsExchange(p.exchangeName),
	)
	if err != nil {
		return errors.Wrap(err, "AMQP publish failed")
	}

	zlog.Debug().Fields(map[string]interface{}{
		"id":          task.ID.String(),
		"exchange":    p.exchangeName,
		"routing_key": p.routingKey,
	}).Msg("a task was enqueued")

	return nil
}
