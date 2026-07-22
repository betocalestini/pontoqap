package integration_test

import (
	"context"
	"testing"

	"github.com/store-platform/store/internal/billing"
	"github.com/store-platform/store/tests/testdb"
)

func TestInstallmentPlanSelectionAndDisabled(t *testing.T) {
	testdb.MigrateUp(t)
	pool := testdb.Pool(t)
	ctx := context.Background()
	if err := testdb.Reset(ctx, pool); err != nil {
		t.Fatal(err)
	}

	svc := billing.NewService(pool, nil, "http://localhost:5173")
	policy, err := svc.GetActiveInstallmentPolicy(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !policy.InstallmentEnabled {
		t.Fatal("expected default policy enabled")
	}

	total := int64(35000)
	max := billing.MaxInstallments(total, policy.Params(), true)
	if max != 3 {
		t.Fatalf("max installments: got %d want 3", max)
	}
	maxOff := billing.MaxInstallments(total, policy.Params(), false)
	if maxOff != 1 {
		t.Fatalf("disabled max: got %d", maxOff)
	}

	amounts := billing.DistributeInstallmentAmounts(total, 3)
	var sum int64
	for _, a := range amounts {
		sum += a
	}
	if sum != total {
		t.Fatalf("sum %d != %d", sum, total)
	}
}
