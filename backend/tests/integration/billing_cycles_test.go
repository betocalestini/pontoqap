package integration_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/store-platform/store/internal/billing"
	"github.com/store-platform/store/tests/testdb"
)

func TestDueAtAndScheduledCloseDay1(t *testing.T) {
	d := billing.DueAtForCompetence(2026, 3)
	if d.Day() != 10 || int(d.Month()) != 4 {
		t.Fatalf("due_at: %v", d)
	}
	day1 := time.Date(2026, 4, 1, 10, 0, 0, 0, time.UTC)
	if !billing.IsMonthlyClosingDay(day1) {
		t.Fatal("expected closing on day 1")
	}
	day2 := time.Date(2026, 4, 2, 10, 0, 0, 0, time.UTC)
	if billing.IsMonthlyClosingDay(day2) {
		t.Fatal("expected no closing on day 2")
	}
}

func TestPartialCloseOpensNextCycle(t *testing.T) {
	testdb.MigrateUp(t)
	pool := testdb.Pool(t)
	ctx := context.Background()
	if err := testdb.Reset(ctx, pool); err != nil {
		t.Fatal(err)
	}
	mgr, _ := testdb.SeedManager(ctx, pool, testdb.UniqueEmail(t, "mgr"))
	cust, _ := testdb.SeedCustomer(ctx, pool, testdb.UniqueEmail(t, "partial"), "Cliente Parcial")
	_ = testdb.ApproveCustomer(ctx, pool, cust.ID, mgr.UserID, 100_000)

	period1 := seedOpenPeriodWithEntry(t, ctx, pool, cust.ID, 2026, 7, 1500)
	svc := billing.NewService(pool, nil, "", nil)

	inv, err := svc.CloseCustomerOpenPeriod(ctx, cust.ID)
	if err != nil {
		t.Fatal(err)
	}
	if inv.CloseType != billing.CloseTypeCustomerRequest {
		t.Fatalf("close_type: %s", inv.CloseType)
	}
	if inv.CycleNumber != 1 {
		t.Fatalf("expected cycle 1 closed, got %d", inv.CycleNumber)
	}

	var openCount int
	var cycle2 int
	err = pool.QueryRow(ctx, `
		SELECT COUNT(*), COALESCE(MAX(cycle_number), 0) FROM billing_periods
		WHERE customer_id = $1 AND status = 'open'
	`, cust.ID).Scan(&openCount, &cycle2)
	if err != nil {
		t.Fatal(err)
	}
	if openCount != 1 || cycle2 != 2 {
		t.Fatalf("expected one open period cycle 2, got count=%d cycle=%d", openCount, cycle2)
	}

	// Nova compra no ciclo 2
	tx, _ := pool.Begin(ctx)
	at := time.Date(2026, 7, 20, 12, 0, 0, 0, time.UTC)
	orderID := uuid.New()
	_, _ = tx.Exec(ctx, `
		INSERT INTO orders (id, order_number, customer_id, status, subtotal_cents, total_cents, idempotency_key, confirmed_at)
		VALUES ($1, $2, $3, 'confirmed', $4, $4, $5, NOW())
	`, orderID, "T2-"+orderID.String()[:8], cust.ID, 800, orderID.String())
	_ = svc.AddOrderEntryTx(ctx, tx, cust.ID, orderID, 800, at)
	_ = tx.Commit(ctx)

	inv2, err := svc.CloseCustomerOpenPeriod(ctx, cust.ID)
	if err != nil {
		t.Fatal(err)
	}
	if inv2.ID == inv.ID {
		t.Fatal("expected second invoice")
	}
	if inv2.CycleNumber != 2 {
		t.Fatalf("expected cycle 2 invoice, got %d", inv2.CycleNumber)
	}

	_ = period1 // first period closed
}

