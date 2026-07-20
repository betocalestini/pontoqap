package billing_test

import (
	"context"
	"testing"
	"time"

	"github.com/store-platform/store/internal/billing"
	"github.com/store-platform/store/tests/testdb"
)

func TestNthBusinessDayMarch2026(t *testing.T) {
	pool := testdb.Pool(t)
	ctx := context.Background()
	fifth, err := billing.NthBusinessDay(ctx, pool, 2026, 3, 5)
	if err != nil {
		t.Fatal(err)
	}
	// Março/2026: 1º = domingo → 5º dia útil = 6
	if fifth.Day() != 6 {
		t.Fatalf("expected 5th business day on 6, got %d", fifth.Day())
	}
}

func TestNthBusinessDayWithHoliday(t *testing.T) {
	testdb.MigrateUp(t)
	pool := testdb.Pool(t)
	ctx := context.Background()
	if err := testdb.Reset(ctx, pool); err != nil {
		t.Fatal(err)
	}
	svc := billing.NewService(pool, nil, "")
	loc, _ := time.LoadLocation("America/Sao_Paulo")
	mgr, err := testdb.SeedManager(ctx, pool, testdb.UniqueEmail(t, "mgr"))
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.UpsertCalendarDay(ctx, billing.CalendarEntry{
		Date:          time.Date(2026, 3, 2, 0, 0, 0, 0, loc),
		Name:          "Feriado teste",
		Scope:         "national",
		IsBusinessDay: false,
	}, mgr.UserID); err != nil {
		t.Fatal(err)
	}
	fifth, err := billing.NthBusinessDay(ctx, pool, 2026, 3, 5)
	if err != nil {
		t.Fatal(err)
	}
	if fifth.Day() != 9 {
		t.Fatalf("expected 5th business day shifted to 9, got %d", fifth.Day())
	}
}
