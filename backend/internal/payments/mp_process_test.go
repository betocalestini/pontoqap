package payments_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/store-platform/store/internal/billing"
	"github.com/store-platform/store/internal/payments"
	"github.com/store-platform/store/internal/payments/mercadopago"
	"github.com/store-platform/store/internal/platform/config"
	"github.com/store-platform/store/tests/testdb"
)

type stubOrderFetcher struct {
	order mercadopago.OrderDetail
}

func (s *stubOrderFetcher) FetchOrder(_ context.Context, orderID string) (mercadopago.OrderDetail, error) {
	o := s.order
	o.ID = orderID
	return o, nil
}

func TestMercadoPagoOrderJobPendingDoesNotSettle(t *testing.T) {
	testdb.MigrateUp(t)
	pool := testdb.Pool(t)
	ctx := context.Background()
	if err := testdb.Reset(ctx, pool); err != nil {
		t.Fatal(err)
	}

	mgr, _ := testdb.SeedManager(ctx, pool, testdb.UniqueEmail(t, "mgr-mp"))
	cust, _ := testdb.SeedCustomer(ctx, pool, testdb.UniqueEmail(t, "c-mp"), "Cliente")
	_ = testdb.ApproveCustomer(ctx, pool, cust.ID, mgr.UserID, 50_000)

	billSvc := billing.NewService(pool, nil, "", nil)
	tx, _ := pool.Begin(ctx)
	at := time.Date(2026, 4, 10, 10, 0, 0, 0, time.UTC)
	periodID, _ := billSvc.EnsureOpenPeriodTx(ctx, tx, cust.ID, at)
	orderID := uuid.New()
	_, _ = tx.Exec(ctx, `
		INSERT INTO orders (id, order_number, customer_id, status, subtotal_cents, total_cents, idempotency_key, confirmed_at)
		VALUES ($1, $2, $3, 'confirmed', 5000, 5000, $4, NOW())
	`, orderID, "T-"+orderID.String()[:8], cust.ID, orderID.String())
	_ = billSvc.AddOrderEntryTx(ctx, tx, cust.ID, orderID, 5000, at)
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

	extRef := mercadopago.ExpectedExternalReferenceForInstallment(instID.String())
	fetcher := &stubOrderFetcher{order: mercadopago.OrderDetail{
		ExternalReference: extRef,
		TotalAmountCents:  inv.TotalCents,
		Payments: []mercadopago.OrderPaymentDetail{{
			Status: "action_required", StatusDetail: "waiting_transfer", AmountCents: inv.TotalCents, PaymentMethod: "pix",
		}},
	}}
	paySvc := payments.NewService(pool, payments.NewSandboxGateway("x"), billSvc, config.PaymentsConfig{Provider: "sandbox"}, nil, &payments.ServiceDeps{OrderFetcher: fetcher})

	chargeID := uuid.New()
	mpOrderID := "ORD-PENDING-1"
	_, err = pool.Exec(ctx, `
		INSERT INTO payment_charges (id, invoice_id, installment_id, provider, external_id, status, amount_cents, expires_at)
		VALUES ($1,$2,$3,'mercadopago',$4,'pending',$5,NOW() + INTERVAL '1 day')
	`, chargeID, inv.ID, instID, mpOrderID, inv.TotalCents)
	if err != nil {
		t.Fatal(err)
	}
	peID := uuid.New()
	_, err = pool.Exec(ctx, `
		INSERT INTO payment_events (id, provider, external_event_id, event_type, payload_hash, processed)
		VALUES ($1,'mercadopago','evt-1','order.updated','hash',false)
	`, peID)
	if err != nil {
		t.Fatal(err)
	}

	if err := paySvc.ProcessMercadoPagoOrderJob(ctx, peID, mpOrderID); err == nil {
		t.Fatal("expected pending order to return retryable error")
	}
	var paid int64
	_ = pool.QueryRow(ctx, `SELECT paid_cents FROM invoices WHERE id = $1`, inv.ID).Scan(&paid)
	if paid != 0 {
		t.Fatalf("expected no settlement, paid_cents=%d", paid)
	}
}

func TestMercadoPagoOrderJobSettlesWhenAccredited(t *testing.T) {
	testdb.MigrateUp(t)
	pool := testdb.Pool(t)
	ctx := context.Background()
	if err := testdb.Reset(ctx, pool); err != nil {
		t.Fatal(err)
	}

	mgr, _ := testdb.SeedManager(ctx, pool, testdb.UniqueEmail(t, "mgr-mp2"))
	cust, _ := testdb.SeedCustomer(ctx, pool, testdb.UniqueEmail(t, "c-mp2"), "Cliente")
	_ = testdb.ApproveCustomer(ctx, pool, cust.ID, mgr.UserID, 50_000)

	billSvc := billing.NewService(pool, nil, "", nil)
	tx, _ := pool.Begin(ctx)
	at := time.Date(2026, 5, 10, 10, 0, 0, 0, time.UTC)
	periodID, _ := billSvc.EnsureOpenPeriodTx(ctx, tx, cust.ID, at)
	orderID := uuid.New()
	_, _ = tx.Exec(ctx, `
		INSERT INTO orders (id, order_number, customer_id, status, subtotal_cents, total_cents, idempotency_key, confirmed_at)
		VALUES ($1, $2, $3, 'confirmed', 5000, 5000, $4, NOW())
	`, orderID, "T-"+orderID.String()[:8], cust.ID, orderID.String())
	_ = billSvc.AddOrderEntryTx(ctx, tx, cust.ID, orderID, 5000, at)
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

	extRef := mercadopago.ExpectedExternalReferenceForInstallment(instID.String())
	fetcher := &stubOrderFetcher{order: mercadopago.OrderDetail{
		ExternalReference: extRef,
		TotalAmountCents:  inv.TotalCents,
		Payments: []mercadopago.OrderPaymentDetail{{
			ID: "PAY-OK", Status: "processed", StatusDetail: "accredited", AmountCents: inv.TotalCents, PaymentMethod: "bank_transfer",
		}},
	}}
	paySvc := payments.NewService(pool, payments.NewSandboxGateway("x"), billSvc, config.PaymentsConfig{Provider: "sandbox"}, nil, &payments.ServiceDeps{OrderFetcher: fetcher})

	mpOrderID := "ORD-PAID-1"
	chargeID := uuid.New()
	_, err = pool.Exec(ctx, `
		INSERT INTO payment_charges (id, invoice_id, installment_id, provider, external_id, status, amount_cents, expires_at)
		VALUES ($1,$2,$3,'mercadopago',$4,'pending',$5,NOW() + INTERVAL '1 day')
	`, chargeID, inv.ID, instID, mpOrderID, inv.TotalCents)
	if err != nil {
		t.Fatal(err)
	}
	peID := uuid.New()
	_, err = pool.Exec(ctx, `
		INSERT INTO payment_events (id, provider, external_event_id, event_type, payload_hash, processed)
		VALUES ($1,'mercadopago','evt-2','order.processed','hash',false)
	`, peID)
	if err != nil {
		t.Fatal(err)
	}

	if err := paySvc.ProcessMercadoPagoOrderJob(ctx, peID, mpOrderID); err != nil {
		t.Fatal(err)
	}
	var invStatus string
	_ = pool.QueryRow(ctx, `SELECT status FROM invoices WHERE id = $1`, inv.ID).Scan(&invStatus)
	if invStatus != "paid" {
		t.Fatalf("invoice status: %s", invStatus)
	}
}
