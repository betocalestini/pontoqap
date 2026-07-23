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

func TestSandboxMultiInstallmentSettlement(t *testing.T) {
	testdb.MigrateUp(t)
	pool := testdb.Pool(t)
	ctx := context.Background()
	if err := testdb.Reset(ctx, pool); err != nil {
		t.Fatal(err)
	}

	mgr, _ := testdb.SeedManager(ctx, pool, testdb.UniqueEmail(t, "mgr-3x"))
	cust, _ := testdb.SeedCustomer(ctx, pool, testdb.UniqueEmail(t, "c-3x"), "Cliente")
	_ = testdb.ApproveCustomer(ctx, pool, cust.ID, mgr.UserID, 200_000)

	billSvc := billing.NewService(pool, nil, "", nil)
	tx, _ := pool.Begin(ctx)
	at := time.Date(2026, 6, 10, 10, 0, 0, 0, time.UTC)
	periodID, _ := billSvc.EnsureOpenPeriodTx(ctx, tx, cust.ID, at)
	orderID := uuid.New()
	_, _ = tx.Exec(ctx, `
		INSERT INTO orders (id, order_number, customer_id, status, subtotal_cents, total_cents, idempotency_key, confirmed_at)
		VALUES ($1, $2, $3, 'confirmed', 35000, 35000, $4, NOW())
	`, orderID, "T-"+orderID.String()[:8], cust.ID, orderID.String())
	_ = billSvc.AddOrderEntryTx(ctx, tx, cust.ID, orderID, 35000, at)
	_ = tx.Commit(ctx)
	inv, err := billSvc.ClosePeriod(ctx, periodID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = billSvc.SelectPaymentPlan(ctx, inv.ID, cust.ID, cust.UserID, 3)
	if err != nil {
		t.Fatal(err)
	}

	secret := "test-webhook-secret"
	paySvc := payments.NewService(pool, payments.NewSandboxGateway(secret), billSvc, config.PaymentsConfig{
		Provider: "sandbox", WebhookSecret: secret,
	}, nil, nil)

	for n := 1; n <= 3; n++ {
		var instID uuid.UUID
		err = pool.QueryRow(ctx, `
			SELECT id FROM invoice_installments WHERE invoice_id = $1 AND installment_number = $2
		`, inv.ID, n).Scan(&instID)
		if err != nil {
			t.Fatal(err)
		}
		charge, err := paySvc.CreateOrReusePixChargeForInstallment(ctx, instID, cust.ID)
		if err != nil {
			t.Fatal(err)
		}
		if err := paySvc.SimulateSandboxPayment(ctx, charge.ID); err != nil {
			t.Fatal(err)
		}
	}

	var status string
	var paid int64
	_ = pool.QueryRow(ctx, `SELECT status, paid_cents FROM invoices WHERE id = $1`, inv.ID).Scan(&status, &paid)
	if status != "paid" || paid != 35000 {
		t.Fatalf("invoice: status=%s paid=%d", status, paid)
	}
}
