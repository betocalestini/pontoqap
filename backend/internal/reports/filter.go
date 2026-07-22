package reports

import (
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
)

var storeLocation *time.Location

func init() {
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		loc = time.UTC
	}
	storeLocation = loc
}

type DateRange struct {
	From time.Time // inclusive, UTC
	To   time.Time // exclusive, UTC
}

type PageFilter struct {
	Limit  int
	Offset int
}

func parsePage(r *http.Request) PageFilter {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if limit <= 0 || limit > 500 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	return PageFilter{Limit: limit, Offset: offset}
}

func parseYearMonth(r *http.Request) (year, month int) {
	y, _ := strconv.Atoi(r.URL.Query().Get("year"))
	m, _ := strconv.Atoi(r.URL.Query().Get("month"))
	if y == 0 {
		now := time.Now().In(storeLocation)
		y, m = now.Year(), int(now.Month())
	}
	if m == 0 {
		m = int(time.Now().In(storeLocation).Month())
	}
	return y, m
}

func monthRangeUTC(year, month int) DateRange {
	start := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, storeLocation)
	end := start.AddDate(0, 1, 0)
	return DateRange{From: start.UTC(), To: end.UTC()}
}

func parseDateRange(r *http.Request) DateRange {
	if fromS := r.URL.Query().Get("date_from"); fromS != "" {
		if toS := r.URL.Query().Get("date_to"); toS != "" {
			from, err1 := time.ParseInLocation("2006-01-02", fromS, storeLocation)
			to, err2 := time.ParseInLocation("2006-01-02", toS, storeLocation)
			if err1 == nil && err2 == nil {
				return DateRange{
					From: from.UTC(),
					To:   to.AddDate(0, 0, 1).UTC(),
				}
			}
		}
	}
	y, m := parseYearMonth(r)
	return monthRangeUTC(y, m)
}

func parseOptionalUUID(q string) *uuid.UUID {
	q = trim(q)
	if q == "" {
		return nil
	}
	id, err := uuid.Parse(q)
	if err != nil {
		return nil
	}
	return &id
}

func trim(s string) string {
	for len(s) > 0 && (s[0] == ' ' || s[0] == '\t') {
		s = s[1:]
	}
	for len(s) > 0 && (s[len(s)-1] == ' ' || s[len(s)-1] == '\t') {
		s = s[:len(s)-1]
	}
	return s
}
