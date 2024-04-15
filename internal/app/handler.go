package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/jasonwvh/webhook-handler/internal/models"
)

type Handler struct {
	storage *SQLiteStorage
	cache   *RedisClient
}

func NewHandler(storage *SQLiteStorage, cache *RedisClient) *Handler {
	return &Handler{
		storage: storage,
		cache:   cache,
	}
}

func (h *Handler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	var workItem models.WorkItem
	if err := json.NewDecoder(r.Body).Decode(&workItem); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if err := h.validateWorkItem(&workItem); err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	if err := h.processWorkItem(&workItem); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) validateWorkItem(workItem *models.WorkItem) error {
	item, _ := h.storage.GetWorkItem(workItem.ID)
	if item != nil {
		err := h.cache.RemovePending(workItem.ID)
		if err != nil {
			return err
		}

		return fmt.Errorf("work item is processed")
	}

	if pending := h.cache.IsPending(workItem.ID); pending {
		return fmt.Errorf("work item is processing")
	}

	return nil
}

func (h *Handler) processWorkItem(workItem *models.WorkItem) error {
	seq, err := h.cache.GetSeq(workItem.URL)
	if err != nil {
		// if url doesn't exist yet, create one
		h.cache.SetSeq(workItem.URL, 0)
	}
	if workItem.Seq != seq+1 {
		// if the url is already processed and it's not the next sequence
		return fmt.Errorf("work item not next in order")
	}
	h.cache.AddPending(workItem.ID)

	// simulate work
	time.Sleep(time.Second)
	resp, err := http.Get(workItem.URL)
	if err != nil {
		return err
	}
	resp.Body.Close()

	h.cache.RemovePending(workItem.ID)
	h.cache.SetSeq(workItem.URL, workItem.Seq)

	return h.storage.StoreWorkItem(workItem)
}
