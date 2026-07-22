package billing_test

import (
	"context"
	"testing"
	"time"

	"github.com/store-platform/store/internal/billing"
	"github.com/store-platform/store/tests/testdb"
)

func TestAddMonthsPreserveDay(t *testing.T) {
	loc, _ := time.LoadLocation("America/Sao_Paulo")
	anchor := time.Date(2026, 1, 31, 0, 0, 0, 0, loc)
	got := billing.AddMonthsPreserveDay(anchor, 1)
	if got.Day() != 28 && got.Day() != 29 {
		t.Fatalf("expected feb last day, got %v", got)
	}
}

func TestBuildInstallmentDueDates(t *testing.T) {
	pool := testdb.Pool(t)
	ctx := context.Background()
	first := time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC)
	dates, err := billing.BuildInstallmentDueDates(ctx, pool, first, 3, 1, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(dates) != 3 {
		t.Fatalf("want 3 dates")
	}
	if dates[0].Day() != 15 || dates[1].Month() != time.September || dates[2].Month() == time.August {
		t.Fatalf("unexpected schedule %v", dates)
	}
}
