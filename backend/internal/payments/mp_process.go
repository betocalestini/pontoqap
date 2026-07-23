package payments

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/store-platform/store/internal/payments/mercadopago"
)

// ProcessMercadoPagoOrderJob syncs an Order from Mercado Pago and settles when accredited.
func (s *Service) ProcessMercadoPagoOrderJob(ctx context.Context, paymentEventID uuid.UUID, orderID string) error {
	if s.orderFetcher == nil {
		return errors.New("mercado pago order fetcher não configurado")
	}
	order, err := s.orderFetcher.FetchOrder(ctx, orderID)
	if err != nil {
		return err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if paymentEventID != uuid.Nil {
		var peProcessed bool
		err = tx.QueryRow(ctx, `
			SELECT processed FROM payment_events WHERE id = $1 FOR UPDATE
		`, paymentEventID).Scan(&peProcessed)
		if err != nil {
			return err
		}
		if peProcessed {
			return tx.Commit(ctx)
		}
	}

	var chargeID uuid.UUID
	var invoiceID uuid.UUID
	var installmentID *uuid.UUID
	var chargeAmount int64
	var chargeStatus string
	var invoiceNumber string
	err = tx.QueryRow(ctx, `
		SELECT pc.id, pc.invoice_id, pc.installment_id, pc.amount_cents, pc.status, i.invoice_number
		FROM payment_charges pc
		JOIN invoices i ON i.id = pc.invoice_id
		WHERE pc.provider = $1 AND pc.external_id = $2
		FOR UPDATE
	`, ProviderMercadoPago, orderID).Scan(&chargeID, &invoiceID, &installmentID, &chargeAmount, &chargeStatus, &invoiceNumber)
	if err == pgx.ErrNoRows {
		if paymentEventID != uuid.Nil {
			_, _ = tx.Exec(ctx, `
				UPDATE payment_events SET processed = TRUE, processed_at = NOW(), error_message = 'charge not found for order'
				WHERE id = $1
			`, paymentEventID)
		}
		return tx.Commit(ctx)
	}
	if err != nil {
		return err
	}

	_, _ = tx.Exec(ctx, `UPDATE payment_charges SET last_synced_at = NOW(), updated_at = NOW() WHERE id = $1`, chargeID)

	expectedRef := invoiceNumber
	if installmentID != nil {
		expectedRef = mercadopago.ExpectedExternalReferenceForInstallment(installmentID.String())
	}
	eval := mercadopago.EvaluateSettlement(order, expectedRef, chargeAmount)

	switch eval.Outcome {
	case mercadopago.SettlementPending:
		if chargeStatus == "pending" {
			return fmt.Errorf("mercado pago order payment pending")
		}
		if paymentEventID != uuid.Nil {
			_, err = tx.Exec(ctx, `
				UPDATE payment_events SET processed = TRUE, processed_at = NOW() WHERE id = $1
			`, paymentEventID)
			if err != nil {
				return err
			}
		}
		return tx.Commit(ctx)

	case mercadopago.SettlementRequiresReview:
		_, err = tx.Exec(ctx, `
			UPDATE payment_charges SET status = 'requires_review', updated_at = NOW() WHERE id = $1
		`, chargeID)
		if err != nil {
			return err
		}
		if paymentEventID != uuid.Nil {
			_, err = tx.Exec(ctx, `
				UPDATE payment_events SET processed = TRUE, processed_at = NOW(), error_message = $2 WHERE id = $1
			`, paymentEventID, eval.Reason)
			if err != nil {
				return err
			}
		}
		s.log.Warn("mercado pago settlement requires review",
			slog.String("order_id", orderID),
			slog.String("charge_id", chargeID.String()),
			slog.String("reason", eval.Reason),
		)
		return tx.Commit(ctx)

	case mercadopago.SettlementSettle:
		if chargeStatus == "paid" {
			if paymentEventID != uuid.Nil {
				_, err = tx.Exec(ctx, `
					UPDATE payment_events SET processed = TRUE, processed_at = NOW() WHERE id = $1
				`, paymentEventID)
				if err != nil {
					return err
				}
			}
			return tx.Commit(ctx)
		}
		paymentExtID := eval.PaymentID
		if paymentExtID == "" {
			paymentExtID = orderID
		}
		settled, err := s.settleChargeTx(ctx, tx, chargeID, ProviderMercadoPago, paymentExtID, eval.AmountCents)
		if err != nil {
			if errors.Is(err, ErrChargeAlreadyPaid) {
				if paymentEventID != uuid.Nil {
					_, _ = tx.Exec(ctx, `UPDATE payment_events SET processed = TRUE, processed_at = NOW() WHERE id = $1`, paymentEventID)
				}
				return tx.Commit(ctx)
			}
			return err
		}
		if paymentEventID != uuid.Nil {
			_, err = tx.Exec(ctx, `
				UPDATE payment_events SET processed = TRUE, processed_at = NOW() WHERE id = $1
			`, paymentEventID)
			if err != nil {
				return err
			}
		}
		if err := tx.Commit(ctx); err != nil {
			return err
		}
		s.log.Info("mercado pago payment settled",
			slog.String("order_id", orderID),
			slog.String("charge_id", chargeID.String()),
			slog.Int64("amount_cents", eval.AmountCents),
			slog.Bool("invoice_paid", settled.invoicePaid),
		)
		s.recordSettlementAudit(ctx, settled)
		return nil
	default:
		return fmt.Errorf("unknown settlement outcome")
	}
}

// SyncMercadoPagoChargeByID re-fetches the Order for a charge (admin/dev).
func (s *Service) SyncMercadoPagoChargeByID(ctx context.Context, chargeID uuid.UUID) error {
	var orderID string
	var provider string
	err := s.pool.QueryRow(ctx, `SELECT external_id, provider FROM payment_charges WHERE id = $1`, chargeID).Scan(&orderID, &provider)
	if err != nil {
		return err
	}
	if provider != ProviderMercadoPago || orderID == "" {
		return fmt.Errorf("cobrança não é Mercado Pago")
	}
	return s.ProcessMercadoPagoOrderJob(ctx, uuid.Nil, orderID)
}
