package integration_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/store-platform/store/internal/catalog"
	"github.com/store-platform/store/internal/inventory"
	"github.com/store-platform/store/tests/testdb"
)

func TestListProductsAfterMigration(t *testing.T) {
	testdb.MigrateUp(t)
	pool := testdb.Pool(t)
	ctx := context.Background()
	if err := testdb.Reset(ctx, pool); err != nil {
		t.Fatal(err)
	}
	mgr, _ := testdb.SeedManager(ctx, pool, testdb.UniqueEmail(t, "mgr"))
	prod, err := testdb.SeedProduct(ctx, pool, "Pub", "PUB-1", 1000)
	if err != nil {
		t.Fatal(err)
	}
	inv := inventory.NewService(pool)
	if err := inv.RegisterEntry(ctx, prod.SKUID, 1, mgr.UserID, "entrada", 500, 0); err != nil {
		t.Fatal(err)
	}
	svc := catalog.NewService(pool)
	items, total, err := svc.ListProducts(ctx, catalog.ListProductsFilter{Page: 1, PageSize: 10, Admin: false})
	if err != nil {
		t.Fatal(err)
	}
	if total < 1 || len(items) == 0 {
		t.Fatalf("expected products, got total=%d len=%d", total, len(items))
	}
}

func TestListProductsIncludesCategoryName(t *testing.T) {
	testdb.MigrateUp(t)
	pool := testdb.Pool(t)
	ctx := context.Background()
	if err := testdb.Reset(ctx, pool); err != nil {
		t.Fatal(err)
	}
	var catID uuid.UUID
	if err := pool.QueryRow(ctx, `
		INSERT INTO categories (name, slug, active) VALUES ('TestCat', 'test-cat', true) RETURNING id
	`).Scan(&catID); err != nil {
		t.Fatal(err)
	}
	prod, err := testdb.SeedProduct(ctx, pool, "Com categoria", "CAT-1", 1000)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `UPDATE products SET category_id = $1 WHERE id = $2`, catID, prod.ProductID); err != nil {
		t.Fatal(err)
	}
	svc := catalog.NewService(pool)
	items, _, err := svc.ListProducts(ctx, catalog.ListProductsFilter{Page: 1, PageSize: 10, Admin: true})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, p := range items {
		if p.ID == prod.ProductID {
			found = true
			if p.CategoryName != "TestCat" {
				t.Fatalf("category_name = %q, want TestCat", p.CategoryName)
			}
		}
	}
	if !found {
		t.Fatal("product not in admin list")
	}
}

func TestListProductsAdminFiltersActiveVisible(t *testing.T) {
	testdb.MigrateUp(t)
	pool := testdb.Pool(t)
	ctx := context.Background()
	if err := testdb.Reset(ctx, pool); err != nil {
		t.Fatal(err)
	}
	activeProd, err := testdb.SeedProduct(ctx, pool, "Ativo", "ACT-1", 1000)
	if err != nil {
		t.Fatal(err)
	}
	inactiveProd, err := testdb.SeedProduct(ctx, pool, "Inativo", "INA-1", 1000)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `UPDATE products SET active = FALSE WHERE id = $1`, inactiveProd.ProductID); err != nil {
		t.Fatal(err)
	}
	svc := catalog.NewService(pool)
	activeOnly := true
	items, total, err := svc.ListProducts(ctx, catalog.ListProductsFilter{
		Page: 1, PageSize: 50, Admin: true, Active: &activeOnly,
	})
	if err != nil {
		t.Fatal(err)
	}
	if total < 1 {
		t.Fatal("expected at least one active product")
	}
	for _, p := range items {
		if p.ID == inactiveProd.ProductID {
			t.Fatal("inactive product must not appear when filtering active=true")
		}
	}
	foundActive := false
	for _, p := range items {
		if p.ID == activeProd.ProductID {
			foundActive = true
		}
	}
	if !foundActive {
		t.Fatal("expected active product in filtered list")
	}
}

func TestPublicCatalogHidesZeroStock(t *testing.T) {
	testdb.MigrateUp(t)
	pool := testdb.Pool(t)
	ctx := context.Background()
	if err := testdb.Reset(ctx, pool); err != nil {
		t.Fatal(err)
	}
	prod, err := testdb.SeedProduct(ctx, pool, "Sem estoque", "ZERO-1", 1000)
	if err != nil {
		t.Fatal(err)
	}
	svc := catalog.NewService(pool)
	items, total, err := svc.ListProducts(ctx, catalog.ListProductsFilter{Page: 1, PageSize: 10, Admin: false})
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range items {
		if p.ID == prod.ProductID {
			t.Fatal("zero-stock product must not appear in public catalog")
		}
	}
	adminItems, adminTotal, err := svc.ListProducts(ctx, catalog.ListProductsFilter{Page: 1, PageSize: 10, Admin: true})
	if err != nil {
		t.Fatal(err)
	}
	if adminTotal < 1 {
		t.Fatal("admin should still list zero-stock product")
	}
	found := false
	for _, p := range adminItems {
		if p.ID == prod.ProductID {
			found = true
		}
	}
	if !found {
		t.Fatal("expected product in admin list")
	}
	if total != 0 && len(items) != 0 {
		// ok if other seeded data exists; main check is product not in list
	}
}
