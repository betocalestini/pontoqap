package payments

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	mpwebhook "github.com/mercadopago/sdk-go/pkg/webhook"

	"github.com/store-platform/store/internal/billing"
	"github.com/store-platform/store/internal/jobs"
	"github.com/store-platform/store/internal/payments/mercadopago"
	"github.com/store-platform/store/internal/platform/config"
	"github.com/store-platform/store/internal/audit"
)

type Service struct {
	pool              *pgxpool.Pool
	gateway           Gateway
	billing           *billing.Service
	jobs              *jobs.Repository
	orderFetcher      OrderFetcher
	audit             *audit.Service
	provider          string
	appEnv            string
	webhookKey        string
	mpWebhookSecret   string
	mpApplicationID   string
	pixExpiration     string
	mpTestAutoApprove bool
	log               *slog.Logger
}

func NewService(pool *pgxpool.Pool, gateway Gateway, bill *billing.Service, payCfg config.PaymentsConfig, log *slog.Logger, deps *ServiceDeps) *Service {
	if log == nil {
		log = slog.Default()
	}
	s := &Service{
		pool:            pool,
		gateway:         gateway,
		billing:         bill,
		provider:        ProviderName(payCfg),
		webhookKey:      payCfg.WebhookSecret,
		mpWebhookSecret: payCfg.MercadoPago.WebhookSecret,
		mpApplicationID: strings.TrimSpace(payCfg.MercadoPago.ApplicationID),
		pixExpiration:   payCfg.MercadoPago.PixExpiration,
		mpTestAutoApprove: payCfg.MercadoPago.TestAutoApprove &&
			strings.EqualFold(payCfg.MercadoPago.Environment, "test"),
		log: log,
	}
	if deps != nil {
		s.jobs = deps.Jobs
		s.orderFetcher = deps.OrderFetcher
		s.audit = deps.Audit
		s.appEnv = deps.AppEnv
	}
	return s
}

var ErrSimulateNotSupported = errors.New("simulação Pix disponível apenas para PAYMENT_PROVIDER=sandbox")

type Charge struct {
	ID            uuid.UUID  `json:"id"`
	InvoiceID     uuid.UUID  `json:"invoice_id"`
	InstallmentID *uuid.UUID `json:"installment_id,omitempty"`
	Provider      string     `json:"provider"`
	Simulatable   bool       `json:"simulatable"`
	Status        string     `json:"status"`
	AmountCents   int64      `json:"amount_cents"`
	QRCodeText    string     `json:"qr_code_text"`
	QRCodeBase64  string     `json:"qr_code_base64,omitempty"`
	TicketURL     string     `json:"ticket_url,omitempty"`
	ExpiresAt     time.Time  `json:"expires_at"`
}

func (s *Service) withChargeMeta(c *Charge) *Charge {
	if c == nil {
		return nil
	}
	c.Provider = s.provider
	c.Simulatable = s.appEnv == "development" && s.provider == "sandbox"
	return c
}

var ErrInvalidWebhookSignature = errors.New("invalid webhook signature")

