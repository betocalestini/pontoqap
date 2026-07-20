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

func TestCatalogListAfterCheckout(t *testing.T) {
	testdb.MigrateUp(t)
	pool := testdb.Pool(t)
	ctx := context.Background()
	if err := testdb.Reset(ctx, pool); err != nil {
		t.Fatal(err)
	}
	mgr, _ := testdb.SeedManager(ctx, pool, testdb.UniqueEmail(t, "mgr"))
	cust, _ := testdb.SeedCustomer(ctx, pool, testdb.UniqueEmail(t, "cli"), "Cliente")
	_ = testdb.ApproveCustomer(ctx, pool, cust.ID, mgr.UserID, 100_000)
	prod, _ := testdb.SeedProduct(ctx, pool, "Macarrão", "MAC-1", 1200)
	inv := inventory.NewService(pool)
	_ = inv.RegisterEntry(ctx, prod.SKUID, 20, mgr.UserID, "entrada", 0)
	cat := catalog.NewService(pool)
	salesSvc := sales.NewService(pool, inv, billing.NewService(pool, nil, ""), cat, customers.NewService(pool, nil))
	_, _ = salesSvc.UpsertCartItem(ctx, cust.ID, prod.SKUID, 3)
	if _, err := salesSvc.Checkout(ctx, cust.ID, "key-1", cust.UserID); err != nil {
		t.Fatal(err)
	}
	_, _, err := cat.ListProducts(ctx, catalog.ListProductsFilter{Page: 1, PageSize: 10, Admin: false})
	if err != nil {
		t.Fatal(err)
	}
}
