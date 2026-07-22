package catalog_test

import (
	"testing"

	"github.com/store-platform/store/internal/catalog"
)

func TestPriceLineRoundsLineTotal(t *testing.T) {
	// Mixed line total before grid rounding (e.g. sum of per-unit amounts).
	const qty = 2
	lineTotal := int64(258 + 290)
	rounded := catalog.RoundSalePriceCents(lineTotal)
	if rounded != 550 {
		t.Fatalf("line total: got %d want 550", rounded)
	}
	unit := rounded / qty
	if unit != 275 {
		t.Fatalf("unit after line round: got %d want 275", unit)
	}
}
