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
		Payer: buildPayer(g.cfg, in.PayerEmail),
	}
	var resp orderResponse
	status, err := g.client.postJSON(ctx, "/v1/orders", in.IdempotencyKey, body, &resp)
	if err != nil {
		switch status {
		case 401:
			return ChargeResult{}, httpStatusFromCall(err,
				"Mercado Pago recusou o Access Token (HTTP 401). Confira a credencial, o header Authorization e a aplicação.")
		case 403:
			return ChargeResult{}, httpStatusFromCall(err,
				"Mercado Pago recusou o acesso à API de Orders (HTTP 403). Consulte código e mensagem retornados pelo provedor.")
		case 429:
			return ChargeResult{}, fmt.Errorf("Mercado Pago indisponível (limite de requisições); tente novamente")
		default:
			return ChargeResult{}, err
		}
	}
	return chargeFromOrderResponse(resp, in.AmountCents, expiration)
}