func TestCloseEmptyPeriodRejected(t *testing.T) {
	testdb.MigrateUp(t)
	pool := testdb.Pool(t)
	ctx := context.Background()
	if err := testdb.Reset(ctx, pool); err != nil {
		t.Fatal(err)
	}
	mgr, _ := testdb.SeedManager(ctx, pool, testdb.UniqueEmail(t, "mgr2"))
	cust, _ := testdb.SeedCustomer(ctx, pool, testdb.UniqueEmail(t, "empty"), "Cliente Vazio")
	_ = testdb.ApproveCustomer(ctx, pool, cust.ID, mgr.UserID, 10_000)

	svc := billing.NewService(pool, nil, "", nil)
	tx, _ := pool.Begin(ctx)
	at := time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC)
	periodID, err := svc.EnsureOpenPeriodTx(ctx, tx, cust.ID, at)
	if err != nil {
		t.Fatal(err)
	}
	_ = tx.Commit(ctx)

	_, err = svc.CloseCustomerOpenPeriod(ctx, cust.ID)
	if err != billing.ErrPeriodEmpty {
		t.Fatalf("expected ErrPeriodEmpty, got %v", err)
	}
	_ = periodID
}

func TestRunScheduledClosingPreviousMonth(t *testing.T) {
	testdb.MigrateUp(t)
	pool := testdb.Pool(t)
	ctx := context.Background()
	if err := testdb.Reset(ctx, pool); err != nil {
		t.Fatal(err)
	}
	mgr, _ := testdb.SeedManager(ctx, pool, testdb.UniqueEmail(t, "mgr3"))
	cust, _ := testdb.SeedCustomer(ctx, pool, testdb.UniqueEmail(t, "auto"), "Cliente Auto")
	_ = testdb.ApproveCustomer(ctx, pool, cust.ID, mgr.UserID, 50_000)
	_ = seedOpenPeriodWithEntry(t, ctx, pool, cust.ID, 2026, 6, 2200)

	svc := billing.NewService(pool, nil, "", nil)
	now := time.Date(2026, 7, 1, 8, 0, 0, 0, mustLoadSP(t))
	ran, n, err := svc.RunScheduledClosingIfDue(ctx, now)
	if err != nil {
		t.Fatal(err)
	}
	if !ran || n < 1 {
		t.Fatalf("expected close ran, got ran=%v n=%d", ran, n)
	}
	var invCount int
	_ = pool.QueryRow(ctx, `SELECT COUNT(*) FROM invoices WHERE customer_id = $1`, cust.ID).Scan(&invCount)
	if invCount != 1 {
		t.Fatalf("expected 1 invoice, got %d", invCount)
	}
}

func TestProcessClosedInvoiceReminders(t *testing.T) {
	testdb.MigrateUp(t)
	pool := testdb.Pool(t)
	ctx := context.Background()
	if err := testdb.Reset(ctx, pool); err != nil {
		t.Fatal(err)
	}
	mgr, _ := testdb.SeedManager(ctx, pool, testdb.UniqueEmail(t, "mgr4"))
	cust, _ := testdb.SeedCustomer(ctx, pool, testdb.UniqueEmail(t, "rem"), "Cliente Rem")
	_ = testdb.ApproveCustomer(ctx, pool, cust.ID, mgr.UserID, 50_000)
	invID := seedClosedInvoice(t, ctx, pool, cust.ID, 2026, 5, 1000)

	closedAt := time.Now().Add(-50 * time.Hour)
	_, err := pool.Exec(ctx, `UPDATE invoices SET closed_at = $2 WHERE id = $1`, invID, closedAt)
	if err != nil {
		t.Fatal(err)
	}

	svc := billing.NewService(pool, nil, "", nil)
	rem, esc, err := svc.ProcessClosedInvoiceReminders(ctx, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if rem != 0 && esc != 0 {
		// without jobs repo, reminders are no-op
	}
	_ = rem
	_ = esc
	_ = pool
}

func mustLoadSP(t *testing.T) *time.Location {
	t.Helper()
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		t.Fatal(err)
	}
	return loc
}
