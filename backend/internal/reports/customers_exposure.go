package reports

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type CustomerExposureRow struct {
	CustomerID           uuid.UUID  `json:"customer_id"`
	CustomerName         string     `json:"customer_name"`
	CustomerEmail        string     `json:"customer_email"`
	Status               string     `json:"status"`
	CreditLimitCents     int64      `json:"credit_limit_cents"`
	CurrentExposureCents int64      `json:"current_exposure_cents"`
	OpenInvoicesCents    int64      `json:"open_invoices_cents"`
	OverdueInvoicesCents int64      `json:"overdue_invoices_cents"`
	AvailableCents       int64      `json:"available_cents"`
	UtilizationPercent   int        `json:"utilization_percent"`
	HasOverdueInvoice    bool       `json:"has_overdue_invoice"`
	LastOrderAt          *time.Time `json:"last_order_at,omitempty"`
	UtilizationBand      string     `json:"utilization_band"`
}

type CustomerExposureFilter struct {
	Status          string
	MinUtilization  int
	OverdueOnly     bool
	LimitExhausted  bool
	BlockedOnly     bool
	PageFilter
}

func utilizationBand(pct int) string {
	switch {
	case pct <= 50:
		return "ate_50"
	case pct <= 80:
		return "51_a_80"
	case pct <= 100:
		return "81_a_100"
	default:
		return "acima_100"
	}
}

func (s *Service) CustomerExposure(ctx context.Context, f CustomerExposureFilter) ([]CustomerExposureRow, int, error) {
	where := []string{"c.status IN ('approved','blocked')"}
	args := []any{}
	n := 1

	if st := strings.TrimSpace(f.Status); st != "" {
		where = append(where, fmt.Sprintf("c.status = $%d", n))
		args = append(args, st)
		n++
	}
	if f.BlockedOnly {
		where = append(where, "c.status = 'blocked'")
	}
	whereSQL := strings.Join(where, " AND ")

	var total int
	countQ := fmt.Sprintf(`SELECT COUNT(*) FROM customers c JOIN users u ON u.id = c.user_id WHERE %s`, whereSQL)
	if err := s.pool.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	limit, offset := f.Limit, f.Offset
	q := fmt.Sprintf(`
		SELECT c.id, u.name, u.email, c.status, c.credit_limit_cents, c.current_exposure_cents,
		       COALESCE((
		         SELECT SUM(GREATEST(i.total_cents - i.paid_cents,0)) FROM invoices i
		         WHERE i.customer_id = c.id AND i.status IN ('open','overdue')
		       ),0),
		       COALESCE((
		         SELECT SUM(GREATEST(i.total_cents - i.paid_cents,0)) FROM invoices i
		         WHERE i.customer_id = c.id AND i.status = 'overdue'
		       ),0),
		       EXISTS (SELECT 1 FROM invoices i WHERE i.customer_id = c.id AND i.status = 'overdue'),
		       (SELECT MAX(o.confirmed_at) FROM orders o WHERE o.customer_id = c.id AND o.status = 'confirmed')
		FROM customers c
		JOIN users u ON u.id = c.user_id
		WHERE %s
		ORDER BY c.current_exposure_cents DESC
		LIMIT $%d OFFSET $%d
	`, whereSQL, n, n+1)
	args = append(args, limit, offset)

	rows, err := s.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []CustomerExposureRow
	for rows.Next() {
		var r CustomerExposureRow
		if err := rows.Scan(
			&r.CustomerID, &r.CustomerName, &r.CustomerEmail, &r.Status,
			&r.CreditLimitCents, &r.CurrentExposureCents,
			&r.OpenInvoicesCents, &r.OverdueInvoicesCents,
			&r.HasOverdueInvoice, &r.LastOrderAt,
		); err != nil {
			return nil, 0, err
		}
		r.AvailableCents = r.CreditLimitCents - r.CurrentExposureCents
		if r.CreditLimitCents > 0 {
			r.UtilizationPercent = int((r.CurrentExposureCents * 100) / r.CreditLimitCents)
		}
		r.UtilizationBand = utilizationBand(r.UtilizationPercent)
		if f.MinUtilization > 0 && r.UtilizationPercent < f.MinUtilization {
			continue
		}
		if f.OverdueOnly && !r.HasOverdueInvoice {
			continue
		}
		if f.LimitExhausted && r.AvailableCents > 0 {
			continue
		}
		out = append(out, r)
	}
	return out, total, rows.Err()
}
