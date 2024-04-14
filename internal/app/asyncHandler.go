package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/jasonwvh/webhook-handler/internal/models"
	"github.com/jasonwvh/webhook-handler/internal/queue"
)

type AsyncHandler struct {
	queue   *queue.RabbitMQQueue
	storage *SQLiteStorage
	cache   *RedisClient
}

func NewAsyncHandler(queue *queue.RabbitMQQueue, storage *SQLiteStorage, cache *RedisClient) *AsyncHandler {
	return &AsyncHandler{
		queue:   queue,
		storage: storage,
		cache:   cache,
	}
}

func (h *AsyncHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	var workItem models.WorkItem
	if err := json.NewDecoder(r.Body).Decode(&workItem); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	// todo: seq checker should be here

	if err := h.validateWorkItem(&workItem); err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	if err := h.enqueueWorkItem(r.Context(), &workItem); err != nil {
		http.Error(w, "failed to queue work item", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *AsyncHandler) validateWorkItem(workItem *models.WorkItem) error {
	item, _ := h.storage.GetWorkItem(workItem.ID)
	if item != nil {
		h.cache.RemovePending(workItem.ID)

		return fmt.Errorf("work item is processed")
	}

	// todo: can just use db
	//seq, err := h.cache.GetSeq(workItem.URL)
	//if err != nil {
	//	// if url doesn't exist yet, create one
	//	log.Printf("seq: %d", seq)
	//	err := h.cache.SetSeq(workItem.URL, 0)
	//	if err != nil {
	//		log.Printf("couldn't set seq: %v", err)
	//	}
	//}
	//if workItem.Seq != seq+1 {
	//	// if the url is already processed and it's not the next sequence
	//	return fmt.Errorf("work item not next in order")
	//}

	return nil
}

func (h *AsyncHandler) enqueueWorkItem(ctx context.Context, workItem *models.WorkItem) error {
	return h.queue.Publish(ctx, workItem)
}
