package inventory

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	platformerrors "github.com/store-platform/store/internal/platform/errors"
)

const DefaultLocationID = "c0000000-0000-4000-8000-000000000001"

type Service struct {
	pool *pgxpool.Pool
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

type Balance struct {
	SKUID            uuid.UUID `json:"sku_id"`
	AvailableQuantity int      `json:"available_quantity"`
}

func (s *Service) RegisterEntry(ctx context.Context, skuID uuid.UUID, quantity int, createdBy uuid.UUID, reason string) error {
	if quantity <= 0 {
		return fmt.Errorf("invalid quantity")
	}
	locID, _ := uuid.Parse(DefaultLocationID)
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var balanceID uuid.UUID
	var prev, version int
	err = tx.QueryRow(ctx, `
		SELECT id, available_quantity, version FROM inventory_balances
		WHERE location_id = $1 AND sku_id = $2 FOR UPDATE
	`, locID, skuID).Scan(&balanceID, &prev, &version)
	if err == pgx.ErrNoRows {
		prev = 0
		version = 0
		balanceID = uuid.New()
		_, err = tx.Exec(ctx, `
			INSERT INTO inventory_balances (id, location_id, sku_id, available_quantity, version)
			VALUES ($1, $2, $3, 0, 0)
		`, balanceID, locID, skuID)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	newBal := prev + quantity
	res, err := tx.Exec(ctx, `
		UPDATE inventory_balances SET available_quantity = $3, version = version + 1, updated_at = NOW()
		WHERE id = $1 AND version = $2
	`, balanceID, version, newBal)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("concurrent inventory update")
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO stock_movements (location_id, sku_id, movement_type, quantity, previous_balance, new_balance, reason, created_by)
		VALUES ($1, $2, 'entry', $3, $4, $5, $6, $7)
	`, locID, skuID, quantity, prev, newBal, reason, createdBy)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *Service) ReserveAndDecrement(ctx context.Context, tx pgx.Tx, skuID uuid.UUID, quantity int, refType string, refID uuid.UUID, createdBy *uuid.UUID) error {
	locID, _ := uuid.Parse(DefaultLocationID)
	var balanceID uuid.UUID
	var prev, version int
	err := tx.QueryRow(ctx, `
		SELECT id, available_quantity, version FROM inventory_balances
		WHERE location_id = $1 AND sku_id = $2 FOR UPDATE
	`, locID, skuID).Scan(&balanceID, &prev, &version)
	if err != nil {
		return ErrInsufficientStock()
	}
	if prev < quantity {
		return ErrInsufficientStock()
	}
	newBal := prev - quantity
	res, err := tx.Exec(ctx, `
		UPDATE inventory_balances SET available_quantity = $3, version = version + 1, updated_at = NOW()
		WHERE id = $1 AND version = $2
	`, balanceID, version, newBal)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("concurrent inventory update")
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO stock_movements (location_id, sku_id, movement_type, quantity, previous_balance, new_balance, reference_type, reference_id, created_by)
		VALUES ($1, $2, 'sale', $3, $4, $5, $6, $7, $8)
	`, locID, skuID, quantity, prev, newBal, refType, refID, createdBy)
	return err
}

func ErrInsufficientStock() error {
	return &AppError{Code: platformerrors.CodeInsufficientStock, Message: "Estoque insuficiente", Status: 422}
}

type AppError struct {
	Code    string
	Message string
	Status  int
}

func (e *AppError) Error() string { return e.Message }
