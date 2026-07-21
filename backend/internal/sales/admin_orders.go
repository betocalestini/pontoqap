package sales

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	platformerrors "github.com/store-platform/store/internal/platform/errors"
)

type AdminOrderListItem struct {
	ID            uuid.UUID  `json:"id"`
	OrderNumber   string     `json:"order_number"`
	Status        string     `json:"status"`
	TotalCents    int64      `json:"total_cents"`
	CustomerID    uuid.UUID  `json:"customer_id"`
	CustomerName  string     `json:"customer_name"`
	CustomerEmail string     `json:"customer_email"`
	ConfirmedAt   *time.Time `json:"confirmed_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

type AdminOrderItem struct {
	ID               uuid.UUID `json:"id"`
	SKUID            uuid.UUID `json:"sku_id"`
	ProductName      string    `json:"product_name"`
	SKUCode          string    `json:"sku_code"`
	UnitPriceCents   int64     `json:"unit_price_cents"`
	Quantity         int       `json:"quantity"`
	TotalCents       int64     `json:"total_cents"`
}

type AdminOrderDetail struct {
	AdminOrderListItem
	Items       []AdminOrderItem `json:"items"`
	CancelledAt *time.Time       `json:"cancelled_at,omitempty"`
}

type AdminOrderFilter struct {
	Status string
	Search string
	Limit  int
	Offset int
}

func (s *Service) AdminListOrders(ctx context.Context, f AdminOrderFilter) ([]AdminOrderListItem, int, error) {
	if f.Limit <= 0 || f.Limit > 100 {
		f.Limit = 50
	}
	if f.Offset < 0 {
		f.Offset = 0
	}
	status := strings.TrimSpace(strings.ToLower(f.Status))
	search := strings.TrimSpace(strings.ToLower(f.Search))

	where := []string{"1=1"}
	args := []any{}
	n := 1
	if status != "" {
		where = append(where, fmt.Sprintf("o.status = $%d", n))
		args = append(args, status)
		n++
	}
	if search != "" {
		where = append(where, fmt.Sprintf("(LOWER(o.order_number) LIKE $%d OR LOWER(u.email) LIKE $%d OR LOWER(u.name) LIKE $%d)", n, n, n))
		args = append(args, "%"+search+"%")
		n++
	}
	whereSQL := strings.Join(where, " AND ")

	var total int
	countQ := fmt.Sprintf(`
		SELECT COUNT(*) FROM orders o
		JOIN customers c ON c.id = o.customer_id
		JOIN users u ON u.id = c.user_id
		WHERE %s
	`, whereSQL)
	if err := s.pool.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	listQ := fmt.Sprintf(`
		SELECT o.id, o.order_number, o.status, o.total_cents, c.id, u.name, u.email, o.confirmed_at, o.created_at
		FROM orders o
		JOIN customers c ON c.id = o.customer_id
		JOIN users u ON u.id = c.user_id
		WHERE %s
		ORDER BY o.created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereSQL, n, n+1)
	args = append(args, f.Limit, f.Offset)
	rows, err := s.pool.Query(ctx, listQ, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []AdminOrderListItem
	for rows.Next() {
		var it AdminOrderListItem
		if err := rows.Scan(&it.ID, &it.OrderNumber, &it.Status, &it.TotalCents, &it.CustomerID, &it.CustomerName, &it.CustomerEmail, &it.ConfirmedAt, &it.CreatedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, it)
	}
	return items, total, rows.Err()
}

