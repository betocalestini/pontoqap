package billing

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const (
	AdjustmentTypeCredit = "credit"
	AdjustmentTypeDebit  = "debit"
)

// AddInvoiceAdjustment applies a signed adjustment to an open or overdue invoice.
// amountCents must be positive; adjustmentType credit reduces total, debit increases total.
func (s *Service) AddInvoiceAdjustment(ctx context.Context, invoiceID, actorID uuid.UUID, adjustmentType string, amountCents int64, reason string) (*InvoiceDetail, error) {
	adjustmentType = strings.TrimSpace(strings.ToLower(adjustmentType))
	reason = strings.TrimSpace(reason)
	if amountCents <= 0 {
		return nil, fmt.Errorf("valor do ajuste deve ser positivo")
	}
	if reason == "" {
		return nil, fmt.Errorf("justificativa obrigatória")
	}
	if adjustmentType != AdjustmentTypeCredit && adjustmentType != AdjustmentTypeDebit {
		return nil, fmt.Errorf("tipo de ajuste inválido")
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var customerID uuid.UUID
	var status string
	var totalCents, paidCents, adjustmentCents int64
	err = tx.QueryRow(ctx, `
		SELECT customer_id, status, total_cents, paid_cents, adjustment_cents
		FROM invoices WHERE id = $1 FOR UPDATE
	`, invoiceID).Scan(&customerID, &status, &totalCents, &paidCents, &adjustmentCents)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("fatura não encontrada")
	}
	if err != nil {
		return nil, err
	}
	if status == "paid" {
		return nil, fmt.Errorf("fatura já quitada")
	}

	signed := amountCents
	if adjustmentType == AdjustmentTypeCredit {
		signed = -amountCents
	}
	newAdjustmentTotal := adjustmentCents + signed
	newTotal := totalCents + signed
	if newTotal < 0 {
		return nil, fmt.Errorf("ajuste excede o total da fatura")
	}
	if newTotal < paidCents {
		return nil, fmt.Errorf("total não pode ficar abaixo do valor já pago")
	}

	adjID := uuid.New()
	_, err = tx.Exec(ctx, `
		INSERT INTO billing_adjustments (id, invoice_id, adjustment_type, amount_cents, reason, created_by)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, adjID, invoiceID, adjustmentType, amountCents, reason, actorID)
	if err != nil {
		return nil, err
	}

	newStatus, markPaidAt := invoiceStatusAfterAdjustment(newTotal, paidCents, status)

	if err := updateInvoiceAfterAdjustmentTx(ctx, tx, invoiceID, newAdjustmentTotal, newTotal, newStatus, markPaidAt); err != nil {
		return nil, err
	}
	if markPaidAt {
		if err := expirePendingPixChargesTx(ctx, tx, invoiceID); err != nil {
			return nil, err
		}
	}

	exposureDelta := newTotal - totalCents
	if exposureDelta != 0 {
		_, err = tx.Exec(ctx, `
			UPDATE customers SET
				current_exposure_cents = GREATEST(0, current_exposure_cents + $2),
				updated_at = NOW()
			WHERE id = $1
		`, customerID, exposureDelta)
		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return s.GetInvoiceDetail(ctx, invoiceID)
}
