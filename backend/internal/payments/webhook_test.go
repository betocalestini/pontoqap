package payments_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/store-platform/store/internal/billing"
	"github.com/store-platform/store/internal/payments"
	"github.com/store-platform/store/tests/testdb"
)

func TestPixWebhookAmountMismatchRejected(t *testing.T) {
	testdb.MigrateUp(t)
	pool := testdb.Pool(t)
	ctx := context.Background()
	if err := testdb.Reset(ctx, pool); err != nil {
		t.Fatal(err)
	}

	mgr, _ := testdb.SeedManager(ctx, pool, testdb.UniqueEmail(t, "mgr-wh"))
	cust, _ := testdb.SeedCustomer(ctx, pool, testdb.UniqueEmail(t, "c-wh"), "Cliente")
	_ = testdb.ApproveCustomer(ctx, pool, cust.ID, mgr.UserID, 50_000)

	billSvc := billing.NewService(pool, nil, "")
	tx, _ := pool.Begin(ctx)
	at := time.Date(2026, 3, 10, 10, 0, 0, 0, time.UTC)
	periodID, _ := billSvc.EnsureOpenPeriodTx(ctx, tx, cust.ID, at)
	orderID := uuid.New()
	_, _ = tx.Exec(ctx, `
		INSERT INTO orders (id, order_number, customer_id, status, subtotal_cents, total_cents, idempotency_key, confirmed_at)
		VALUES ($1, $2, $3, 'confirmed', 2000, 2000, $4, NOW())
	`, orderID, "T-"+orderID.String()[:8], cust.ID, orderID.String())
	_ = billSvc.AddOrderEntryTx(ctx, tx, cust.ID, orderID, 2000, at)
	_ = tx.Commit(ctx)
	inv, err := billSvc.ClosePeriod(ctx, periodID)
	if err != nil {
		t.Fatal(err)
	}

	secret := "test-webhook-secret"
	paySvc := payments.NewService(pool, payments.NewSandboxGateway(secret), billSvc, secret)
	charge, err := paySvc.CreateOrReusePixCharge(ctx, inv.ID)
	if err != nil {
		t.Fatal(err)
	}

	var extID string
	if err := pool.QueryRow(ctx, `SELECT external_id FROM payment_charges WHERE id = $1`, charge.ID).Scan(&extID); err != nil {
		t.Fatal(err)
	}
	body := []byte(`{"event_id":"evt-mismatch-1","event_type":"payment.confirmed","payment_id":"` + extID + `","amount_cents":1}`)
	sig := payments.SignPayloadForTest(body, secret)
	if err := paySvc.ProcessWebhook(ctx, body, sig); err == nil {
		t.Fatal("expected amount mismatch error")
	}

	var paid int64
	_ = pool.QueryRow(ctx, `SELECT paid_cents FROM invoices WHERE id = $1`, inv.ID).Scan(&paid)
	if paid != 0 {
		t.Fatalf("invoice should remain unpaid, paid_cents=%d", paid)
	}
}
