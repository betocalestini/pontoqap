package catalog

import (
	"context"
	"math"
	"strconv"

	"github.com/google/uuid"
)

const settingDefaultMargin = "default_margin_percent"

// SalePriceFromCost applies markup on cost: price = cost * (1 + margin%/100).
func SalePriceFromCost(costCents int64, marginPercent float64) int64 {
	if costCents <= 0 {
		return 0
	}
	f := float64(costCents) * (1 + marginPercent/100)
	return int64(math.Round(f))
}

func (s *Service) GetDefaultMarginPercent(ctx context.Context) (float64, error) {
	var val string
	err := s.pool.QueryRow(ctx, `SELECT value FROM store_settings WHERE key = $1`, settingDefaultMargin).Scan(&val)
	if err != nil {
		return 30, err
	}
	f, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return 30, err
	}
	return f, nil
}

func (s *Service) SetDefaultMarginPercent(ctx context.Context, margin float64) error {
	if margin < 0 || margin > 1000 {
		return ErrValidation("Margem inválida")
	}
	_, err := s.pool.Exec(ctx, `
		INSERT INTO store_settings (key, value, updated_at) VALUES ($1, $2, NOW())
		ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = NOW()
	`, settingDefaultMargin, strconv.FormatFloat(margin, 'f', 2, 64))
	return err
}

func (s *Service) ProductMarginPercent(ctx context.Context, productID uuid.UUID) (float64, error) {
	var m float64
	err := s.pool.QueryRow(ctx, `SELECT margin_percent FROM products WHERE id = $1`, productID).Scan(&m)
	return m, err
}

func (s *Service) effectiveMarginForProduct(ctx context.Context, productID uuid.UUID) (float64, error) {
	p, err := s.GetProduct(ctx, productID, false)
	if err != nil || p == nil {
		return 0, err
	}
	return EffectiveMarginPercent(*p), nil
}

// RecalculateSKU updates sale price from lot average cost and product margin.
func (s *Service) RecalculateSKU(ctx context.Context, skuID uuid.UUID, changedBy uuid.UUID, reason string, avgCostFn func(context.Context, uuid.UUID) (int64, bool, error)) (bool, error) {
	var productID uuid.UUID
	var currentPrice int64
	err := s.pool.QueryRow(ctx, `
		SELECT product_id, sale_price_cents FROM skus WHERE id = $1
	`, skuID).Scan(&productID, &currentPrice)
	if err != nil {
		return false, err
	}
	margin, err := s.effectiveMarginForProduct(ctx, productID)
	if err != nil {
		return false, err
	}
	avgCost, ok, err := avgCostFn(ctx, skuID)
	if err != nil {
		return false, err
	}
	if !ok || avgCost <= 0 {
		return false, nil
	}
	newPrice := SalePriceFromCost(avgCost, margin)
	if newPrice <= 0 {
		return false, nil
	}
	if newPrice == currentPrice {
		return false, nil
	}
	if err := s.ChangeSKUPrice(ctx, skuID, newPrice, changedBy, reason); err != nil {
		return false, err
	}
	return true, nil
}

func (s *Service) RecalculateProductSKUs(ctx context.Context, productID uuid.UUID, changedBy uuid.UUID, reason string, avgCostFn func(context.Context, uuid.UUID) (int64, bool, error)) error {
	rows, err := s.pool.Query(ctx, `SELECT id FROM skus WHERE product_id = $1`, productID)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var skuID uuid.UUID
		if err := rows.Scan(&skuID); err != nil {
			return err
		}
		if _, err := s.RecalculateSKU(ctx, skuID, changedBy, reason, avgCostFn); err != nil {
			return err
		}
	}
	return rows.Err()
}

func (s *Service) RepriceAllProducts(ctx context.Context, marginPercent float64, changedBy uuid.UUID, avgCostFn func(context.Context, uuid.UUID) (int64, bool, error)) (int, error) {
	if marginPercent < 0 || marginPercent > 1000 {
		return 0, ErrValidation("Margem inválida")
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)
	res, err := tx.Exec(ctx, `UPDATE products SET margin_percent = $1, updated_at = NOW()`, marginPercent)
	if err != nil {
		return 0, err
	}
	n := int(res.RowsAffected())
	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}
	rows, err := s.pool.Query(ctx, `SELECT id FROM products`)
	if err != nil {
		return n, err
	}
	defer rows.Close()
	for rows.Next() {
		var pid uuid.UUID
		if err := rows.Scan(&pid); err != nil {
			return n, err
		}
		if err := s.RecalculateProductSKUs(ctx, pid, changedBy, "bulk:margem", avgCostFn); err != nil {
			return n, err
		}
	}
	return n, rows.Err()
}

type ValidationError struct{ Msg string }

func (e ValidationError) Error() string { return e.Msg }

func ErrValidation(msg string) error { return ValidationError{Msg: msg} }
