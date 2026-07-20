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

func TestCollaboratorCheckoutLowerPrice(t *testing.T) {
	testdb.MigrateUp(t)
	pool := testdb.Pool(t)
	ctx := context.Background()
	if err := testdb.Reset(ctx, pool); err != nil {
		t.Fatal(err)
	}
	mgr, _ := testdb.SeedManager(ctx, pool, testdb.UniqueEmail(t, "mgr"))
	cust, _ := testdb.SeedCustomer(ctx, pool, testdb.UniqueEmail(t, "cli"), "Colab")
	_ = testdb.ApproveCustomer(ctx, pool, cust.ID, mgr.UserID, 500_000)
	prod, _ := testdb.SeedProduct(ctx, pool, "Item", "IT-1", 0)
	inv := inventory.NewService(pool)
	cat := catalog.NewService(pool)
	custSvc := customers.NewService(pool, nil)
	_ = inv.RegisterEntry(ctx, prod.SKUID, 10, mgr.UserID, "e", 1000)
	_, _ = pool.Exec(ctx, `UPDATE products SET margin_percent = 50 WHERE id = $1`, prod.ProductID)
	_, _ = cat.RecalculateSKU(ctx, prod.SKUID, mgr.UserID, "t", inv.WeightedAverageCostCents)
	var catID uuid.UUID
	_ = pool.QueryRow(ctx, `SELECT id FROM collaborator_categories WHERE slug = 'funcionario'`).Scan(&catID)
	_, _ = pool.Exec(ctx, `UPDATE customers SET collaborator_category_id = $2 WHERE id = $1`, cust.ID, catID)
	salesSvc := sales.NewService(pool, inv, billing.NewService(pool, nil, ""), cat, custSvc)
	cart, _ := salesSvc.UpsertCartItem(ctx, cust.ID, prod.SKUID, 1)
	retail := cart.Items[0].LineTotalCents
	if retail >= 1500 {
		t.Fatalf("expected collab price below retail 1500, got %d", retail)
	}
}

func TestBlockedCustomerCheckout(t *testing.T) {
	testdb.MigrateUp(t)
	pool := testdb.Pool(t)
	ctx := context.Background()
	if err := testdb.Reset(ctx, pool); err != nil {
		t.Fatal(err)
	}
	mgr, _ := testdb.SeedManager(ctx, pool, testdb.UniqueEmail(t, "mgr"))
	cust, _ := testdb.SeedCustomer(ctx, pool, testdb.UniqueEmail(t, "cli"), "Bloq")
	_ = testdb.ApproveCustomer(ctx, pool, cust.ID, mgr.UserID, 100_000)
	prod, _ := testdb.SeedProduct(ctx, pool, "X", "X-1", 1000)
	inv := inventory.NewService(pool)
	_ = inv.RegisterEntry(ctx, prod.SKUID, 5, mgr.UserID, "e", 100)
	custSvc := customers.NewService(pool, nil)
	_ = custSvc.Block(ctx, cust.ID, "teste")
	salesSvc := sales.NewService(pool, inv, billing.NewService(pool, nil, ""), catalog.NewService(pool), custSvc)
	_, _ = salesSvc.UpsertCartItem(ctx, cust.ID, prod.SKUID, 1)
	_, err := salesSvc.Checkout(ctx, cust.ID, "blocked-key", cust.UserID)
	if err == nil {
		t.Fatal("expected blocked error")
	}
}
