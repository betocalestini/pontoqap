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

func TestClearCartRemovesAllItems(t *testing.T) {
	testdb.MigrateUp(t)
	pool := testdb.Pool(t)
	ctx := context.Background()
	if err := testdb.Reset(ctx, pool); err != nil {
		t.Fatal(err)
	}
	mgr, _ := testdb.SeedManager(ctx, pool, testdb.UniqueEmail(t, "mgr-cart"))
	cust, _ := testdb.SeedCustomer(ctx, pool, testdb.UniqueEmail(t, "cust-cart"), "Cliente Cart")
	_ = testdb.ApproveCustomer(ctx, pool, cust.ID, mgr.UserID, 100_000)
	prod, _ := testdb.SeedProduct(ctx, pool, "Item Cart", "CRT-1", 500)
	inv := inventory.NewService(pool)
	_ = inv.RegisterEntry(ctx, prod.SKUID, 10, mgr.UserID, "entrada", 5000, 0)

	salesSvc := sales.NewService(pool, inv, billing.NewService(pool, nil, "", nil), catalog.NewService(pool), customers.NewService(pool, nil))
	cart, err := salesSvc.AddCartItem(ctx, cust.ID, prod.SKUID, 2)
	if err != nil || len(cart.Items) != 1 {
		t.Fatalf("add cart: %v items=%d", err, len(cart.Items))
	}

	cleared, err := salesSvc.ClearCart(ctx, cust.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(cleared.Items) != 0 {
		t.Fatalf("expected empty cart, got %d items", len(cleared.Items))
	}

	again, err := salesSvc.GetOrCreateCart(ctx, cust.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(again.Items) != 0 {
		t.Fatalf("expected empty cart on reload, got %d items", len(again.Items))
	}
}

func TestSetCartItemQuantityZeroRemovesLine(t *testing.T) {
	testdb.MigrateUp(t)
	pool := testdb.Pool(t)
	ctx := context.Background()
	if err := testdb.Reset(ctx, pool); err != nil {
		t.Fatal(err)
	}
	mgr, _ := testdb.SeedManager(ctx, pool, testdb.UniqueEmail(t, "mgr-cart2"))
	cust, _ := testdb.SeedCustomer(ctx, pool, testdb.UniqueEmail(t, "cust-cart2"), "Cliente Cart 2")
	_ = testdb.ApproveCustomer(ctx, pool, cust.ID, mgr.UserID, 100_000)
	prod, _ := testdb.SeedProduct(ctx, pool, "Remover", "RM-1", 300)
	inv := inventory.NewService(pool)
	_ = inv.RegisterEntry(ctx, prod.SKUID, 5, mgr.UserID, "entrada", 1500, 0)

	salesSvc := sales.NewService(pool, inv, billing.NewService(pool, nil, "", nil), catalog.NewService(pool), customers.NewService(pool, nil))
	_, err := salesSvc.AddCartItem(ctx, cust.ID, prod.SKUID, 3)
	if err != nil {
		t.Fatal(err)
	}
	cart, err := salesSvc.SetCartItemQuantity(ctx, cust.ID, prod.SKUID, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(cart.Items) != 0 {
		t.Fatalf("expected no items after qty 0, got %d", len(cart.Items))
	}
}
