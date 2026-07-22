package devseed

import (
	"math/rand"
	"testing"

	"github.com/google/uuid"
)

func TestPickAffordableLine(t *testing.T) {
	products := []productSeed{
		{SKUID: uuid.New(), SalePriceCents: 100_000},
		{SKUID: uuid.New(), SalePriceCents: 500},
	}
	rng := rand.New(rand.NewSource(1))
	_, qty, ok := pickAffordableLine(products, 600, rng, 3)
	if !ok || qty < 1 {
		t.Fatalf("expected affordable line, ok=%v qty=%d", ok, qty)
	}
	_, _, ok = pickAffordableLine(products, 100, rng, 3)
	if ok {
		t.Fatal("expected no line when limit too low for cheapest")
	}
}
