package mercadopago

import (
	"fmt"
	"strconv"
	"strings"
)

// DecimalStringToCents parses Mercado Pago decimal amounts (e.g. "350.50").
func DecimalStringToCents(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty amount")
	}
	parts := strings.Split(s, ".")
	if len(parts) > 2 {
		return 0, fmt.Errorf("invalid amount %q", s)
	}
	whole, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, err
	}
	var frac int64
	if len(parts) == 2 {
		fracStr := parts[1]
		if len(fracStr) > 2 {
			fracStr = fracStr[:2]
		}
		for len(fracStr) < 2 {
			fracStr += "0"
		}
		frac, err = strconv.ParseInt(fracStr, 10, 64)
		if err != nil {
			return 0, err
		}
	}
	if whole < 0 || (whole == 0 && strings.HasPrefix(strings.TrimSpace(s), "-")) {
		return 0, fmt.Errorf("negative amount")
	}
	return whole*100 + frac, nil
}

// CentsToDecimalString converts integer cents to a decimal string for the Mercado Pago API.
func CentsToDecimalString(cents int64) string {
	if cents < 0 {
		cents = -cents
	}
	return fmt.Sprintf("%d.%02d", cents/100, cents%100)
}
