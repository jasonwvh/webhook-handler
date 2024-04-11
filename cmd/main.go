package main

import (
	"log"
	"net/http"

	"github.com/jasonwvh/webhook-handler/internal/app"
	"github.com/jasonwvh/webhook-handler/internal/config"
	"github.com/jasonwvh/webhook-handler/internal/queue"
)

// In main.go
func main() {
	conf, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
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
	if err != nil {
		log.Fatalf("failed to create cache: %v", err)
	}

	handler := app.NewHandler(storage)
	asyncHandler := app.NewAsyncHandler(que, storage, cache)

	webhookProcessor := app.NewWebhookProcessor(storage, que, cache)
	go webhookProcessor.ProcessWebhooks()

	http.HandleFunc("/webhook", handler.HandleWebhook)
	http.HandleFunc("/async-webhook", asyncHandler.HandleWebhook)

	log.Printf("Starting webhook handler on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
