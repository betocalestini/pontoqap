package billing

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var saoPaulo *time.Location

func init() {
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		loc = time.FixedZone("BRT", -3*3600)
	}
	saoPaulo = loc
}

type CalendarEntry struct {
	Date          time.Time `json:"date"`
	Name          string    `json:"name"`
	Scope         string    `json:"scope"`
	IsBusinessDay bool      `json:"is_business_day"`
}

func calendarDateOnly(t time.Time) time.Time {
	d := t.In(saoPaulo)
	return time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, time.UTC)
}

func (s *Service) UpsertCalendarDay(ctx context.Context, entry CalendarEntry, actorID uuid.UUID) error {
	date := calendarDateOnly(entry.Date)
	_, err := s.pool.Exec(ctx, `
		INSERT INTO business_calendar (date, name, scope, is_business_day, created_by)
		VALUES ($1::date, $2, $3, $4, $5)
		ON CONFLICT (date) DO UPDATE SET
			name = EXCLUDED.name,
			scope = EXCLUDED.scope,
			is_business_day = EXCLUDED.is_business_day,
			updated_at = NOW()
	`, date, entry.Name, entry.Scope, entry.IsBusinessDay, actorID)
	return err
}

func (s *Service) ListCalendar(ctx context.Context, from, to time.Time) ([]CalendarEntry, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT date, name, scope, is_business_day FROM business_calendar
		WHERE date BETWEEN $1::date AND $2::date ORDER BY date
	`, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []CalendarEntry
	for rows.Next() {
		var e CalendarEntry
		if err := rows.Scan(&e.Date, &e.Name, &e.Scope, &e.IsBusinessDay); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func IsBusinessDay(ctx context.Context, pool *pgxpool.Pool, day time.Time) (bool, error) {
	date := calendarDateOnly(day)
	var isBusiness bool
	err := pool.QueryRow(ctx, `
		SELECT is_business_day FROM business_calendar WHERE date = $1::date
	`, date).Scan(&isBusiness)
	if err == pgx.ErrNoRows {
		wd := date.Weekday()
		return wd != time.Saturday && wd != time.Sunday, nil
	}
	if err != nil {
		return false, err
	}
	return isBusiness, nil
}

func NthBusinessDay(ctx context.Context, pool *pgxpool.Pool, year, month, n int) (time.Time, error) {
	if n < 1 {
		n = 1
	}
	count := 0
	for day := 1; day <= 31; day++ {
		d := time.Date(year, time.Month(month), day, 0, 0, 0, 0, saoPaulo)
		if d.Month() != time.Month(month) {
			break
		}
		ok, err := IsBusinessDay(ctx, pool, d)
		if err != nil {
			return time.Time{}, err
		}
		if ok {
			count++
			if count == n {
				return d, nil
			}
		}
	}
	return time.Time{}, ErrNoBusinessDay
}

func IsMonthlyClosingDay(ctx context.Context, pool *pgxpool.Pool, now time.Time) (bool, error) {
	now = now.In(saoPaulo)
	fifth, err := NthBusinessDay(ctx, pool, now.Year(), int(now.Month()), 5)
	if err != nil {
		return false, err
	}
	return sameDate(now, fifth), nil
}

func sameDate(a, b time.Time) bool {
	a = a.In(saoPaulo)
	b = b.In(saoPaulo)
	return a.Year() == b.Year() && a.Month() == b.Month() && a.Day() == b.Day()
}

func PreviousMonth(year, month int) (int, int) {
	if month == 1 {
		return year - 1, 12
	}
	return year, month - 1
}
