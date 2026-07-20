package catalog

import (
	"context"

	"github.com/google/uuid"
)

// PriceLineInput is used for cart/checkout line pricing with shared promo quota per product.
type PriceLineInput struct {
	ProductID    uuid.UUID
	SKUID        uuid.UUID
	Quantity     int
	SalePrice    int64
	PromoActive  bool
	PromoRemain  int
}

// PriceLineResult holds computed line amounts and promo units consumed for quota simulation.
type PriceLineResult struct {
	LineTotalCents int64
	UnitPriceCents int64
	PromoUnits     int
}

func (s *Service) PriceLine(ctx context.Context, in PriceLineInput, avgCostFn func(context.Context, uuid.UUID) (int64, bool, error)) (PriceLineResult, error) {
	regular, err := s.RegularSalePriceCents(ctx, in.SKUID, avgCostFn)
	if err != nil {
		return PriceLineResult{}, err
	}
	if regular <= 0 {
		regular = in.SalePrice
	}
	lineTotal, unitAvg, promoUnits := SplitLinePriceCents(in.Quantity, in.PromoRemain, in.PromoActive, in.SalePrice, regular)
	return PriceLineResult{
		LineTotalCents: lineTotal,
		UnitPriceCents: unitAvg,
		PromoUnits:     promoUnits,
	}, nil
}
