package payments

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/store-platform/store/internal/audit"
	"github.com/store-platform/store/internal/jobs"
)

func (s *Service) auditLog(ctx context.Context, action, entityType string, entityID uuid.UUID, meta map[string]any) {
	if s.audit == nil {
		return
	}
	_ = s.audit.Log(ctx, nil, action, entityType, &entityID, nil, meta)
}

func (s *Service) recordSettlementAudit(ctx context.Context, settled settleResult) {
	if s.audit == nil {
		return
	}
	var invStatus string
	_ = s.pool.QueryRow(ctx, `SELECT status FROM invoices WHERE id = $1`, settled.invoiceID).Scan(&invStatus)
	if settled.installmentID != nil {
		s.auditLog(ctx, "INSTALLMENT_PAID", "invoice_installment", *settled.installmentID, map[string]any{
			"invoice_id":   settled.invoiceID.String(),
			"amount_cents": settled.amountCents,
		})
	}
	switch invStatus {
	case "paid":
		s.auditLog(ctx, "INVOICE_FULLY_PAID", "invoice", settled.invoiceID, map[string]any{
			"amount_cents": settled.amountCents,
		})
	case "partially_paid":
		s.auditLog(ctx, "INVOICE_PARTIALLY_PAID", "invoice", settled.invoiceID, map[string]any{
			"amount_cents": settled.amountCents,
		})
	}
}

// ServiceDeps optional dependencies for payments.Service.
type ServiceDeps struct {
	Jobs         *jobs.Repository
	OrderFetcher OrderFetcher
	Audit        *audit.Service
	AppEnv       string
}

func (s *Service) enqueueMercadoPagoOrderJob(ctx context.Context, paymentEventID uuid.UUID, orderID string) {
	s.enqueueMercadoPagoOrderJobAt(ctx, paymentEventID, orderID, time.Now())
}

func (s *Service) enqueueMercadoPagoOrderJobAt(ctx context.Context, paymentEventID uuid.UUID, orderID string, availableAt time.Time) {
	if s.jobs == nil || orderID == "" {
		return
	}
	payload := map[string]any{"order_id": orderID}
	if paymentEventID != uuid.Nil {
		payload["payment_event_id"] = paymentEventID.String()
	}
	_ = s.jobs.EnqueueAvailableAt(ctx, jobs.TypeMercadoPagoOrder, payload, availableAt)
}

func (s *Service) scheduleMercadoPagoSettlementAfterPixCharge(ctx context.Context, mpOrderID string) {
	if s.provider != ProviderMercadoPago || !s.mpTestAutoApprove || mpOrderID == "" {
		return
	}
	// APRO: MP costuma accreditar em poucos segundos; reconcilia sem depender só do webhook (túnel/assinatura).
	s.enqueueMercadoPagoOrderJobAt(ctx, uuid.Nil, mpOrderID, time.Now().Add(12*time.Second))
	s.log.Info("mercado pago settlement sync scheduled",
		slog.String("order_id", mpOrderID),
		slog.Duration("delay", 12*time.Second),
	)
}
