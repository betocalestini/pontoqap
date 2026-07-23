package payments

import (
	"context"
	"log/slog"

	"github.com/store-platform/store/internal/payments/mercadopago"
	"github.com/store-platform/store/internal/platform/config"
)

// OrderFetcher loads Order state from Mercado Pago (GET /v1/orders/{id}).
type OrderFetcher interface {
	FetchOrder(ctx context.Context, orderID string) (mercadopago.OrderDetail, error)
}

// NewMercadoPagoOrderFetcher returns a fetcher when MP credentials are configured.
func NewMercadoPagoOrderFetcher(cfg config.PaymentsConfig, log *slog.Logger) (OrderFetcher, error) {
	if cfg.MercadoPago.AccessToken == "" {
		return nil, nil
	}
	return mercadopago.NewGateway(mercadopago.ConfigFromPayments(cfg), log)
}
