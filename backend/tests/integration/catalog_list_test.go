package integration_test

import (
	"context"
	"testing"

	"github.com/store-platform/store/internal/catalog"
	"github.com/store-platform/store/tests/testdb"
)

func TestListProductsAfterMigration(t *testing.T) {
	testdb.MigrateUp(t)
	pool := testdb.Pool(t)
	ctx := context.Background()
	if err := testdb.Reset(ctx, pool); err != nil {
		t.Fatal(err)
	}
	_, err := testdb.SeedProduct(ctx, pool, "Pub", "PUB-1", 1000)
	if err != nil {
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
