package messaging

import (
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

// EventConsumer consumes events from RabbitMQ and forwards them via a channel.
type EventConsumer struct {
	conn    *amqp.Connection
	ch      *amqp.Channel
	queue   string
	msgs    <-chan amqp.Delivery
}

type ConsumerConfig struct {
	URL          string
	Exchange     string
	ExchangeType string
	RoutingKey   string
	Queue        string // optional durable queue name; if empty, an exclusive auto-delete queue is created
}

func NewEventConsumer(cfg ConsumerConfig) (*EventConsumer, error) {
	conn, err := amqp.Dial(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("rabbitmq dial: %w", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("rabbitmq channel: %w", err)
	}
	if err := ch.ExchangeDeclare(cfg.Exchange, cfg.ExchangeType, true, false, false, false, nil); err != nil {
		_ = ch.Close(); _ = conn.Close()
		return nil, fmt.Errorf("exchange declare: %w", err)
	}
	qName := cfg.Queue
	var q amqp.Queue
	if qName == "" {
		// ephemeral queue for this consumer
		q, err = ch.QueueDeclare("", false, true, true, false, nil)
	} else {
		q, err = ch.QueueDeclare(qName, true, false, false, false, nil)
	}
	if err != nil {
		_ = ch.Close(); _ = conn.Close()
		return nil, fmt.Errorf("queue declare: %w", err)
	}
	if err := ch.QueueBind(q.Name, cfg.RoutingKey, cfg.Exchange, false, nil); err != nil {
		_ = ch.Close(); _ = conn.Close()
		return nil, fmt.Errorf("queue bind: %w", err)
	}
	msgs, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		_ = ch.Close(); _ = conn.Close()
		return nil, fmt.Errorf("consume: %w", err)
	}
	return &EventConsumer{conn: conn, ch: ch, queue: q.Name, msgs: msgs}, nil
}

func (c *EventConsumer) Messages() <-chan amqp.Delivery { return c.msgs }

func (c *EventConsumer) Close() error {
	if c.ch != nil { _ = c.ch.Close() }
	if c.conn != nil { return c.conn.Close() }
	return nil
}

// DecodeEvent attempts to unmarshal an Event from a Delivery body.
func DecodeEvent(d amqp.Delivery) (Event, error) {
	var e Event
	if err := json.Unmarshal(d.Body, &e); err != nil {
		return Event{}, err
	}
	return e, nil
}
