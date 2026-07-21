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
	CycleNumber     int       `json:"cycle_number"`
	Status          string    `json:"status"`
	TotalCents      int64     `json:"total_cents"`
	EntryCount      int       `json:"entry_count"`
}

func (s *Service) GetOpenPeriodSummary(ctx context.Context, customerID uuid.UUID) (*OpenPeriodSummary, error) {
	var p OpenPeriodSummary
	err := s.pool.QueryRow(ctx, `
		SELECT bp.id, bp.reference_year, bp.reference_month, bp.cycle_number, bp.status,
		       COALESCE(SUM(be.amount_cents), 0)::bigint,
		       COUNT(be.id)::int
		FROM billing_periods bp
		LEFT JOIN billing_entries be ON be.billing_period_id = bp.id
		WHERE bp.customer_id = $1 AND bp.status = 'open'
		GROUP BY bp.id, bp.reference_year, bp.reference_month, bp.cycle_number, bp.status
		ORDER BY bp.reference_year DESC, bp.reference_month DESC
		LIMIT 1
	`, customerID).Scan(&p.BillingPeriodID, &p.ReferenceYear, &p.ReferenceMonth, &p.CycleNumber, &p.Status, &p.TotalCents, &p.EntryCount)
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
		limit = 20
	}
	rows, err := s.pool.Query(ctx, `
		SELECT i.id, i.invoice_number, i.customer_id, i.billing_period_id, i.status,
		       i.total_cents, i.paid_cents, i.due_at, i.close_type,
		       bp.reference_year, bp.reference_month, bp.cycle_number
		FROM invoices i
		JOIN billing_periods bp ON bp.id = i.billing_period_id
		WHERE i.customer_id = $1 ORDER BY i.created_at DESC LIMIT $2
	`, customerID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Invoice
	for rows.Next() {
		var inv Invoice
		if err := rows.Scan(&inv.ID, &inv.InvoiceNumber, &inv.CustomerID, &inv.BillingPeriodID,
			&inv.Status, &inv.TotalCents, &inv.PaidCents, &inv.DueAt, &inv.CloseType,
			&inv.ReferenceYear, &inv.ReferenceMonth, &inv.CycleNumber); err != nil {
			return nil, err
		}
		out = append(out, inv)
	}
	return out, rows.Err()
}

// GetOpenPeriodDetail retorna a competência aberta do cliente com lançamentos e produtos.
func (s *Service) GetOpenPeriodDetail(ctx context.Context, customerID uuid.UUID) (*OpenPeriodDetail, error) {
	summary, err := s.GetOpenPeriodSummary(ctx, customerID)
	if err != nil {
		return nil, err
	}
	if summary == nil {
		return nil, ErrNoOpenPeriod
	}
	entries, err := s.loadEntryViewsForPeriod(ctx, summary.BillingPeriodID)
	if err != nil {
		return nil, err
	}
	return &OpenPeriodDetail{Period: *summary, Entries: entries}, nil
}
