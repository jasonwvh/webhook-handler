package app

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/jasonwvh/webhook-handler/internal/executor"
	"github.com/jasonwvh/webhook-handler/internal/models"
	"github.com/jasonwvh/webhook-handler/internal/queue"
)

type WebhookProcessor struct {
	storage  *SQLiteStorage
	queue    *queue.RabbitMQQueue
	cache    *RedisClient
	executor *executor.Executor
}

func NewWebhookProcessor(storage *SQLiteStorage, queue *queue.RabbitMQQueue, cache *RedisClient) *WebhookProcessor {
	return &WebhookProcessor{
		storage:  storage,
		queue:    queue,
		cache:    cache,
		executor: executor.NewExecutor(10),
	}
}

func (p *WebhookProcessor) ProcessWebhooks() {
	for {
		msgs, err := p.queue.Receive(context.Background())
		if err != nil {
			continue
		}

		go func() {
			for d := range msgs {
				var workItem models.WorkItem
				if err := json.Unmarshal(d.Body, &workItem); err != nil {
					continue
				}

				var retry int32
				if d.Headers != nil {
					if val, ok := d.Headers["retry"]; ok {
						retry = val.(int32)
					}
					if retry > 3 {
						d.Ack(false)
						continue
					}
				}

				log.Printf("Received a message: %s with retry %d", d.Body, retry)
				p.executor.Submit(func() {
					if err := p.processWorkItem(&workItem); err != nil {
						time.Sleep(5 * time.Second)
						p.queue.Publish(context.Background(), workItem, retry+1)
					}
					d.Ack(false)

					p.cache.RemovePending(workItem.ID)
					p.cache.SetSeq(workItem.URL, workItem.Seq)
				})
			}
		}()
	}
}

func (p *WebhookProcessor) StopWebhooks() {
	p.executor.Stop()
}

func (p *WebhookProcessor) processWorkItem(workItem *models.WorkItem) error {
	seq, err := p.cache.GetSeq(workItem.URL)
	if err != nil {
		// if url doesn't exist yet, create one
		p.cache.SetSeq(workItem.URL, workItem.Seq)
	}
	seqInt, _ := strconv.Atoi(seq)
	if err == nil && workItem.Seq != seqInt+1 {
		// if the url is already processed and it's not the next sequence
		return fmt.Errorf("work item not next in order")
	}
	p.cache.AddPending(workItem.ID)

	// Simulate work
	time.Sleep(5 * time.Second)

	resp, err := http.Get(workItem.URL)
	if err != nil {
		log.Printf("Error processing work item %d: %v\n", workItem.ID, err)
		return err
	}
	resp.Body.Close()

	return p.storage.StoreWorkItem(workItem)
}
