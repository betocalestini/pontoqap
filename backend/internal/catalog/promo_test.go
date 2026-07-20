package catalog_test

import (
	"testing"

	"github.com/store-platform/store/internal/catalog"
)

func TestSplitLinePriceCents(t *testing.T) {
	total, avg, promo := catalog.SplitLinePriceCents(5, 3, true, 1000, 1500)
	if promo != 3 {
		t.Fatalf("promo units: got %d", promo)
	}
	want := int64(3*1000 + 2*1500)
	if total != want {
		t.Fatalf("total: got %d want %d", total, want)
	}
	if avg != want/5 {
		t.Fatalf("avg: got %d want %d", avg, want/5)
	}
}

func TestEffectiveMarginPercent(t *testing.T) {
	m := 30.0
	pm := 10.0
	p := catalog.Product{MarginPercent: m, PromoActive: true, PromoQuantityRemaining: 2, PromoMarginPercent: &pm}
	if catalog.EffectiveMarginPercent(p) != pm {
		t.Fatal("expected promo margin")
	}
	p.PromoQuantityRemaining = 0
	if catalog.EffectiveMarginPercent(p) != m {
		t.Fatal("expected normal margin")
	}
}
