package jobs

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/store-platform/store/internal/billing"
)

type Handler struct {
	Billing *billing.Service
	Log     *slog.Logger
}

func (h *Handler) Handle(ctx context.Context, job Job) error {
	switch job.Type {
	case TypeMonthlyClose:
		var p struct {
			Year  int `json:"year"`
			Month int `json:"month"`
		}
		if len(job.Payload) > 0 {
			_ = json.Unmarshal(job.Payload, &p)
		}
		if p.Year == 0 {
			now := time.Now()
			p.Year, p.Month = billing.PreviousMonth(now.Year(), int(now.Month()))
		}
		_, err := h.Billing.CloseOpenPeriodsForReference(ctx, p.Year, p.Month)
		return err
	case TypeMarkOverdue:
		_, err := h.Billing.MarkOverdueInvoices(ctx, time.Now())
		return err
	default:
		h.Log.Warn("unknown job type", "type", job.Type)
		return nil
	}
}

type Runner struct {
	Repo    *Repository
	Handler *Handler
	WorkerID string
	Batch   int
}

func (r *Runner) Tick(ctx context.Context) error {
	jobs, err := r.Repo.Acquire(ctx, r.WorkerID, r.Batch)
	if err != nil {
		return err
	}
	for _, job := range jobs {
		err := r.Handler.Handle(ctx, job)
		if err != nil {
			_ = r.Repo.Fail(ctx, job.ID, err, 30*time.Second)
			continue
		}
		_ = r.Repo.Complete(ctx, job.ID)
	}
	return nil
}

func (r *Runner) ScheduleDailyMaintenance(ctx context.Context) error {
	ran, n, err := r.Handler.Billing.RunScheduledClosingIfDue(ctx, time.Now())
	if err != nil {
		return err
	}
	if ran {
		r.Handler.Log.Info("monthly closing executed", "periods_closed", n)
	}
	return r.Repo.Enqueue(ctx, TypeMarkOverdue, map[string]any{})
}
