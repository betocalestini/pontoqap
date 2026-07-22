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
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	mpwebhook "github.com/mercadopago/sdk-go/pkg/webhook"

	"github.com/store-platform/store/internal/billing"
	"github.com/store-platform/store/internal/platform/config"
)

type Service struct {
	pool            *pgxpool.Pool
	gateway         Gateway
	billing         *billing.Service
	provider        string
	webhookKey      string
	mpWebhookSecret string
	pixExpiration   string
	log             *slog.Logger
}

func NewService(pool *pgxpool.Pool, gateway Gateway, bill *billing.Service, payCfg config.PaymentsConfig, log *slog.Logger) *Service {
	if log == nil {
		log = slog.Default()
	}
	return &Service{
		pool:            pool,
		gateway:         gateway,
		billing:         bill,
		provider:        ProviderName(payCfg),
		webhookKey:      payCfg.WebhookSecret,
		mpWebhookSecret: payCfg.MercadoPago.WebhookSecret,
		pixExpiration:   payCfg.MercadoPago.PixExpiration,
		log:             log,
	}
}

var ErrInvalidWebhookSignature = errors.New("invalid webhook signature")

type Charge struct {
	ID            uuid.UUID  `json:"id"`
	InvoiceID     uuid.UUID  `json:"invoice_id"`
	InstallmentID *uuid.UUID `json:"installment_id,omitempty"`
	Status        string     `json:"status"`
	AmountCents   int64      `json:"amount_cents"`
	QRCodeText    string     `json:"qr_code_text"`
	ExpiresAt     time.Time  `json:"expires_at"`
}

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
		return &existing, nil
	}
	if err != pgx.ErrNoRows {
		return nil, err
	}

	email, err := s.payerEmailForInvoice(ctx, invoiceID)
	if err != nil {
		return nil, err
	}

	res, err := s.gateway.CreatePixCharge(ctx, PixChargeInput{
		InvoiceID:         invoiceID,
		AmountCents:       remaining,
		ExternalReference: inv.InvoiceNumber,
		PayerEmail:        email,
		IdempotencyKey:    uuid.NewString(),
		ExpirationISO:     s.pixExpiration,
	})
	if err != nil {
		s.log.Warn("pix charge gateway failed",
			slog.String("invoice_id", invoiceID.String()),
			slog.String("reason", err.Error()),
		)
		return nil, err
	}
	chargeID := uuid.New()
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
	return &Charge{
		ID: chargeID, InvoiceID: invoiceID, Status: "pending",
		AmountCents: res.AmountCents, QRCodeText: res.QRCodeText, ExpiresAt: res.ExpiresAt,
	}, nil
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

	var existing Charge
	var existingInstID *uuid.UUID
	err = s.pool.QueryRow(ctx, `
		SELECT id, invoice_id, installment_id, status, amount_cents, COALESCE(qr_code_text,''), expires_at
		FROM payment_charges
		WHERE installment_id = $1 AND provider = $2 AND status = 'pending' AND expires_at > NOW()
		ORDER BY created_at DESC LIMIT 1
	`, installmentID, s.provider).Scan(&existing.ID, &existing.InvoiceID, &existingInstID, &existing.Status, &existing.AmountCents,
		&existing.QRCodeText, &existing.ExpiresAt)
	if err == nil {
		existing.InstallmentID = existingInstID
		s.log.Info("pix charge reused",
			slog.String("invoice_id", existing.InvoiceID.String()),
			slog.String("installment_id", installmentID.String()),
			slog.String("charge_id", existing.ID.String()),
		)
		return &existing, nil
	}
	if err != pgx.ErrNoRows {
		return nil, err
	}

	email, err := s.payerEmailForInvoice(ctx, inst.InvoiceID)
	if err != nil {
		return nil, err
	}

	amount := inst.RemainingCents
	extRef := fmt.Sprintf("INSTALLMENT-%s", installmentID.String())
	res, err := s.gateway.CreatePixCharge(ctx, PixChargeInput{
		InvoiceID:         inst.InvoiceID,
		AmountCents:       amount,
		ExternalReference: extRef,
		PayerEmail:        email,
		IdempotencyKey:    uuid.NewString(),
		ExpirationISO:     s.pixExpiration,
	})
	if err != nil {
		s.log.Warn("pix charge gateway failed",
			slog.String("invoice_id", inst.InvoiceID.String()),
			slog.String("installment_id", installmentID.String()),
			slog.String("reason", err.Error()),
		)
		return nil, err
	}
	chargeID := uuid.New()
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
	return &Charge{
		ID: chargeID, InvoiceID: inst.InvoiceID, InstallmentID: &instID, Status: "pending",
		AmountCents: res.AmountCents, QRCodeText: res.QRCodeText, ExpiresAt: res.ExpiresAt,
	}, nil
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

	_, err = tx.Exec(ctx, `
		INSERT INTO payments (id, invoice_id, installment_id, payment_charge_id, provider, external_payment_id, amount_cents, status, settled_at)
		VALUES ($1,$2,$3,$4,'sandbox',$5,$6,'settled',NOW())
	`, uuid.New(), invoiceID, installmentID, chargeID, paymentExtID, amountCents)
	if err != nil {
		return out, err
	}
	_, err = tx.Exec(ctx, `UPDATE payment_charges SET status = 'paid', paid_at = NOW(), updated_at = NOW() WHERE id = $1`, chargeID)
	if err != nil {
		return out, err
	}

	if installmentID != nil {
		if err := s.billing.ApplyInstallmentPaymentTx(ctx, tx, *installmentID, amountCents); err != nil {
			return out, err
		}
	} else {
		_, err = tx.Exec(ctx, `
			UPDATE invoices SET paid_cents = paid_cents + $2,
				status = CASE WHEN paid_cents + $2 >= total_cents THEN 'paid' ELSE status END,
				paid_at = CASE WHEN paid_cents + $2 >= total_cents THEN NOW() ELSE paid_at END,
				updated_at = NOW()
			WHERE id = $1
		`, invoiceID, amountCents)
		if err != nil {
			return out, err
		}
		_, err = tx.Exec(ctx, `
			UPDATE customers SET current_exposure_cents = GREATEST(0, current_exposure_cents - $2), updated_at = NOW()
			WHERE id = (SELECT customer_id FROM invoices WHERE id = $1)
		`, invoiceID, amountCents)
		if err != nil {
			return out, err
		}
	}

	var invStatus string
	if err := tx.QueryRow(ctx, `SELECT status FROM invoices WHERE id = $1`, invoiceID).Scan(&invStatus); err == nil && invStatus == "paid" {
		out.Settled = true
	}

	_, err = tx.Exec(ctx, `UPDATE payment_events SET processed = TRUE, processed_at = NOW() WHERE id = $1`, peID)
	if err != nil {
		return out, err
	}
	if err := tx.Commit(ctx); err != nil {
		return out, err
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
	err := s.pool.QueryRow(ctx, `SELECT external_id, amount_cents FROM payment_charges WHERE id = $1`, chargeID).Scan(&extID, &amount)
	if err != nil {
		return err
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
func (s *Service) ReceiveMercadoPagoOrderWebhook(ctx context.Context, xSignature, xRequestID, dataID string, body []byte) (MercadoPagoWebhookResult, error) {
	var out MercadoPagoWebhookResult
	if s.mpWebhookSecret == "" {
		return out, errors.New("webhook Mercado Pago não configurado")
	}
	if err := mpwebhook.ValidateSignature(xSignature, xRequestID, dataID, s.mpWebhookSecret); err != nil {
		return out, fmt.Errorf("%w: %v", ErrInvalidWebhookSignature, err)
	}
	payload, err := parseMercadoPagoOrderWebhook(body)
	if err != nil {
		return out, err
	}
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
	return out, nil
}
