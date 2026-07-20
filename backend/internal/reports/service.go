package reports

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	pool *pgxpool.Pool
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

type Dashboard struct {
	SalesMonthCents      int64 `json:"sales_month_cents"`
	ReceivedMonthCents   int64 `json:"received_month_cents"`
	OpenReceivablesCents int64 `json:"open_receivables_cents"`
	OverdueInvoices      int64 `json:"overdue_invoices"`
	OrdersMonth          int64 `json:"orders_month"`
}

func (s *Service) Dashboard(ctx context.Context, year, month int) (Dashboard, error) {
	var d Dashboard
	start := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, 0)

	_ = s.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(total_cents),0), COUNT(*) FROM orders
		WHERE status = 'confirmed' AND confirmed_at >= $1 AND confirmed_at < $2
	`, start, end).Scan(&d.SalesMonthCents, &d.OrdersMonth)

	_ = s.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(amount_cents),0) FROM payments
		WHERE status = 'settled' AND settled_at >= $1 AND settled_at < $2
	`, start, end).Scan(&d.ReceivedMonthCents)

	_ = s.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(total_cents - paid_cents),0) FROM invoices
		WHERE status IN ('open','overdue')
	`).Scan(&d.OpenReceivablesCents)

	_ = s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM invoices WHERE status = 'overdue'
	`).Scan(&d.OverdueInvoices)

	return d, nil
}

type RankRow struct {
	Label      string `json:"label"`
	TotalCents int64  `json:"total_cents"`
	Quantity   int64  `json:"quantity"`
}

func (s *Service) TopProducts(ctx context.Context, year, month int, limit int) ([]RankRow, error) {
	if limit <= 0 {
		limit = 10
	}
	start := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, 0)
	rows, err := s.pool.Query(ctx, `
		SELECT oi.product_name_snapshot, SUM(oi.total_cents), SUM(oi.quantity)
		FROM order_items oi
		JOIN orders o ON o.id = oi.order_id
		WHERE o.confirmed_at >= $1 AND o.confirmed_at < $2
		GROUP BY oi.product_name_snapshot
		ORDER BY SUM(oi.total_cents) DESC
		LIMIT $3
	`, start, end, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []RankRow
	for rows.Next() {
		var r RankRow
		if err := rows.Scan(&r.Label, &r.TotalCents, &r.Quantity); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *Service) InventoryReport(ctx context.Context) ([]RankRow, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT s.code, COALESCE(ib.available_quantity,0)
		FROM skus s
		LEFT JOIN inventory_balances ib ON ib.sku_id = s.id
		ORDER BY s.code
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []RankRow
	for rows.Next() {
		var r RankRow
		if err := rows.Scan(&r.Label, &r.Quantity); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}
