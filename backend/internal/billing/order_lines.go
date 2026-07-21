package billing

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// InvoiceProductLine é uma linha de produto de um pedido faturado.
type InvoiceProductLine struct {
	ProductName    string `json:"product_name"`
	SkuCode        string `json:"sku_code"`
	Quantity       int    `json:"quantity"`
	UnitPriceCents int64  `json:"unit_price_cents"`
	TotalCents     int64  `json:"total_cents"`
}

// BillingEntryView é um lançamento visível ao cliente (competência aberta).
type BillingEntryView struct {
	ID          uuid.UUID            `json:"id"`
	Description string               `json:"description"`
	AmountCents int64                `json:"amount_cents"`
	OccurredAt  time.Time            `json:"occurred_at"`
	OrderNumber string               `json:"order_number,omitempty"`
	Products    []InvoiceProductLine `json:"products"`
}

// OpenPeriodDetail competência em aberto com lançamentos.
type OpenPeriodDetail struct {
	Period  OpenPeriodSummary  `json:"period"`
	Entries []BillingEntryView `json:"entries"`
}

func (s *Service) loadProductsForOrder(ctx context.Context, orderID uuid.UUID) ([]InvoiceProductLine, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT product_name_snapshot, sku_code_snapshot, quantity, unit_price_cents, total_cents
		FROM order_items WHERE order_id = $1 ORDER BY created_at
	`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []InvoiceProductLine
	for rows.Next() {
		var line InvoiceProductLine
		if err := rows.Scan(&line.ProductName, &line.SkuCode, &line.Quantity, &line.UnitPriceCents, &line.TotalCents); err != nil {
			return nil, err
		}
		out = append(out, line)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if out == nil {
		out = []InvoiceProductLine{}
	}
	return out, nil
}

func (s *Service) loadEntryViewsForPeriod(ctx context.Context, periodID uuid.UUID) ([]BillingEntryView, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT be.id, be.description, be.amount_cents, be.occurred_at, be.order_id, o.order_number
		FROM billing_entries be
		LEFT JOIN orders o ON o.id = be.order_id
		WHERE be.billing_period_id = $1
		ORDER BY be.occurred_at ASC, be.created_at ASC
	`, periodID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []BillingEntryView
	orderCache := make(map[uuid.UUID][]InvoiceProductLine)
	for rows.Next() {
		var e BillingEntryView
		var orderID *uuid.UUID
		var orderNumber *string
		if err := rows.Scan(&e.ID, &e.Description, &e.AmountCents, &e.OccurredAt, &orderID, &orderNumber); err != nil {
			return nil, err
		}
		if orderNumber != nil {
			e.OrderNumber = *orderNumber
		}
		if orderID != nil {
			lines, ok := orderCache[*orderID]
			if !ok {
				var err error
				lines, err = s.loadProductsForOrder(ctx, *orderID)
				if err != nil {
					return nil, err
				}
				orderCache[*orderID] = lines
			}
			e.Products = lines
		} else {
			e.Products = []InvoiceProductLine{}
		}
		out = append(out, e)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if out == nil {
		out = []BillingEntryView{}
	}
	return out, nil
}

func (s *Service) attachProductsToInvoiceItems(ctx context.Context, items []InvoiceItem, orderIDs []uuid.UUID) error {
	if len(items) != len(orderIDs) {
		return nil
	}
	cache := make(map[uuid.UUID][]InvoiceProductLine)
	for i := range items {
		oid := orderIDs[i]
		if oid == uuid.Nil {
			items[i].Products = []InvoiceProductLine{}
			continue
		}
		lines, ok := cache[oid]
		if !ok {
			var err error
			lines, err = s.loadProductsForOrder(ctx, oid)
			if err != nil {
				return err
			}
			cache[oid] = lines
		}
		items[i].Products = lines
	}
	return nil
}
