package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jasonwvh/webhook-handler/internal/app"
	"github.com/jasonwvh/webhook-handler/internal/config"
	"github.com/jasonwvh/webhook-handler/internal/models"
	"github.com/jasonwvh/webhook-handler/internal/queue"
)

func TestAsyncMain(t *testing.T) {
	// Mock the dependencies
	conf := &config.Config{
		RabbitMQHost:     "localhost",
		RabbitMQUser:     "user",
		RabbitMQPassword: "password",
		SQLiteDBPath:     "../data/test.db",
	}

	storage, err := app.NewSQLiteStorage(conf.SQLiteDBPath)
	if err != nil {
		log.Fatalf("failed to create storage: %v", err)
	}
	defer func(storage *app.SQLiteStorage) {
		err := storage.Close()
		if err != nil {
			panic(err)
		}
	}(storage)

	que, err := queue.NewRabbitMQQueue(conf.RabbitMQHost, conf.RabbitMQUser, conf.RabbitMQPassword)
	if err != nil {
		log.Fatalf("failed to create queue: %v", err)
	}
	defer func(q *queue.RabbitMQQueue) {
		err := q.Close()
		if err != nil {
			panic(err)
		}
	}(que)

	cache := app.NewRedisClient(conf.RedisHost)

	handler := app.NewAsyncHandler(que, storage, cache)

	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(handler.HandleWebhook))
	defer server.Close()

	// Marshal the struct into JSON
	body, err := json.Marshal(models.WorkItem{ID: 201, URL: "http://yahoo.com", Seq: 1})
	if err != nil {
		t.Errorf("Failed to marshal JSON: %v", err)
		return
	}

	// Test the webhook handler
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, server.URL+"/async-webhook", bytes.NewBuffer(body))
	if err != nil {
		t.Errorf("failed to create request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status code 200 OK, got %d", resp.StatusCode)
	}

}