func (s *Service) CreateOrReusePixCharge(ctx context.Context, invoiceID uuid.UUID) (*Charge, error) {
	requires, err := s.billing.InvoiceRequiresInstallmentPix(ctx, invoiceID)
	if err != nil {
		return nil, err
	}
	if requires {
		return nil, errors.New("selecione o plano de pagamento e gere Pix pela parcela em aberto")
	}
	inv, err := s.billing.GetInvoice(ctx, invoiceID)
	if err != nil || inv == nil {
		return nil, fmt.Errorf("invoice not found")
	}
	remaining := inv.RemainingCents()
	if remaining <= 0 {
		return nil, errors.New("invoice already settled")
	}

	var existing Charge
	err = s.pool.QueryRow(ctx, `
		SELECT id, invoice_id, status, amount_cents, COALESCE(qr_code_text,''), expires_at
		FROM payment_charges
		WHERE invoice_id = $1 AND provider = $2 AND installment_id IS NULL AND status = 'pending' AND expires_at > NOW()
		ORDER BY created_at DESC LIMIT 1
	`, invoiceID, s.provider).Scan(&existing.ID, &existing.InvoiceID, &existing.Status, &existing.AmountCents,
		&existing.QRCodeText, &existing.ExpiresAt)
	if err == nil {
		s.log.Info("pix charge reused",
			slog.String("invoice_id", invoiceID.String()),
			slog.String("charge_id", existing.ID.String()),
		)
		return s.withChargeMeta(&existing), nil
	}
	if err != pgx.ErrNoRows {
		return nil, err
	}
	if err := s.expireStalePendingPixChargesForInvoice(ctx, invoiceID); err != nil {
		return nil, err
	}

	email, err := s.payerEmailForInvoice(ctx, invoiceID)
	if err != nil {
		return nil, err
	}

	chargeID := uuid.New()
	res, err := s.gateway.CreatePixCharge(ctx, PixChargeInput{
		InvoiceID:         invoiceID,
		AmountCents:       remaining,
		ExternalReference: inv.InvoiceNumber,
		PayerEmail:        email,
		IdempotencyKey:    fmt.Sprintf("pix-invoice-%s-%s", invoiceID.String(), chargeID.String()),
		ExpirationISO:     s.pixExpiration,
	})
	if err != nil {
		s.logPixChargeGatewayFailed(err,
			slog.String("invoice_id", invoiceID.String()),
		)
		return nil, err
	}
	_, err = s.pool.Exec(ctx, `
		INSERT INTO payment_charges (
			id, invoice_id, provider, external_id, txid, status, amount_cents, qr_code_text, expires_at
		) VALUES ($1,$2,$3,$4,$5,'pending',$6,$7,$8)
	`, chargeID, invoiceID, s.provider, res.ExternalID, res.TxID, res.AmountCents, res.QRCodeText, res.ExpiresAt)
	if err != nil {
		return nil, err
	}
	s.log.Info("pix charge created",
		slog.String("provider", s.provider),
		slog.String("invoice_id", invoiceID.String()),
		slog.String("charge_id", chargeID.String()),
		slog.String("external_id", res.ExternalID),
	)
	s.auditLog(ctx, "PIX_CHARGE_CREATED", "payment_charge", chargeID, map[string]any{
		"invoice_id":  invoiceID.String(),
		"provider":    s.provider,
		"amount_cents": res.AmountCents,
	})
	s.scheduleMercadoPagoSettlementAfterPixCharge(ctx, res.ExternalID)
	return s.withChargeMeta(&Charge{
		ID: chargeID, InvoiceID: invoiceID, Status: "pending",
		AmountCents: res.AmountCents, QRCodeText: res.QRCodeText,
		QRCodeBase64: res.QRCodeBase64, TicketURL: res.TicketURL, ExpiresAt: res.ExpiresAt,
	}), nil
}

