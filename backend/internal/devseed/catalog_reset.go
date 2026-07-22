package devseed

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// resetDemoCatalog removes catalog, inventory and demo commerce data so seed can reload CSV.
// Does not touch users, customers or RBAC bootstrap.
func resetDemoCatalog(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
		TRUNCATE TABLE
			forecast_snapshots,
			payment_events, payments, payment_charges,
			billing_adjustments, invoice_items, invoices, billing_entries, billing_periods,
			order_return_items, order_returns, order_items, orders,
			cart_items, carts,
			stock_movements, inventory_lots, inventory_balances,
			price_history, product_images, skus, products, categories
		RESTART IDENTITY CASCADE
	`)
	return err
}
