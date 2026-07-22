package billing_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/store-platform/store/internal/billing"
	"github.com/store-platform/store/tests/testdb"
)

func TestEffectiveDueAtAfterFirstInstallmentPaid(t *testing.T) {
	testdb.MigrateUp(t)
	pool := testdb.Pool(t)
	ctx := context.Background()
	if err := testdb.Reset(ctx, pool); err != nil {
		t.Fatal(err)
	}
	if err := testdb.EnsureDefaultInstallmentPolicy(ctx, pool); err != nil {
		t.Fatal(err)
	}

	svc := billing.NewService(pool, nil, "", nil)

	customerID := uuid.New()
	userID := uuid.New()
	_, err := pool.Exec(ctx, `
		INSERT INTO users (id, name, email, password_hash, status)
		VALUES ($1, 'Test', $2, 'hash', 'active')
	`, userID, testdb.UniqueEmail(t, "due"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO customers (id, user_id, status, credit_limit_cents, current_exposure_cents)
		VALUES ($1, $2, 'approved', 500000, 0)
	`, customerID, userID)
	if err != nil {
		t.Fatal(err)
	}

	year, month := 2026, 7
	periodID := uuid.New()
	_, err = pool.Exec(ctx, `
		INSERT INTO billing_periods (id, customer_id, reference_year, reference_month, cycle_number, status, opened_at)
		VALUES ($1, $2, $3, $4, 1, 'open', NOW())
	`, periodID, customerID, year, month)
	if err != nil {
		t.Fatal(err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO billing_entries (billing_period_id, entry_type, description, amount_cents, occurred_at)
		VALUES ($1, 'order', 'Pedido teste', 30000, NOW())
	`, periodID)
	if err != nil {
		t.Fatal(err)
	}

	inv, err := svc.ClosePeriodWithType(ctx, periodID, "customer_request")
	if err != nil {
		t.Fatal(err)
	}
	if inv.TotalCents < 30000 {
		t.Fatalf("expected invoice >= 30000, got %d", inv.TotalCents)
	}

	_, err = svc.SelectPaymentPlan(ctx, inv.ID, customerID, userID, 3)
	if err != nil {
		t.Fatal(err)
	}

	installments, err := svc.ListInvoiceInstallments(ctx, inv.ID, customerID)
	if err != nil || len(installments) != 3 {
		t.Fatalf("installments: %v err %v", len(installments), err)
	}
	firstDue := installments[0].DueDate
	secondDue := installments[1].DueDate

	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.ApplyInstallmentPaymentTx(ctx, tx, installments[0].ID, installments[0].AmountCents); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}

	list, err := svc.ListInvoicesByCustomerLimit(ctx, customerID, 10)
	if err != nil {
		t.Fatal(err)
	}
	var found *billing.Invoice
	for i := range list {
		if list[i].ID == inv.ID {
			found = &list[i]
			break
		}
	}
	if found == nil {
		t.Fatal("invoice not in list")
	}
	if !sameCalendarDay(found.DueAt, secondDue) {
		t.Fatalf("list due_at: got %v want second installment %v (first was %v)", found.DueAt, secondDue, firstDue)
	}

	detail, err := svc.GetInvoiceDetail(ctx, inv.ID)
	if err != nil || detail == nil {
		t.Fatal(err)
	}
	if detail.DueAt == nil || !sameCalendarDay(*detail.DueAt, secondDue) {
		t.Fatalf("detail due_at: %v want %v", detail.DueAt, secondDue)
	}
}

func sameCalendarDay(a time.Time, b time.Time) bool {
	ay, am, ad := a.In(time.UTC).Date()
	by, bm, bd := b.In(time.UTC).Date()
	return ay == by && am == bm && ad == bd
}
