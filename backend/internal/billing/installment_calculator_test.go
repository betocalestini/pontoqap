package billing_test

import (
	"testing"

	"github.com/store-platform/store/internal/billing"
)

func defaultPolicy() billing.InstallmentPolicyParams {
	return billing.InstallmentPolicyParams{
		MinimumInvoiceAmountCents:     30000,
		MinimumInstallmentAmountCents: 10000,
		MaximumInstallments:           10,
		InstallmentEnabled:            true,
	}
}

func TestMaxInstallmentsLimits(t *testing.T) {
	p := defaultPolicy()
	cases := []struct {
		total int64
		want  int
	}{
		{29999, 1},
		{30000, 3},
		{39999, 3},
		{40000, 4},
		{99999, 9},
		{100000, 10},
		{200000, 10},
	}
	for _, c := range cases {
		got := billing.MaxInstallments(c.total, p, true)
		if got != c.want {
			t.Fatalf("total %d: got %d want %d", c.total, got, c.want)
		}
	}
}

func TestMaxInstallmentsDisabled(t *testing.T) {
	p := defaultPolicy()
	if got := billing.MaxInstallments(500000, p, false); got != 1 {
		t.Fatalf("expected 1 when disabled, got %d", got)
	}
}

func TestDistributeCentavos(t *testing.T) {
	sum := func(a []int64) int64 {
		var s int64
		for _, v := range a {
			s += v
		}
		return s
	}
	cases := []struct {
		total int64
		n     int
		want  []int64
	}{
		{31000, 3, []int64{10333, 10333, 10334}},
		{35000, 3, []int64{11666, 11667, 11667}},
	}
	p := defaultPolicy()
	for _, c := range cases {
		got := billing.DistributeInstallmentAmounts(c.total, c.n)
		if sum(got) != c.total {
			t.Fatalf("sum mismatch %d", c.total)
		}
		for _, g := range got {
			if g < p.MinimumInstallmentAmountCents && c.n > 1 {
				t.Fatalf("below minimum: %v", got)
			}
		}
		if len(c.want) > 0 {
			for i := range c.want {
				if got[i] != c.want[i] {
					t.Fatalf("total %d n %d: got %v want %v", c.total, c.n, got, c.want)
				}
			}
		}
	}
}
