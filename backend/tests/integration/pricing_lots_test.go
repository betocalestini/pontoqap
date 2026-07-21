package integration_test

import (
	"context"
	"testing"

	"github.com/store-platform/store/internal/catalog"
	"github.com/store-platform/store/internal/inventory"
	"github.com/store-platform/store/tests/testdb"
)

func TestEntryCreatesLotAndRecalculatesPrice(t *testing.T) {
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
	prod, err := testdb.SeedProduct(ctx, pool, "Margem Teste", "MARG-1", 0)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = pool.Exec(ctx, `UPDATE products SET margin_percent = 50 WHERE id = $1`, prod.ProductID)

	inv := inventory.NewService(pool)
	cat := catalog.NewService(pool)
	if err := inv.RegisterEntry(ctx, prod.SKUID, 10, mgr.UserID, "compra", 10_000, 0); err != nil {
		t.Fatal(err)
	}
	changed, err := cat.RecalculateSKU(ctx, prod.SKUID, mgr.UserID, "test", inv.WeightedAverageCostCents)
	if err != nil || !changed {
		t.Fatalf("recalc: changed=%v err=%v", changed, err)
	}
	var price int64
	if err := pool.QueryRow(ctx, `SELECT sale_price_cents FROM skus WHERE id = $1`, prod.SKUID).Scan(&price); err != nil {
		t.Fatal(err)
	}
	if price != 1500 {
		t.Fatalf("expected sale 1500 (50%% markup on 1000), got %d", price)
	}
}

func TestRepriceAllUpdatesMarginsAndPrices(t *testing.T) {
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
	prod, err := testdb.SeedProduct(ctx, pool, "Bulk", "BLK-1", 0)
	if err != nil {
		t.Fatal(err)
	}
	inv := inventory.NewService(pool)
	cat := catalog.NewService(pool)
	if err := inv.RegisterEntry(ctx, prod.SKUID, 5, mgr.UserID, "compra", 10_000, 0); err != nil {
		t.Fatal(err)
	}
	_, _ = cat.RecalculateSKU(ctx, prod.SKUID, mgr.UserID, "setup", inv.WeightedAverageCostCents)
	n, err := cat.RepriceAllProducts(ctx, 100, mgr.UserID, inv.WeightedAverageCostCents)
	if err != nil || n < 1 {
		t.Fatalf("reprice-all: n=%d err=%v", n, err)
	}
	var margin float64
	var price int64
	if err := pool.QueryRow(ctx, `SELECT margin_percent FROM products WHERE id = $1`, prod.ProductID).Scan(&margin); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, `SELECT sale_price_cents FROM skus WHERE id = $1`, prod.SKUID).Scan(&price); err != nil {
		t.Fatal(err)
	}
	if margin != 100 {
		t.Fatalf("margin want 100 got %v", margin)
	}
	if price != 4000 {
		t.Fatalf("price want 4000 got %d", price)
	}
}
