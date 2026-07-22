package reports

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/store-platform/store/internal/inventory"
)

type InventoryMovementRow struct {
	ID              uuid.UUID  `json:"id"`
	CreatedAt       time.Time  `json:"created_at"`
	ProductName     string     `json:"product_name"`
	SKUCode         string     `json:"sku_code"`
	MovementType    string     `json:"movement_type"`
	Quantity        int        `json:"quantity"`
	PreviousBalance int        `json:"previous_balance"`
	NewBalance      int        `json:"new_balance"`
	Reason          *string    `json:"reason,omitempty"`
	ReferenceType   *string    `json:"reference_type,omitempty"`
	ReferenceID     *uuid.UUID `json:"reference_id,omitempty"`
	CreatedByEmail  *string    `json:"created_by_email,omitempty"`
}

type InventoryMovementsFilter struct {
	DateRange
	SKUID        *uuid.UUID
	ProductID    *uuid.UUID
	MovementType string
	UserID       *uuid.UUID
	OrderID      *uuid.UUID
	ManualOnly   bool
	PageFilter
}

func (s *Service) InventoryMovements(ctx context.Context, f InventoryMovementsFilter) ([]InventoryMovementRow, int, error) {
	locID, _ := uuid.Parse(inventory.DefaultLocationID)
	where := []string{"sm.location_id = $1", "sm.created_at >= $2", "sm.created_at < $3"}
	args := []any{locID, f.From, f.To}
	n := 4

	if f.SKUID != nil {
		where = append(where, fmt.Sprintf("sm.sku_id = $%d", n))
		args = append(args, *f.SKUID)
		n++
	}
	if f.ProductID != nil {
		where = append(where, fmt.Sprintf("s.product_id = $%d", n))
		args = append(args, *f.ProductID)
		n++
	}
	if mt := strings.TrimSpace(f.MovementType); mt != "" {
		where = append(where, fmt.Sprintf("sm.movement_type = $%d", n))
		args = append(args, mt)
		n++
	}
	if f.UserID != nil {
		where = append(where, fmt.Sprintf("sm.created_by = $%d", n))
		args = append(args, *f.UserID)
		n++
	}
	if f.OrderID != nil {
		where = append(where, fmt.Sprintf("sm.reference_type = 'order' AND sm.reference_id = $%d", n))
		args = append(args, *f.OrderID)
		n++
	}
	if f.ManualOnly {
		where = append(where, "sm.movement_type IN ('entry','loss','damage','adjustment','return')")
	}

	whereSQL := strings.Join(where, " AND ")

	var total int
	countQ := `
		SELECT COUNT(*) FROM stock_movements sm
		JOIN skus s ON s.id = sm.sku_id
		WHERE ` + whereSQL
	if err := s.pool.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	limit, offset := f.Limit, f.Offset
	q := `
		SELECT sm.id, sm.created_at, p.name, s.code, sm.movement_type, sm.quantity,
		       sm.previous_balance, sm.new_balance, sm.reason, sm.reference_type, sm.reference_id,
		       u.email
		FROM stock_movements sm
		JOIN skus s ON s.id = sm.sku_id
		JOIN products p ON p.id = s.product_id
		LEFT JOIN users u ON u.id = sm.created_by
		WHERE ` + whereSQL + fmt.Sprintf(" ORDER BY sm.created_at DESC LIMIT $%d OFFSET $%d", n, n+1)
	args = append(args, limit, offset)

	rows, err := s.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []InventoryMovementRow
	for rows.Next() {
		var r InventoryMovementRow
		if err := rows.Scan(
			&r.ID, &r.CreatedAt, &r.ProductName, &r.SKUCode, &r.MovementType, &r.Quantity,
			&r.PreviousBalance, &r.NewBalance, &r.Reason, &r.ReferenceType, &r.ReferenceID,
			&r.CreatedByEmail,
		); err != nil {
			return nil, 0, err
		}
		out = append(out, r)
	}
	return out, total, rows.Err()
}
