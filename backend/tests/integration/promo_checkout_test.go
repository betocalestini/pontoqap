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

func TestCheckoutPromoSplitPricing(t *testing.T) {
	testdb.MigrateUp(t)
	pool := testdb.Pool(t)
	ctx := context.Background()
	if err := testdb.Reset(ctx, pool); err != nil {
		t.Fatal(err)
	}
	mgr, _ := testdb.SeedManager(ctx, pool, testdb.UniqueEmail(t, "mgr"))
	cust, _ := testdb.SeedCustomer(ctx, pool, testdb.UniqueEmail(t, "cli"), "Cliente")
	_ = testdb.ApproveCustomer(ctx, pool, cust.ID, mgr.UserID, 500_000)
	prod, _ := testdb.SeedProduct(ctx, pool, "Promo Split", "PRM-1", 0)
	inv := inventory.NewService(pool)
	cat := catalog.NewService(pool)
	if err := inv.RegisterEntry(ctx, prod.SKUID, 20, mgr.UserID, "entrada", 20_000, 0); err != nil {
		t.Fatal(err)
	}
	_, _ = cat.RecalculateSKU(ctx, prod.SKUID, mgr.UserID, "setup", inv.WeightedAverageCostCents)
	pm := 10.0
	_, err := pool.Exec(ctx, `
		UPDATE products SET margin_percent = 50, promo_active = TRUE, promo_margin_percent = $2,
			promo_quantity_total = 3, promo_quantity_remaining = 3
		WHERE id = $1
	`, prod.ProductID, pm)
	if err != nil {
		t.Fatal(err)
	}
	_ = cat.RecalculateProductSKUs(ctx, prod.ProductID, mgr.UserID, "promo", inv.WeightedAverageCostCents)
	var promoPrice int64
	_ = pool.QueryRow(ctx, `SELECT sale_price_cents FROM skus WHERE id = $1`, prod.SKUID).Scan(&promoPrice)
	if promoPrice != 1100 {
		t.Fatalf("promo price want 1100 got %d", promoPrice)
	}
	regular := catalog.SalePriceFromCost(1000, 50)
	salesSvc := sales.NewService(pool, inv, billing.NewService(pool, nil, "", nil), cat, customers.NewService(pool, nil), nil)
	_, _ = salesSvc.UpsertCartItem(ctx, cust.ID, prod.SKUID, 5)
	order, err := salesSvc.Checkout(ctx, cust.ID, "promo-split-key", cust.UserID)
	if err != nil {
		t.Fatal(err)
	}
	wantTotal := 3*promoPrice + 2*regular
	if order.TotalCents != wantTotal {
		t.Fatalf("total want %d got %d", wantTotal, order.TotalCents)
	}
	var rem int
	_ = pool.QueryRow(ctx, `SELECT promo_quantity_remaining FROM products WHERE id = $1`, prod.ProductID).Scan(&rem)
	if rem != 0 {
		t.Fatalf("remaining want 0 got %d", rem)
	}
}

func TestListProductsPromoFirst(t *testing.T) {
	testdb.MigrateUp(t)
	pool := testdb.Pool(t)
	ctx := context.Background()
	if err := testdb.Reset(ctx, pool); err != nil {
		t.Fatal(err)
	}
	mgr, _ := testdb.SeedManager(ctx, pool, testdb.UniqueEmail(t, "mgr"))
	inv := inventory.NewService(pool)
	z, _ := testdb.SeedProduct(ctx, pool, "Zebra", "Z-1", 100)
	p2, _ := testdb.SeedProduct(ctx, pool, "Abacaxi", "A-1", 100)
	_ = inv.RegisterEntry(ctx, z.SKUID, 5, mgr.UserID, "e", 250, 0)
	_ = inv.RegisterEntry(ctx, p2.SKUID, 5, mgr.UserID, "e", 250, 0)
	_, _ = pool.Exec(ctx, `
		UPDATE products SET promo_active = TRUE, promo_margin_percent = 10,
			promo_quantity_total = 5, promo_quantity_remaining = 5 WHERE id = $1
	`, p2.ProductID)
	cat := catalog.NewService(pool)
	items, _, err := cat.ListProducts(ctx, catalog.ListProductsFilter{Page: 1, PageSize: 10, Admin: false})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) < 2 || items[0].ID != p2.ProductID {
		if len(items) == 0 {
			t.Fatal("expected products in public catalog")
		}
		t.Fatalf("expected promo product first, got %v", items[0].Name)
	}
}
