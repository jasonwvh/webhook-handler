package app

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
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
		http.Error(w, "invalid work item", http.StatusBadRequest)
		return
	}

	if val, _ := h.storage.GetWorkItem(workItem.ID); val != nil {
		http.Error(w, "work item already processed", http.StatusConflict)

		h.cache.RemoveKey(strconv.Itoa(workItem.ID))
		return
	}

	if val, _ := h.cache.GetValue(strconv.Itoa(workItem.ID)); val == "pending" {
		http.Error(w, "work item already pending", http.StatusConflict)
		return
	}

	h.cache.SetValue(strconv.Itoa(workItem.ID), "pending")

	if err := h.processWorkItem(&workItem); err != nil {
		http.Error(w, "failed to process work item", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) processWorkItem(workItem *models.WorkItem) error {
	// Simulate work
	time.Sleep(time.Second)

	resp, err := http.Get(workItem.URL)
	if err != nil {
		log.Printf("Error processing work item %d: %v\n", workItem.ID, err)
		return err
	}
	resp.Body.Close()

	h.cache.RemoveKey(strconv.Itoa(workItem.ID))
	return h.storage.StoreWorkItem(workItem)
}
