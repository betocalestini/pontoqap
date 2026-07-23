package mercadopago

import "strings"

const testAutoApproveEmail = "test_user_br@testuser.com"

func buildPayer(cfg Config, customerEmail string) payerPayload {
	payer := payerPayload{Email: customerEmail}
	if cfg.TestAutoApprove && strings.EqualFold(strings.TrimSpace(cfg.Environment), "test") {
		payer.Email = testAutoApproveEmail
		payer.FirstName = "APRO"
	}
	return payer
}

// BuildPayerForTest exposes payer logic for unit tests.
func BuildPayerForTest(cfg Config, customerEmail string) payerPayload {
	return buildPayer(cfg, customerEmail)
}
