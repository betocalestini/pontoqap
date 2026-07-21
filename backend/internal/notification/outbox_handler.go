package notification

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/store-platform/store/internal/platform/config"
)

const (
	EventUserVerification        = "user.verification_requested"
	EventInvoiceClosed           = "invoice.closed"
	EventInvoicePaymentReminder  = "invoice.payment_reminder"
	EventInvoicePaymentEscalation = "invoice.payment_escalation"
	EventAdminInvitation         = "admin.invitation_sent"
)

type OutboxHandler struct {
	Mailer Mailer
	Log    interface {
		Error(msg string, args ...any)
	}
}

func (h *OutboxHandler) Handle(ctx context.Context, eventType string, payload json.RawMessage) error {
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
		return h.Mailer.Send(p.To, subj, text, html)
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
		return h.Mailer.Send(p.To, subj, text, html)
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
		return h.Mailer.Send(p.To, subj, text, html)
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
		return h.Mailer.Send(p.To, subj, text, html)
	default:
		return nil
	}
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
