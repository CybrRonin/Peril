package pubsub

import (
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType int

const (
	SimpleQueueDurable SimpleQueueType = iota
	SimpleQueueTransient
)

type Acktype int

const (
	Ack Acktype = iota
	NackRequeue
	NackDiscard
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
	queueTable := amqp.Table{
		"x-dead-letter-exchange": "peril_dlx",
	}
	newQueue, err := ch.QueueDeclare(queueName, durable, !durable, !durable, false, queueTable)
	if err != nil {
		return &amqp.Channel{}, amqp.Queue{}, fmt.Errorf("unable to declare queue: %v", err)
	}

	err = ch.QueueBind(newQueue.Name, key, exchange, false, nil)
	if err != nil {
		return &amqp.Channel{}, amqp.Queue{}, fmt.Errorf("unable to bind queue: %v", err)
	}

	return ch, newQueue, nil
}

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
	handler func(T) Acktype,
) error {
	ch, queue, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return fmt.Errorf("could not declare and bind queue: %v", err)
	}

	delivChan, err := ch.Consume(queue.Name, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("could not consume messages : %v", err)
	}

	go func() {
		defer ch.Close()
		for msg := range delivChan {
			var data T
			err := json.Unmarshal(msg.Body, &data)
			if err != nil {
				fmt.Printf("could not unmarshal message: %v\n", err)
				continue
			}

			switch handler(data) {
			case Ack:
				msg.Ack(false)
				fmt.Println("Message acknowledged!")
			case NackRequeue:
				msg.Nack(false, true)
				fmt.Println("Message negatively acknowledged but requeued!")
			case NackDiscard:
				msg.Nack(false, false)
				fmt.Println("Message negatively acknowledged and discarded!")
			}
		}
	}()

	return nil
}
