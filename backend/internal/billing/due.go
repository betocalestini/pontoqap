package billing

import "time"

// DueAtForCompetence returns payment due at day 10 of the month following competence.
func DueAtForCompetence(refYear, refMonth int) time.Time {
	y, m := refYear, refMonth+1
	if m > 12 {
		m = 1
		y++
	}
	return time.Date(y, time.Month(m), 10, 23, 59, 59, 0, saoPaulo)
}

// IsMonthlyClosingDay is true on the 1st calendar day in America/Sao_Paulo.
func IsMonthlyClosingDay(now time.Time) bool {
	now = now.In(saoPaulo)
	return now.Day() == 1
}
