package main

import (
	"context"
	"log"
	"time"

	"github.com/store-platform/store/internal/billing"
	"github.com/store-platform/store/internal/jobs"
	"github.com/store-platform/store/internal/notification"
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

	jobRepo := jobs.NewRepository(pool)
	billSvc := billing.NewService(pool, jobRepo, cfg.App.StoreWebURL, logger)
	mailer := notification.NewSMTPMailer(cfg.SMTP)
	outbox := &notification.OutboxHandler{Mailer: mailer, Log: logger}
	runner := &jobs.Runner{
		Repo:     jobRepo,
		WorkerID: cfg.Worker.WorkerID,
		Batch:    5,
		Handler: &jobs.Handler{
			Billing: billSvc,
			Log:     logger,
			Outbox:  outbox,
		},
	}

	logger.Info("worker started", "id", cfg.Worker.WorkerID)
	ticker := time.NewTicker(cfg.Worker.PollInterval)
	defer ticker.Stop()

	lastMaintenance := time.Time{}
	for range ticker.C {
		if time.Since(lastMaintenance) > 12*time.Hour {
			if err := runner.ScheduleDailyMaintenance(ctx); err != nil {
				logger.Error("daily maintenance failed", "error", err)
			}
			lastMaintenance = time.Now()
		}
		if err := runner.Tick(ctx); err != nil {
			logger.Error("job tick failed", "error", err)
		}
	}
}
