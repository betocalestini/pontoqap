package payments

import (
	"context"
	"log/slog"

	"github.com/store-platform/store/internal/payments/mercadopago"
	"github.com/store-platform/store/internal/platform/config"
)

type mercadoPagoGateway struct {
	inner *mercadopago.Gateway
}

func (a *mercadoPagoGateway) CreatePixCharge(ctx context.Context, in PixChargeInput) (ChargeResult, error) {
	res, err := a.inner.CreatePixCharge(ctx, mercadopago.PixChargeInput{
		InvoiceID:         in.InvoiceID,
		AmountCents:       in.AmountCents,
		ExternalReference: in.ExternalReference,
		PayerEmail:        in.PayerEmail,
		IdempotencyKey:    in.IdempotencyKey,
		ExpirationISO:     in.ExpirationISO,
	})
	if err != nil {
		return ChargeResult{}, err
	}
	return ChargeResult{
		Provider:    ProviderMercadoPago,
		ExternalID:  res.ExternalID,
		TxID:        res.TxID,
		QRCodeText:  res.QRCodeText,
		ExpiresAt:   res.ExpiresAt,
		AmountCents: res.AmountCents,
	}, nil
}

func (a *mercadoPagoGateway) VerifyWebhookSignature(payload []byte, signature string) bool {
	return a.inner.VerifyWebhookSignature(payload, signature)
}

func (a *mercadoPagoGateway) ParseWebhookEvent(payload []byte) (string, string, string, int64, error) {
	return a.inner.ParseWebhookEvent(payload)
}

func newMercadoPagoGateway(cfg config.PaymentsConfig, log *slog.Logger) (Gateway, error) {
	gw, err := mercadopago.NewGateway(mercadopago.ConfigFromPayments(cfg), log)
	if err != nil {
		return nil, err
	}
	return &mercadoPagoGateway{inner: gw}, nil
}
