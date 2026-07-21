package billing

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/store-platform/store/internal/notification"
)

// ProcessClosedInvoiceReminders envia lembrete D+2 e escalada D+3 após fechamento sem pagamento.
func (s *Service) ProcessClosedInvoiceReminders(ctx context.Context, now time.Time) (reminders int, escalations int, err error) {
	if s.jobs == nil {
		return 0, 0, nil
	}
	reminders, err = s.processPaymentReminders(ctx, now)
	if err != nil {
		return reminders, escalations, err
	}
	escalations, err = s.processPaymentEscalations(ctx, now)
	return reminders, escalations, err
}

func (s *Service) processPaymentReminders(ctx context.Context, now time.Time) (int, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT i.id, i.invoice_number, i.customer_id, i.total_cents, i.due_at
		FROM invoices i
		WHERE i.status IN ('open', 'overdue')
		  AND i.paid_cents < i.total_cents
		  AND i.closed_at IS NOT NULL
		  AND i.reminder_sent_at IS NULL
		  AND i.closed_at + interval '48 hours' <= $1
	`, now)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var n int
	for rows.Next() {
		var invID, customerID uuid.UUID
		var invNumber string
		var total int64
		var dueAt time.Time
		if err := rows.Scan(&invID, &invNumber, &customerID, &total, &dueAt); err != nil {
			return n, err
		}
		tx, err := s.pool.Begin(ctx)
		if err != nil {
			return n, err
		}
		if err := s.publishInvoiceNotifyTx(ctx, tx, invID, customerID, invNumber, total, dueAt, notification.EventInvoicePaymentReminder); err != nil {
			tx.Rollback(ctx)
			return n, err
		}
		_, err = tx.Exec(ctx, `UPDATE invoices SET reminder_sent_at = $2, updated_at = NOW() WHERE id = $1`, invID, now)
		if err != nil {
			tx.Rollback(ctx)
			return n, err
		}
		if err := tx.Commit(ctx); err != nil {
			return n, err
		}
		n++
	}
	return n, rows.Err()
}

func (s *Service) processPaymentEscalations(ctx context.Context, now time.Time) (int, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT i.id, i.invoice_number, i.customer_id, i.total_cents, i.due_at
		FROM invoices i
		WHERE i.status IN ('open', 'overdue')
		  AND i.paid_cents < i.total_cents
		  AND i.closed_at IS NOT NULL
		  AND i.escalation_sent_at IS NULL
		  AND i.closed_at + interval '72 hours' <= $1
	`, now)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var n int
	for rows.Next() {
		var invID, customerID uuid.UUID
		var invNumber string
		var total int64
		var dueAt time.Time
		if err := rows.Scan(&invID, &invNumber, &customerID, &total, &dueAt); err != nil {
			return n, err
		}
		tx, err := s.pool.Begin(ctx)
		if err != nil {
			return n, err
		}
		_, err = tx.Exec(ctx, `
			UPDATE invoices SET status = 'overdue', updated_at = NOW()
			WHERE id = $1 AND paid_cents < total_cents
		`, invID)
		if err != nil {
			tx.Rollback(ctx)
			return n, err
		}
		if err := s.publishInvoiceNotifyTx(ctx, tx, invID, customerID, invNumber, total, dueAt, notification.EventInvoicePaymentEscalation); err != nil {
			tx.Rollback(ctx)
			return n, err
		}
		_, err = tx.Exec(ctx, `UPDATE invoices SET escalation_sent_at = $2, updated_at = NOW() WHERE id = $1`, invID, now)
		if err != nil {
			tx.Rollback(ctx)
			return n, err
		}
		if err := tx.Commit(ctx); err != nil {
			return n, err
		}
		n++
	}
	return n, rows.Err()
}

func (s *Service) publishInvoiceNotifyTx(ctx context.Context, tx pgx.Tx, invID, customerID uuid.UUID, invNumber string, total int64, dueAt time.Time, eventType string) error {
	var email, name string
	err := tx.QueryRow(ctx, `
		SELECT u.email, u.name FROM users u
		JOIN customers c ON c.user_id = u.id
		WHERE c.id = $1
	`, customerID).Scan(&email, &name)
	if err != nil {
		return err
	}
	invoiceURL := notification.BuildInvoiceURL(s.storeWebURL, invID.String())
	payload := map[string]any{
		"to":             email,
		"name":           name,
		"invoice_number": invNumber,
		"total_cents":    total,
		"due_at":         dueAt.Format(time.RFC3339),
		"invoice_url":    invoiceURL,
	}
	return s.jobs.PublishOutbox(ctx, tx, eventType, "invoice", invID, payload)
}
