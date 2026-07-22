package reports

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/store-platform/store/internal/inventory"
)

type InventoryPositionRow struct {
	ProductID         uuid.UUID  `json:"product_id"`
	ProductName       string     `json:"product_name"`
	SKUID             uuid.UUID  `json:"sku_id"`
	SKUCode           string     `json:"sku_code"`
	CategoryName      string     `json:"category_name,omitempty"`
	AvailableQuantity int        `json:"available_quantity"`
	MinimumStock      int        `json:"minimum_stock"`
	GapToMinimum      int        `json:"gap_to_minimum"`
	SalePriceCents    int64      `json:"sale_price_cents"`
	UnitCostCents     *int64     `json:"unit_cost_cents,omitempty"`
	StockValueCents   int64      `json:"stock_value_cents"`
	LastEntryAt       *time.Time `json:"last_entry_at,omitempty"`
	LastSaleAt        *time.Time `json:"last_sale_at,omitempty"`
	Situation         string     `json:"situation"`
	Active            bool       `json:"active"`
}

type InventoryPositionFilter struct {
	CategoryID   *uuid.UUID
	Situation    string
	ProductID    *uuid.UUID
	BelowMinimum bool
	ZeroStock    bool
	ActiveOnly   *bool
	PageFilter
}

func stockSituation(active bool, qty, min int) string {
	if !active {
		return "INATIVO"
	}
	if qty <= 0 {
		return "ZERADO"
	}
	if qty < min {
		return "BAIXO"
	}
	return "NORMAL"
}

func (s *Service) InventoryPosition(ctx context.Context, f InventoryPositionFilter) ([]InventoryPositionRow, int, error) {
	locID, _ := uuid.Parse(inventory.DefaultLocationID)
	where := []string{"1=1"}
	args := []any{locID}
	n := 2

	if f.ProductID != nil {
		where = append(where, fmt.Sprintf("p.id = $%d", n))
		args = append(args, *f.ProductID)
		n++
	}
	if f.CategoryID != nil {
		where = append(where, fmt.Sprintf("p.category_id = $%d", n))
		args = append(args, *f.CategoryID)
		n++
	}
	if f.ActiveOnly != nil {
		where = append(where, fmt.Sprintf("s.active = $%d", n))
		args = append(args, *f.ActiveOnly)
		n++
	}
	if f.BelowMinimum {
		where = append(where, "COALESCE(ib.available_quantity,0) < s.minimum_stock")
	}
	if f.ZeroStock {
		where = append(where, "COALESCE(ib.available_quantity,0) = 0")
	}
	if f.Situation != "" {
		switch f.Situation {
		case "INATIVO":
			where = append(where, "NOT s.active")
		case "ZERADO":
			where = append(where, "s.active AND COALESCE(ib.available_quantity,0) = 0")
		case "BAIXO":
			where = append(where, "s.active AND COALESCE(ib.available_quantity,0) > 0 AND COALESCE(ib.available_quantity,0) < s.minimum_stock")
		case "NORMAL":
			where = append(where, "s.active AND COALESCE(ib.available_quantity,0) >= s.minimum_stock")
		}
	}
	whereSQL := strings.Join(where, " AND ")

	fromSQL := `
		FROM skus s
		JOIN products p ON p.id = s.product_id
		LEFT JOIN categories cat ON cat.id = p.category_id
		LEFT JOIN inventory_balances ib ON ib.sku_id = s.id AND ib.location_id = $1
		WHERE ` + whereSQL

	var total int
	if err := s.pool.QueryRow(ctx, `SELECT COUNT(*)`+fromSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	limit, offset := f.Limit, f.Offset
	q := `
		SELECT p.id, p.name, s.id, s.code, COALESCE(cat.name,''),
		       COALESCE(ib.available_quantity,0), s.minimum_stock,
		       s.sale_price_cents, s.cost_price_cents,
		       (SELECT MAX(sm.created_at) FROM stock_movements sm
		        WHERE sm.sku_id = s.id AND sm.movement_type IN ('entry','initial_stock')),
		       (SELECT MAX(sm.created_at) FROM stock_movements sm
		        WHERE sm.sku_id = s.id AND sm.movement_type = 'sale'),
		       s.active
	` + fromSQL + fmt.Sprintf(" ORDER BY p.name, s.code LIMIT $%d OFFSET $%d", n, n+1)
	args = append(args, limit, offset)

	rows, err := s.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []InventoryPositionRow
	for rows.Next() {
		var r InventoryPositionRow
		var cost *int64
		if err := rows.Scan(
			&r.ProductID, &r.ProductName, &r.SKUID, &r.SKUCode, &r.CategoryName,
			&r.AvailableQuantity, &r.MinimumStock,
			&r.SalePriceCents, &cost,
			&r.LastEntryAt, &r.LastSaleAt, &r.Active,
		); err != nil {
			return nil, 0, err
		}
		r.UnitCostCents = cost
		r.GapToMinimum = r.MinimumStock - r.AvailableQuantity
		if r.GapToMinimum < 0 {
			r.GapToMinimum = 0
		}
		unit := r.SalePriceCents
		if cost != nil && *cost > 0 {
			unit = *cost
		}
		r.StockValueCents = int64(r.AvailableQuantity) * unit
		r.Situation = stockSituation(r.Active, r.AvailableQuantity, r.MinimumStock)
		out = append(out, r)
	}
	return out, total, rows.Err()
}
