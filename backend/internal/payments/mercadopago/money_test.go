package mercadopago_test

import (
	"testing"

	"github.com/store-platform/store/internal/payments/mercadopago"
)

func TestCentsToDecimalString(t *testing.T) {
	tests := []struct {
		cents int64
		want  string
	}{
		{100, "1.00"},
		{1990, "19.90"},
		{35050, "350.50"},
		{0, "0.00"},
	}
	for _, tc := range tests {
		if got := mercadopago.CentsToDecimalString(tc.cents); got != tc.want {
			t.Fatalf("CentsToDecimalString(%d) = %q, want %q", tc.cents, got, tc.want)
		}
	}
}
