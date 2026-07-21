package inventory

import "fmt"

// EntryUnitCostCents derives per-unit cost from purchase total, other expenses and quantity.
func EntryUnitCostCents(totalPaidCents, otherExpensesCents int64, quantity int) (int64, error) {
	if quantity <= 0 {
		return 0, fmt.Errorf("invalid quantity")
	}
	if totalPaidCents < 0 || otherExpensesCents < 0 {
		return 0, fmt.Errorf("invalid cost")
	}
	total := totalPaidCents + otherExpensesCents
	q := int64(quantity)
	return (total + q/2) / q, nil
}
