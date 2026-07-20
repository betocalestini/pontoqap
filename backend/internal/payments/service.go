package payments

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/store-platform/store/internal/billing"
)

type Service struct {
	pool       *pgxpool.Pool
	gateway    Gateway
	billing    *billing.Service
	webhookKey string
}

func NewService(pool *pgxpool.Pool, gateway Gateway, bill *billing.Service, webhookSecret string) *Service {
	return &Service{pool: pool, gateway: gateway, billing: bill, webhookKey: webhookSecret}
}

type Charge struct {
	ID          uuid.UUID `json:"id"`
	InvoiceID   uuid.UUID `json:"invoice_id"`
	Status      string    `json:"status"`
	AmountCents int64     `json:"amount_cents"`
	QRCodeText  string    `json:"qr_code_text"`
	ExpiresAt   time.Time `json:"expires_at"`
}

func (s *Service) CreateOrReusePixCharge(ctx context.Context, invoiceID uuid.UUID) (*Charge, error) {
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
		WHERE invoice_id = $1 AND status = 'pending' AND expires_at > NOW()
		ORDER BY created_at DESC LIMIT 1
	`, invoiceID).Scan(&existing.ID, &existing.InvoiceID, &existing.Status, &existing.AmountCents,
		&existing.QRCodeText, &existing.ExpiresAt)
	if err == nil {
		return &existing, nil
	}
	if err != pgx.ErrNoRows {
		return nil, err
	}

	res, err := s.gateway.CreatePixCharge(ctx, invoiceID, remaining, inv.InvoiceNumber)
	if err != nil {
		return nil, err
	}
	chargeID := uuid.New()
	_, err = s.pool.Exec(ctx, `
		INSERT INTO payment_charges (
			id, invoice_id, provider, external_id, txid, status, amount_cents, qr_code_text, expires_at
		) VALUES ($1,$2,$3,$4,$5,'pending',$6,$7,$8)
	`, chargeID, invoiceID, res.Provider, res.ExternalID, res.TxID, res.AmountCents, res.QRCodeText, res.ExpiresAt)
	if err != nil {
		return nil, err
	}
	return &Charge{
		ID: chargeID, InvoiceID: invoiceID, Status: "pending",
		AmountCents: res.AmountCents, QRCodeText: res.QRCodeText, ExpiresAt: res.ExpiresAt,
	}, nil
}

func (s *Service) ProcessWebhook(ctx context.Context, payload []byte, signature string) error {
	if !s.gateway.VerifyWebhookSignature(payload, signature) {
		return errors.New("invalid webhook signature")
	}
	eventID, eventType, paymentExtID, amountCents, err := s.gateway.ParseWebhookEvent(payload)
	if err != nil {
		return err
	}
	if eventType != "payment.confirmed" {
		return nil
	}

	hash := sha256.Sum256(payload)
	hashStr := hex.EncodeToString(hash[:])

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
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
		return nil
	}
	if err != nil {
		return err
	}

	var chargeID, invoiceID uuid.UUID
	var chargeAmount int64
	err = tx.QueryRow(ctx, `
		SELECT id, invoice_id, amount_cents FROM payment_charges
		WHERE provider = 'sandbox' AND external_id = $1 FOR UPDATE
	`, paymentExtID).Scan(&chargeID, &invoiceID, &chargeAmount)
	if err != nil {
		return err
	}
	if amountCents != chargeAmount {
		_, _ = tx.Exec(ctx, `UPDATE payment_events SET error_message = 'amount mismatch', processed = TRUE WHERE id = $1`, peID)
		return errors.New("payment amount mismatch")
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO payments (id, invoice_id, payment_charge_id, provider, external_payment_id, amount_cents, status, settled_at)
		VALUES ($1,$2,$3,'sandbox',$4,$5,'settled',NOW())
	`, uuid.New(), invoiceID, chargeID, paymentExtID, amountCents)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `UPDATE payment_charges SET status = 'paid', paid_at = NOW(), updated_at = NOW() WHERE id = $1`, chargeID)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `
		UPDATE invoices SET paid_cents = paid_cents + $2,
			status = CASE WHEN paid_cents + $2 >= total_cents THEN 'paid' ELSE status END,
			paid_at = CASE WHEN paid_cents + $2 >= total_cents THEN NOW() ELSE paid_at END,
			updated_at = NOW()
		WHERE id = $1
	`, invoiceID, amountCents)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `
		UPDATE customers SET current_exposure_cents = GREATEST(0, current_exposure_cents - $2), updated_at = NOW()
		WHERE id = (SELECT customer_id FROM invoices WHERE id = $1)
	`, invoiceID, amountCents)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `UPDATE payment_events SET processed = TRUE, processed_at = NOW() WHERE id = $1`, peID)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
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
	return s.ProcessWebhook(ctx, body, sig)
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
