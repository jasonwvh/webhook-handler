package app

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/jasonwvh/webhook-handler/internal/models"
	"github.com/jasonwvh/webhook-handler/internal/queue"
)

type AsyncHandler struct {
	queue   *queue.RabbitMQQueue
	storage *SQLiteStorage
}

func NewAsyncHandler(queue *queue.RabbitMQQueue, storage *SQLiteStorage) *AsyncHandler {
	return &AsyncHandler{
		queue:   queue,
		storage: storage,
	}
}

func (h *AsyncHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	var workItem models.WorkItem
	if err := json.NewDecoder(r.Body).Decode(&workItem); err != nil {
		http.Error(w, "invalid work item", http.StatusBadRequest)
		return
	}

	if _, err := h.storage.GetWorkItem(context.Background(), workItem.ID); err == nil {
		http.Error(w, "work item already processed", http.StatusConflict)
		return
	}

	if err := h.enqueueWorkItem(r.Context(), workItem); err != nil {
		http.Error(w, "failed to enqueue work item", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *AsyncHandler) enqueueWorkItem(ctx context.Context, workItem models.WorkItem) error {
	return h.queue.Publish(ctx, workItem)
}
