package payments

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var ErrChargeAlreadyPaid = errors.New("payment charge already paid")

// settleResult describes invoice state after settlement.
type settleResult struct {
	invoiceID     uuid.UUID
	installmentID *uuid.UUID
	invoicePaid   bool
	amountCents   int64
}

// settleChargeTx records payment and updates invoice/installment. Charge row must be locked by caller.
func (s *Service) settleChargeTx(ctx context.Context, tx pgx.Tx, chargeID uuid.UUID, provider, externalPaymentID string, amountCents int64) (settleResult, error) {
	var out settleResult
	var invoiceID uuid.UUID
	var installmentID *uuid.UUID
	var chargeStatus string
	var chargeAmount int64
	err := tx.QueryRow(ctx, `
		SELECT invoice_id, installment_id, status, amount_cents
		FROM payment_charges WHERE id = $1 FOR UPDATE
	`, chargeID).Scan(&invoiceID, &installmentID, &chargeStatus, &chargeAmount)
	if err != nil {
		return out, err
	}
	if chargeStatus == "paid" {
		return out, ErrChargeAlreadyPaid
	}
	if amountCents != chargeAmount {
		return out, errors.New("payment amount mismatch")
	}

	out.invoiceID = invoiceID
	out.installmentID = installmentID
	out.amountCents = amountCents

	_, err = tx.Exec(ctx, `
		INSERT INTO payments (id, invoice_id, installment_id, payment_charge_id, provider, external_payment_id, amount_cents, status, settled_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,'settled',NOW())
		ON CONFLICT (provider, external_payment_id) WHERE external_payment_id IS NOT NULL DO NOTHING
	`, uuid.New(), invoiceID, installmentID, chargeID, provider, externalPaymentID, amountCents)
	if err != nil {
		return out, err
	}
	_, err = tx.Exec(ctx, `UPDATE payment_charges SET status = 'paid', paid_at = NOW(), updated_at = NOW() WHERE id = $1`, chargeID)
	if err != nil {
		return out, err
	}

	if installmentID != nil {
		if err := s.billing.ApplyInstallmentPaymentTx(ctx, tx, *installmentID, amountCents); err != nil {
			return out, err
		}
	} else {
		_, err = tx.Exec(ctx, `
			UPDATE invoices SET paid_cents = paid_cents + $2,
				status = CASE WHEN paid_cents + $2 >= total_cents THEN 'paid' ELSE status END,
				paid_at = CASE WHEN paid_cents + $2 >= total_cents THEN NOW() ELSE paid_at END,
				updated_at = NOW()
			WHERE id = $1
		`, invoiceID, amountCents)
		if err != nil {
			return out, err
		}
		_, err = tx.Exec(ctx, `
			UPDATE customers SET current_exposure_cents = GREATEST(0, current_exposure_cents - $2), updated_at = NOW()
			WHERE id = (SELECT customer_id FROM invoices WHERE id = $1)
		`, invoiceID, amountCents)
		if err != nil {
			return out, err
		}
	}

	var invStatus string
	if err := tx.QueryRow(ctx, `SELECT status FROM invoices WHERE id = $1`, invoiceID).Scan(&invStatus); err == nil && invStatus == "paid" {
		out.invoicePaid = true
	}
	return out, nil
}
