package payments

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type PixChargeInput struct {
	InvoiceID         uuid.UUID
	AmountCents       int64
	ExternalReference string
	PayerEmail        string
	IdempotencyKey    string
	ExpirationISO     string
}

type ChargeResult struct {
	ChargeID    uuid.UUID `json:"charge_id"`
	Provider    string    `json:"provider"`
	ExternalID  string    `json:"external_id"`
	TxID        string    `json:"txid"`
	QRCodeText  string    `json:"qr_code_text"`
	ExpiresAt   time.Time `json:"expires_at"`
	AmountCents int64     `json:"amount_cents"`
}

type Gateway interface {
	CreatePixCharge(ctx context.Context, in PixChargeInput) (ChargeResult, error)
	VerifyWebhookSignature(payload []byte, signature string) bool
	ParseWebhookEvent(payload []byte) (externalEventID, eventType string, externalPaymentID string, amountCents int64, err error)
}
