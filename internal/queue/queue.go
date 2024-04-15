package queue

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jasonwvh/webhook-handler/internal/models"
	"github.com/rabbitmq/amqp091-go"
)

type RabbitMQQueue struct {
	conn *amqp091.Connection
	ch   *amqp091.Channel
	q    amqp091.Queue
}

func NewRabbitMQQueue(host, user, password string) (*RabbitMQQueue, error) {
	conn, err := amqp091.Dial(fmt.Sprintf("amqp://%s:%s@%s/", user, password, host))
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	err = ch.ExchangeDeclare(
		"webhookx", // name
		"fanout",   // type
		true,       // durable
		false,      // auto-deleted
		false,      // internal
		false,      // no-wait
		nil,        // arguments
	)
	if err != nil {
		panic(err)
	}

	q, err := ch.QueueDeclare(
		"webhook_queue", // name
		true,            // durable
		false,           // delete when unused
		false,           // exclusive
		false,           // no-wait
		nil,             // arguments
	)
	if err != nil {
		return nil, err
	}

	err = ch.QueueBind(
		q.Name,     // queue name
		"",         // routing key
		"webhookx", // exchange
		false,
		nil,
	)
	if err != nil {
		panic(err)
	}

	return &RabbitMQQueue{
		conn: conn,
		ch:   ch,
		q:    q,
	}, nil
}

func (q *RabbitMQQueue) Publish(ctx context.Context, workItem models.WorkItem) error {
	payload, err := json.Marshal(workItem)
	if err != nil {
		return err
	}

	return q.ch.PublishWithContext(ctx,
		"",       // exchange
		q.q.Name, // routing key
		false,    // mandatory
		false,    // immediate
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        payload,
		})
}

func (q *RabbitMQQueue) Receive(ctx context.Context) (<-chan amqp091.Delivery, error) {
	msgs, err := q.ch.ConsumeWithContext(ctx,
		q.q.Name, // queue
		"",       // consumer
		true,     // auto-ack
		false,    // exclusive
		false,    // no-local
		false,    // no-wait
		nil,      // args
	)

	if err != nil {
		return nil, err
	}

	return msgs, nil
}

func (q *RabbitMQQueue) Close() error {
	if err := q.ch.Close(); err != nil {
		return err
	}
	return q.conn.Close()
}
