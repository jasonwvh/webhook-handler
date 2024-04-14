package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
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

		// msg, err := p.queue.Receive(context.Background())
		// if err != nil {
		// 	log.Printf("err: %v", err.Error())
		// 	return
		// }

		// var workItem models.WorkItem
		// if err := json.Unmarshal(msg.Body, &workItem); err != nil {
		// 	log.Printf("err: %v", err.Error())
		// 	return
		// }

		p.executor.Submit(func() {
			log.Println("submitting func")
			if err := p.validateWorkItem(workItem); err != nil {
				log.Printf("err: %v", err.Error())
				return
			}

			if err := p.processWorkItem(workItem); err != nil {
				log.Printf("failed to process work item: %v", err)

				// todo: need dlq
				err := p.queue.Publish(context.Background(), workItem)
				if err != nil {
					log.Printf("failed to retry work item: %v", err)
				}
				return
			}

			// err := msg.Ack(false)
			//if err != nil {
			//	log.Printf("error acknowledging message: %v", err)
			//}
		})
	}
}

func (p *WebhookProcessor) validateWorkItem(workItem *models.WorkItem) error {
	if pending := p.cache.IsPending(workItem.ID); pending {
		return fmt.Errorf("work item is processing")
	}

	return nil
}

func (p *WebhookProcessor) processWorkItem(workItem *models.WorkItem) error {
	// todo: can just use db
	seq, err := p.cache.GetSeq(workItem.URL)
	if err != nil {
		// if url doesn't exist yet, create one
		log.Printf("seq: %d", seq)
		err := p.cache.SetSeq(workItem.URL, 0)
		if err != nil {
			log.Printf("couldn't set seq: %v", err)
		}
	}
	if workItem.Seq != seq+1 {
		// if the url is already processed and it's not the next sequence
		return fmt.Errorf("work item not next in order")
	}
	err = p.cache.AddPending(workItem.ID)
	if err != nil {
		return err
	}

	// simulate work
	time.Sleep(time.Second)
	resp, err := http.Get(workItem.URL)
	if err != nil {
		log.Printf("Error processing work item %d: %v\n", workItem.ID, err)
		return err
	}
	resp.Body.Close()

	err = p.cache.RemovePending(workItem.ID)
	if err != nil {
		return err
	}
	err = p.cache.SetSeq(workItem.URL, workItem.Seq)
	if err != nil {
		return err
	}

	return p.storage.StoreWorkItem(workItem)
}
