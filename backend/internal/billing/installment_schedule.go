package billing

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func lastDayOfMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

// AddMonthsPreserveDay advances months from anchor, clamping to last valid day of target month.
func AddMonthsPreserveDay(anchor time.Time, months int) time.Time {
	loc := anchor.Location()
	y, m, d := anchor.In(loc).Date()
	tm := int(m) + months
	for tm > 12 {
		tm -= 12
		y++
	}
	for tm < 1 {
		tm += 12
		y--
	}
	month := time.Month(tm)
	ld := lastDayOfMonth(y, month)
	if d > ld {
		d = ld
	}
	return time.Date(y, month, d, 0, 0, 0, 0, loc)
}

// BuildInstallmentDueDates returns due dates for each installment (1..count) from invoice due_at.
func BuildInstallmentDueDates(
	ctx context.Context,
	pool *pgxpool.Pool,
	firstDue time.Time,
	count int,
	intervalMonths int,
	adjustBusinessDay bool,
) ([]time.Time, error) {
	if intervalMonths < 1 {
		intervalMonths = 1
	}
	out := make([]time.Time, count)
	anchor := calendarDateOnly(firstDue)
	for i := 0; i < count; i++ {
		raw := AddMonthsPreserveDay(anchor, i*intervalMonths)
		if adjustBusinessDay {
			adj, err := NextBusinessDay(ctx, pool, raw)
			if err != nil {
				return nil, err
			}
			out[i] = adj
		} else {
			out[i] = raw
		}
	}
	return out, nil
}