func (s *Service) CreateOrReusePixChargeForInstallment(ctx context.Context, installmentID, customerID uuid.UUID) (*Charge, error) {
	inst, err := s.billing.GetInstallmentForCustomer(ctx, installmentID, customerID)
	if err != nil || inst == nil {
		return nil, fmt.Errorf("parcela não encontrada")
	}
	if err := s.billing.ValidateInstallmentForPix(ctx, installmentID); err != nil {
		return nil, err
	}
	inv, err := s.billing.GetInvoice(ctx, inst.InvoiceID)
	if err != nil || inv == nil {
		return nil, fmt.Errorf("invoice not found")
	}
	if inv.RemainingCents() <= 0 {
		return nil, errors.New("invoice already settled")
	}

	if charge, err := s.reusePendingPixChargeForInstallment(ctx, installmentID); err != nil {
		return nil, err
	} else if charge != nil {
		return charge, nil
	}
	if err := s.expireStalePendingPixChargesForInstallment(ctx, installmentID); err != nil {
		return nil, err
	}

	email, err := s.payerEmailForInvoice(ctx, inst.InvoiceID)
	if err != nil {
		return nil, err
	}

	amount := inst.RemainingCents
	extRef := fmt.Sprintf("INSTALLMENT-%s", installmentID.String())
	chargeID := uuid.New()
	res, err := s.gateway.CreatePixCharge(ctx, PixChargeInput{
		InvoiceID:         inst.InvoiceID,
		AmountCents:       amount,
		ExternalReference: extRef,
		PayerEmail:        email,
		IdempotencyKey:    fmt.Sprintf("pix-installment-%s-%s", installmentID.String(), chargeID.String()),
		ExpirationISO:     s.pixExpiration,
	})
	if err != nil {
		s.logPixChargeGatewayFailed(err,
			slog.String("invoice_id", inst.InvoiceID.String()),
			slog.String("installment_id", installmentID.String()),
		)
		return nil, err
	}
	_, err = s.pool.Exec(ctx, `
		INSERT INTO payment_charges (
			id, invoice_id, installment_id, provider, external_id, txid, status, amount_cents, qr_code_text, expires_at
		) VALUES ($1,$2,$3,$4,$5,$6,'pending',$7,$8,$9)
	`, chargeID, inst.InvoiceID, installmentID, s.provider, res.ExternalID, res.TxID, res.AmountCents, res.QRCodeText, res.ExpiresAt)
	if err != nil {
		return nil, err
	}
	_ = s.billing.MarkInstallmentPixActive(ctx, installmentID)
	instID := installmentID
	s.log.Info("pix charge created",
		slog.String("provider", s.provider),
		slog.String("invoice_id", inst.InvoiceID.String()),
		slog.String("installment_id", installmentID.String()),
		slog.String("charge_id", chargeID.String()),
		slog.String("external_id", res.ExternalID),
	)
	s.auditLog(ctx, "PIX_CHARGE_CREATED", "payment_charge", chargeID, map[string]any{
		"invoice_id":     inst.InvoiceID.String(),
		"installment_id": installmentID.String(),
		"provider":       s.provider,
		"amount_cents":   res.AmountCents,
	})
	s.scheduleMercadoPagoSettlementAfterPixCharge(ctx, res.ExternalID)
	return s.withChargeMeta(&Charge{
		ID: chargeID, InvoiceID: inst.InvoiceID, InstallmentID: &instID, Status: "pending",
		AmountCents: res.AmountCents, QRCodeText: res.QRCodeText,
		QRCodeBase64: res.QRCodeBase64, TicketURL: res.TicketURL, ExpiresAt: res.ExpiresAt,
	}), nil
}

var ErrPixChargeNotFound = errors.New("no active pix charge")

func (s *Service) GetPendingPixChargeForInstallment(ctx context.Context, installmentID, customerID uuid.UUID) (*Charge, error) {
	inst, err := s.billing.GetInstallmentForCustomer(ctx, installmentID, customerID)
	if err != nil || inst == nil {
		return nil, fmt.Errorf("parcela não encontrada")
	}
	if err := s.billing.ValidateInstallmentForPix(ctx, installmentID); err != nil {
		return nil, err
	}
	charge, err := s.reusePendingPixChargeForInstallment(ctx, installmentID)
	if err != nil {
		return nil, err
	}
	if charge == nil {
		return nil, ErrPixChargeNotFound
	}
	return charge, nil
}

