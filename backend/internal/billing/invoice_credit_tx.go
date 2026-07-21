package billing

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// applyInvoiceCreditAdjustmentTx aplica crédito na fatura (mesma transação).
func (s *Service) applyInvoiceCreditAdjustmentTx(
	ctx context.Context,
	tx pgx.Tx,
	invoiceID, customerID, actorID uuid.UUID,
	amountCents int64,
	reason string,
) error {
	if amountCents <= 0 {
		return fmt.Errorf("valor do ajuste deve ser positivo")
	}
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return fmt.Errorf("justificativa obrigatória")
	}

	var status string
	var totalCents, paidCents, adjustmentCents int64
	err := tx.QueryRow(ctx, `
		SELECT status, total_cents, paid_cents, adjustment_cents
		FROM invoices WHERE id = $1 FOR UPDATE
	`, invoiceID).Scan(&status, &totalCents, &paidCents, &adjustmentCents)
	if err == pgx.ErrNoRows {
		return fmt.Errorf("fatura não encontrada")
	}
	if err != nil {
		return err
	}
	if status == "paid" {
		return fmt.Errorf("fatura já quitada")
	}

	signed := -amountCents
	newAdjustmentTotal := adjustmentCents + signed
	newTotal := totalCents + signed
	if newTotal < 0 {
		return fmt.Errorf("ajuste excede o total da fatura")
	}
	if newTotal < paidCents {
		return fmt.Errorf("total não pode ficar abaixo do valor já pago")
	}

	adjID := uuid.New()
	_, err = tx.Exec(ctx, `
		INSERT INTO billing_adjustments (id, invoice_id, adjustment_type, amount_cents, reason, created_by)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, adjID, invoiceID, AdjustmentTypeCredit, amountCents, reason, actorID)
	if err != nil {
		return err
	}

	newStatus := status
	if newTotal <= paidCents && paidCents > 0 {
		newStatus = "paid"
	}

	_, err = tx.Exec(ctx, `
		UPDATE invoices SET
			adjustment_cents = $2,
			total_cents = $3,
			status = $4,
			updated_at = NOW()
		WHERE id = $1
	`, invoiceID, newAdjustmentTotal, newTotal, newStatus)
	if err != nil {
		return err
	}

	exposureDelta := newTotal - totalCents
	if exposureDelta != 0 {
		_, err = tx.Exec(ctx, `
			UPDATE customers SET
				current_exposure_cents = GREATEST(0, current_exposure_cents + $2),
				updated_at = NOW()
			WHERE id = $1
		`, customerID, exposureDelta)
	}
	return err
}
