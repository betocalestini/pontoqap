package inventory

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *Service) createLotTx(ctx context.Context, tx pgx.Tx, locID, skuID uuid.UUID, qty int, unitCostCents int64, movementID *uuid.UUID) error {
	if qty <= 0 {
		return nil
	}
	if unitCostCents < 0 {
		return fmt.Errorf("invalid unit cost")
	}
	_, err := tx.Exec(ctx, `
		INSERT INTO inventory_lots (location_id, sku_id, quantity_remaining, unit_cost_cents, source_movement_id)
		VALUES ($1, $2, $3, $4, $5)
	`, locID, skuID, qty, unitCostCents, movementID)
	return err
}

func (s *Service) consumeLotsFIFOtx(ctx context.Context, tx pgx.Tx, locID, skuID uuid.UUID, quantity int) error {
	remaining := quantity
	rows, err := tx.Query(ctx, `
		SELECT id, quantity_remaining FROM inventory_lots
		WHERE location_id = $1 AND sku_id = $2 AND quantity_remaining > 0
		ORDER BY created_at ASC
		FOR UPDATE
	`, locID, skuID)
	if err != nil {
		return err
	}
	defer rows.Close()
	type lotRow struct {
		id  uuid.UUID
		qty int
	}
	var lots []lotRow
	for rows.Next() {
		var r lotRow
		if err := rows.Scan(&r.id, &r.qty); err != nil {
			return err
		}
		lots = append(lots, r)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	for _, lot := range lots {
		if remaining <= 0 {
			break
		}
		take := lot.qty
		if take > remaining {
			take = remaining
		}
		newQty := lot.qty - take
		_, err := tx.Exec(ctx, `
			UPDATE inventory_lots SET quantity_remaining = $2 WHERE id = $1
		`, lot.id, newQty)
		if err != nil {
			return err
		}
		remaining -= take
	}
	if remaining > 0 {
		return fmt.Errorf("insufficient lot quantity")
	}
	return nil
}

func (s *Service) weightedAverageCostTx(ctx context.Context, tx pgx.Tx, locID, skuID uuid.UUID) (int64, bool, error) {
	var num, den int64
	err := tx.QueryRow(ctx, `
		SELECT COALESCE(SUM(quantity_remaining::bigint * unit_cost_cents), 0),
		       COALESCE(SUM(quantity_remaining), 0)
		FROM inventory_lots
		WHERE location_id = $1 AND sku_id = $2 AND quantity_remaining > 0
	`, locID, skuID).Scan(&num, &den)
	if err != nil {
		return 0, false, err
	}
	if den == 0 {
		return 0, false, nil
	}
	return (num + den/2) / den, true, nil
}

// WeightedAverageCostCents returns average unit cost from remaining lots.
func (s *Service) WeightedAverageCostCents(ctx context.Context, skuID uuid.UUID) (int64, bool, error) {
	locID, _ := uuid.Parse(DefaultLocationID)
	var num, den int64
	err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(quantity_remaining::bigint * unit_cost_cents), 0),
		       COALESCE(SUM(quantity_remaining), 0)
		FROM inventory_lots
		WHERE location_id = $1 AND sku_id = $2 AND quantity_remaining > 0
	`, locID, skuID).Scan(&num, &den)
	if err != nil {
		return 0, false, err
	}
	if den == 0 {
		return 0, false, nil
	}
	return (num + den/2) / den, true, nil
}

// SetRemainingLotsUnitCost updates unit cost on all open lots for the default location (admin cost correction).
func (s *Service) SetRemainingLotsUnitCost(ctx context.Context, skuID uuid.UUID, unitCostCents int64) error {
	if unitCostCents < 0 {
		return fmt.Errorf("invalid unit cost")
	}
	locID, _ := uuid.Parse(DefaultLocationID)
	_, err := s.pool.Exec(ctx, `
		UPDATE inventory_lots
		SET unit_cost_cents = $3
		WHERE location_id = $1 AND sku_id = $2 AND quantity_remaining > 0
	`, locID, skuID, unitCostCents)
	return err
}

func (s *Service) skuFallbackCostTx(ctx context.Context, tx pgx.Tx, skuID uuid.UUID) int64 {
	var cost *int64
	_ = tx.QueryRow(ctx, `SELECT cost_price_cents FROM skus WHERE id = $1`, skuID).Scan(&cost)
	if cost != nil {
		return *cost
	}
	return 0
}
