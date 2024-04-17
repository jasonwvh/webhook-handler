package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jasonwvh/webhook-handler/internal/app"
	"github.com/jasonwvh/webhook-handler/internal/config"
	"github.com/jasonwvh/webhook-handler/internal/queue"
)

var signalChan chan (os.Signal) = make(chan os.Signal, 1)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conf, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	storage, err := app.NewSQLiteStorage(conf.SQLiteDBPath)
	if err != nil {
		log.Fatalf("failed to create storage: %v", err)
	}
	defer storage.Close()

	que, err := queue.NewRabbitMQQueue(conf.RabbitMQHost, conf.RabbitMQUser, conf.RabbitMQPassword)
	if err != nil {
		log.Fatalf("failed to create queue: %v", err)
	}
	defer que.Close()

	cache := app.NewRedisClient(conf.RedisHost)

	handler := app.NewHandler(storage, cache)
	asyncHandler := app.NewAsyncHandler(que, storage, cache)

	webhookProcessor := app.NewWebhookProcessor(storage, que, cache)
	go webhookProcessor.ProcessWebhooks()

	server := &http.Server{
		Addr: ":8080",
	}

	http.HandleFunc("/webhook", handler.HandleWebhook)
	http.HandleFunc("/async-webhook", asyncHandler.HandleWebhook)
	http.HandleFunc("/terminate", termHandler)

	go func() {
		fmt.Println("Server running on port 8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	ctxTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctxTimeout); err != nil {
		fmt.Printf("Error during server shutdown: %v\n", err)
	}

	webhookProcessor.Stop()

	fmt.Println("Server stopped gracefully.")
}

func termHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Goodbye World!\n")
	signalChan <- syscall.SIGTERM
}
