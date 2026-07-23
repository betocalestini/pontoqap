package mercadopago

import (
	"context"
	"fmt"
	"strings"
)

// OrderDetail is the subset of GET /v1/orders/{id} used for settlement.
type OrderDetail struct {
	ID                string
	ExternalReference string
	TotalAmountCents  int64
	Payments          []OrderPaymentDetail
}

type OrderPaymentDetail struct {
	ID            string
	Status        string
	StatusDetail  string
	AmountCents   int64
	PaymentMethod string
}

func (g *Gateway) FetchOrder(ctx context.Context, orderID string) (OrderDetail, error) {
	if orderID == "" {
		return OrderDetail{}, fmt.Errorf("order id vazio")
	}
	var raw orderDetailResponse
	path := "/v1/orders/" + orderID
	if _, err := g.client.getJSON(ctx, path, &raw); err != nil {
		return OrderDetail{}, err
	}
	return raw.toDetail()
}

type orderDetailResponse struct {
	ID                string `json:"id"`
	ExternalReference string `json:"external_reference"`
	TotalAmount       string `json:"total_amount"`
	Transactions      struct {
		Payments []struct {
			ID            string `json:"id"`
			Status        string `json:"status"`
			StatusDetail  string `json:"status_detail"`
			Amount        string `json:"amount"`
			PaymentMethod struct {
				ID   string `json:"id"`
				Type string `json:"type"`
			} `json:"payment_method"`
		} `json:"payments"`
	} `json:"transactions"`
}

func (o orderDetailResponse) toDetail() (OrderDetail, error) {
	total, err := DecimalStringToCents(o.TotalAmount)
	if err != nil {
		return OrderDetail{}, fmt.Errorf("total_amount: %w", err)
	}
	d := OrderDetail{
		ID:                o.ID,
		ExternalReference: o.ExternalReference,
		TotalAmountCents:  total,
	}
	for _, p := range o.Transactions.Payments {
		amt, err := DecimalStringToCents(p.Amount)
		if err != nil {
			return OrderDetail{}, err
		}
		pm := strings.ToLower(strings.TrimSpace(p.PaymentMethod.Type))
		if pm == "" {
			pm = strings.ToLower(strings.TrimSpace(p.PaymentMethod.ID))
		}
		d.Payments = append(d.Payments, OrderPaymentDetail{
			ID:            p.ID,
			Status:        strings.ToLower(strings.TrimSpace(p.Status)),
			StatusDetail:  strings.ToLower(strings.TrimSpace(p.StatusDetail)),
			AmountCents:   amt,
			PaymentMethod: pm,
		})
	}
	return d, nil
}
