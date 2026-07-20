package catalog

import (
	"context"

	"github.com/google/uuid"
)

// PriceLineInput is used for cart/checkout line pricing with shared promo quota per product.
type PriceLineInput struct {
	ProductID          uuid.UUID
	SKUID              uuid.UUID
	Quantity           int
	SalePrice          int64
	PromoActive        bool
	PromoRemain        int
	CollaboratorMargin *float64
}

// PriceLineResult holds computed line amounts and promo units consumed for quota simulation.
type PriceLineResult struct {
	LineTotalCents int64
	UnitPriceCents int64
	PromoUnits     int
}

func applyCollaboratorMin(promoUnit, regularUnit, collabUnit int64, hasCollab bool) (int64, int64) {
	if !hasCollab || collabUnit <= 0 {
		return promoUnit, regularUnit
	}
	effPromo := promoUnit
	if collabUnit < effPromo {
		effPromo = collabUnit
	}
	effRegular := regularUnit
	if collabUnit < effRegular {
		effRegular = collabUnit
	}
	return effPromo, effRegular
}

// ApplyCollaboratorMinForTest exposes pricing helper for tests.
func ApplyCollaboratorMinForTest(promoUnit, regularUnit, collabUnit int64, hasCollab bool) (int64, int64) {
	return applyCollaboratorMin(promoUnit, regularUnit, collabUnit, hasCollab)
}

func (s *Service) PriceLine(ctx context.Context, in PriceLineInput, avgCostFn func(context.Context, uuid.UUID) (int64, bool, error)) (PriceLineResult, error) {
	regular, err := s.RegularSalePriceCents(ctx, in.SKUID, avgCostFn)
	if err != nil {
		return PriceLineResult{}, err
	}
	if regular <= 0 {
		regular = in.SalePrice
	}
	promoQty := 0
	if in.PromoActive && in.PromoRemain > 0 && in.SalePrice > 0 {
		promoQty = in.Quantity
		if promoQty > in.PromoRemain {
			promoQty = in.PromoRemain
		}
	}
	regQty := in.Quantity - promoQty

	collabUnit := int64(0)
	hasCollab := false
	if in.CollaboratorMargin != nil {
		avgCost, ok, err := avgCostFn(ctx, in.SKUID)
		if err != nil {
			return PriceLineResult{}, err
		}
		if ok && avgCost > 0 {
			collabUnit = SalePriceFromCost(avgCost, *in.CollaboratorMargin)
			hasCollab = true
		}
	}
	effPromo, effRegular := applyCollaboratorMin(in.SalePrice, regular, collabUnit, hasCollab)
	lineTotal := int64(promoQty)*effPromo + int64(regQty)*effRegular
	var unitAvg int64
	if in.Quantity > 0 {
		unitAvg = lineTotal / int64(in.Quantity)
	}
	return PriceLineResult{
		LineTotalCents: lineTotal,
		UnitPriceCents: unitAvg,
		PromoUnits:     promoQty,
	}, nil
}
