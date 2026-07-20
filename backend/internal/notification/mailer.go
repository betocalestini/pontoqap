package notification

import (
	"fmt"
	"net/smtp"
	"strings"

	"github.com/store-platform/store/internal/platform/config"
)

type Mailer interface {
	Send(to, subject, bodyText, bodyHTML string) error
}

type SMTPMailer struct {
	cfg config.SMTPConfig
}

func NewSMTPMailer(cfg config.SMTPConfig) *SMTPMailer {
	return &SMTPMailer{cfg: cfg}
}

func (m *SMTPMailer) Send(to, subject, bodyText, bodyHTML string) error {
	if m.cfg.Host == "" {
		return fmt.Errorf("smtp not configured")
	}
	from := m.cfg.From
	if from == "" {
		from = "noreply@store.local"
	}
	msg := buildMIME(from, to, subject, bodyText, bodyHTML)
	addr := fmt.Sprintf("%s:%d", m.cfg.Host, m.cfg.Port)
	var auth smtp.Auth
	if m.cfg.User != "" {
		auth = smtp.PlainAuth("", m.cfg.User, m.cfg.Password, m.cfg.Host)
	}
	return smtp.SendMail(addr, auth, extractAddr(from), []string{to}, []byte(msg))
}

func extractAddr(from string) string {
	if i := strings.LastIndex(from, "<"); i >= 0 {
		return strings.Trim(from[i+1:], "> ")
	}
	return from
}

func buildMIME(from, to, subject, text, html string) string {
	boundary := "store-boundary"
	var b strings.Builder
	b.WriteString(fmt.Sprintf("From: %s\r\n", from))
	b.WriteString(fmt.Sprintf("To: %s\r\n", to))
	b.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=%s\r\n\r\n", boundary))
	b.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	b.WriteString("Content-Type: text/plain; charset=UTF-8\r\n\r\n")
	b.WriteString(text)
	b.WriteString("\r\n")
	if html != "" {
		b.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		b.WriteString("Content-Type: text/html; charset=UTF-8\r\n\r\n")
		b.WriteString(html)
		b.WriteString("\r\n")
	}
	b.WriteString(fmt.Sprintf("--%s--\r\n", boundary))
	return b.String()
}
