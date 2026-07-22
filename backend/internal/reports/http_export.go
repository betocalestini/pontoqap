package reports

import (
	"net/http"
	"time"

	"github.com/google/uuid"
)

func ParseYearMonthHTTP(r *http.Request) (int, int) {
	return parseYearMonth(r)
}

func ParseDateRangeHTTP(r *http.Request) DateRange {
	return parseDateRange(r)
}

func ParsePageHTTP(r *http.Request) PageFilter {
	return parsePage(r)
}

func ParseOptionalUUIDHTTP(q string) *uuid.UUID {
	return parseOptionalUUID(q)
}

func Int64QueryHTTP(r *http.Request, key string) *int64 {
	return int64Query(r, key)
}

func BoolQueryHTTP(r *http.Request, key string) *bool {
	return boolQuery(r, key)
}

func FormatTimeHTTP(t *time.Time) string {
	return formatTime(t)
}

func WriteCSV(w http.ResponseWriter, filename string, header []string, rows [][]string) {
	writeCSV(w, filename, header, rows)
}
