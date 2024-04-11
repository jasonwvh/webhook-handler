package app

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

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
		http.Error(w, "invalid work item", http.StatusBadRequest)
		return
	}

	// workItemString, err := json.Marshal(workItem)
	// if err != nil {
	// 	http.Error(w, "unable to marshal work item", http.StatusBadRequest)
	// 	return
	// }
	if val, _ := h.storage.GetWorkItem(context.Background(), workItem.ID); val != nil {
		if err := h.cache.RemoveKey(strconv.Itoa(workItem.ID)); err != nil {
			log.Printf("failed to delete pending item: %v", err)
			return
		}

		http.Error(w, "work item already processed", http.StatusConflict)
		return
	}

	log.Print(strconv.Itoa(workItem.ID))
	cVal, _ := h.cache.GetValue(strconv.Itoa(workItem.ID))
	if cVal == "pending" {
		http.Error(w, "work item already pending", http.StatusConflict)
		return
	}

	err := h.cache.SetValue(strconv.Itoa(workItem.ID), "pending")
	if err != nil {
		http.Error(w, "unable to store work item", http.StatusInternalServerError)
		return
	}

	if err := h.enqueueWorkItem(r.Context(), workItem); err != nil {
		http.Error(w, "failed to enqueue work item", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *AsyncHandler) enqueueWorkItem(ctx context.Context, workItem models.WorkItem) error {
	log.Print("publishing work item")
	return h.queue.Publish(ctx, workItem)
}
