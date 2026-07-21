package billing

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type AdminInvoiceListRow struct {
	ID              uuid.UUID  `json:"id"`
	InvoiceNumber   string     `json:"invoice_number"`
	CustomerID      uuid.UUID  `json:"customer_id"`
	CustomerName    string     `json:"customer_name"`
	CustomerEmail   string     `json:"customer_email"`
	ReferenceYear   int        `json:"reference_year"`
	ReferenceMonth  int        `json:"reference_month"`
	Status          string     `json:"status"`
	TotalCents      int64      `json:"total_cents"`
	PaidCents       int64      `json:"paid_cents"`
	RemainingCents  int64      `json:"remaining_cents"`
	DueAt           *time.Time `json:"due_at,omitempty"`
	ClosedAt        *time.Time `json:"closed_at,omitempty"`
}

type AdminInvoiceFilter struct {
	Status     string
	CustomerID *uuid.UUID
	Year       int
	Month      int
	Search     string
	Limit      int
	Offset     int
}

type InvoiceItem struct {
	ID               uuid.UUID            `json:"id"`
	Description      string               `json:"description"`
	Quantity         int                  `json:"quantity"`
	UnitPriceCents   int64                `json:"unit_price_cents"`
	TotalCents       int64                `json:"total_cents"`
	Products         []InvoiceProductLine `json:"products,omitempty"`
}

type InvoiceAdjustment struct {
	ID             uuid.UUID `json:"id"`
	AdjustmentType string    `json:"adjustment_type"`
	AmountCents    int64     `json:"amount_cents"`
	Reason         string    `json:"reason"`
	CreatedAt      time.Time `json:"created_at"`
}

type InvoiceDetail struct {
	ID               uuid.UUID           `json:"id"`
	InvoiceNumber    string              `json:"invoice_number"`
	CustomerID       uuid.UUID           `json:"customer_id"`
	CustomerName     string              `json:"customer_name,omitempty"`
	CustomerEmail    string              `json:"customer_email,omitempty"`
	BillingPeriodID  uuid.UUID           `json:"billing_period_id"`
	ReferenceYear    int                 `json:"reference_year"`
	ReferenceMonth   int                 `json:"reference_month"`
	CycleNumber      int                 `json:"cycle_number"`
	CloseType        string              `json:"close_type,omitempty"`
	Status           string              `json:"status"`
	SubtotalCents    int64               `json:"subtotal_cents"`
	CreditCents      int64               `json:"credit_cents"`
	AdjustmentCents  int64               `json:"adjustment_cents"`
	TotalCents       int64               `json:"total_cents"`
	PaidCents        int64               `json:"paid_cents"`
	RemainingCents   int64               `json:"remaining_cents"`
	DueAt            *time.Time          `json:"due_at,omitempty"`
	ClosedAt         *time.Time          `json:"closed_at,omitempty"`
	PaidAt           *time.Time          `json:"paid_at,omitempty"`
	Items            []InvoiceItem       `json:"items"`
	Adjustments      []InvoiceAdjustment `json:"adjustments"`
}

type AdminBillingSummary struct {
	OpenReceivablesCents       int64 `json:"open_receivables_cents"`
	OverdueInvoicesCount       int64 `json:"overdue_invoices_count"`
	OpenPeriodsCount           int   `json:"open_periods_count"`
	OpenPeriodsTotalCents      int64 `json:"open_periods_total_cents"`
	ScheduledClosingToday      bool  `json:"scheduled_closing_today"`
	ScheduledMonthlyCloseToday bool  `json:"scheduled_monthly_close_today"`
}

func (s *Service) AdminBillingSummary(ctx context.Context, now time.Time) (*AdminBillingSummary, error) {
	var sum AdminBillingSummary
	err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(total_cents - paid_cents), 0)
		FROM invoices WHERE status IN ('open', 'overdue') AND total_cents > paid_cents
	`).Scan(&sum.OpenReceivablesCents)
	if err != nil {
		return nil, err
	}
	err = s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM invoices WHERE status = 'overdue'`).Scan(&sum.OverdueInvoicesCount)
	if err != nil {
		return nil, err
	}
	err = s.pool.QueryRow(ctx, `
		SELECT COUNT(DISTINCT bp.id)::int, COALESCE(SUM(be.amount_cents), 0)::bigint
		FROM billing_periods bp
		LEFT JOIN billing_entries be ON be.billing_period_id = bp.id
		WHERE bp.status = 'open'
	`).Scan(&sum.OpenPeriodsCount, &sum.OpenPeriodsTotalCents)
	if err != nil {
		return nil, err
	}
	due := IsMonthlyClosingDay(now)
	sum.ScheduledClosingToday = due
	sum.ScheduledMonthlyCloseToday = due
	return &sum, nil
}

