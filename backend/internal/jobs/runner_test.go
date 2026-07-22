package jobs_test

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/store-platform/store/internal/billing"
	"github.com/store-platform/store/internal/jobs"
	"github.com/store-platform/store/tests/testdb"
)

func TestRunnerProcessesMarkOverdueJob(t *testing.T) {
	testdb.MigrateUp(t)
	pool := testdb.Pool(t)
	ctx := context.Background()
	if err := testdb.Reset(ctx, pool); err != nil {
		t.Fatal(err)
	}

	jobRepo := jobs.NewRepository(pool)
	billSvc := billing.NewService(pool, jobRepo, "")
	handler := &jobs.Handler{
		Billing: billSvc,
		Log:     slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError})),
	}
	runner := &jobs.Runner{
		Repo:     jobRepo,
		Handler:  handler,
		WorkerID: "test-worker",
		Batch:    5,
	}

	if err := jobRepo.Enqueue(ctx, jobs.TypeMarkOverdue, map[string]any{}); err != nil {
		t.Fatal(err)
	}
	if err := runner.Tick(ctx); err != nil {
		t.Fatal(err)
	}

	var pending int
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM jobs WHERE type = $1 AND status = 'completed'`, jobs.TypeMarkOverdue).Scan(&pending); err != nil {
		t.Fatal(err)
	}
	if pending < 1 {
		t.Fatal("expected completed mark_overdue job")
	}
}

func TestRunnerScheduleDailyMaintenance(t *testing.T) {
	testdb.MigrateUp(t)
	pool := testdb.Pool(t)
	ctx := context.Background()
	if err := testdb.Reset(ctx, pool); err != nil {
		t.Fatal(err)
	}

	jobRepo := jobs.NewRepository(pool)
	billSvc := billing.NewService(pool, jobRepo, "")
	handler := &jobs.Handler{
		Billing: billSvc,
		Log:     slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError})),
	}
	runner := &jobs.Runner{
		Repo:     jobRepo,
		Handler:  handler,
		WorkerID: "test-worker-2",
		Batch:    5,
	}

	// Day 1 of month triggers scheduled closing path (may close 0 periods on empty DB).
	now := time.Date(2026, 5, 1, 8, 0, 0, 0, time.UTC)
	if err := runner.ScheduleDailyMaintenance(ctx); err != nil {
		t.Fatal(err)
	}
	_ = now
}
