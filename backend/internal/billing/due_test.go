package billing

import (
	"testing"
	"time"
)

func TestDueAtForCompetence(t *testing.T) {
	d := DueAtForCompetence(2026, 3)
	if d.Year() != 2026 || int(d.Month()) != 4 || d.Day() != 10 {
		t.Fatalf("got %v", d)
	}
	d2 := DueAtForCompetence(2026, 12)
	if d2.Year() != 2027 || int(d2.Month()) != 1 || d2.Day() != 10 {
		t.Fatalf("got %v", d2)
	}
}

func TestIsMonthlyClosingDay(t *testing.T) {
	d1 := time.Date(2026, 4, 1, 12, 0, 0, 0, saoPaulo)
	if !IsMonthlyClosingDay(d1) {
		t.Fatal("expected day 1")
	}
	d2 := time.Date(2026, 4, 2, 12, 0, 0, 0, saoPaulo)
	if IsMonthlyClosingDay(d2) {
		t.Fatal("expected not day 1")
	}
}
