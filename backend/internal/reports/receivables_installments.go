package reports

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type ReceivableInstallmentRow struct {
	InvoiceID         uuid.UUID  `json:"invoice_id"`
	InvoiceNumber     string     `json:"invoice_number"`
	CustomerName      string     `json:"customer_name"`
	InstallmentID     uuid.UUID  `json:"installment_id"`
	InstallmentNumber int        `json:"installment_number"`
	DueDate           time.Time  `json:"due_date"`
	AmountCents       int64      `json:"amount_cents"`
	RemainingCents    int64      `json:"remaining_cents"`
	Status            string     `json:"status"`
}

func (s *Service) ReceivablesInstallments(ctx context.Context, f ReceivablesFilter, now time.Time) ([]ReceivableInstallmentRow, int, error) {
	where := []string{"ii.remaining_cents > 0", "ii.status NOT IN ('paid', 'canceled')"}
	args := []any{}
	n := 1
	if f.CustomerID != nil {
		where = append(where, fmt.Sprintf("i.customer_id = $%d", n))
		args = append(args, *f.CustomerID)
		n++
	}
	if f.OverdueOnly {
		where = append(where, "ii.status = 'overdue'")
	}
	whereSQL := strings.Join(where, " AND ")
	var total int
	if err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM invoice_installments ii
		JOIN invoices i ON i.id = ii.invoice_id
		WHERE `+whereSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	limit, offset := f.Limit, f.Offset
	q := fmt.Sprintf(`
		SELECT i.id, i.invoice_number, u.name, ii.id, ii.installment_number,
		       ii.due_date, ii.amount_cents, ii.remaining_cents, ii.status
		FROM invoice_installments ii
		JOIN invoices i ON i.id = ii.invoice_id
		JOIN customers c ON c.id = i.customer_id
		JOIN users u ON u.id = c.user_id
		WHERE %s
		ORDER BY ii.due_date, ii.installment_number
		LIMIT $%d OFFSET $%d
	`, whereSQL, n, n+1)
	args = append(args, limit, offset)
	rows, err := s.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var out []ReceivableInstallmentRow
	for rows.Next() {
		var r ReceivableInstallmentRow
		if err := rows.Scan(
			&r.InvoiceID, &r.InvoiceNumber, &r.CustomerName, &r.InstallmentID,
			&r.InstallmentNumber, &r.DueDate, &r.AmountCents, &r.RemainingCents, &r.Status,
		); err != nil {
			return nil, 0, err
		}
		_ = now
		out = append(out, r)
	}
	return out, total, rows.Err()
}
