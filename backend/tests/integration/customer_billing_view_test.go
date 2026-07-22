package integration_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/store-platform/store/internal/billing"
	"github.com/store-platform/store/tests/testdb"
)

func seedOpenPeriodWithOrderItems(t *testing.T, ctx context.Context, custID, skuID uuid.UUID, year, month int, productName string, qty int, unitCents int64) uuid.UUID {
	t.Helper()
	pool := testdb.Pool(t)
	svc := billing.NewService(pool, nil, "", nil)
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	at := time.Date(year, time.Month(month), 10, 12, 0, 0, 0, time.UTC)
	periodID, err := svc.EnsureOpenPeriodTx(ctx, tx, custID, at)
	if err != nil {
		t.Fatal(err)
	}
	orderID := uuid.New()
	total := unitCents * int64(qty)
	ordNum := "V-" + orderID.String()[:8]
	if _, err := tx.Exec(ctx, `
		INSERT INTO orders (id, order_number, customer_id, status, subtotal_cents, total_cents, idempotency_key, confirmed_at)
		VALUES ($1, $2, $3, 'confirmed', $4, $4, $5, NOW())
	`, orderID, ordNum, custID, total, orderID.String()); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO order_items (order_id, sku_id, product_name_snapshot, sku_code_snapshot, unit_price_cents, quantity, total_cents)
		VALUES ($1, $2, $3, 'SKU-1', $4, $5, $6)
	`, orderID, skuID, productName, unitCents, qty, total); err != nil {
		t.Fatal(err)
	}
	if err := svc.AddOrderEntryTx(ctx, tx, custID, orderID, total, at); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	return periodID
}

func TestOpenPeriodAndInvoiceDetailProducts(t *testing.T) {
	testdb.MigrateUp(t)
	pool := testdb.Pool(t)
	ctx := context.Background()
	if err := testdb.Reset(ctx, pool); err != nil {
		t.Fatal(err)
	}
	mgr, _ := testdb.SeedManager(ctx, pool, testdb.UniqueEmail(t, "mgr"))
	cust, _ := testdb.SeedCustomer(ctx, pool, testdb.UniqueEmail(t, "view"), "Cliente View")
	_ = testdb.ApproveCustomer(ctx, pool, cust.ID, mgr.UserID, 100_000)
	prod, err := testdb.SeedProduct(ctx, pool, "Arroz 5kg", testdb.UniqueEmail(t, "sku"), 1200)
	if err != nil {
		t.Fatal(err)
	}

	_ = seedOpenPeriodWithOrderItems(t, ctx, cust.ID, prod.SKUID, 2026, 7, "Arroz 5kg", 2, 1200)

	svc := billing.NewService(pool, nil, "", nil)
	open, err := svc.GetOpenPeriodDetail(ctx, cust.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(open.Entries) != 1 || len(open.Entries[0].Products) != 1 {
		t.Fatalf("open period entries: %+v", open.Entries)
	}
	if open.Entries[0].Products[0].ProductName != "Arroz 5kg" || open.Entries[0].Products[0].Quantity != 2 {
		t.Fatalf("products: %+v", open.Entries[0].Products)
	}

	inv, err := svc.CloseCustomerOpenPeriod(ctx, cust.ID)
	if err != nil {
		t.Fatal(err)
	}
	detail, err := svc.GetInvoiceDetail(ctx, inv.ID)
	if err != nil || detail == nil {
		t.Fatal(err)
	}
	if len(detail.Items) != 1 || len(detail.Items[0].Products) != 1 {
		t.Fatalf("invoice items: %+v", detail.Items)
	}
	if detail.Items[0].Products[0].ProductName != "Arroz 5kg" {
		t.Fatalf("invoice product: %+v", detail.Items[0].Products)
	}
}