func (s *Service) reusePendingPixChargeForInstallment(ctx context.Context, installmentID uuid.UUID) (*Charge, error) {
	var existing Charge
	var existingInstID *uuid.UUID
	err := s.pool.QueryRow(ctx, `
		SELECT id, invoice_id, installment_id, status, amount_cents, COALESCE(qr_code_text,''), expires_at
		FROM payment_charges
		WHERE installment_id = $1 AND provider = $2 AND status = 'pending' AND expires_at > NOW()
		ORDER BY created_at DESC LIMIT 1
	`, installmentID, s.provider).Scan(&existing.ID, &existing.InvoiceID, &existingInstID, &existing.Status, &existing.AmountCents,
		&existing.QRCodeText, &existing.ExpiresAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	existing.InstallmentID = existingInstID
	s.log.Info("pix charge reused",
		slog.String("invoice_id", existing.InvoiceID.String()),
		slog.String("installment_id", installmentID.String()),
		slog.String("charge_id", existing.ID.String()),
	)
	return s.withChargeMeta(&existing), nil
}

func (s *Service) expireStalePendingPixChargesForInstallment(ctx context.Context, installmentID uuid.UUID) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE payment_charges SET status = 'expired', updated_at = NOW()
		WHERE installment_id = $1 AND provider = $2 AND status = 'pending' AND expires_at <= NOW()
	`, installmentID, s.provider)
	return err
}

func (s *Service) expireStalePendingPixChargesForInvoice(ctx context.Context, invoiceID uuid.UUID) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE payment_charges SET status = 'expired', updated_at = NOW()
		WHERE invoice_id = $1 AND provider = $2 AND installment_id IS NULL AND status = 'pending' AND expires_at <= NOW()
	`, invoiceID, s.provider)
	return err
}

// SandboxWebhookResult descreve o processamento do webhook sandbox Pix.
type SandboxWebhookResult struct {
	Inserted       bool
	Duplicate      bool
	Ignored        bool
	AmountMismatch bool
	Settled        bool
	ExternalEventID string
	EventType      string
	InvoiceID      uuid.UUID
	InstallmentID  *uuid.UUID
	ChargeID       uuid.UUID
	AmountCents    int64
}

func (s *Service) ProcessWebhook(ctx context.Context, payload []byte, signature string) (SandboxWebhookResult, error) {
	var out SandboxWebhookResult
	if !s.gateway.VerifyWebhookSignature(payload, signature) {
		return out, fmt.Errorf("%w", ErrInvalidWebhookSignature)
	}
	eventID, eventType, paymentExtID, amountCents, err := s.gateway.ParseWebhookEvent(payload)
	if err != nil {
		return out, err
	}
	out.ExternalEventID = eventID
	out.EventType = eventType
	if eventType != "payment.confirmed" {
		out.Ignored = true
		return out, nil
	}

	hash := sha256.Sum256(payload)
	hashStr := hex.EncodeToString(hash[:])

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return out, err
	}
	defer tx.Rollback(ctx)

	var peID uuid.UUID
	err = tx.QueryRow(ctx, `
		INSERT INTO payment_events (provider, external_event_id, event_type, payload_hash, processed)
		VALUES ('sandbox', $1, $2, $3, FALSE)
		ON CONFLICT (provider, external_event_id) DO NOTHING
		RETURNING id
	`, eventID, eventType, hashStr).Scan(&peID)
	if err == pgx.ErrNoRows {
		out.Duplicate = true
		return out, nil
	}
	if err != nil {
		return out, err
	}
	out.Inserted = true

	var chargeID, invoiceID uuid.UUID
	var installmentID *uuid.UUID
	var chargeAmount int64
	err = tx.QueryRow(ctx, `
		SELECT id, invoice_id, installment_id, amount_cents FROM payment_charges
		WHERE provider = 'sandbox' AND external_id = $1 FOR UPDATE
	`, paymentExtID).Scan(&chargeID, &invoiceID, &installmentID, &chargeAmount)
	if err != nil {
		return out, err
	}
	out.ChargeID = chargeID
	out.InvoiceID = invoiceID
	out.InstallmentID = installmentID
	out.AmountCents = amountCents

	if amountCents != chargeAmount {
		_, _ = tx.Exec(ctx, `UPDATE payment_events SET error_message = 'amount mismatch', processed = TRUE WHERE id = $1`, peID)
		out.AmountMismatch = true
		s.log.Warn("sandbox pix webhook amount mismatch",
			slog.String("invoice_id", invoiceID.String()),
			slog.String("charge_id", chargeID.String()),
			slog.Int64("expected_cents", chargeAmount),
			slog.Int64("received_cents", amountCents),
		)
		return out, errors.New("payment amount mismatch")
	}

	settled, settleErr := s.settleChargeTx(ctx, tx, chargeID, "sandbox", paymentExtID, amountCents)
	if settleErr != nil {
		if errors.Is(settleErr, ErrChargeAlreadyPaid) {
			out.Settled = false
		} else {
			return out, settleErr
		}
	} else {
		out.Settled = settled.invoicePaid
	}

	_, err = tx.Exec(ctx, `UPDATE payment_events SET processed = TRUE, processed_at = NOW() WHERE id = $1`, peID)
	if err != nil {
		return out, err
	}
	if err := tx.Commit(ctx); err != nil {
		return out, err
	}
	if settleErr == nil {
		s.recordSettlementAudit(ctx, settled)
	}
	if out.Settled {
		attrs := []any{
			slog.String("invoice_id", invoiceID.String()),
			slog.String("charge_id", chargeID.String()),
			slog.Int64("amount_cents", amountCents),
		}
		if installmentID != nil {
			attrs = append(attrs, slog.String("installment_id", installmentID.String()))
		}
		s.log.Info("sandbox pix payment settled", attrs...)
	}
	return out, nil
}

