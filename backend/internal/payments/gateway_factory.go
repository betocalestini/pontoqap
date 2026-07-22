package payments

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/store-platform/store/internal/platform/config"
)

const ProviderMercadoPago = "mercadopago"

func NewGateway(cfg config.PaymentsConfig, log *slog.Logger) (Gateway, error) {
	switch strings.ToLower(strings.TrimSpace(cfg.Provider)) {
	case "mercadopago":
		return newMercadoPagoGateway(cfg, log)
	case "sandbox", "":
		return NewSandboxGateway(cfg.WebhookSecret), nil
	default:
		return nil, fmt.Errorf("unsupported PAYMENT_PROVIDER: %q", cfg.Provider)
	}
}

func ProviderName(cfg config.PaymentsConfig) string {
	if strings.EqualFold(cfg.Provider, "mercadopago") {
		return ProviderMercadoPago
	}
	return "sandbox"
}
