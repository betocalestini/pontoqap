package forecasting

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	pool *pgxpool.Pool
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

type Snapshot struct {
	SKUID                     uuid.UUID `json:"sku_id"`
	SKUCode                   string    `json:"sku_code"`
	ProductName               string    `json:"product_name,omitempty"`
	ReferenceMonth            time.Time `json:"reference_month"`
	SalesLast3Months          int64     `json:"sales_last_3_months"`
	CurrentStock              int       `json:"current_stock"`
	ForecastQuantity          int       `json:"forecast_quantity"`
	SafetyStockQuantity       int       `json:"safety_stock_quantity"`
	SuggestedPurchaseQuantity int       `json:"suggested_purchase_quantity"`
	ConfidenceLevel           string    `json:"confidence_level"`
	Method                    string    `json:"method"`
}

// GenerateMonthlySnapshots usa média simples dos últimos 3 meses de vendas por SKU.
func (s *Service) GenerateMonthlySnapshots(ctx context.Context, ref time.Time) (int, error) {
	ref = time.Date(ref.Year(), ref.Month(), 1, 0, 0, 0, 0, time.UTC)
	rows, err := s.pool.Query(ctx, `
		SELECT oi.sku_id, s.code, COALESCE(SUM(oi.quantity),0)::bigint
		FROM order_items oi
		JOIN orders o ON o.id = oi.order_id
		JOIN skus s ON s.id = oi.sku_id
		WHERE o.confirmed_at >= $1::timestamptz - interval '3 months'
		  AND o.confirmed_at < $1::timestamptz + interval '1 month'
		GROUP BY oi.sku_id, s.code
	`, ref)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var skuID uuid.UUID
		var code string
		var qty int64
		if err := rows.Scan(&skuID, &code, &qty); err != nil {
			return count, err
		}
		forecast := int(qty / 3)
		if forecast < 0 {
			forecast = 0
		}
		safety := forecast / 5
		var stock int
		_ = s.pool.QueryRow(ctx, `SELECT COALESCE(SUM(available_quantity),0) FROM inventory_balances WHERE sku_id = $1`, skuID).Scan(&stock)
		suggest := forecast + safety - stock
		if suggest < 0 {
			suggest = 0
		}
		conf := "medium"
		if qty == 0 {
			conf = "low"
		}
		params, _ := json.Marshal(map[string]any{"months": 3, "total_qty": qty})
		_, err = s.pool.Exec(ctx, `
			INSERT INTO forecast_snapshots (
				sku_id, reference_month, forecast_quantity, safety_stock_quantity,
				suggested_purchase_quantity, confidence_level, method, parameters
			) VALUES ($1,$2,$3,$4,$5,$6,'weighted_avg_3m',$7)
		`, skuID, ref, forecast, safety, suggest, conf, params)
		if err != nil {
			return count, err
		}
		count++
	}
	return count, rows.Err()
}

func (s *Service) ListLatest(ctx context.Context, limit int) ([]Snapshot, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.pool.Query(ctx, `
		SELECT fs.sku_id, s.code, p.name, fs.reference_month, fs.forecast_quantity,
		       fs.safety_stock_quantity, fs.suggested_purchase_quantity,
		       fs.confidence_level, fs.method,
		       COALESCE((
		         SELECT SUM(oi.quantity) FROM order_items oi
		         JOIN orders o ON o.id = oi.order_id
		         WHERE oi.sku_id = fs.sku_id AND o.status = 'confirmed'
		           AND o.confirmed_at >= fs.reference_month - interval '3 months'
		           AND o.confirmed_at < fs.reference_month + interval '1 month'
		       ),0),
		       COALESCE((SELECT SUM(available_quantity) FROM inventory_balances WHERE sku_id = fs.sku_id),0)
		FROM forecast_snapshots fs
		JOIN skus s ON s.id = fs.sku_id
		JOIN products p ON p.id = s.product_id
		ORDER BY fs.created_at DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Snapshot
	for rows.Next() {
		var sn Snapshot
		if err := rows.Scan(&sn.SKUID, &sn.SKUCode, &sn.ProductName, &sn.ReferenceMonth, &sn.ForecastQuantity,
			&sn.SafetyStockQuantity, &sn.SuggestedPurchaseQuantity, &sn.ConfidenceLevel, &sn.Method,
			&sn.SalesLast3Months, &sn.CurrentStock); err != nil {
			return nil, err
		}
		out = append(out, sn)
	}
	return out, rows.Err()
}
