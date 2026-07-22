package catalog_test

import (
	"testing"

	"github.com/store-platform/store/internal/catalog"
)

func TestRoundSalePriceCents(t *testing.T) {
	cases := []struct {
		in, want int64
	}{
		{258, 250},
		{290, 300},
		{1300, 1300},
		{0, 0},
		{-10, -10},
		{25, 50},
		{24, 0},
	}
	for _, c := range cases {
		if got := catalog.RoundSalePriceCents(c.in); got != c.want {
			t.Fatalf("RoundSalePriceCents(%d) = %d, want %d", c.in, got, c.want)
		}
	}
}
