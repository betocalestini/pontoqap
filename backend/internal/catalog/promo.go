package catalog

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// EffectiveMarginPercent returns margin used for catalog sale_price (promo while quota remains).
func EffectiveMarginPercent(p Product) float64 {
	if p.PromoActive && p.PromoQuantityRemaining > 0 && p.PromoMarginPercent != nil {
		return *p.PromoMarginPercent
	}
	return p.MarginPercent
}

// IsOnPromotion reports whether the product should show promo pricing in the store.
func (p Product) IsOnPromotion() bool {
	return p.PromoActive && p.PromoQuantityRemaining > 0
}

// SplitLinePriceCents computes mixed promo + regular totals (option A).
func SplitLinePriceCents(qty int, promoRemaining int, promoActive bool, promoUnitPrice, regularUnitPrice int64) (lineTotal, unitAvg int64, promoUnits int) {
	if qty <= 0 {
		return 0, 0, 0
	}
	promoUnits = 0
	if promoActive && promoRemaining > 0 && promoUnitPrice > 0 {
		promoUnits = qty
		if promoUnits > promoRemaining {
			promoUnits = promoRemaining
		}
	}
	regularUnits := qty - promoUnits
	lineTotal = int64(promoUnits)*promoUnitPrice + int64(regularUnits)*regularUnitPrice
	unitAvg = lineTotal / int64(qty)
	return lineTotal, unitAvg, promoUnits
}

func (s *Service) RegularSalePriceCents(ctx context.Context, skuID uuid.UUID, avgCostFn func(context.Context, uuid.UUID) (int64, bool, error)) (int64, error) {
	var productID uuid.UUID
	var margin float64
	err := s.pool.QueryRow(ctx, `
		SELECT product_id, margin_percent FROM skus s JOIN products p ON p.id = s.product_id WHERE s.id = $1
	`, skuID).Scan(&productID, &margin)
	if err != nil {
		return 0, err
	}
	avgCost, ok, err := avgCostFn(ctx, skuID)
	if err != nil {
		return 0, err
	}
	if !ok || avgCost <= 0 {
		return 0, nil
	}
	return SalePriceFromCost(avgCost, margin), nil
}

// ConsumePromoQuotaTx decrements promo remaining for a product; returns units charged at promo price.
func (s *Service) ConsumePromoQuotaTx(ctx context.Context, tx pgx.Tx, productID uuid.UUID, quantity int) (promoUnits int, promoEnded bool, err error) {
	var active bool
	var remaining int
	err = tx.QueryRow(ctx, `
		SELECT promo_active, promo_quantity_remaining FROM products WHERE id = $1 FOR UPDATE
	`, productID).Scan(&active, &remaining)
	if err != nil {
		return 0, false, err
	}
	if !active || remaining <= 0 {
		return 0, false, nil
	}
	promoUnits = quantity
	if promoUnits > remaining {
		promoUnits = remaining
	}
	newRem := remaining - promoUnits
	promoEnded = newRem <= 0
	_, err = tx.Exec(ctx, `
		UPDATE products SET
			promo_quantity_remaining = $2,
			promo_active = CASE WHEN $2 <= 0 THEN FALSE ELSE promo_active END,
			updated_at = NOW()
		WHERE id = $1
	`, productID, newRem)
	if err != nil {
		return 0, false, err
	}
	return promoUnits, promoEnded, nil
}

func validatePromoMargin(m float64) error {
	if m < 0 || m > 1000 {
		return ErrValidation("Margem promocional inválida")
	}
	return nil
}
