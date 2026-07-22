package catalog

import "math"

// SalePriceRoundStepCents is the sale price grid step (R$ 0,50).
const SalePriceRoundStepCents = 50

// RoundSalePriceCents rounds to the nearest multiple of SalePriceRoundStepCents.
func RoundSalePriceCents(cents int64) int64 {
	if cents <= 0 {
		return cents
	}
	return int64(math.Round(float64(cents)/SalePriceRoundStepCents)) * SalePriceRoundStepCents
}
