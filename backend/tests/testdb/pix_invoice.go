package testdb

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/store-platform/store/internal/billing"
)

// ActivateSingleInstallmentPlan confirma plano 1× (fluxo atual de Pix na fatura fechada).
func ActivateSingleInstallmentPlan(ctx context.Context, bill *billing.Service, invoiceID, customerID, userID uuid.UUID) error {
	_, err := bill.SelectPaymentPlan(ctx, invoiceID, customerID, userID, 1)
	return err
}

// FirstOpenInstallmentID retorna a parcela em aberto (número 1 após plano 1×).
func FirstOpenInstallmentID(ctx context.Context, pool *pgxpool.Pool, invoiceID uuid.UUID) (uuid.UUID, error) {
	var id uuid.UUID
	err := pool.QueryRow(ctx, `
		SELECT id FROM invoice_installments
		WHERE invoice_id = $1 AND installment_number = 1
	`, invoiceID).Scan(&id)
	if err == pgx.ErrNoRows {
		return uuid.Nil, err
	}
	return id, err
}