func (s *Service) AdminGetOrder(ctx context.Context, orderID uuid.UUID) (*AdminOrderDetail, error) {
	var d AdminOrderDetail
	err := s.pool.QueryRow(ctx, `
		SELECT o.id, o.order_number, o.status, o.total_cents, c.id, u.name, u.email, o.confirmed_at, o.created_at, o.cancelled_at
		FROM orders o
		JOIN customers c ON c.id = o.customer_id
		JOIN users u ON u.id = c.user_id
		WHERE o.id = $1
	`, orderID).Scan(&d.ID, &d.OrderNumber, &d.Status, &d.TotalCents, &d.CustomerID, &d.CustomerName, &d.CustomerEmail, &d.ConfirmedAt, &d.CreatedAt, &d.CancelledAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errNotFound()
	}
	if err != nil {
		return nil, err
	}
	rows, err := s.pool.Query(ctx, `
		SELECT id, sku_id, product_name_snapshot, sku_code_snapshot, unit_price_cents, quantity, total_cents
		FROM order_items WHERE order_id = $1 ORDER BY product_name_snapshot
	`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var it AdminOrderItem
		if err := rows.Scan(&it.ID, &it.SKUID, &it.ProductName, &it.SKUCode, &it.UnitPriceCents, &it.Quantity, &it.TotalCents); err != nil {
			return nil, err
		}
		d.Items = append(d.Items, it)
	}
	return &d, rows.Err()
}

func (s *Service) AdminCancelOrder(ctx context.Context, orderID uuid.UUID, actorUserID uuid.UUID) (*AdminOrderDetail, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var customerID uuid.UUID
	var status string
	var total int64
	err = tx.QueryRow(ctx, `
		SELECT customer_id, status, total_cents FROM orders WHERE id = $1 FOR UPDATE
	`, orderID).Scan(&customerID, &status, &total)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errNotFound()
	}
	if err != nil {
		return nil, err
	}
	if status != "confirmed" {
		return nil, &AppError{Code: platformerrors.CodeValidation, Message: "Pedido não pode ser cancelado", Status: 400}
	}

	itemRows, err := tx.Query(ctx, `SELECT sku_id, quantity FROM order_items WHERE order_id = $1`, orderID)
	if err != nil {
		return nil, err
	}
	type line struct {
		skuID uuid.UUID
		qty   int
	}
	var lines []line
	for itemRows.Next() {
		var l line
		if err := itemRows.Scan(&l.skuID, &l.qty); err != nil {
			itemRows.Close()
			return nil, err
		}
		lines = append(lines, l)
	}
	itemRows.Close()
	if err := itemRows.Err(); err != nil {
		return nil, err
	}

	now := time.Now()
	billRes, err := s.billing.ApplyOrderCancellationBillingTx(ctx, tx, customerID, orderID, total, actorUserID, now)
	if err != nil {
		return nil, mapBillingCancelError(err)
	}

	for _, l := range lines {
		if err := s.inventory.RestoreFromSaleTx(ctx, tx, l.skuID, l.qty, orderID, actorUserID); err != nil {
			return nil, err
		}
	}

	if billRes.ExposureDeltaCents != 0 {
		_, err = tx.Exec(ctx, `
			UPDATE customers SET current_exposure_cents = GREATEST(0, current_exposure_cents + $2), updated_at = NOW()
			WHERE id = $1
		`, customerID, billRes.ExposureDeltaCents)
		if err != nil {
			return nil, err
		}
	}

	_, err = tx.Exec(ctx, `
		UPDATE orders SET status = 'cancelled', cancelled_at = $2, updated_at = NOW() WHERE id = $1
	`, orderID, now)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return s.AdminGetOrder(ctx, orderID)
}

func mapBillingCancelError(err error) error {
	msg := err.Error()
	switch {
	case strings.Contains(msg, "fatura já quitada"):
		return &AppError{Code: platformerrors.CodeValidation, Message: "Fatura já quitada; não é possível estornar este pedido", Status: 409}
	case strings.Contains(msg, "ajuste excede"):
		return &AppError{Code: platformerrors.CodeValidation, Message: msg, Status: 400}
	case strings.Contains(msg, "fatura não encontrada"):
		return &AppError{Code: platformerrors.CodeValidation, Message: msg, Status: 404}
	default:
		return err
	}
}

func errNotFound() error {
	return &AppError{Code: platformerrors.CodeNotFound, Message: "Não encontrado", Status: 404}
}
