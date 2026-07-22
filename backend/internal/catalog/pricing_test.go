package catalog_test

import (
	"testing"

	"github.com/store-platform/store/internal/catalog"
)

func TestSalePriceFromCost(t *testing.T) {
	if got := catalog.SalePriceFromCost(1000, 30); got != 1300 {
		t.Fatalf("expected 1300, got %d", got)
	}
	if got := catalog.SalePriceFromCost(1050, 25); got != 1300 {
		t.Fatalf("expected 1300, got %d", got)
	}
	if got := catalog.SalePriceFromCost(0, 30); got != 0 {
		t.Fatalf("expected 0, got %d", got)
	}
}