func (s *Service) ListInvoicesAdmin(ctx context.Context, f AdminInvoiceFilter) ([]AdminInvoiceListRow, int, error) {
	limit := f.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	offset := f.Offset
	if offset < 0 {
		offset = 0
	}

	where := "WHERE 1=1"
	args := []any{}
	arg := 1

	if st := strings.TrimSpace(f.Status); st != "" {
		where += fmt.Sprintf(" AND i.status = $%d", arg)
		args = append(args, st)
		arg++
	}
	if f.CustomerID != nil {
		where += fmt.Sprintf(" AND i.customer_id = $%d", arg)
		args = append(args, *f.CustomerID)
		arg++
	}
	if f.Year > 0 {
		where += fmt.Sprintf(" AND bp.reference_year = $%d", arg)
		args = append(args, f.Year)
		arg++
	}
	if f.Month > 0 {
		where += fmt.Sprintf(" AND bp.reference_month = $%d", arg)
		args = append(args, f.Month)
		arg++
	}
	if q := strings.TrimSpace(f.Search); q != "" {
		where += fmt.Sprintf(" AND (i.invoice_number ILIKE $%d OR u.name ILIKE $%d OR u.email ILIKE $%d)", arg, arg, arg)
		args = append(args, "%"+q+"%")
		arg++
	}

	countQ := `
		SELECT COUNT(*) FROM invoices i
		JOIN billing_periods bp ON bp.id = i.billing_period_id
		JOIN customers c ON c.id = i.customer_id
		JOIN users u ON u.id = c.user_id
	` + where
	var total int
	if err := s.pool.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	q := `
		SELECT i.id, i.invoice_number, i.customer_id, u.name, u.email,
		       bp.reference_year, bp.reference_month, i.status,
		       i.total_cents, i.paid_cents, i.due_at, i.closed_at
		FROM invoices i
		JOIN billing_periods bp ON bp.id = i.billing_period_id
		JOIN customers c ON c.id = i.customer_id
		JOIN users u ON u.id = c.user_id
	` + where + fmt.Sprintf(" ORDER BY i.created_at DESC LIMIT $%d OFFSET $%d", arg, arg+1)
	args = append(args, limit, offset)

	rows, err := s.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []AdminInvoiceListRow
	for rows.Next() {
		var r AdminInvoiceListRow
		var dueAt, closedAt *time.Time
		if err := rows.Scan(
			&r.ID, &r.InvoiceNumber, &r.CustomerID, &r.CustomerName, &r.CustomerEmail,
			&r.ReferenceYear, &r.ReferenceMonth, &r.Status,
			&r.TotalCents, &r.PaidCents, &dueAt, &closedAt,
		); err != nil {
			return nil, 0, err
		}
		r.DueAt = dueAt
		r.ClosedAt = closedAt
		r.RemainingCents = r.TotalCents - r.PaidCents
		if r.RemainingCents < 0 {
			r.RemainingCents = 0
		}
		out = append(out, r)
	}
	return out, total, rows.Err()
}

func (s *Service) GetInvoiceDetail(ctx context.Context, id uuid.UUID) (*InvoiceDetail, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT i.id, i.invoice_number, i.customer_id, u.name, u.email,
		       i.billing_period_id, bp.reference_year, bp.reference_month, bp.cycle_number,
		       i.status, i.subtotal_cents, i.credit_cents, i.adjustment_cents,
		       i.total_cents, i.paid_cents, i.due_at, i.closed_at, i.paid_at, i.close_type
		FROM invoices i
		JOIN billing_periods bp ON bp.id = i.billing_period_id
		JOIN customers c ON c.id = i.customer_id
		JOIN users u ON u.id = c.user_id
		WHERE i.id = $1
	`, id)
	var d InvoiceDetail
	var dueAt, closedAt, paidAt *time.Time
	err := row.Scan(
		&d.ID, &d.InvoiceNumber, &d.CustomerID, &d.CustomerName, &d.CustomerEmail,
		&d.BillingPeriodID, &d.ReferenceYear, &d.ReferenceMonth, &d.CycleNumber,
		&d.Status, &d.SubtotalCents, &d.CreditCents, &d.AdjustmentCents,
		&d.TotalCents, &d.PaidCents, &dueAt, &closedAt, &paidAt, &d.CloseType,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	d.DueAt = dueAt
	d.ClosedAt = closedAt
	d.PaidAt = paidAt
	d.RemainingCents = d.TotalCents - d.PaidCents
	if d.RemainingCents < 0 {
		d.RemainingCents = 0
	}

	itemRows, err := s.pool.Query(ctx, `
		SELECT ii.id, ii.description, ii.quantity, ii.unit_price_cents, ii.total_cents, COALESCE(be.order_id, '00000000-0000-0000-0000-000000000000'::uuid)
		FROM invoice_items ii
		LEFT JOIN billing_entries be ON be.id = ii.billing_entry_id
		WHERE ii.invoice_id = $1 ORDER BY ii.created_at
	`, id)
	if err != nil {
		return nil, err
	}
	defer itemRows.Close()
	var orderIDs []uuid.UUID
	for itemRows.Next() {
		var it InvoiceItem
		var orderID uuid.UUID
		if err := itemRows.Scan(&it.ID, &it.Description, &it.Quantity, &it.UnitPriceCents, &it.TotalCents, &orderID); err != nil {
			return nil, err
		}
		d.Items = append(d.Items, it)
		orderIDs = append(orderIDs, orderID)
	}
	if err := itemRows.Err(); err != nil {
		return nil, err
	}
	if err := s.attachProductsToInvoiceItems(ctx, d.Items, orderIDs); err != nil {
		return nil, err
	}

	adjRows, err := s.pool.Query(ctx, `
		SELECT id, adjustment_type, amount_cents, reason, created_at
		FROM billing_adjustments WHERE invoice_id = $1 ORDER BY created_at
	`, id)
	if err != nil {
		return nil, err
	}
	defer adjRows.Close()
	for adjRows.Next() {
		var a InvoiceAdjustment
		if err := adjRows.Scan(&a.ID, &a.AdjustmentType, &a.AmountCents, &a.Reason, &a.CreatedAt); err != nil {
			return nil, err
		}
		d.Adjustments = append(d.Adjustments, a)
	}
	if err := adjRows.Err(); err != nil {
		return nil, err
	}
	if d.Items == nil {
		d.Items = []InvoiceItem{}
	}
	if d.Adjustments == nil {
		d.Adjustments = []InvoiceAdjustment{}
	}
	return &d, nil
}
