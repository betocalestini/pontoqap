package mercadopago

import "fmt"

// CentsToDecimalString converts integer cents to a decimal string for the Mercado Pago API.
func CentsToDecimalString(cents int64) string {
	if cents < 0 {
		cents = -cents
	}
	return fmt.Sprintf("%d.%02d", cents/100, cents%100)
}
