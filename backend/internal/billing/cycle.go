package billing

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// OpenNextCycleTx abre um novo período após fechamento parcial (mesma competência se ainda no mês).
func (s *Service) OpenNextCycleTx(ctx context.Context, tx pgx.Tx, customerID uuid.UUID, closedRefYear, closedRefMonth int, at time.Time) (uuid.UUID, error) {
	now := at.In(saoPaulo)
	refYear, refMonth := closedRefYear, closedRefMonth
	if now.Year() != refYear || int(now.Month()) != refMonth {
		refYear, refMonth = now.Year(), int(now.Month())
	}
	var maxCycle int
	err := tx.QueryRow(ctx, `
		SELECT COALESCE(MAX(cycle_number), 0) FROM billing_periods
		WHERE customer_id = $1 AND reference_year = $2 AND reference_month = $3
	`, customerID, refYear, refMonth).Scan(&maxCycle)
	if err != nil {
		return uuid.Nil, err
	}
	periodID := uuid.New()
	_, err = tx.Exec(ctx, `
		INSERT INTO billing_periods (id, customer_id, reference_year, reference_month, cycle_number, status, opened_at)
		VALUES ($1, $2, $3, $4, $5, 'open', $6)
	`, periodID, customerID, refYear, refMonth, maxCycle+1, at)
	return periodID, err
}

// CloseCustomerOpenPeriod fecha o ciclo aberto do cliente e abre o próximo.
func (s *Service) CloseCustomerOpenPeriod(ctx context.Context, customerID uuid.UUID) (*Invoice, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var periodID uuid.UUID
	var refYear, refMonth int
	err = tx.QueryRow(ctx, `
		SELECT id, reference_year, reference_month FROM billing_periods
		WHERE customer_id = $1 AND status = 'open'
		FOR UPDATE
	`, customerID).Scan(&periodID, &refYear, &refMonth)
	if err == pgx.ErrNoRows {
		return nil, ErrNoOpenPeriod
	}
	if err != nil {
		return nil, err
	}

	inv, err := s.closePeriodTx(ctx, tx, periodID, CloseTypeCustomerRequest)
	if err != nil {
		return nil, err
	}
	if _, err := s.OpenNextCycleTx(ctx, tx, customerID, refYear, refMonth, time.Now()); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return inv, nil
}
