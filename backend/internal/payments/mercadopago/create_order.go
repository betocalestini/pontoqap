package mercadopago

import (
	"context"
	"fmt"
)

func (g *Gateway) createOrder(ctx context.Context, in PixChargeInput) (ChargeResult, error) {
	amount := CentsToDecimalString(in.AmountCents)
	expiration := in.ExpirationISO
	if expiration == "" {
		expiration = g.cfg.PixExpiration
	}
	body := createOrderRequest{
		Type:              "online",
		TotalAmount:       amount,
		ExternalReference: in.ExternalReference,
		ProcessingMode:    "automatic",
		Transactions: transactionPayload{
			Payments: []paymentPayload{{
				Amount:         amount,
				ExpirationTime: expiration,
				PaymentMethod: paymentMethodPayload{
					ID:   "pix",
					Type: "bank_transfer",
				},
			}},
		},
		Payer: payerPayload{Email: in.PayerEmail},
	}
	var resp orderResponse
	status, err := g.client.postJSON(ctx, "/v1/orders", in.IdempotencyKey, body, &resp)
	if err != nil {
		if status == 401 || status == 403 {
			return ChargeResult{}, fmt.Errorf("credenciais Mercado Pago inválidas")
		}
		if status == 429 {
			return ChargeResult{}, fmt.Errorf("Mercado Pago indisponível (limite de requisições); tente novamente")
		}
		return ChargeResult{}, err
	}
	return chargeFromOrderResponse(resp, in.AmountCents, expiration)
}
