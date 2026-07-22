package reports

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type ReceivableRow struct {
	ID              uuid.UUID  `json:"id"`
	InvoiceNumber   string     `json:"invoice_number"`
	CustomerID      uuid.UUID  `json:"customer_id"`
	CustomerName    string     `json:"customer_name"`
	ReferenceYear   int        `json:"reference_year"`
	ReferenceMonth  int        `json:"reference_month"`
	ClosedAt        *time.Time `json:"closed_at,omitempty"`
	DueAt           *time.Time `json:"due_at,omitempty"`
	TotalCents      int64      `json:"total_cents"`
	PaidCents       int64      `json:"paid_cents"`
	RemainingCents  int64      `json:"remaining_cents"`
	Status          string     `json:"status"`
	DaysOverdue     int        `json:"days_overdue"`
	AgingBucket     string     `json:"aging_bucket"`
	HasActivePix    bool       `json:"has_active_pix"`
}

type ReceivablesSummary struct {
	TotalBilledCents   int64 `json:"total_billed_cents"`
	TotalReceivedCents int64 `json:"total_received_cents"`
	TotalOpenCents     int64 `json:"total_open_cents"`
	TotalOverdueCents  int64 `json:"total_overdue_cents"`
	OverdueCustomers   int   `json:"overdue_customers"`
	DueSoonCents       int64 `json:"due_soon_cents"`
	Bucket1To7Cents    int64 `json:"bucket_1_7_cents"`
	Bucket8To30Cents   int64 `json:"bucket_8_30_cents"`
	BucketOver30Cents  int64 `json:"bucket_over_30_cents"`
}

type ReceivablesFilter struct {
	Year         int
	Month        int
	Status       string
	CustomerID   *uuid.UUID
	OverdueOnly  bool
	PartialOnly  bool
	MinRemaining *int64
	PageFilter
}

func agingBucket(now time.Time, due *time.Time, status string, remaining int64) (int, string) {
	if remaining <= 0 || status == "paid" {
		return 0, "pago"
	}
	if due == nil {
		return 0, "a_vencer"
	}
	d := due.In(storeLocation)
	n := now.In(storeLocation)
	if !d.Before(n) {
		return 0, "a_vencer"
	}
	days := int(n.Sub(d).Hours() / 24)
	if days <= 0 {
		days = 1
	}
	switch {
	case days <= 7:
		return days, "1_a_7"
	case days <= 30:
		return days, "8_a_30"
	default:
		return days, "mais_30"
	}
}

func (s *Service) ReceivablesInvoices(ctx context.Context, f ReceivablesFilter, now time.Time) ([]ReceivableRow, ReceivablesSummary, int, error) {
	where := []string{"1=1"}
	args := []any{}
	n := 1

	if st := strings.TrimSpace(f.Status); st != "" {
		where = append(where, fmt.Sprintf("i.status = $%d", n))
		args = append(args, st)
		n++
	}
	if f.CustomerID != nil {
		where = append(where, fmt.Sprintf("i.customer_id = $%d", n))
		args = append(args, *f.CustomerID)
		n++
	}
	if f.Year > 0 {
		where = append(where, fmt.Sprintf("bp.reference_year = $%d", n))
		args = append(args, f.Year)
		n++
	}
	if f.Month > 0 {
		where = append(where, fmt.Sprintf("bp.reference_month = $%d", n))
		args = append(args, f.Month)
		n++
	}
	if f.OverdueOnly {
		where = append(where, "i.status = 'overdue'")
	}
	if f.PartialOnly {
		where = append(where, "i.paid_cents > 0 AND i.paid_cents < i.total_cents")
	}
	if f.MinRemaining != nil {
		where = append(where, fmt.Sprintf("(i.total_cents - i.paid_cents) >= $%d", n))
		args = append(args, *f.MinRemaining)
		n++
	}

	whereSQL := strings.Join(where, " AND ")

	var summary ReceivablesSummary
	sumQ := fmt.Sprintf(`
		SELECT
		  COALESCE(SUM(i.total_cents),0),
		  COALESCE(SUM(i.paid_cents),0),
		  COALESCE(SUM(GREATEST(i.total_cents - i.paid_cents,0)),0),
		  COALESCE(SUM(CASE WHEN i.status = 'overdue' THEN GREATEST(i.total_cents - i.paid_cents,0) ELSE 0 END),0),
		  COUNT(DISTINCT CASE WHEN i.status = 'overdue' THEN i.customer_id END)::int
		FROM invoices i
		JOIN billing_periods bp ON bp.id = i.billing_period_id
		WHERE %s
	`, whereSQL)
	if err := s.pool.QueryRow(ctx, sumQ, args...).Scan(
		&summary.TotalBilledCents, &summary.TotalReceivedCents,
		&summary.TotalOpenCents, &summary.TotalOverdueCents, &summary.OverdueCustomers,
	); err != nil {
		return nil, summary, 0, err
	}

	var total int
	countQ := fmt.Sprintf(`
		SELECT COUNT(*) FROM invoices i
		JOIN billing_periods bp ON bp.id = i.billing_period_id
		WHERE %s
	`, whereSQL)
	if err := s.pool.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, summary, 0, err
	}

	limit, offset := f.Limit, f.Offset
	q := fmt.Sprintf(`
		SELECT i.id, i.invoice_number, i.customer_id, u.name,
		       bp.reference_year, bp.reference_month, i.closed_at, i.due_at,
		       i.total_cents, i.paid_cents, i.status,
		       EXISTS (
		         SELECT 1 FROM payment_charges pc
		         WHERE pc.invoice_id = i.id AND pc.status = 'pending'
		       )
		FROM invoices i
		JOIN billing_periods bp ON bp.id = i.billing_period_id
		JOIN customers c ON c.id = i.customer_id
		JOIN users u ON u.id = c.user_id
		WHERE %s
		ORDER BY i.due_at NULLS LAST, i.created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereSQL, n, n+1)
	args = append(args, limit, offset)

	rows, err := s.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, summary, 0, err
	}
	defer rows.Close()

	var out []ReceivableRow
	for rows.Next() {
		var r ReceivableRow
		if err := rows.Scan(
			&r.ID, &r.InvoiceNumber, &r.CustomerID, &r.CustomerName,
			&r.ReferenceYear, &r.ReferenceMonth, &r.ClosedAt, &r.DueAt,
			&r.TotalCents, &r.PaidCents, &r.Status, &r.HasActivePix,
		); err != nil {
			return nil, summary, 0, err
		}
		r.RemainingCents = r.TotalCents - r.PaidCents
		if r.RemainingCents < 0 {
			r.RemainingCents = 0
		}
		r.DaysOverdue, r.AgingBucket = agingBucket(now, r.DueAt, r.Status, r.RemainingCents)
		switch r.AgingBucket {
		case "a_vencer":
			summary.DueSoonCents += r.RemainingCents
		case "1_a_7":
			summary.Bucket1To7Cents += r.RemainingCents
		case "8_a_30":
			summary.Bucket8To30Cents += r.RemainingCents
		case "mais_30":
			summary.BucketOver30Cents += r.RemainingCents
		}
		out = append(out, r)
	}
	return out, summary, total, rows.Err()
}
