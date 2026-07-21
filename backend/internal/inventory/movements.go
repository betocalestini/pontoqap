package inventory

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const (
	MovementEntry        = "entry"
	MovementInitialStock   = "initial_stock"
	MovementSale           = "sale"
	MovementLoss           = "loss"
	MovementDamage         = "damage"
	MovementAdjustment     = "adjustment"
	MovementReturn         = "return"
)

type Movement struct {
	ID               uuid.UUID  `json:"id"`
	SKUID            uuid.UUID  `json:"sku_id"`
	ProductName      string     `json:"product_name,omitempty"`
	SKUCode          string     `json:"sku_code,omitempty"`
	MovementType     string     `json:"movement_type"`
	Quantity         int        `json:"quantity"`
	PreviousBalance  int        `json:"previous_balance"`
	NewBalance       int        `json:"new_balance"`
	ReferenceType    *string    `json:"reference_type,omitempty"`
	ReferenceID      *uuid.UUID `json:"reference_id,omitempty"`
	Reason           *string    `json:"reason,omitempty"`
	UnitCostCents    *int64     `json:"unit_cost_cents,omitempty"`
	CreatedBy        *uuid.UUID `json:"created_by,omitempty"`
	CreatedByEmail   *string    `json:"created_by_email,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
}

type BalanceRow struct {
	SKUID             uuid.UUID `json:"sku_id"`
	SKUCode           string    `json:"sku_code"`
	ProductID         uuid.UUID `json:"product_id"`
	ProductName       string    `json:"product_name"`
	MinimumStock      int       `json:"minimum_stock"`
	AvailableQuantity int       `json:"available_quantity"`
}

type movementApply struct {
	movementType  string
	quantityDelta int
	reason        string
	refType       string
	refID         *uuid.UUID
	createdBy     uuid.UUID
	unitCostCents *int64
}

func (s *Service) RegisterEntry(ctx context.Context, skuID uuid.UUID, quantity int, createdBy uuid.UUID, reason string, totalPaidCents, otherExpensesCents int64) error {
	unitCostCents, err := EntryUnitCostCents(totalPaidCents, otherExpensesCents, quantity)
	if err != nil {
		return err
	}
	cost := unitCostCents
	return s.applyMovement(ctx, skuID, movementApply{
		movementType:  MovementEntry,
		quantityDelta: quantity,
		reason:        reason,
		createdBy:     createdBy,
		unitCostCents: &cost,
	})
}

func (s *Service) RegisterInitialStock(ctx context.Context, skuID uuid.UUID, quantity int, createdBy uuid.UUID, unitCostCents int64) error {
	var cost *int64
	if unitCostCents >= 0 {
		c := unitCostCents
		cost = &c
	}
	return s.applyMovement(ctx, skuID, movementApply{
		movementType:  MovementInitialStock,
		quantityDelta: quantity,
		reason:        "Estoque inicial",
		createdBy:     createdBy,
		unitCostCents: cost,
	})
}

func (s *Service) RegisterOutbound(ctx context.Context, skuID uuid.UUID, quantity int, movementType, reason string, createdBy uuid.UUID) error {
	if quantity <= 0 {
		return fmt.Errorf("invalid quantity")
	}
	movementType = strings.TrimSpace(movementType)
	if movementType != MovementLoss && movementType != MovementDamage {
		return fmt.Errorf("invalid movement type")
	}
	return s.applyMovement(ctx, skuID, movementApply{
		movementType:  movementType,
		quantityDelta: -quantity,
		reason:        reason,
		createdBy:     createdBy,
	})
}

func (s *Service) RegisterAdjustment(ctx context.Context, skuID uuid.UUID, physicalCount int, reason string, createdBy uuid.UUID) error {
	if strings.TrimSpace(reason) == "" {
		return fmt.Errorf("reason required")
	}
	if physicalCount < 0 {
		return fmt.Errorf("invalid physical count")
	}
	locID, _ := uuid.Parse(DefaultLocationID)
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	prev, err := s.currentBalance(ctx, tx, locID, skuID)
	if err != nil {
		return err
	}
	delta := physicalCount - prev
	if delta == 0 {
		return fmt.Errorf("no adjustment needed")
	}
	if err := s.applyMovementTx(ctx, tx, locID, skuID, movementApply{
		movementType:  MovementAdjustment,
		quantityDelta: delta,
		reason:        reason,
		createdBy:     createdBy,
	}); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *Service) applyMovement(ctx context.Context, skuID uuid.UUID, in movementApply) error {
	if in.quantityDelta == 0 {
		return fmt.Errorf("invalid quantity")
	}
	locID, _ := uuid.Parse(DefaultLocationID)
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if err := s.applyMovementTx(ctx, tx, locID, skuID, in); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *Service) applyMovementTx(ctx context.Context, tx pgx.Tx, locID, skuID uuid.UUID, in movementApply) error {
	if in.quantityDelta == 0 {
		return fmt.Errorf("invalid quantity")
	}
	qtyRecorded := in.quantityDelta
	if qtyRecorded < 0 {
		qtyRecorded = -qtyRecorded
	}
	if in.quantityDelta < 0 {
		if err := s.consumeLotsFIFOtx(ctx, tx, locID, skuID, qtyRecorded); err != nil {
			return err
		}
	}

	balanceID, prev, version, err := s.lockBalance(ctx, tx, locID, skuID)
	if err != nil {
		return err
	}
	newBal := prev + in.quantityDelta
	if newBal < 0 {
		return ErrInsufficientStock()
	}
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
	var refType *string
	if in.refType != "" {
		refType = &in.refType
	}
	var reason *string
	if in.reason != "" {
		reason = &in.reason
	}
	var movementID uuid.UUID
	err = tx.QueryRow(ctx, `
		INSERT INTO stock_movements (location_id, sku_id, movement_type, quantity, previous_balance, new_balance, reference_type, reference_id, reason, created_by, unit_cost_cents)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id
	`, locID, skuID, in.movementType, qtyRecorded, prev, newBal, refType, in.refID, reason, in.createdBy, in.unitCostCents).Scan(&movementID)
	if err != nil {
		return err
	}
	if in.movementType == MovementEntry && in.unitCostCents != nil {
		_, err = tx.Exec(ctx, `
			UPDATE skus SET cost_price_cents = $2, updated_at = NOW() WHERE id = $1
		`, skuID, *in.unitCostCents)
		if err != nil {
			return err
		}
	}
	if in.quantityDelta > 0 {
		unitCost := int64(0)
		switch in.movementType {
		case MovementEntry:
			if in.unitCostCents != nil {
				unitCost = *in.unitCostCents
			}
		case MovementInitialStock:
			if in.unitCostCents != nil {
				unitCost = *in.unitCostCents
			} else {
				unitCost = s.skuFallbackCostTx(ctx, tx, skuID)
			}
		case MovementAdjustment:
			if avg, ok, err := s.weightedAverageCostTx(ctx, tx, locID, skuID); err != nil {
				return err
			} else if ok {
				unitCost = avg
			} else {
				unitCost = s.skuFallbackCostTx(ctx, tx, skuID)
			}
		default:
			unitCost = s.skuFallbackCostTx(ctx, tx, skuID)
		}
		if err := s.createLotTx(ctx, tx, locID, skuID, qtyRecorded, unitCost, &movementID); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) lockBalance(ctx context.Context, tx pgx.Tx, locID, skuID uuid.UUID) (balanceID uuid.UUID, prev, version int, err error) {
	err = tx.QueryRow(ctx, `
		SELECT id, available_quantity, version FROM inventory_balances
		WHERE location_id = $1 AND sku_id = $2 FOR UPDATE
	`, locID, skuID).Scan(&balanceID, &prev, &version)
	if err == pgx.ErrNoRows {
		balanceID = uuid.New()
		prev = 0
		version = 0
		_, err = tx.Exec(ctx, `
			INSERT INTO inventory_balances (id, location_id, sku_id, available_quantity, version)
			VALUES ($1, $2, $3, 0, 0)
		`, balanceID, locID, skuID)
	}
	return balanceID, prev, version, err
}

func (s *Service) currentBalance(ctx context.Context, tx pgx.Tx, locID, skuID uuid.UUID) (int, error) {
	var prev int
	err := tx.QueryRow(ctx, `
		SELECT available_quantity FROM inventory_balances
		WHERE location_id = $1 AND sku_id = $2 FOR UPDATE
	`, locID, skuID).Scan(&prev)
	if err == pgx.ErrNoRows {
		return 0, nil
	}
	return prev, err
}

func (s *Service) ListBalances(ctx context.Context) ([]BalanceRow, error) {
	locID, _ := uuid.Parse(DefaultLocationID)
	rows, err := s.pool.Query(ctx, `
		SELECT s.id, s.code, p.id, p.name, s.minimum_stock, COALESCE(ib.available_quantity, 0)
		FROM skus s
		JOIN products p ON p.id = s.product_id
		LEFT JOIN inventory_balances ib ON ib.sku_id = s.id AND ib.location_id = $1
		ORDER BY p.name, s.code
	`, locID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []BalanceRow
	for rows.Next() {
		var r BalanceRow
		if err := rows.Scan(&r.SKUID, &r.SKUCode, &r.ProductID, &r.ProductName, &r.MinimumStock, &r.AvailableQuantity); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *Service) ListMovements(ctx context.Context, filter MovementFilter) ([]Movement, int, error) {
	limit := filter.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}
	locID, _ := uuid.Parse(DefaultLocationID)
	where := "WHERE sm.location_id = $1"
	args := []any{locID}
	arg := 2
	if filter.SKUID != nil {
		where += fmt.Sprintf(" AND sm.sku_id = $%d", arg)
		args = append(args, *filter.SKUID)
		arg++
	}
	if filter.ProductID != nil {
		where += fmt.Sprintf(" AND s.product_id = $%d", arg)
		args = append(args, *filter.ProductID)
		arg++
	}

	var total int
	countQ := `
		SELECT COUNT(*) FROM stock_movements sm
		JOIN skus s ON s.id = sm.sku_id
	` + where
	if err := s.pool.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	q := `
		SELECT sm.id, sm.sku_id, p.name, s.code, sm.movement_type, sm.quantity,
		       sm.previous_balance, sm.new_balance, sm.reference_type, sm.reference_id,
		       sm.reason, sm.created_by, u.email, sm.created_at, sm.unit_cost_cents
		FROM stock_movements sm
		JOIN skus s ON s.id = sm.sku_id
		JOIN products p ON p.id = s.product_id
		LEFT JOIN users u ON u.id = sm.created_by
	` + where
	q += fmt.Sprintf(" ORDER BY sm.created_at DESC LIMIT $%d OFFSET $%d", arg, arg+1)
	args = append(args, limit, offset)

	rows, err := s.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var out []Movement
	for rows.Next() {
		var m Movement
		if err := rows.Scan(
			&m.ID, &m.SKUID, &m.ProductName, &m.SKUCode, &m.MovementType, &m.Quantity,
			&m.PreviousBalance, &m.NewBalance, &m.ReferenceType, &m.ReferenceID,
			&m.Reason, &m.CreatedBy, &m.CreatedByEmail, &m.CreatedAt, &m.UnitCostCents,
		); err != nil {
			return nil, 0, err
		}
		out = append(out, m)
	}
	return out, total, rows.Err()
}

type MovementFilter struct {
	SKUID     *uuid.UUID
	ProductID *uuid.UUID
	Limit     int
	Offset    int
}
