package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/jasonwvh/webhook-handler/internal/models"
)

type Handler struct {
	storage *SQLiteStorage
}

func NewHandler(storage *SQLiteStorage) *Handler {
	return &Handler{
		storage: storage,
	}
}

func (h *Handler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	var workItem models.WorkItem
	if err := json.NewDecoder(r.Body).Decode(&workItem); err != nil {
		http.Error(w, "invalid work item", http.StatusBadRequest)
		return
	}

	if err := h.processWorkItem(r.Context(), workItem); err != nil {
		http.Error(w, "failed to process work item", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) processWorkItem(ctx context.Context, workItem models.WorkItem) error {
	if _, err := h.storage.GetWorkItem(context.Background(), workItem.ID); err == nil {
		return fmt.Errorf("work item already processed")
	}

	// Simulate work
	time.Sleep(time.Second)
	return h.storage.StoreWorkItem(ctx, workItem)
}
