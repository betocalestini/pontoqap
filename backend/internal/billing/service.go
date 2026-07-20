package billing

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	pool *pgxpool.Pool
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

func (s *Service) EnsureOpenPeriodTx(ctx context.Context, tx pgx.Tx, customerID uuid.UUID, at time.Time) (uuid.UUID, error) {
	year, month := at.Year(), int(at.Month())
	var periodID uuid.UUID
	err := tx.QueryRow(ctx, `
		SELECT id FROM billing_periods
		WHERE customer_id = $1 AND reference_year = $2 AND reference_month = $3
	`, customerID, year, month).Scan(&periodID)
	if err == nil {
		return periodID, nil
	}
	if err != pgx.ErrNoRows {
		return uuid.Nil, err
	}
	periodID = uuid.New()
	_, err = tx.Exec(ctx, `
		INSERT INTO billing_periods (id, customer_id, reference_year, reference_month, status, opened_at)
		VALUES ($1, $2, $3, $4, 'open', $5)
	`, periodID, customerID, year, month, at)
	return periodID, err
}

func (s *Service) AddOrderEntryTx(ctx context.Context, tx pgx.Tx, customerID, orderID uuid.UUID, amount int64, at time.Time) error {
	periodID, err := s.EnsureOpenPeriodTx(ctx, tx, customerID, at)
	if err != nil {
		return err
	}
	desc := fmt.Sprintf("Pedido %s", orderID.String())
	_, err = tx.Exec(ctx, `
		INSERT INTO billing_entries (billing_period_id, entry_type, order_id, description, amount_cents, occurred_at)
		VALUES ($1, 'order', $2, $3, $4, $5)
	`, periodID, orderID, desc, amount, at)
	return err
}
