package notification

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/store-platform/store/internal/platform/config"
	"github.com/store-platform/store/internal/platform/logging"
)

const (
	EventUserVerification         = "user.verification_requested"
	EventInvoiceClosed            = "invoice.closed"
	EventInvoicePaymentReminder   = "invoice.payment_reminder"
	EventInvoicePaymentEscalation = "invoice.payment_escalation"
	EventAdminInvitation          = "admin.invitation_sent"
)

type OutboxHandler struct {
	Mailer Mailer
	Log    *slog.Logger
}

func (h *OutboxHandler) Handle(ctx context.Context, eventType string, payload json.RawMessage) error {
	if h.Log == nil {
		h.Log = slog.Default()
	}
	switch eventType {
	case EventUserVerification:
		var p struct {
			To        string `json:"to"`
			Name      string `json:"name"`
			VerifyURL string `json:"verify_url"`
		}
		if err := json.Unmarshal(payload, &p); err != nil {
			return err
		}
		subj, text, html := VerifyEmailContent(p.Name, p.VerifyURL)
		return h.send(ctx, eventType, p.To, subj, text, html)
	case EventInvoiceClosed:
		var p struct {
			To            string `json:"to"`
			Name          string `json:"name"`
			InvoiceNumber string `json:"invoice_number"`
			RefYear       int    `json:"ref_year"`
			RefMonth      int    `json:"ref_month"`
			TotalCents    int64  `json:"total_cents"`
			DueAt         string `json:"due_at"`
			InvoiceURL    string `json:"invoice_url"`
		}
		if err := json.Unmarshal(payload, &p); err != nil {
			return err
		}
		due, err := time.Parse(time.RFC3339, p.DueAt)
		if err != nil {
			return fmt.Errorf("due_at: %w", err)
		}
		subj, text, html := InvoiceClosedContent(p.Name, p.InvoiceNumber, p.RefYear, p.RefMonth, p.TotalCents, due, p.InvoiceURL)
		return h.send(ctx, eventType, p.To, subj, text, html)
	case EventInvoicePaymentReminder, EventInvoicePaymentEscalation:
		var p struct {
			To            string `json:"to"`
			Name          string `json:"name"`
			InvoiceNumber string `json:"invoice_number"`
			TotalCents    int64  `json:"total_cents"`
			DueAt         string `json:"due_at"`
			InvoiceURL    string `json:"invoice_url"`
		}
		if err := json.Unmarshal(payload, &p); err != nil {
			return err
		}
		due, err := time.Parse(time.RFC3339, p.DueAt)
		if err != nil {
			return fmt.Errorf("due_at: %w", err)
		}
		escalation := eventType == EventInvoicePaymentEscalation
		subj, text, html := InvoicePaymentFollowUpContent(p.Name, p.InvoiceNumber, p.TotalCents, due, p.InvoiceURL, escalation)
		return h.send(ctx, eventType, p.To, subj, text, html)
	case EventAdminInvitation:
		var p struct {
			To        string `json:"to"`
			Name      string `json:"name"`
			InviteURL string `json:"invite_url"`
		}
		if err := json.Unmarshal(payload, &p); err != nil {
			return err
		}
		subj, text, html := AdminInviteContent(p.Name, p.InviteURL)
		return h.send(ctx, eventType, p.To, subj, text, html)
	default:
		return nil
	}
}

func (h *OutboxHandler) send(ctx context.Context, eventType, to, subj, text, html string) error {
	err := h.Mailer.Send(to, subj, text, html)
	masked := logging.MaskEmail(to)
	if err != nil {
		h.Log.Error("outbox email failed",
			slog.String("event_type", eventType),
			slog.String("to", masked),
			slog.String("error", err.Error()),
		)
		return err
	}
	h.Log.Info("outbox email sent",
		slog.String("event_type", eventType),
		slog.String("to", masked),
	)
	_ = ctx
	return nil
}

func BuildVerifyURL(storeWebURL, rawToken string) string {
	base := config.TrimTrailingSlash(storeWebURL)
	if base == "" {
		base = "http://localhost:5173"
	}
	return fmt.Sprintf("%s/verificar-email?token=%s", base, rawToken)
}

func BuildInvoiceURL(storeWebURL, invoiceID string) string {
	base := config.TrimTrailingSlash(storeWebURL)
	if base == "" {
		base = "http://localhost:5173"
	}
	return fmt.Sprintf("%s/faturas/%s", base, invoiceID)
}

func BuildAdminInviteURL(adminWebURL, rawToken string) string {
	base := config.TrimTrailingSlash(adminWebURL)
	if base == "" {
		base = "http://localhost:5174"
	}
	return fmt.Sprintf("%s/convite/aceitar?token=%s", base, rawToken)
}
