package customers_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/store-platform/store/internal/customers"
)

func TestAvailableLimit(t *testing.T) {
	c := customers.Customer{
		ID:                   uuid.New(),
		CreditLimitCents:     10_000,
		CurrentExposureCents: 3_500,
	}
	if got := customers.NewService(nil, nil).AvailableLimit(c); got != 6_500 {
		t.Fatalf("expected 6500, got %d", got)
	}
}

func TestAvailableLimitNeverNegative(t *testing.T) {
	c := customers.Customer{
		CreditLimitCents:     1_000,
		CurrentExposureCents: 5_000,
	}
	if got := customers.NewService(nil, nil).AvailableLimit(c); got != 0 {
		t.Fatalf("expected 0, got %d", got)
	}
}
