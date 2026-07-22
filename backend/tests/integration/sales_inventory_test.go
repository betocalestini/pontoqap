package integration_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/store-platform/store/internal/billing"
	"github.com/store-platform/store/internal/catalog"
	"github.com/store-platform/store/internal/customers"
	"github.com/store-platform/store/internal/inventory"
	"github.com/store-platform/store/internal/sales"
	"github.com/store-platform/store/tests/testdb"
)

func TestInventoryEntryAndNegativeStockBlocked(t *testing.T) {
	testdb.MigrateUp(t)
	pool := testdb.Pool(t)
	ctx := context.Background()
	if err := testdb.Reset(ctx, pool); err != nil {
		t.Fatal(err)
	}

	mgr, err := testdb.SeedManager(ctx, pool, testdb.UniqueEmail(t, "mgr"))
	if err != nil {
		t.Fatal(err)
	}
	prod, err := testdb.SeedProduct(ctx, pool, "Feijão", "FEJ-1", 800)
	if err != nil {
		t.Fatal(err)
	}

	inv := inventory.NewService(pool)
	if err := inv.RegisterEntry(ctx, prod.SKUID, 5, mgr.UserID, "entrada", 0, 0); err != nil {
		t.Fatal(err)
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	orderID := uuid.New()
	err = inv.ReserveAndDecrement(ctx, tx, prod.SKUID, 10, "order", orderID, &mgr.UserID)
	if err == nil {
		t.Fatal("expected insufficient stock error")
	}
	_ = tx.Rollback(ctx)
}

func TestCheckoutReducesStockAndCreatesBillingEntry(t *testing.T) {
	testdb.MigrateUp(t)
	pool := testdb.Pool(t)
	ctx := context.Background()
	if err := testdb.Reset(ctx, pool); err != nil {
		t.Fatal(err)
	}

	mgr, err := testdb.SeedManager(ctx, pool, testdb.UniqueEmail(t, "mgr"))
	if err != nil {
		t.Fatal(err)
	}
	cust, err := testdb.SeedCustomer(ctx, pool, testdb.UniqueEmail(t, "cli"), "Cliente")
	if err != nil {
		t.Fatal(err)
	}
	if err := testdb.ApproveCustomer(ctx, pool, cust.ID, mgr.UserID, 100_000); err != nil {
		t.Fatal(err)
	}
	prod, err := testdb.SeedProduct(ctx, pool, "Óleo", "OLE-1", 2000)
	if err != nil {
		t.Fatal(err)
	}
	inv := inventory.NewService(pool)
	if err := inv.RegisterEntry(ctx, prod.SKUID, 10, mgr.UserID, "entrada", 0, 0); err != nil {
		t.Fatal(err)
	}

	salesSvc := sales.NewService(pool, inv, billing.NewService(pool, nil, "", nil), catalog.NewService(pool), customers.NewService(pool, nil))
	if _, err := salesSvc.UpsertCartItem(ctx, cust.ID, prod.SKUID, 2); err != nil {
		t.Fatal(err)
	}
	order, err := salesSvc.Checkout(ctx, cust.ID, "idem-checkout-1", cust.UserID)
	if err != nil {
		t.Fatal(err)
	}
	if order.TotalCents != 4000 {
		t.Fatalf("total expected 4000, got %d", order.TotalCents)
	}

	var qty int
	if err := pool.QueryRow(ctx, `
		SELECT available_quantity FROM inventory_balances WHERE sku_id = $1
	`, prod.SKUID).Scan(&qty); err != nil {
		t.Fatal(err)
	}
	if qty != 8 {
		t.Fatalf("stock expected 8, got %d", qty)
	}

	var exposure int64
	if err := pool.QueryRow(ctx, `SELECT current_exposure_cents FROM customers WHERE id = $1`, cust.ID).Scan(&exposure); err != nil {
		t.Fatal(err)
	}
	if exposure != 4000 {
		t.Fatalf("exposure expected 4000, got %d", exposure)
	}

	var entries int
	if err := pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM billing_entries be
		JOIN billing_periods bp ON bp.id = be.billing_period_id
		WHERE bp.customer_id = $1 AND be.order_id = $2
	`, cust.ID, order.ID).Scan(&entries); err != nil {
		t.Fatal(err)
	}
	if entries != 1 {
		t.Fatalf("expected 1 billing entry, got %d", entries)
	}
}

func TestCheckoutIdempotency(t *testing.T) {
	testdb.MigrateUp(t)
	pool := testdb.Pool(t)
	ctx := context.Background()
	if err := testdb.Reset(ctx, pool); err != nil {
		t.Fatal(err)
	}

	mgr, _ := testdb.SeedManager(ctx, pool, testdb.UniqueEmail(t, "mgr"))
	cust, _ := testdb.SeedCustomer(ctx, pool, testdb.UniqueEmail(t, "cli"), "Cliente")
	_ = testdb.ApproveCustomer(ctx, pool, cust.ID, mgr.UserID, 100_000)
	prod, _ := testdb.SeedProduct(ctx, pool, "Sal", "SAL-1", 500)
	inv := inventory.NewService(pool)
	_ = inv.RegisterEntry(ctx, prod.SKUID, 5, mgr.UserID, "entrada", 0, 0)

	salesSvc := sales.NewService(pool, inv, billing.NewService(pool, nil, "", nil), catalog.NewService(pool), customers.NewService(pool, nil))
	_, _ = salesSvc.UpsertCartItem(ctx, cust.ID, prod.SKUID, 1)

	o1, err := salesSvc.Checkout(ctx, cust.ID, "same-key", cust.UserID)
	if err != nil {
		t.Fatal(err)
	}
	o2, err := salesSvc.Checkout(ctx, cust.ID, "same-key", cust.UserID)
	if err != nil {
		t.Fatal(err)
	}
	if o1.ID != o2.ID {
		t.Fatalf("idempotent checkout should return same order: %s vs %s", o1.ID, o2.ID)
	}
	var orderCount int
	_ = pool.QueryRow(ctx, `SELECT COUNT(*) FROM orders WHERE idempotency_key = 'same-key'`).Scan(&orderCount)
	if orderCount != 1 {
		t.Fatalf("expected 1 order row, got %d", orderCount)
	}
}

func TestCheckoutInsufficientLimit(t *testing.T) {
	testdb.MigrateUp(t)
	pool := testdb.Pool(t)
	ctx := context.Background()
	if err := testdb.Reset(ctx, pool); err != nil {
		t.Fatal(err)
	}

	mgr, _ := testdb.SeedManager(ctx, pool, testdb.UniqueEmail(t, "mgr"))
	cust, _ := testdb.SeedCustomer(ctx, pool, testdb.UniqueEmail(t, "cli"), "Cliente")
	_ = testdb.ApproveCustomer(ctx, pool, cust.ID, mgr.UserID, 100) // R$ 1,00
	prod, _ := testdb.SeedProduct(ctx, pool, "Caro", "CAR-1", 5000)
	inv := inventory.NewService(pool)
	_ = inv.RegisterEntry(ctx, prod.SKUID, 5, mgr.UserID, "entrada", 0, 0)

	salesSvc := sales.NewService(pool, inv, billing.NewService(pool, nil, "", nil), catalog.NewService(pool), customers.NewService(pool, nil))
	_, _ = salesSvc.UpsertCartItem(ctx, cust.ID, prod.SKUID, 1)
	_, err := salesSvc.Checkout(ctx, cust.ID, "limit-key", cust.UserID)
	if err == nil {
		t.Fatal("expected insufficient limit")
	}
	ae := sales.AsAppError(err)
	if ae == nil {
		t.Fatalf("expected app error, got %v", err)
	}
}
