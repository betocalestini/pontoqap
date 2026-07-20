package integration_test

import (
	"context"
	"testing"

	"github.com/store-platform/store/internal/catalog"
	"github.com/store-platform/store/internal/customers"
	"github.com/store-platform/store/tests/testdb"
)

func TestCatalogCreateProductAndChangePrice(t *testing.T) {
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
	_ = mgr

	svc := catalog.NewService(pool)
	p, err := svc.CreateProduct(ctx, catalog.CreateProductInput{
		Name: "Arroz", SKUCode: "ARZ-001", SalePrice: 1500,
	})
	if err != nil || len(p.SKUs) != 1 {
		t.Fatalf("create product: %v", err)
	}
	skuID := p.SKUs[0].ID

	if err := svc.ChangeSKUPrice(ctx, skuID, 1800, mgr.UserID, "reajuste"); err != nil {
		t.Fatal(err)
	}
	var price int64
	if err := pool.QueryRow(ctx, `SELECT sale_price_cents FROM skus WHERE id = $1`, skuID).Scan(&price); err != nil {
		t.Fatal(err)
	}
	if price != 1800 {
		t.Fatalf("expected price 1800, got %d", price)
	}
	var historyCount int
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM price_history WHERE sku_id = $1`, skuID).Scan(&historyCount); err != nil {
		t.Fatal(err)
	}
	if historyCount != 1 {
		t.Fatalf("expected 1 price history row, got %d", historyCount)
	}
}

func TestCustomerApproveAndLimit(t *testing.T) {
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

	svc := customers.NewService(pool)
	if err := svc.Approve(ctx, cust.ID, mgr.UserID, 50_000); err != nil {
		t.Fatal(err)
	}
	got, err := svc.GetByID(ctx, cust.ID)
	if err != nil || got.Status != "approved" || got.CreditLimitCents != 50_000 {
		t.Fatalf("approve: %+v err=%v", got, err)
	}
	if svc.AvailableLimit(*got) != 50_000 {
		t.Fatal("available limit should match credit after approve")
	}
}
