package reports

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type ExceptionRow struct {
	OccurredAt   time.Time  `json:"occurred_at"`
	EventType    string     `json:"event_type"`
	EntityType   string     `json:"entity_type"`
	EntityID     uuid.UUID  `json:"entity_id"`
	Label        string     `json:"label"`
	AmountCents  *int64     `json:"amount_cents,omitempty"`
	Quantity     *int       `json:"quantity,omitempty"`
	Reason       string     `json:"reason,omitempty"`
	ActorEmail   string     `json:"actor_email,omitempty"`
	Status       string     `json:"status,omitempty"`
}

type ExceptionsFilter struct {
	DateRange
	EventType  string
	CustomerID *uuid.UUID
	ProductID  *uuid.UUID
	ManualOnly bool
	PageFilter
}

func (s *Service) Exceptions(ctx context.Context, f ExceptionsFilter) ([]ExceptionRow, int, error) {
	// Union of key exception sources; count approximated from union subquery
	union := `
		SELECT o.cancelled_at AS occurred_at, 'order_cancelled' AS event_type, 'order' AS entity_type,
		       o.id AS entity_id, o.order_number AS label, o.total_cents AS amount_cents, NULL::int AS quantity,
		       '' AS reason, '' AS actor_email, o.status
		FROM orders o
		WHERE o.status = 'cancelled' AND o.cancelled_at >= $1 AND o.cancelled_at < $2
		UNION ALL
		SELECT sm.created_at, sm.movement_type, 'stock_movement', sm.id, s.code,
		       NULL, sm.quantity, COALESCE(sm.reason,''), COALESCE(u.email,''), sm.movement_type
		FROM stock_movements sm
		JOIN skus s ON s.id = sm.sku_id
		LEFT JOIN users u ON u.id = sm.created_by
		WHERE sm.movement_type IN ('loss','damage','adjustment')
		  AND sm.created_at >= $1 AND sm.created_at < $2
		UNION ALL
		SELECT ba.created_at, 'invoice_adjustment', 'invoice', i.id, i.invoice_number,
		       ba.amount_cents, NULL, ba.reason, COALESCE(u.email,''), ba.adjustment_type
		FROM billing_adjustments ba
		JOIN invoices i ON i.id = ba.invoice_id
		LEFT JOIN users u ON u.id = ba.created_by
		WHERE ba.created_at >= $1 AND ba.created_at < $2
	`
	args := []any{f.From, f.To}

	var total int
	if err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM (`+union+`) x`, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	limit, offset := f.Limit, f.Offset
	q := fmt.Sprintf(`SELECT * FROM (%s) x ORDER BY occurred_at DESC LIMIT $3 OFFSET $4`, union)
	args = append(args, limit, offset)

	rows, err := s.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []ExceptionRow
	for rows.Next() {
		var r ExceptionRow
		var amt *int64
		var qty *int
		var reason, actor, status string
		if err := rows.Scan(
			&r.OccurredAt, &r.EventType, &r.EntityType, &r.EntityID, &r.Label,
			&amt, &qty, &reason, &actor, &status,
		); err != nil {
			return nil, 0, err
		}
		r.AmountCents = amt
		r.Quantity = qty
		r.Reason = reason
		r.ActorEmail = actor
		r.Status = status
		if f.EventType != "" && r.EventType != f.EventType {
			continue
		}
		out = append(out, r)
	}
	return out, total, rows.Err()
}

// ExceptionsSummary returns aggregate counts for the period.
func (s *Service) ExceptionsSummary(ctx context.Context, dr DateRange) (map[string]int64, error) {
	out := map[string]int64{}
	var n int64
	_ = s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM orders WHERE status = 'cancelled' AND cancelled_at >= $1 AND cancelled_at < $2
	`, dr.From, dr.To).Scan(&n)
	out["cancelled_orders"] = n
	_ = s.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(quantity),0) FROM stock_movements
		WHERE movement_type IN ('loss','damage') AND created_at >= $1 AND created_at < $2
	`, dr.From, dr.To).Scan(&n)
	out["stock_loss_units"] = n
	return out, nil
}
