package billing

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/store-platform/store/internal/notification"
)

type Invoice struct {
	ID              uuid.UUID `json:"id"`
	InvoiceNumber   string    `json:"invoice_number"`
	CustomerID      uuid.UUID `json:"customer_id"`
	BillingPeriodID uuid.UUID `json:"billing_period_id"`
	Status          string    `json:"status"`
	TotalCents      int64     `json:"total_cents"`
	PaidCents       int64     `json:"paid_cents"`
	DueAt           time.Time `json:"due_at"`
}

func (inv Invoice) RemainingCents() int64 {
	r := inv.TotalCents - inv.PaidCents
	if r < 0 {
		return 0
	}
	return r
}

// ClosePeriod fecha um período aberto e gera fatura (idempotente).
func (s *Service) ClosePeriod(ctx context.Context, periodID uuid.UUID) (*Invoice, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var customerID uuid.UUID
	var status string
	var refYear, refMonth int
	err = tx.QueryRow(ctx, `
		SELECT customer_id, status, reference_year, reference_month
		FROM billing_periods WHERE id = $1 FOR UPDATE
	`, periodID).Scan(&customerID, &status, &refYear, &refMonth)
	if err == pgx.ErrNoRows {
		return nil, ErrPeriodNotFound
	}
	if err != nil {
		return nil, err
	}

	var existing Invoice
	err = tx.QueryRow(ctx, `
		SELECT id, invoice_number, customer_id, billing_period_id, status, total_cents, paid_cents, due_at
		FROM invoices WHERE billing_period_id = $1
	`, periodID).Scan(&existing.ID, &existing.InvoiceNumber, &existing.CustomerID,
		&existing.BillingPeriodID, &existing.Status, &existing.TotalCents, &existing.PaidCents, &existing.DueAt)
	if err == nil {
		return &existing, nil
	}
	if err != pgx.ErrNoRows {
		return nil, err
	}

	if status == "closed" {
		return nil, ErrPeriodNotOpen
	}

	rows, err := tx.Query(ctx, `
		SELECT id, description, amount_cents FROM billing_entries WHERE billing_period_id = $1
	`, periodID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subtotal int64
	type entry struct {
		id          uuid.UUID
		description string
		amount      int64
	}
	var entries []entry
	for rows.Next() {
		var e entry
		if err := rows.Scan(&e.id, &e.description, &e.amount); err != nil {
			return nil, err
		}
		subtotal += e.amount
		entries = append(entries, e)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	creditCents := int64(0)
	adjustmentCents := int64(0)
	total := subtotal - creditCents + adjustmentCents
	if total < 0 {
		total = 0
	}

	invID := uuid.New()
	invNumber := fmt.Sprintf("INV-%04d%02d-%s", refYear, refMonth, invID.String()[:8])
	dueAt := time.Date(refYear, time.Month(refMonth), 1, 0, 0, 0, 0, saoPaulo).AddDate(0, 2, 0) // vencimento simplificado: +2 meses

	now := time.Now()
	_, err = tx.Exec(ctx, `
		INSERT INTO invoices (
			id, invoice_number, customer_id, billing_period_id, status,
			subtotal_cents, credit_cents, adjustment_cents, total_cents, paid_cents,
			due_at, closed_at
		) VALUES ($1,$2,$3,$4,'open',$5,$6,$7,$8,0,$9,$10)
	`, invID, invNumber, customerID, periodID, subtotal, creditCents, adjustmentCents, total, dueAt, now)
	if err != nil {
		return nil, err
	}

	for _, e := range entries {
		_, err = tx.Exec(ctx, `
			INSERT INTO invoice_items (invoice_id, billing_entry_id, description, quantity, unit_price_cents, total_cents)
			VALUES ($1, $2, $3, 1, $4, $4)
		`, invID, e.id, e.description, e.amount)
		if err != nil {
			return nil, err
		}
	}

	_, err = tx.Exec(ctx, `
		UPDATE billing_periods SET status = 'closed', closed_at = $2, updated_at = NOW() WHERE id = $1
	`, periodID, now)
	if err != nil {
		return nil, err
	}

	if s.jobs != nil {
		var email, name string
		err = tx.QueryRow(ctx, `
			SELECT u.email, u.name FROM users u
			JOIN customers c ON c.user_id = u.id
			WHERE c.id = $1
		`, customerID).Scan(&email, &name)
		if err != nil {
			return nil, err
		}
		invoiceURL := notification.BuildInvoiceURL(s.storeWebURL, invID.String())
		payload := map[string]any{
			"to":             email,
			"name":           name,
			"invoice_number": invNumber,
			"ref_year":       refYear,
			"ref_month":      refMonth,
			"total_cents":    total,
			"due_at":         dueAt.Format(time.RFC3339),
			"invoice_url":    invoiceURL,
		}
		if err := s.jobs.PublishOutbox(ctx, tx, notification.EventInvoiceClosed, "invoice", invID, payload); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &Invoice{
		ID: invID, InvoiceNumber: invNumber, CustomerID: customerID,
		BillingPeriodID: periodID, Status: "open", TotalCents: total, PaidCents: 0, DueAt: dueAt,
	}, nil
}

// CloseOpenPeriodsForReference fecha todos os períodos abertos de um ano/mês de competência.
func (s *Service) CloseOpenPeriodsForReference(ctx context.Context, year, month int) (int, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id FROM billing_periods
		WHERE reference_year = $1 AND reference_month = $2 AND status = 'open'
	`, year, month)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	var closed int
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return closed, err
		}
		if _, err := s.ClosePeriod(ctx, id); err != nil && err != ErrInvoiceExists {
			return closed, err
		}
		closed++
	}
	return closed, rows.Err()
}

// RunScheduledClosingIfDue executa fechamento do mês anterior no 5º dia útil.
func (s *Service) RunScheduledClosingIfDue(ctx context.Context, now time.Time) (bool, int, error) {
	due, err := IsMonthlyClosingDay(ctx, s.pool, now)
	if err != nil {
		return false, 0, err
	}
	if !due {
		return false, 0, nil
	}
	now = now.In(saoPaulo)
	py, pm := PreviousMonth(now.Year(), int(now.Month()))
	n, err := s.CloseOpenPeriodsForReference(ctx, py, pm)
	return true, n, err
}

func (s *Service) ListInvoicesByCustomer(ctx context.Context, customerID uuid.UUID) ([]Invoice, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, invoice_number, customer_id, billing_period_id, status, total_cents, paid_cents, due_at
		FROM invoices WHERE customer_id = $1 ORDER BY created_at DESC
	`, customerID)
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

func (s *Service) GetInvoice(ctx context.Context, id uuid.UUID) (*Invoice, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, invoice_number, customer_id, billing_period_id, status, total_cents, paid_cents, due_at
		FROM invoices WHERE id = $1
	`, id)
	var inv Invoice
	err := row.Scan(&inv.ID, &inv.InvoiceNumber, &inv.CustomerID, &inv.BillingPeriodID,
		&inv.Status, &inv.TotalCents, &inv.PaidCents, &inv.DueAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

func (s *Service) MarkOverdueInvoices(ctx context.Context, now time.Time) (int64, error) {
	tag, err := s.pool.Exec(ctx, `
		UPDATE invoices SET status = 'overdue', updated_at = NOW()
		WHERE status = 'open' AND due_at < $1 AND paid_cents < total_cents
	`, now)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}
