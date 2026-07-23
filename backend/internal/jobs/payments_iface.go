package jobs

import (
	"context"

	"github.com/google/uuid"
)

// MercadoPagoPayments processes Mercado Pago order sync jobs.
type MercadoPagoPayments interface {
	ProcessMercadoPagoOrderJob(ctx context.Context, paymentEventID uuid.UUID, orderID string) error
}
