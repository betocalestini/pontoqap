package reports

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type SalesOrderRow struct {
	ID              uuid.UUID  `json:"id"`
	OrderNumber     string     `json:"order_number"`
	ConfirmedAt     *time.Time `json:"confirmed_at,omitempty"`
	CustomerID      uuid.UUID  `json:"customer_id"`
	CustomerName    string     `json:"customer_name"`
	ItemCount       int        `json:"item_count"`
	SubtotalCents   int64      `json:"subtotal_cents"`
	DiscountCents   int64      `json:"discount_cents"`
	TotalCents      int64      `json:"total_cents"`
	Status          string     `json:"status"`
	ReferenceYear   int        `json:"reference_year,omitempty"`
	ReferenceMonth  int        `json:"reference_month,omitempty"`
}

type SalesSummary struct {
	TotalSalesCents int64 `json:"total_sales_cents"`
	OrderCount      int   `json:"order_count"`
	AverageTicket   int64 `json:"average_ticket_cents"`
	ProductUnits    int64 `json:"product_units"`
}

type SalesOrdersFilter struct {
	DateRange
	Status      string
	CustomerID  *uuid.UUID
	ProductID   *uuid.UUID
	CategoryID  *uuid.UUID
	MinTotal    *int64
	MaxTotal    *int64
	Cancelled   *bool
	PageFilter
}

func (s *Service) SalesOrders(ctx context.Context, f SalesOrdersFilter) ([]SalesOrderRow, SalesSummary, int, error) {
	where := []string{"1=1"}
	args := []any{}
	n := 1

	if f.Status != "" {
		where = append(where, fmt.Sprintf("o.status = $%d", n))
		args = append(args, f.Status)
		n++
	} else if f.Cancelled != nil && *f.Cancelled {
		where = append(where, "o.status = 'cancelled'")
	} else if f.Cancelled != nil && !*f.Cancelled {
		where = append(where, "o.status = 'confirmed'")
	} else {
		where = append(where, "o.confirmed_at >= $"+fmt.Sprint(n)+" AND o.confirmed_at < $"+fmt.Sprint(n+1))
		args = append(args, f.From, f.To)
		n += 2
	}

	if f.CustomerID != nil {
		where = append(where, fmt.Sprintf("o.customer_id = $%d", n))
		args = append(args, *f.CustomerID)
		n++
	}
	if f.ProductID != nil {
		where = append(where, fmt.Sprintf(`EXISTS (
			SELECT 1 FROM order_items oi JOIN skus sk ON sk.id = oi.sku_id
			WHERE oi.order_id = o.id AND sk.product_id = $%d
		)`, n))
		args = append(args, *f.ProductID)
		n++
	}
	if f.CategoryID != nil {
		where = append(where, fmt.Sprintf(`EXISTS (
			SELECT 1 FROM order_items oi
			JOIN skus sk ON sk.id = oi.sku_id
			JOIN products p ON p.id = sk.product_id
			WHERE oi.order_id = o.id AND p.category_id = $%d
		)`, n))
		args = append(args, *f.CategoryID)
		n++
	}
	if f.MinTotal != nil {
		where = append(where, fmt.Sprintf("o.total_cents >= $%d", n))
		args = append(args, *f.MinTotal)
		n++
	}
	if f.MaxTotal != nil {
		where = append(where, fmt.Sprintf("o.total_cents <= $%d", n))
		args = append(args, *f.MaxTotal)
		n++
	}

	whereSQL := strings.Join(where, " AND ")

	var summary SalesSummary
	sumQ := fmt.Sprintf(`
		SELECT COALESCE(SUM(o.total_cents),0), COUNT(*)::int,
		       COALESCE(SUM(oi.quantity),0)
		FROM orders o
		LEFT JOIN order_items oi ON oi.order_id = o.id
		JOIN customers c ON c.id = o.customer_id
		JOIN users u ON u.id = c.user_id
		WHERE %s
	`, whereSQL)
	if err := s.pool.QueryRow(ctx, sumQ, args...).Scan(&summary.TotalSalesCents, &summary.OrderCount, &summary.ProductUnits); err != nil {
		return nil, summary, 0, err
	}
	if summary.OrderCount > 0 {
		summary.AverageTicket = summary.TotalSalesCents / int64(summary.OrderCount)
	}

	var total int
	countQ := fmt.Sprintf(`
		SELECT COUNT(*) FROM orders o
		JOIN customers c ON c.id = o.customer_id
		WHERE %s
	`, whereSQL)
	if err := s.pool.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, summary, 0, err
	}

	limit := f.Limit
	offset := f.Offset
	q := fmt.Sprintf(`
		SELECT o.id, o.order_number, o.confirmed_at, o.customer_id, u.name,
		       (SELECT COUNT(*)::int FROM order_items WHERE order_id = o.id),
		       o.subtotal_cents, o.discount_cents, o.total_cents, o.status,
		       COALESCE(bp.reference_year, 0), COALESCE(bp.reference_month, 0)
		FROM orders o
		JOIN customers c ON c.id = o.customer_id
		JOIN users u ON u.id = c.user_id
		LEFT JOIN billing_entries be ON be.order_id = o.id
		LEFT JOIN billing_periods bp ON bp.id = be.billing_period_id
		WHERE %s
		GROUP BY o.id, u.name, bp.reference_year, bp.reference_month
		ORDER BY COALESCE(o.confirmed_at, o.created_at) DESC
		LIMIT $%d OFFSET $%d
	`, whereSQL, n, n+1)
	args = append(args, limit, offset)

	rows, err := s.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, summary, 0, err
	}
	defer rows.Close()

	var out []SalesOrderRow
	for rows.Next() {
		var r SalesOrderRow
		if err := rows.Scan(
			&r.ID, &r.OrderNumber, &r.ConfirmedAt, &r.CustomerID, &r.CustomerName,
			&r.ItemCount, &r.SubtotalCents, &r.DiscountCents, &r.TotalCents, &r.Status,
			&r.ReferenceYear, &r.ReferenceMonth,
		); err != nil {
			return nil, summary, 0, err
		}
		out = append(out, r)
	}
	return out, summary, total, rows.Err()
}
