package billing

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/store-platform/store/internal/jobs"
)

type Service struct {
	pool          *pgxpool.Pool
	jobs          *jobs.Repository
	storeWebURL   string
}

func NewService(pool *pgxpool.Pool, jobRepo *jobs.Repository, storeWebURL string) *Service {
	return &Service{pool: pool, jobs: jobRepo, storeWebURL: storeWebURL}
}

func (s *Service) EnsureOpenPeriodTx(ctx context.Context, tx pgx.Tx, customerID uuid.UUID, at time.Time) (uuid.UUID, error) {
	var periodID uuid.UUID
	err := tx.QueryRow(ctx, `
		SELECT id FROM billing_periods
		WHERE customer_id = $1 AND status = 'open'
		FOR UPDATE
	`, customerID).Scan(&periodID)
	if err == nil {
		return periodID, nil
	}
	if err != pgx.ErrNoRows {
		return uuid.Nil, err
	}
	at = at.In(saoPaulo)
	year, month := at.Year(), int(at.Month())
	var maxCycle int
	err = tx.QueryRow(ctx, `
		SELECT COALESCE(MAX(cycle_number), 0) FROM billing_periods
		WHERE customer_id = $1 AND reference_year = $2 AND reference_month = $3
	`, customerID, year, month).Scan(&maxCycle)
	if err != nil {
		return uuid.Nil, err
	}
	periodID = uuid.New()
	_, err = tx.Exec(ctx, `
		INSERT INTO billing_periods (id, customer_id, reference_year, reference_month, cycle_number, status, opened_at)
		VALUES ($1, $2, $3, $4, $5, 'open', $6)
	`, periodID, customerID, year, month, maxCycle+1, at)
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

// AddOrderCancellationTx estorna o valor do pedido no período em aberto original.
func (s *Service) AddOrderCancellationTx(ctx context.Context, tx pgx.Tx, orderID uuid.UUID, amount int64, at time.Time) error {
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
		return err
	}
	if periodStatus != "open" {
		return fmt.Errorf("período de faturamento já fechado")
	}
	desc := fmt.Sprintf("Cancelamento pedido %s", orderID.String())
	_, err = tx.Exec(ctx, `
		INSERT INTO billing_entries (billing_period_id, entry_type, order_id, description, amount_cents, occurred_at)
		VALUES ($1, 'order_cancellation', $2, $3, $4, $5)
	`, periodID, orderID, desc, -amount, at)
	return err
}
