package payments

import (
	"fmt"
	"log/slog"

	"github.com/store-platform/store/internal/platform/config"
)

const ProviderMercadoPago = "mercadopago"

func NewGateway(cfg config.PaymentsConfig, log *slog.Logger) (Gateway, error) {
	switch config.NormalizePaymentProvider(cfg.Provider) {
	case "mercadopago":
		return newMercadoPagoGateway(cfg, log)
	case "sandbox", "":
		return NewSandboxGateway(cfg.WebhookSecret), nil
	default:
		return nil, fmt.Errorf("unsupported PAYMENT_PROVIDER: %q", cfg.Provider)
	}
}

func ProviderName(cfg config.PaymentsConfig) string {
	if config.IsMercadoPagoProvider(cfg.Provider) {
		return ProviderMercadoPago
	}
	return "sandbox"
}
