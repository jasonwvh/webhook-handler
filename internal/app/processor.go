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
			log.Printf("failed to receive work item: %v", err)
			continue
		}

		go func() {
			for d := range msgs {
				// log.Printf("Received a message: %s", d.Body)

				var workItem models.WorkItem
				if err := json.Unmarshal(d.Body, &workItem); err != nil {
					continue
				}

				p.executor.Submit(func() {
					if err := p.processWorkItem(&workItem); err != nil {
						// log.Printf("failed to process work item: %v", err)

						time.Sleep(2 * time.Second)
						p.queue.Publish(context.Background(), workItem)
						return
					}

					p.cache.RemovePending(workItem.ID)
					p.cache.SetSeq(workItem.URL, workItem.Seq)
				})
			}
		}()
	}
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
	time.Sleep(time.Second)

	resp, err := http.Get(workItem.URL)
	if err != nil {
		log.Printf("Error processing work item %d: %v\n", workItem.ID, err)
		return err
	}
	resp.Body.Close()

	return p.storage.StoreWorkItem(workItem)
}
