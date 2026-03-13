package pubsub

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType int

const (
	SimpleQueueDurable SimpleQueueType = iota
	SimpleQueueTransient
)

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType,
) (*amqp.Channel, amqp.Queue, error) {
	ch, err := conn.Channel()
	if err != nil {
		return &amqp.Channel{}, amqp.Queue{}, fmt.Errorf("unable to create channel: %v", err)
	}

	durable := true
	if queueType == SimpleQueueTransient {
		durable = false
	}
	newQueue, err := ch.QueueDeclare(queueName, durable, !durable, !durable, false, nil)
	if err != nil {
		return &amqp.Channel{}, amqp.Queue{}, fmt.Errorf("unable to declare queue: %v", err)
	}

	err = ch.QueueBind(newQueue.Name, key, exchange, false, nil)
	if err != nil {
		return &amqp.Channel{}, amqp.Queue{}, fmt.Errorf("unable to bind queue: %v", err)
	}

	return ch, newQueue, nil
}