func (s *Service) SimulateSandboxPayment(ctx context.Context, chargeID uuid.UUID) error {
	var extID string
	var amount int64
	var provider string
	err := s.pool.QueryRow(ctx, `SELECT external_id, amount_cents, provider FROM payment_charges WHERE id = $1`, chargeID).Scan(&extID, &amount, &provider)
	if err != nil {
		return err
	}
	if provider == ProviderMercadoPago {
		return fmt.Errorf("%w; aguarde webhook/worker (APRO) ou use POST /api/v1/admin/payment-charges/%s/sync", ErrSimulateNotSupported, chargeID)
	}
	body, _ := json.Marshal(map[string]any{
		"event_id":     uuid.NewString(),
		"event_type":   "payment.confirmed",
		"payment_id":   extID,
		"amount_cents": amount,
	})
	sig := signPayload(body, s.webhookKey)
	_, err = s.ProcessWebhook(ctx, body, sig)
	return err
}

func signPayload(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

// SignPayloadForTest expõe assinatura HMAC para testes de integração.
func SignPayloadForTest(payload []byte, secret string) string {
	return signPayload(payload, secret)
}

func (s *Service) payerEmailForInvoice(ctx context.Context, invoiceID uuid.UUID) (string, error) {
	var email string
	err := s.pool.QueryRow(ctx, `
		SELECT COALESCE(NULLIF(TRIM(u.email), ''), '')
		FROM invoices i
		JOIN customers c ON c.id = i.customer_id
		JOIN users u ON u.id = c.user_id
		WHERE i.id = $1
	`, invoiceID).Scan(&email)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", fmt.Errorf("fatura não encontrada")
		}
		return "", err
	}
	if email == "" {
		return "", errors.New("e-mail do cliente é obrigatório para gerar Pix")
	}
	return email, nil
}

// MercadoPagoWebhookResult descreve o resultado da recepção (sem baixa financeira).
type MercadoPagoWebhookResult struct {
	Inserted  bool
	EventID   string
	OrderID   string
	EventType string
}

