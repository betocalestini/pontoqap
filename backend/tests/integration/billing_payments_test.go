package integration_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/store-platform/store/internal/billing"
	"github.com/store-platform/store/internal/payments"
	"github.com/store-platform/store/internal/platform/config"
	"github.com/store-platform/store/tests/testdb"
)

func TestBillingCloseIsIdempotent(t *testing.T) {
	testdb.MigrateUp(t)
	pool := testdb.Pool(t)
	ctx := context.Background()
	if err := testdb.Reset(ctx, pool); err != nil {
		t.Fatal(err)
	}

	mgr, _ := testdb.SeedManager(ctx, pool, testdb.UniqueEmail(t, "mgr"))
	cust, _ := testdb.SeedCustomer(ctx, pool, testdb.UniqueEmail(t, "c"), "Cliente")
	_ = testdb.ApproveCustomer(ctx, pool, cust.ID, mgr.UserID, 50_000)

	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	svc := billing.NewService(pool, nil, "", nil)
	at := time.Date(2026, 2, 10, 10, 0, 0, 0, time.UTC)
	periodID, err := svc.EnsureOpenPeriodTx(ctx, tx, cust.ID, at)
	if err != nil {
		t.Fatal(err)
	}
	orderID := uuid.New()
	ordNum := "T-" + orderID.String()[:8]
	if _, err := tx.Exec(ctx, `
		INSERT INTO orders (id, order_number, customer_id, status, subtotal_cents, total_cents, idempotency_key, confirmed_at)
		VALUES ($1, $2, $3, 'confirmed', 2500, 2500, $4, NOW())
	`, orderID, ordNum, cust.ID, orderID.String()); err != nil {
		t.Fatal(err)
	}
	if err := svc.AddOrderEntryTx(ctx, tx, cust.ID, orderID, 2500, at); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}

	inv1, err := svc.ClosePeriod(ctx, periodID)
	if err != nil {
		t.Fatal(err)
	}
	inv2, err := svc.ClosePeriod(ctx, periodID)
	if err != nil {
		t.Fatal(err)
	}
	if inv1.ID != inv2.ID || inv1.TotalCents != 2500 {
		t.Fatalf("expected same invoice with total 2500, got %+v %+v", inv1, inv2)
	}
}

func TestPixWebhookDuplicateIsIgnored(t *testing.T) {
	testdb.MigrateUp(t)
	pool := testdb.Pool(t)
	ctx := context.Background()
	if err := testdb.Reset(ctx, pool); err != nil {
		t.Fatal(err)
	}

	mgr, _ := testdb.SeedManager(ctx, pool, testdb.UniqueEmail(t, "mgr"))
	cust, _ := testdb.SeedCustomer(ctx, pool, testdb.UniqueEmail(t, "c"), "Cliente")
	_ = testdb.ApproveCustomer(ctx, pool, cust.ID, mgr.UserID, 50_000)

	billSvc := billing.NewService(pool, nil, "", nil)
	tx, _ := pool.Begin(ctx)
	at := time.Date(2026, 2, 10, 10, 0, 0, 0, time.UTC)
	periodID, _ := billSvc.EnsureOpenPeriodTx(ctx, tx, cust.ID, at)
	orderID := uuid.New()
	if _, err := tx.Exec(ctx, `
		INSERT INTO orders (id, order_number, customer_id, status, subtotal_cents, total_cents, idempotency_key, confirmed_at)
		VALUES ($1, $2, $3, 'confirmed', 1000, 1000, $4, NOW())
	`, orderID, "T-"+orderID.String()[:8], cust.ID, orderID.String()); err != nil {
		t.Fatal(err)
	}
	if err := billSvc.AddOrderEntryTx(ctx, tx, cust.ID, orderID, 1000, at); err != nil {
		t.Fatal(err)
	}
	_ = tx.Commit(ctx)
	inv, err := billSvc.ClosePeriod(ctx, periodID)
	if err != nil {
		t.Fatal(err)
	}
	if err := testdb.ActivateSingleInstallmentPlan(ctx, billSvc, inv.ID, cust.ID, cust.UserID); err != nil {
		t.Fatal(err)
	}
	instID, err := testdb.FirstOpenInstallmentID(ctx, pool, inv.ID)
	if err != nil {
		t.Fatal(err)
	}

	secret := "test-webhook-secret"
	paySvc := payments.NewService(pool, payments.NewSandboxGateway(secret), billSvc, config.PaymentsConfig{
		Provider:      "sandbox",
		WebhookSecret: secret,
	}, nil)
	charge, err := paySvc.CreateOrReusePixChargeForInstallment(ctx, instID, cust.ID)
	if err != nil {
		t.Fatal(err)
	}

	var extID string
	if err := pool.QueryRow(ctx, `SELECT external_id FROM payment_charges WHERE id = $1`, charge.ID).Scan(&extID); err != nil {
		t.Fatal(err)
	}
	body := []byte(`{"event_id":"evt-dup-1","event_type":"payment.confirmed","payment_id":"` + extID + `","amount_cents":1000}`)
	sig := payments.SignPayloadForTest(body, secret)

	if _, err := paySvc.ProcessWebhook(ctx, body, sig); err != nil {
		t.Fatal(err)
	}
	if _, err := paySvc.ProcessWebhook(ctx, body, sig); err != nil {
		t.Fatal(err)
	}

	var paid int64
	_ = pool.QueryRow(ctx, `SELECT paid_cents FROM invoices WHERE id = $1`, inv.ID).Scan(&paid)
	if paid != 1000 {
		t.Fatalf("expected paid_cents 1000, got %d", paid)
	}
}
