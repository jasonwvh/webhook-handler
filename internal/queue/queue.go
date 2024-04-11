package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

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

func (q *RabbitMQQueue) Receive(ctx context.Context) (models.WorkItem, error) {
	msgs, err := q.ch.ConsumeWithContext(ctx,
		q.q.Name, // queue
		"",       // consumer
		false,    // auto-ack
		false,    // exclusive
		false,    // no-local
		false,    // no-wait
		nil,      // args
	)

	if err != nil {
		return models.WorkItem{}, err
	}

	go func() {
		for d := range msgs {
			log.Printf("received a message: %s", d.Body)
		}
	}()

	data := <-msgs
	var workItem models.WorkItem
	if err := json.Unmarshal(data.Body, &workItem); err != nil {
		return models.WorkItem{}, err
	}

	return workItem, nil
}

func (q *RabbitMQQueue) Close() error {
	if err := q.ch.Close(); err != nil {
		return err
	}
	return q.conn.Close()
}