// ReceiveMercadoPagoOrderWebhook valida a notificação Order e persiste em payment_events (sem baixa financeira).
func (s *Service) ReceiveMercadoPagoOrderWebhook(ctx context.Context, xSignature, xRequestID, queryDataID string, body []byte) (MercadoPagoWebhookResult, error) {
	var out MercadoPagoWebhookResult
	if s.mpWebhookSecret == "" {
		return out, errors.New("webhook Mercado Pago não configurado")
	}
	dataIDForSig, err := validateMercadoPagoWebhookSignature(xSignature, xRequestID, queryDataID, body, s.mpWebhookSecret)
	if err != nil {
		s.logMercadoPagoWebhookSignatureFailed(err, queryDataID, dataIDForSig, xRequestID, xSignature, body, s.mpWebhookSecret)
		return out, fmt.Errorf("%w: %v", ErrInvalidWebhookSignature, err)
	}
	payload, err := parseMercadoPagoOrderWebhook(body)
	if err != nil {
		return out, err
	}
	dataID := strings.TrimSpace(queryDataID)
	if dataID == "" {
		dataID = payload.Data.ID
	}
	out.OrderID = dataID

	eventID := xRequestID
	if eventID == "" {
		eventID = payload.Data.ID + ":" + payload.Action
	}
	out.EventID = eventID
	eventType := payload.Action
	if eventType == "" {
		eventType = payload.Type
	}
	out.EventType = eventType

	hash := sha256.Sum256(body)
	hashStr := hex.EncodeToString(hash[:])

	var peID uuid.UUID
	err = s.pool.QueryRow(ctx, `
		INSERT INTO payment_events (provider, external_event_id, event_type, payload_hash, processed)
		VALUES ($1, $2, $3, $4, FALSE)
		ON CONFLICT (provider, external_event_id) DO NOTHING
		RETURNING id
	`, ProviderMercadoPago, eventID, eventType, hashStr).Scan(&peID)
	if err == pgx.ErrNoRows {
		out.Inserted = false
		return out, nil
	}
	if err != nil {
		return out, err
	}
	out.Inserted = true
	if isMercadoPagoLookupableOrderID(dataID) {
		s.enqueueMercadoPagoOrderJob(ctx, peID, dataID)
	} else {
		s.log.Info("mercado pago webhook skipped settlement job",
			slog.String("order_id", dataID),
			slog.String("reason", "not_a_lookupable_order_id"),
		)
	}
	if s.audit != nil {
		s.auditLog(ctx, "MERCADO_PAGO_WEBHOOK_RECEIVED", "payment_event", peID, map[string]any{
			"order_id":   dataID,
			"event_type": eventType,
		})
	}
	return out, nil
}

func (s *Service) logPixChargeGatewayFailed(err error, attrs ...any) {
	var mpErr mercadopago.HTTPStatusError
	if errors.As(err, &mpErr) {
		attrs = append(attrs, slog.Int("mp_http_status", mpErr.Status))
		if mpErr.MPRequestID != "" {
			attrs = append(attrs, slog.String("mp_request_id", mpErr.MPRequestID))
		}
		if mpErr.ErrorCode != "" {
			attrs = append(attrs, slog.String("mp_error_code", mpErr.ErrorCode))
		}
		if mpErr.ErrorMessage != "" {
			attrs = append(attrs, slog.String("mp_error_message", mpErr.ErrorMessage))
		}
	}
	attrs = append(attrs, slog.String("reason", err.Error()))
	s.log.Warn("pix charge gateway failed", attrs...)
}

func (s *Service) logMercadoPagoWebhookSignatureFailed(err error, queryDataID, dataIDUsed, xRequestID, xSignature string, body []byte, secret string) {
	attrs := []any{
		slog.String("data_id_query", queryDataID),
		slog.String("data_id_used_for_signature", dataIDUsed),
		slog.String("mp_request_id", xRequestID),
		slog.Bool("x_signature_present", strings.TrimSpace(xSignature) != ""),
		slog.Bool("x_request_id_present", strings.TrimSpace(xRequestID) != ""),
		slog.Int("webhook_secret_len", len(strings.TrimSpace(secret))),
		slog.Int("data_id_candidates", len(mercadoPagoWebhookSignatureDataIDCandidates(queryDataID, body))),
	}
	if appID := mercadoPagoWebhookApplicationID(body); appID != "" {
		attrs = append(attrs, slog.String("webhook_body_application_id", appID))
		if s.mpApplicationID != "" && appID != s.mpApplicationID {
			attrs = append(attrs,
				slog.String("configured_application_id", s.mpApplicationID),
				slog.String("hint", "Access Token and webhook secret are from different MP apps; copy both from the same application"),
			)
		}
	}
	var sigErr *mpwebhook.SignatureError
	if errors.As(err, &sigErr) {
		attrs = append(attrs, slog.String("mp_signature_reason", string(sigErr.Reason)))
	}
	attrs = append(attrs, slog.String("error", err.Error()))
	s.log.Warn("mercado pago webhook signature failed", attrs...)
}
