package devseed_test

import (
	"testing"

	"github.com/store-platform/store/internal/devseed"
)

func TestExpandDomain(t *testing.T) {
	got := devseed.ExpandDomainForTest("cli@{domain}", "demo.loja.local")
	if got != "cli@demo.loja.local" {
		t.Fatalf("got %q", got)
	}
}

func TestParseBoolForTest(t *testing.T) {
	for _, tc := range []struct {
		in   string
		want bool
	}{
		{"true", true},
		{"sim", true},
		{"false", false},
		{"não", false},
	} {
		got, err := devseed.ParseBoolForTest(tc.in)
		if err != nil || got != tc.want {
			t.Fatalf("%q: got %v err %v", tc.in, got, err)
		}
	}
}

func TestParseStockQtyForTest(t *testing.T) {
	got, err := devseed.ParseStockQtyForTest("0", 50)
	if err != nil || got != 50 {
		t.Fatalf("zero: got %d err %v", got, err)
	}
	got, err = devseed.ParseStockQtyForTest("", 40)
	if err != nil || got != 40 {
		t.Fatalf("empty: got %d err %v", got, err)
	}
	got, err = devseed.ParseStockQtyForTest("12", 50)
	if err != nil || got != 12 {
		t.Fatalf("explicit: got %d err %v", got, err)
	}
}

func TestLoadProductsCSV(t *testing.T) {
	dir := testDataDir(t)
	rows, err := devseed.LoadProductsCSVForTest(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 {
		t.Fatalf("want 2 rows got %d", len(rows))
	}
	if rows[0].UnitCostCents != 1000 || rows[0].StockQty != 10 {
		t.Fatalf("row0: %+v", rows[0])
	}
}
