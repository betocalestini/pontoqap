package notification

import (
	"fmt"
	"time"
)

func VerifyEmailContent(name, verifyURL string) (subject, text, html string) {
	subject = "Confirme seu e-mail — Store"
	text = fmt.Sprintf("Olá, %s!\n\nConfirme seu cadastro acessando o link abaixo (válido por 24 horas):\n\n%s\n\nSe você não solicitou este cadastro, ignore este e-mail.\n", name, verifyURL)
	html = fmt.Sprintf(`<p>Olá, %s!</p><p>Confirme seu cadastro clicando no link abaixo (válido por 24 horas):</p><p><a href="%s">Confirmar e-mail</a></p><p>Se você não solicitou este cadastro, ignore este e-mail.</p>`, name, verifyURL)
	return subject, text, html
}

func InvoiceClosedContent(name, invoiceNumber string, refYear, refMonth int, totalCents int64, dueAt time.Time, invoiceURL string) (subject, text, html string) {
	total := formatBRL(totalCents)
	due := dueAt.Format("02/01/2006")
	subject = fmt.Sprintf("Fatura %s — Store", invoiceNumber)
	text = fmt.Sprintf("Olá, %s!\n\nSua fatura %s referente a %02d/%d foi fechada.\nTotal: %s\nVencimento: %s\n\nConsulte e pague em: %s\n", name, invoiceNumber, refMonth, refYear, total, due, invoiceURL)
	html = fmt.Sprintf(`<p>Olá, %s!</p><p>Sua fatura <strong>%s</strong> (%02d/%d) foi fechada.</p><p>Total: <strong>%s</strong><br>Vencimento: %s</p><p><a href="%s">Ver fatura</a></p>`, name, invoiceNumber, refMonth, refYear, total, due, invoiceURL)
	return subject, text, html
}

func formatBRL(cents int64) string {
	reais := cents / 100
	frac := cents % 100
	if frac < 0 {
		frac = -frac
	}
	return fmt.Sprintf("R$ %d,%02d", reais, frac)
}
