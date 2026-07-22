package billing

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// SQL expression: effective due_at from next unpaid installment, else invoices.due_at.
const sqlEffectiveInvoiceDueAt = `COALESCE(
  (
    SELECT (
      (MIN(ii.due_date)::timestamp AT TIME ZONE 'America/Sao_Paulo')
      + interval '23 hours 59 minutes 59 seconds'
    ) AT TIME ZONE 'America/Sao_Paulo'
    FROM invoice_installments ii
    INNER JOIN invoice_payment_plans pp ON pp.id = ii.payment_plan_id
    WHERE ii.invoice_id = i.id
      AND pp.status = 'active'
      AND ii.remaining_cents > 0
      AND ii.status NOT IN ('paid', 'canceled')
  ),
  i.due_at
)`

func dueAtFromInstallmentDate(d time.Time) time.Time {
	loc := saoPaulo
	y, m, day := d.In(loc).Date()
	return time.Date(y, m, day, 23, 59, 59, 0, loc)
}

func (s *Service) nextUnpaidInstallmentDueDateTx(ctx context.Context, tx pgx.Tx, invoiceID uuid.UUID) (*time.Time, error) {
	var dueDate *time.Time
	err := tx.QueryRow(ctx, `
		SELECT MIN(ii.due_date)
		FROM invoice_installments ii
		INNER JOIN invoice_payment_plans pp ON pp.id = ii.payment_plan_id
		WHERE ii.invoice_id = $1
		  AND pp.status = 'active'
		  AND ii.remaining_cents > 0
		  AND ii.status NOT IN ('paid', 'canceled')
	`, invoiceID).Scan(&dueDate)
	if err != nil {
		return nil, err
	}
	if dueDate == nil || dueDate.IsZero() {
		return nil, nil
	}
	t := dueAtFromInstallmentDate(*dueDate)
	return &t, nil
}

func (s *Service) NextUnpaidInstallmentDueDate(ctx context.Context, invoiceID uuid.UUID) (*time.Time, error) {
	var dueDate *time.Time
	err := s.pool.QueryRow(ctx, `
		SELECT MIN(ii.due_date)
		FROM invoice_installments ii
		INNER JOIN invoice_payment_plans pp ON pp.id = ii.payment_plan_id
		WHERE ii.invoice_id = $1
		  AND pp.status = 'active'
		  AND ii.remaining_cents > 0
		  AND ii.status NOT IN ('paid', 'canceled')
	`, invoiceID).Scan(&dueDate)
	if err != nil {
		return nil, err
	}
	if dueDate == nil || dueDate.IsZero() {
		return nil, nil
	}
	t := dueAtFromInstallmentDate(*dueDate)
	return &t, nil
}

func (s *Service) EffectiveDueAt(ctx context.Context, invoiceID uuid.UUID, stored time.Time) (time.Time, error) {
	next, err := s.NextUnpaidInstallmentDueDate(ctx, invoiceID)
	if err != nil {
		return stored, err
	}
	if next != nil {
		return *next, nil
	}
	return stored, nil
}
