package integration_test

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/google/uuid"

	"github.com/store-platform/store/internal/billing"
	"github.com/store-platform/store/internal/catalog"
	"github.com/store-platform/store/internal/customers"
	"github.com/store-platform/store/internal/inventory"
	"github.com/store-platform/store/internal/sales"
	"github.com/store-platform/store/tests/testdb"
)

// TestConcurrentCheckoutSingleStock ensures only one checkout succeeds when stock is 1 (BK-0415).
func TestConcurrentCheckoutSingleStock(t *testing.T) {
	testdb.MigrateUp(t)
	pool := testdb.Pool(t)
	ctx := context.Background()
	if err := testdb.Reset(ctx, pool); err != nil {
		t.Fatal(err)
	}

	mgr, _ := testdb.SeedManager(ctx, pool, testdb.UniqueEmail(t, "mgr-conc"))
	prod, _ := testdb.SeedProduct(ctx, pool, "Item Único", "UNI-1", 1000)
	inv := inventory.NewService(pool)
	if err := inv.RegisterEntry(ctx, prod.SKUID, 1, mgr.UserID, "entrada", 0, 0); err != nil {
		t.Fatal(err)
	}

	billSvc := billing.NewService(pool, nil, "", nil)
	salesSvc := sales.NewService(pool, inv, billSvc, catalog.NewService(pool), customers.NewService(pool, nil), nil)

	var buyers [2]testdb.Customer
	for i := range buyers {
		c, err := testdb.SeedCustomer(ctx, pool, testdb.UniqueEmail(t, "conc-cli"), "C")
		if err != nil {
			t.Fatal(err)
		}
		if err := testdb.ApproveCustomer(ctx, pool, c.ID, mgr.UserID, 100_000); err != nil {
			t.Fatal(err)
		}
		if _, err := salesSvc.UpsertCartItem(ctx, c.ID, prod.SKUID, 1); err != nil {
			t.Fatal(err)
		}
		buyers[i] = c
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	successes := 0
	for i, b := range buyers {
		wg.Add(1)
		go func(idx int, custID, userID uuid.UUID) {
			defer wg.Done()
			_, err := salesSvc.Checkout(ctx, custID, fmt.Sprintf("conc-idem-%d", idx), userID)
			if err == nil {
				mu.Lock()
				successes++
				mu.Unlock()
			}
		}(i, b.ID, b.UserID)
	}
	wg.Wait()

	if successes != 1 {
		t.Fatalf("expected exactly 1 successful checkout, got %d", successes)
	}
	var qty int
	if err := pool.QueryRow(ctx, `SELECT available_quantity FROM inventory_balances WHERE sku_id = $1`, prod.SKUID).Scan(&qty); err != nil {
		t.Fatal(err)
	}
	if qty != 0 {
		t.Fatalf("expected stock 0, got %d", qty)
	}
}
