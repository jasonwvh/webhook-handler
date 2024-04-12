package app

import (
	"context"
	"log"
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
		workItem, err := p.queue.Receive(context.Background())
		if err != nil {
			log.Printf("failed to receive work item: %v", err)
			continue
		}

		p.executor.Submit(func() {
			if err := p.processWorkItem(workItem); err != nil {
				log.Printf("failed to process work item: %v", err)
				return
			}

			if err := p.cache.RemoveKey(strconv.Itoa(workItem.ID)); err != nil {
				log.Printf("failed to delete pending item: %v", err)
				return
			}
		})
	}
}

func (p *WebhookProcessor) processWorkItem(workItem *models.WorkItem) error {
	// Simulate work
	time.Sleep(time.Second)

	return p.storage.StoreWorkItem(workItem)
}
