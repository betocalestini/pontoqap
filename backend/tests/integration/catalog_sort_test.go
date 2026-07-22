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

func TestListProductsCategoryPromocoes(t *testing.T) {
	testdb.MigrateUp(t)
	pool := testdb.Pool(t)
	ctx := context.Background()
	if err := testdb.Reset(ctx, pool); err != nil {
		t.Fatal(err)
	}
	mgr, _ := testdb.SeedManager(ctx, pool, testdb.UniqueEmail(t, "mgr"))
	inv := inventory.NewService(pool)
	cat := catalog.NewService(pool)
	regular, _ := testdb.SeedProduct(ctx, pool, "Normal", "N-1", 100)
	promo, _ := testdb.SeedProduct(ctx, pool, "Em promo", "PR-1", 100)
	_ = inv.RegisterEntry(ctx, regular.SKUID, 5, mgr.UserID, "e", 250, 0)
	_ = inv.RegisterEntry(ctx, promo.SKUID, 5, mgr.UserID, "e", 250, 0)
	_, _ = pool.Exec(ctx, `
		UPDATE products SET promo_active = TRUE, promo_margin_percent = 10,
			promo_quantity_total = 5, promo_quantity_remaining = 5 WHERE id = $1
	`, promo.ProductID)
	items, total, err := cat.ListProducts(ctx, catalog.ListProductsFilter{
		Page: 1, PageSize: 20, Admin: false, Category: "promocoes",
	})
	if err != nil {
		t.Fatal(err)
	}
	if total != 1 || len(items) != 1 || items[0].ID != promo.ProductID {
		t.Fatalf("promocoes filter: total=%d len=%d first=%s", total, len(items), firstName(items))
	}
}

func TestListProductsSortNameNotPromoFirst(t *testing.T) {
	testdb.MigrateUp(t)
	pool := testdb.Pool(t)
	ctx := context.Background()
	if err := testdb.Reset(ctx, pool); err != nil {
		t.Fatal(err)
	}
	mgr, _ := testdb.SeedManager(ctx, pool, testdb.UniqueEmail(t, "mgr"))
	inv := inventory.NewService(pool)
	cat := catalog.NewService(pool)
	z, _ := testdb.SeedProduct(ctx, pool, "Zebra", "Z-1", 100)
	p2, _ := testdb.SeedProduct(ctx, pool, "Abacaxi", "A-1", 100)
	_ = inv.RegisterEntry(ctx, z.SKUID, 5, mgr.UserID, "e", 250, 0)
	_ = inv.RegisterEntry(ctx, p2.SKUID, 5, mgr.UserID, "e", 250, 0)
	_, _ = pool.Exec(ctx, `
		UPDATE products SET promo_active = TRUE, promo_margin_percent = 10,
			promo_quantity_total = 5, promo_quantity_remaining = 5 WHERE id = $1
	`, p2.ProductID)
	items, _, err := cat.ListProducts(ctx, catalog.ListProductsFilter{
		Page: 1, PageSize: 10, Admin: false, Sort: "name",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) < 2 || items[0].ID != p2.ProductID {
		t.Fatalf("sort=name want Abacaxi first (A–Z), promo not boosted; got %v", firstName(items))
	}
}

func TestListProductsSortPriceAsc(t *testing.T) {
	testdb.MigrateUp(t)
	pool := testdb.Pool(t)
	ctx := context.Background()
	if err := testdb.Reset(ctx, pool); err != nil {
		t.Fatal(err)
	}
	mgr, _ := testdb.SeedManager(ctx, pool, testdb.UniqueEmail(t, "mgr"))
	inv := inventory.NewService(pool)
	cat := catalog.NewService(pool)
	cheap, _ := testdb.SeedProduct(ctx, pool, "Barato", "CH-1", 500)
	dear, _ := testdb.SeedProduct(ctx, pool, "Caro", "CR-1", 5000)
	_ = inv.RegisterEntry(ctx, cheap.SKUID, 5, mgr.UserID, "e", 250, 0)
	_ = inv.RegisterEntry(ctx, dear.SKUID, 5, mgr.UserID, "e", 250, 0)
	items, _, err := cat.ListProducts(ctx, catalog.ListProductsFilter{
		Page: 1, PageSize: 10, Admin: false, Sort: "price_asc",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) < 2 || items[0].ID != cheap.ProductID {
		t.Fatalf("price_asc want cheap first, got %v", firstName(items))
	}
}

func TestListProductsSortPurchasesByCustomer(t *testing.T) {
	testdb.MigrateUp(t)
	pool := testdb.Pool(t)
	ctx := context.Background()
	if err := testdb.Reset(ctx, pool); err != nil {
		t.Fatal(err)
	}
	mgr, _ := testdb.SeedManager(ctx, pool, testdb.UniqueEmail(t, "mgr"))
	cust, _ := testdb.SeedCustomer(ctx, pool, testdb.UniqueEmail(t, "cli"), "Cliente")
	_ = testdb.ApproveCustomer(ctx, pool, cust.ID, mgr.UserID, 500_000)
	inv := inventory.NewService(pool)
	catSvc := catalog.NewService(pool)
	salesSvc := sales.NewService(pool, inv, billing.NewService(pool, nil, "", nil), catSvc, customers.NewService(pool, nil))
	a, _ := testdb.SeedProduct(ctx, pool, "Prod A", "PA-1", 1000)
	b, _ := testdb.SeedProduct(ctx, pool, "Prod B", "PB-1", 1000)
	_ = inv.RegisterEntry(ctx, a.SKUID, 20, mgr.UserID, "e", 500, 0)
	_ = inv.RegisterEntry(ctx, b.SKUID, 20, mgr.UserID, "e", 500, 0)
	_, _ = salesSvc.UpsertCartItem(ctx, cust.ID, a.SKUID, 1)
	_, _ = salesSvc.Checkout(ctx, cust.ID, "sort-purch-a", cust.UserID)
	_, _ = salesSvc.UpsertCartItem(ctx, cust.ID, b.SKUID, 5)
	_, _ = salesSvc.Checkout(ctx, cust.ID, "sort-purch-b", cust.UserID)
	items, _, err := catSvc.ListProducts(ctx, catalog.ListProductsFilter{
		Page: 1, PageSize: 10, Admin: false, Sort: "purchases", CustomerID: &cust.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) < 2 || items[0].ID != b.ProductID {
		t.Fatalf("purchases sort want B first, got %v", firstName(items))
	}
}

func firstName(items []catalog.Product) string {
	if len(items) == 0 {
		return ""
	}
	return items[0].Name
}
