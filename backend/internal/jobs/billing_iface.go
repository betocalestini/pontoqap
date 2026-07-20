package jobs

import (
	"context"
	"time"
)

// BillingWorker evita ciclo de importação billing ↔ jobs.
type BillingWorker interface {
	CloseOpenPeriodsForReference(ctx context.Context, year, month int) (int, error)
	MarkOverdueInvoices(ctx context.Context, now time.Time) (int64, error)
	RunScheduledClosingIfDue(ctx context.Context, now time.Time) (bool, int, error)
}
