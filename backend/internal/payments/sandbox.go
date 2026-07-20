package payments

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type SandboxGateway struct {
	Secret string
}

func NewSandboxGateway(secret string) *SandboxGateway {
	return &SandboxGateway{Secret: secret}
}

func (g *SandboxGateway) CreatePixCharge(ctx context.Context, invoiceID uuid.UUID, amountCents int64, description string) (ChargeResult, error) {
	extID := uuid.NewString()
	txid := fmt.Sprintf("SBX%s", uuid.NewString()[:8])
	exp := time.Now().Add(24 * time.Hour)
	qr := fmt.Sprintf("00020126580014br.gov.bcb.pix0136%s520400005303986540%d5802BR5925STORE PLATFORM6009SAO PAULO62070503***6304ABCD",
		txid, amountCents/100)
	return ChargeResult{
		ChargeID:    uuid.New(),
		Provider:    "sandbox",
		ExternalID:  extID,
		TxID:        txid,
		QRCodeText:  qr,
		ExpiresAt:   exp,
		AmountCents: amountCents,
	}, nil
}

func (g *SandboxGateway) VerifyWebhookSignature(payload []byte, signature string) bool {
	mac := hmac.New(sha256.New, []byte(g.Secret))
	mac.Write(payload)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}

type sandboxWebhook struct {
	EventID       string `json:"event_id"`
	EventType     string `json:"event_type"`
	PaymentID     string `json:"payment_id"`
	AmountCents   int64  `json:"amount_cents"`
}

func (g *SandboxGateway) ParseWebhookEvent(payload []byte) (string, string, string, int64, error) {
	var ev sandboxWebhook
	if err := json.Unmarshal(payload, &ev); err != nil {
		return "", "", "", 0, err
	}
	return ev.EventID, ev.EventType, ev.PaymentID, ev.AmountCents, nil
}
