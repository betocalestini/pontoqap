package billing

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// OpenPeriodSummary é a competência em aberto (compras ainda não faturadas).
type OpenPeriodSummary struct {
	BillingPeriodID uuid.UUID `json:"billing_period_id"`
	ReferenceYear   int       `json:"reference_year"`
	ReferenceMonth  int       `json:"reference_month"`
	Status          string    `json:"status"`
	TotalCents      int64     `json:"total_cents"`
	EntryCount      int       `json:"entry_count"`
}

func (s *Service) GetOpenPeriodSummary(ctx context.Context, customerID uuid.UUID) (*OpenPeriodSummary, error) {
	var p OpenPeriodSummary
	err := s.pool.QueryRow(ctx, `
		SELECT bp.id, bp.reference_year, bp.reference_month, bp.status,
		       COALESCE(SUM(be.amount_cents), 0)::bigint,
		       COUNT(be.id)::int
		FROM billing_periods bp
		LEFT JOIN billing_entries be ON be.billing_period_id = bp.id
		WHERE bp.customer_id = $1 AND bp.status = 'open'
		GROUP BY bp.id, bp.reference_year, bp.reference_month, bp.status
		ORDER BY bp.reference_year DESC, bp.reference_month DESC
		LIMIT 1
	`, customerID).Scan(&p.BillingPeriodID, &p.ReferenceYear, &p.ReferenceMonth, &p.Status, &p.TotalCents, &p.EntryCount)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (s *Service) ListInvoicesByCustomerLimit(ctx context.Context, customerID uuid.UUID, limit int) ([]Invoice, error) {
	if limit <= 0 {
		limit = 3
	}
	rows, err := s.pool.Query(ctx, `
		SELECT id, invoice_number, customer_id, billing_period_id, status, total_cents, paid_cents, due_at
		FROM invoices WHERE customer_id = $1 ORDER BY created_at DESC LIMIT $2
	`, customerID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Invoice
	for rows.Next() {
		var inv Invoice
		if err := rows.Scan(&inv.ID, &inv.InvoiceNumber, &inv.CustomerID, &inv.BillingPeriodID,
			&inv.Status, &inv.TotalCents, &inv.PaidCents, &inv.DueAt); err != nil {
			return nil, err
		}
		out = append(out, inv)
	}
	return out, rows.Err()
}
