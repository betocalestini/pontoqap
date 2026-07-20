package integration_test

import (
	"context"
	"testing"

	"github.com/store-platform/store/internal/billing"
	"github.com/store-platform/store/internal/catalog"
	"github.com/store-platform/store/internal/customers"
	"github.com/store-platform/store/internal/inventory"
	"github.com/store-platform/store/internal/sales"
	"github.com/store-platform/store/tests/testdb"
)

func TestRegisterAdjustmentUpdatesBalance(t *testing.T) {
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
	prod, err := testdb.SeedProduct(ctx, pool, "Arroz", "ARR-1", 1200)
	if err != nil {
		t.Fatal(err)
	}
	inv := inventory.NewService(pool)
	if err := inv.RegisterEntry(ctx, prod.SKUID, 10, mgr.UserID, "entrada teste", 500); err != nil {
		t.Fatal(err)
	}
	if err := inv.RegisterAdjustment(ctx, prod.SKUID, 7, "contagem", mgr.UserID); err != nil {
		t.Fatal(err)
	}
	var qty int
	if err := pool.QueryRow(ctx, `SELECT available_quantity FROM inventory_balances WHERE sku_id = $1`, prod.SKUID).Scan(&qty); err != nil {
		t.Fatal(err)
	}
	if qty != 7 {
		t.Fatalf("expected balance 7, got %d", qty)
	}
}

func TestRegisterInitialStock(t *testing.T) {
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
	prod, err := testdb.SeedProduct(ctx, pool, "Macarrão", "MAC-1", 900)
	if err != nil {
		t.Fatal(err)
	}
	inv := inventory.NewService(pool)
	if err := inv.RegisterInitialStock(ctx, prod.SKUID, 15, mgr.UserID, 500); err != nil {
		t.Fatal(err)
	}
	var qty int
	var movType string
	if err := pool.QueryRow(ctx, `SELECT available_quantity FROM inventory_balances WHERE sku_id = $1`, prod.SKUID).Scan(&qty); err != nil {
		t.Fatal(err)
	}
	if qty != 15 {
		t.Fatalf("expected balance 15, got %d", qty)
	}
	if err := pool.QueryRow(ctx, `
		SELECT movement_type FROM stock_movements WHERE sku_id = $1 ORDER BY created_at DESC LIMIT 1
	`, prod.SKUID).Scan(&movType); err != nil {
		t.Fatal(err)
	}
	if movType != inventory.MovementInitialStock {
		t.Fatalf("expected initial_stock movement, got %s", movType)
	}
}

func TestCheckoutCreatesSaleMovement(t *testing.T) {
	testdb.MigrateUp(t)
	pool := testdb.Pool(t)
	ctx := context.Background()
	if err := testdb.Reset(ctx, pool); err != nil {
		t.Fatal(err)
	}
	mgr, _ := testdb.SeedManager(ctx, pool, testdb.UniqueEmail(t, "mgr"))
	cust, _ := testdb.SeedCustomer(ctx, pool, testdb.UniqueEmail(t, "cli"), "Cliente")
	_ = testdb.ApproveCustomer(ctx, pool, cust.ID, mgr.UserID, 100_000)
	prod, _ := testdb.SeedProduct(ctx, pool, "Café", "CAF-1", 1500)
	inv := inventory.NewService(pool)
	_ = inv.RegisterEntry(ctx, prod.SKUID, 5, mgr.UserID, "entrada", 0)

	salesSvc := sales.NewService(pool, inv, billing.NewService(pool, nil, ""), catalog.NewService(pool), customers.NewService(pool, nil))
	_, _ = salesSvc.UpsertCartItem(ctx, cust.ID, prod.SKUID, 1)
	order, err := salesSvc.Checkout(ctx, cust.ID, "sale-mov-key", cust.UserID)
	if err != nil {
		t.Fatal(err)
	}

	var movType string
	var refID *string
	if err := pool.QueryRow(ctx, `
		SELECT movement_type, reference_id::text FROM stock_movements
		WHERE sku_id = $1 AND movement_type = 'sale' ORDER BY created_at DESC LIMIT 1
	`, prod.SKUID).Scan(&movType, &refID); err != nil {
		t.Fatal(err)
	}
	if movType != "sale" {
		t.Fatalf("expected sale, got %s", movType)
	}
	if refID == nil || *refID != order.ID.String() {
		t.Fatalf("expected reference_id %s, got %v", order.ID, refID)
	}
}
