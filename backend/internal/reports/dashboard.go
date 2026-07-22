package reports

import (
	"context"
	"time"
)

type DashboardSeriesPoint struct {
	Year          int   `json:"year"`
	Month         int   `json:"month"`
	SalesCents    int64 `json:"sales_cents"`
	ReceivedCents int64 `json:"received_cents"`
}

func (s *Service) DashboardSeries(ctx context.Context, months int) ([]DashboardSeriesPoint, error) {
	if months <= 0 || months > 24 {
		months = 6
	}
	now := time.Now().In(storeLocation)
	// Start at first day of month (months-1) ago
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, storeLocation).AddDate(0, -(months - 1), 0)

	var out []DashboardSeriesPoint
	for i := 0; i < months; i++ {
		t := start.AddDate(0, i, 0)
		y, m := t.Year(), int(t.Month())
		rng := monthRangeUTC(y, m)
		var pt DashboardSeriesPoint
		pt.Year = y
		pt.Month = m
		_ = s.pool.QueryRow(ctx, `
			SELECT COALESCE(SUM(total_cents),0) FROM orders
			WHERE status = 'confirmed' AND confirmed_at >= $1 AND confirmed_at < $2
		`, rng.From, rng.To).Scan(&pt.SalesCents)
		_ = s.pool.QueryRow(ctx, `
			SELECT COALESCE(SUM(amount_cents),0) FROM payments
			WHERE status = 'settled' AND settled_at >= $1 AND settled_at < $2
		`, rng.From, rng.To).Scan(&pt.ReceivedCents)
		out = append(out, pt)
	}
	return out, nil
}

type RankRow struct {
	Label      string `json:"label"`
	TotalCents int64  `json:"total_cents"`
	Quantity   int64  `json:"quantity"`
	ID         string `json:"id,omitempty"`
}

type Dashboard struct {
	Year                 int       `json:"year"`
	Month                int       `json:"month"`
	SalesMonthCents      int64     `json:"sales_month_cents"`
	ReceivedMonthCents   int64     `json:"received_month_cents"`
	OpenReceivablesCents int64     `json:"open_receivables_cents"`
	OverdueInvoices      int64     `json:"overdue_invoices"`
	OverdueAmountCents   int64     `json:"overdue_amount_cents"`
	OrdersMonth          int64     `json:"orders_month"`
	AverageTicketCents   int64     `json:"average_ticket_cents"`
	BuyersCount          int64     `json:"buyers_count"`
	LowStockSKUs         int64     `json:"low_stock_skus"`
	CancelledOrdersMonth int64     `json:"cancelled_orders_month"`
	CancelledAmountCents int64     `json:"cancelled_amount_cents"`
	StockLossesMonth     int64     `json:"stock_losses_month"`
	TopProducts          []RankRow `json:"top_products"`
	TopCustomers         []RankRow `json:"top_customers"`
}

func (s *Service) Dashboard(ctx context.Context, year, month int) (Dashboard, error) {
	var d Dashboard
	d.Year = year
	d.Month = month
	rng := monthRangeUTC(year, month)

	_ = s.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(total_cents),0), COUNT(*), COUNT(DISTINCT customer_id)
		FROM orders
		WHERE status = 'confirmed' AND confirmed_at >= $1 AND confirmed_at < $2
	`, rng.From, rng.To).Scan(&d.SalesMonthCents, &d.OrdersMonth, &d.BuyersCount)
	if d.OrdersMonth > 0 {
		d.AverageTicketCents = d.SalesMonthCents / d.OrdersMonth
	}

	_ = s.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(amount_cents),0) FROM payments
		WHERE status = 'settled' AND settled_at >= $1 AND settled_at < $2
	`, rng.From, rng.To).Scan(&d.ReceivedMonthCents)

	_ = s.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(total_cents - paid_cents),0) FROM invoices
		WHERE status IN ('open','overdue')
	`).Scan(&d.OpenReceivablesCents)

	_ = s.pool.QueryRow(ctx, `
		SELECT COUNT(*), COALESCE(SUM(total_cents - paid_cents),0) FROM invoices WHERE status = 'overdue'
	`).Scan(&d.OverdueInvoices, &d.OverdueAmountCents)

	_ = s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM skus s
		LEFT JOIN inventory_balances ib ON ib.sku_id = s.id
		WHERE s.active AND COALESCE(ib.available_quantity,0) < s.minimum_stock
	`).Scan(&d.LowStockSKUs)

	_ = s.pool.QueryRow(ctx, `
		SELECT COUNT(*), COALESCE(SUM(total_cents),0) FROM orders
		WHERE status = 'cancelled' AND cancelled_at >= $1 AND cancelled_at < $2
	`, rng.From, rng.To).Scan(&d.CancelledOrdersMonth, &d.CancelledAmountCents)

	_ = s.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(quantity),0) FROM stock_movements
		WHERE movement_type IN ('loss','damage') AND created_at >= $1 AND created_at < $2
	`, rng.From, rng.To).Scan(&d.StockLossesMonth)

	topP, _ := s.TopProducts(ctx, year, month, 5)
	d.TopProducts = topP
	topC, _ := s.TopCustomers(ctx, year, month, 5)
	d.TopCustomers = topC

	return d, nil
}

func (s *Service) TopProducts(ctx context.Context, year, month int, limit int) ([]RankRow, error) {
	if limit <= 0 {
		limit = 10
	}
	rng := monthRangeUTC(year, month)
	rows, err := s.pool.Query(ctx, `
		SELECT oi.product_name_snapshot, SUM(oi.total_cents), SUM(oi.quantity)
		FROM order_items oi
		JOIN orders o ON o.id = oi.order_id
		WHERE o.status = 'confirmed' AND o.confirmed_at >= $1 AND o.confirmed_at < $2
		GROUP BY oi.product_name_snapshot
		ORDER BY SUM(oi.total_cents) DESC
		LIMIT $3
	`, rng.From, rng.To, limit)
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

func (s *Service) TopCustomers(ctx context.Context, year, month int, limit int) ([]RankRow, error) {
	if limit <= 0 {
		limit = 10
	}
	rng := monthRangeUTC(year, month)
	rows, err := s.pool.Query(ctx, `
		SELECT u.name, SUM(o.total_cents), COUNT(*), c.id::text
		FROM orders o
		JOIN customers c ON c.id = o.customer_id
		JOIN users u ON u.id = c.user_id
		WHERE o.status = 'confirmed' AND o.confirmed_at >= $1 AND o.confirmed_at < $2
		GROUP BY c.id, u.name
		ORDER BY SUM(o.total_cents) DESC
		LIMIT $3
	`, rng.From, rng.To, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []RankRow
	for rows.Next() {
		var r RankRow
		if err := rows.Scan(&r.Label, &r.TotalCents, &r.Quantity, &r.ID); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// Legacy inventory list (deprecated; use InventoryPosition).
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
