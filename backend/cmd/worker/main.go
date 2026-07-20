package main

import (
	"context"
	"log"
	"time"

	"github.com/store-platform/store/internal/platform/config"
	"github.com/store-platform/store/internal/platform/database"
	"github.com/store-platform/store/internal/platform/logging"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	logger := logging.New(cfg.LogLevel)
	ctx := context.Background()
	pool, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	logger.Info("worker started", "id", cfg.Worker.WorkerID)
	ticker := time.NewTicker(cfg.Worker.PollInterval)
	defer ticker.Stop()

	for range ticker.C {
		// Placeholder: monthly closing, Pix reconciliation, outbox (EP-11 / EP-12)
		_ = pool.Ping(ctx)
	}
}
