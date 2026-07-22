package reports

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type PixReconciliationRow struct {
	CustomerID           uuid.UUID  `json:"customer_id"`
	CustomerName         string     `json:"customer_name"`
	InvoiceID            uuid.UUID  `json:"invoice_id"`
	InvoiceNumber        string     `json:"invoice_number"`
	ChargeID             *uuid.UUID `json:"charge_id,omitempty"`
	TxID                 string     `json:"txid,omitempty"`
	Provider             string     `json:"provider,omitempty"`
	ChargeAmountCents    int64      `json:"charge_amount_cents"`
	ReceivedAmountCents  int64      `json:"received_amount_cents"`
	PaidAt               *time.Time `json:"paid_at,omitempty"`
	ChargeStatus         string     `json:"charge_status,omitempty"`
	InvoiceStatus        string     `json:"invoice_status"`
	ReconciliationStatus string     `json:"reconciliation_status"`
}

type PixReconciliationFilter struct {
	DateRange
	CustomerID   *uuid.UUID
	InvoiceID    *uuid.UUID
	Status       string
	DivergenceOnly bool
	PageFilter
}

func (s *Service) PixReconciliation(ctx context.Context, f PixReconciliationFilter) ([]PixReconciliationRow, int, error) {
	where := []string{"i.created_at >= $1", "i.created_at < $2"}
	args := []any{f.From, f.To}
	n := 3

	if f.CustomerID != nil {
		where = append(where, fmt.Sprintf("i.customer_id = $%d", n))
		args = append(args, *f.CustomerID)
		n++
	}
	if f.InvoiceID != nil {
		where = append(where, fmt.Sprintf("i.id = $%d", n))
		args = append(args, *f.InvoiceID)
		n++
	}
	whereSQL := strings.Join(where, " AND ")

	var total int
	countQ := fmt.Sprintf(`SELECT COUNT(*) FROM invoices i WHERE %s`, whereSQL)
	if err := s.pool.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	limit, offset := f.Limit, f.Offset
	q := fmt.Sprintf(`
		SELECT i.customer_id, u.name, i.id, i.invoice_number,
		       pc.id, COALESCE(pc.txid,''), COALESCE(pc.provider,''),
		       COALESCE(pc.amount_cents,0), COALESCE(pay.amount_cents,0),
		       COALESCE(pc.paid_at, pay.settled_at), COALESCE(pc.status,''),
		       i.status,
		       (SELECT COUNT(*) FROM payments p2 WHERE p2.payment_charge_id = pc.id) AS pay_count
		FROM invoices i
		JOIN customers c ON c.id = i.customer_id
		JOIN users u ON u.id = c.user_id
		LEFT JOIN LATERAL (
			SELECT * FROM payment_charges pch
			WHERE pch.invoice_id = i.id
			ORDER BY pch.created_at DESC LIMIT 1
		) pc ON true
		LEFT JOIN payments pay ON pay.payment_charge_id = pc.id AND pay.status = 'settled'
		WHERE %s
		ORDER BY i.created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereSQL, n, n+1)
	args = append(args, limit, offset)

	rows, err := s.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []PixReconciliationRow
	for rows.Next() {
		var r PixReconciliationRow
		var chargeID *uuid.UUID
		var payCount int
		if err := rows.Scan(
			&r.CustomerID, &r.CustomerName, &r.InvoiceID, &r.InvoiceNumber,
			&chargeID, &r.TxID, &r.Provider,
			&r.ChargeAmountCents, &r.ReceivedAmountCents,
			&r.PaidAt, &r.ChargeStatus, &r.InvoiceStatus, &payCount,
		); err != nil {
			return nil, 0, err
		}
		r.ChargeID = chargeID
		r.ReconciliationStatus = classifyPix(r, payCount)
		if f.Status != "" && r.ReconciliationStatus != f.Status {
			continue
		}
		if f.DivergenceOnly && r.ReconciliationStatus == "CONCILIADO" {
			continue
		}
		out = append(out, r)
	}
	return out, total, rows.Err()
}

func classifyPix(r PixReconciliationRow, payCount int) string {
	if r.ChargeID == nil {
		if r.InvoiceStatus == "paid" {
			return "PAGAMENTO_SEM_FATURA"
		}
		return "PENDENTE"
	}
	if payCount > 1 {
		return "EVENTO_DUPLICADO"
	}
	if r.ChargeStatus == "expired" {
		return "EXPIRADO"
	}
	if r.ReceivedAmountCents > 0 && r.ChargeAmountCents > 0 && r.ReceivedAmountCents != r.ChargeAmountCents {
		return "VALOR_DIVERGENTE"
	}
	if r.InvoiceStatus == "paid" && r.ReceivedAmountCents >= r.ChargeAmountCents && r.ChargeAmountCents > 0 {
		return "CONCILIADO"
	}
	if r.ChargeStatus == "pending" || r.ChargeStatus == "active" {
		return "PENDENTE"
	}
	return "PENDENTE"
}
