package mercadopago

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	mpwebhook "github.com/mercadopago/sdk-go/pkg/webhook"

	"github.com/google/uuid"
)

const ProviderName = "mercadopago"

type PixChargeInput struct {
	InvoiceID         uuid.UUID
	AmountCents       int64
	ExternalReference string
	PayerEmail        string
	IdempotencyKey    string
	ExpirationISO     string
}

type ChargeResult struct {
	ExternalID  string
	TxID        string
	QRCodeText  string
	ExpiresAt   time.Time
	AmountCents int64
}

type Gateway struct {
	cfg    Config
	client *apiClient
}

func NewGateway(cfg Config, log *slog.Logger) (*Gateway, error) {
	if cfg.AccessToken == "" {
		return nil, fmt.Errorf("mercado pago: access token vazio")
	}
	return &Gateway{
		cfg:    cfg,
		client: newAPIClient(cfg, log),
	}, nil
}

func (g *Gateway) CreatePixCharge(ctx context.Context, in PixChargeInput) (ChargeResult, error) {
	if in.PayerEmail == "" {
		return ChargeResult{}, fmt.Errorf("e-mail do pagador é obrigatório para Pix")
	}
	if in.ExternalReference == "" {
		return ChargeResult{}, fmt.Errorf("referência externa é obrigatória")
	}
	if in.IdempotencyKey == "" {
		return ChargeResult{}, fmt.Errorf("chave de idempotência é obrigatória")
	}
	return g.createOrder(ctx, in)
}

func (g *Gateway) VerifyWebhookSignature(payload []byte, signature string) bool {
	return false
}

func (g *Gateway) ParseWebhookEvent(payload []byte) (string, string, string, int64, error) {
	return "", "", "", 0, errors.New("use webhook Mercado Pago dedicado")
}

// ValidateOrderWebhook checks x-signature from Mercado Pago Order notifications.
func (g *Gateway) ValidateOrderWebhook(xSignature, xRequestID, dataID string) error {
	if g.cfg.WebhookSecret == "" {
		return fmt.Errorf("MERCADO_PAGO_WEBHOOK_SECRET não configurado")
	}
	return mpwebhook.ValidateSignature(xSignature, xRequestID, dataID, g.cfg.WebhookSecret)
}

// OrderWebhookPayload is the minimal notification body for Order topic.
type OrderWebhookPayload struct {
	Type   string `json:"type"`
	Action string `json:"action"`
	Data   struct {
		ID string `json:"id"`
	} `json:"data"`
}

func ParseOrderWebhookPayload(body []byte) (OrderWebhookPayload, error) {
	var p OrderWebhookPayload
	if err := json.Unmarshal(body, &p); err != nil {
		return OrderWebhookPayload{}, err
	}
	if p.Type != "order" {
		return OrderWebhookPayload{}, fmt.Errorf("tipo de notificação inválido")
	}
	if p.Data.ID == "" {
		return OrderWebhookPayload{}, fmt.Errorf("data.id ausente")
	}
	return p, nil
}
