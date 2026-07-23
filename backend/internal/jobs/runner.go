package jobs

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/store-platform/store/internal/notification"
)

type Handler struct {
	Billing           BillingWorker
	Payments          MercadoPagoPayments
	Log               *slog.Logger
	Outbox            *notification.OutboxHandler
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
			y, m := now.Year(), int(now.Month())
			if m == 1 {
				p.Year, p.Month = y-1, 12
			} else {
				p.Year, p.Month = y, m-1
			}
		}
		closed, err := h.Billing.CloseOpenPeriodsForReference(ctx, p.Year, p.Month, "monthly_auto")
		if err == nil {
			h.Log.Info("billing monthly close job completed",
				slog.String("job_id", job.ID.String()),
				slog.Int("year", p.Year),
				slog.Int("month", p.Month),
				slog.Int("closed_periods", closed),
			)
		}
		return err
	case TypeMercadoPagoOrder:
		if h.Payments == nil {
			h.Log.Warn("mercado pago job skipped: payments handler not configured")
			return nil
		}
		var p struct {
			PaymentEventID string `json:"payment_event_id"`
			OrderID        string `json:"order_id"`
		}
		if len(job.Payload) > 0 {
			_ = json.Unmarshal(job.Payload, &p)
		}
		var peID uuid.UUID
		if p.PaymentEventID != "" {
			peID, _ = uuid.Parse(p.PaymentEventID)
		}
		return h.Payments.ProcessMercadoPagoOrderJob(ctx, peID, p.OrderID)
	case TypeMarkOverdue:
		n, err := h.Billing.MarkOverdueInvoices(ctx, time.Now())
		if err == nil {
			h.Log.Info("billing mark overdue job completed",
				slog.String("job_id", job.ID.String()),
				slog.Int64("updated", n),
			)
		}
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
	if r.Handler.Outbox != nil {
		_ = r.Repo.ProcessOutbox(ctx, r.Batch, func(eventType string, payload json.RawMessage) error {
			return r.Handler.Outbox.Handle(ctx, eventType, payload)
		})
	}
	jobs, err := r.Repo.Acquire(ctx, r.WorkerID, r.Batch)
	if err != nil {
		return err
	}
	for _, job := range jobs {
		start := time.Now()
		err := r.Handler.Handle(ctx, job)
		dur := time.Since(start).Milliseconds()
		if err != nil {
			r.Handler.Log.Error("job failed",
				slog.String("job_id", job.ID.String()),
				slog.String("type", job.Type),
				slog.Int64("duration_ms", dur),
				slog.String("error", err.Error()),
			)
			_ = r.Repo.Fail(ctx, job.ID, err, 30*time.Second)
			continue
		}
		r.Handler.Log.Info("job completed",
			slog.String("job_id", job.ID.String()),
			slog.String("type", job.Type),
			slog.Int64("duration_ms", dur),
		)
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
		now := time.Now()
		py, pm := previousCalendarMonth(now.Year(), int(now.Month()))
		r.Handler.Log.Info("monthly closing executed", "year", py, "month", pm, "periods_closed", n)
	}
	rem, esc, err := r.Handler.Billing.ProcessClosedInvoiceReminders(ctx, time.Now())
	if err != nil {
		return err
	}
	if rem > 0 || esc > 0 {
		r.Handler.Log.Info("invoice payment reminders processed", "reminders", rem, "escalations", esc)
	}
	return r.Repo.Enqueue(ctx, TypeMarkOverdue, map[string]any{})
}

func previousCalendarMonth(year, month int) (int, int) {
	if month == 1 {
		return year - 1, 12
	}
	return year, month - 1
}
