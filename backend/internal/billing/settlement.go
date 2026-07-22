package billing

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// invoiceFullySettled reports whether the invoice has no remaining balance.
func invoiceFullySettled(totalCents, paidCents int64) bool {
	return totalCents <= paidCents
}

// invoiceStatusAfterAdjustment returns the new status and whether paid_at should be set.
func invoiceStatusAfterAdjustment(newTotal, paidCents int64, currentStatus string) (status string, markPaidAt bool) {
	if invoiceFullySettled(newTotal, paidCents) {
		return "paid", true
	}
	return currentStatus, false
}

func expirePendingPixChargesTx(ctx context.Context, tx pgx.Tx, invoiceID uuid.UUID) error {
	_, err := tx.Exec(ctx, `
		UPDATE payment_charges SET status = 'expired', updated_at = NOW()
		WHERE invoice_id = $1 AND status = 'pending'
	`, invoiceID)
	return err
}

func updateInvoiceAfterAdjustmentTx(
	ctx context.Context,
	tx pgx.Tx,
	invoiceID uuid.UUID,
	newAdjustmentTotal, newTotal int64,
	newStatus string,
	markPaidAt bool,
) error {
	if markPaidAt {
		_, err := tx.Exec(ctx, `
			UPDATE invoices SET
				adjustment_cents = $2,
				total_cents = $3,
				status = $4,
				paid_at = COALESCE(paid_at, NOW()),
				updated_at = NOW()
			WHERE id = $1
		`, invoiceID, newAdjustmentTotal, newTotal, newStatus)
		return err
	}
	_, err := tx.Exec(ctx, `
		UPDATE invoices SET
			adjustment_cents = $2,
			total_cents = $3,
			status = $4,
			updated_at = NOW()
		WHERE id = $1
	`, invoiceID, newAdjustmentTotal, newTotal, newStatus)
	return err
}
