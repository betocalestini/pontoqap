package billing

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// OrderCancelBillingResult indica ajuste de exposição a aplicar fora do faturamento (período aberto).
type OrderCancelBillingResult struct {
	ExposureDeltaCents int64
}

// ApplyOrderCancellationBillingTx estorna financeiramente o pedido (competência aberta ou fatura fechada).
func (s *Service) ApplyOrderCancellationBillingTx(
	ctx context.Context,
	tx pgx.Tx,
	customerID, orderID uuid.UUID,
	amount int64,
	actorID uuid.UUID,
	at time.Time,
) (OrderCancelBillingResult, error) {
	var res OrderCancelBillingResult
	var periodID uuid.UUID
	var periodStatus string
	err := tx.QueryRow(ctx, `
		SELECT be.billing_period_id, bp.status
		FROM billing_entries be
		JOIN billing_periods bp ON bp.id = be.billing_period_id
		WHERE be.order_id = $1 AND be.entry_type = 'order'
		ORDER BY be.created_at ASC
		LIMIT 1
	`, orderID).Scan(&periodID, &periodStatus)
	if err != nil {
		return res, err
	}

	reason := fmt.Sprintf("Cancelamento pedido %s", orderID.String())

	if periodStatus == "open" {
		desc := reason
		_, err = tx.Exec(ctx, `
			INSERT INTO billing_entries (billing_period_id, entry_type, order_id, description, amount_cents, occurred_at)
			VALUES ($1, 'order_cancellation', $2, $3, $4, $5)
		`, periodID, orderID, desc, -amount, at)
		if err != nil {
			return res, err
		}
		res.ExposureDeltaCents = -amount
		return res, nil
	}

	var invoiceID uuid.UUID
	var invStatus string
	err = tx.QueryRow(ctx, `
		SELECT id, status FROM invoices WHERE billing_period_id = $1
	`, periodID).Scan(&invoiceID, &invStatus)
	if err == pgx.ErrNoRows {
		return res, fmt.Errorf("fatura não encontrada para o período do pedido")
	}
	if err != nil {
		return res, err
	}

	if invStatus == "paid" {
		openPeriodID, err := s.EnsureOpenPeriodTx(ctx, tx, customerID, at)
		if err != nil {
			return res, err
		}
		_, err = tx.Exec(ctx, `
			INSERT INTO billing_entries (billing_period_id, entry_type, order_id, description, amount_cents, occurred_at)
			VALUES ($1, 'order_cancellation', $2, $3, $4, $5)
		`, openPeriodID, orderID, reason, -amount, at)
		if err != nil {
			return res, err
		}
		return res, nil
	}

	if err := s.applyInvoiceCreditAdjustmentTx(ctx, tx, invoiceID, customerID, actorID, amount, reason); err != nil {
		return res, err
	}
	return res, nil
}
