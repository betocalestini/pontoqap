package mercadopago

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type createOrderRequest struct {
	Type              string              `json:"type"`
	TotalAmount       string              `json:"total_amount"`
	ExternalReference string              `json:"external_reference"`
	ProcessingMode    string              `json:"processing_mode"`
	Transactions      transactionPayload  `json:"transactions"`
	Payer             payerPayload        `json:"payer"`
}

type transactionPayload struct {
	Payments []paymentPayload `json:"payments"`
}

type paymentPayload struct {
	Amount         string               `json:"amount"`
	ExpirationTime string               `json:"expiration_time"`
	PaymentMethod  paymentMethodPayload `json:"payment_method"`
}

type paymentMethodPayload struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

type payerPayload struct {
	Email string `json:"email"`
}

type orderResponse struct {
	ID           string `json:"id"`
	Transactions struct {
		Payments []struct {
			ID               string `json:"id"`
			ReferenceID      string `json:"reference_id"`
			DateOfExpiration string `json:"date_of_expiration"`
			ExpirationTime   string `json:"expiration_time"`
			PaymentMethod    struct {
				QRCode       string `json:"qr_code"`
				ReferenceID  string `json:"reference_id"`
				Reference    string `json:"reference"`
			} `json:"payment_method"`
		} `json:"payments"`
	} `json:"transactions"`
}

func parseOrderResponse(raw []byte) (orderResponse, error) {
	var o orderResponse
	if err := json.Unmarshal(raw, &o); err != nil {
		return orderResponse{}, err
	}
	return o, nil
}

func chargeFromOrderResponse(o orderResponse, amountCents int64, expirationISO string) (ChargeResult, error) {
	if o.ID == "" {
		return ChargeResult{}, fmt.Errorf("mercado pago: resposta sem order id")
	}
	qr := ""
	txid := ""
	exp := expirationFromISO(expirationISO)
	if len(o.Transactions.Payments) > 0 {
		p := o.Transactions.Payments[0]
		qr = p.PaymentMethod.QRCode
		txid = firstNonEmpty(p.ReferenceID, p.PaymentMethod.ReferenceID, p.PaymentMethod.Reference, p.ID)
		if p.DateOfExpiration != "" {
			if t, err := time.Parse(time.RFC3339, p.DateOfExpiration); err == nil {
				exp = t
			}
		} else if p.ExpirationTime != "" {
			exp = expirationFromISO(p.ExpirationTime)
		}
	}
	if qr == "" {
		return ChargeResult{}, fmt.Errorf("mercado pago: resposta sem QR Code Pix")
	}
	return ChargeResult{
		ExternalID:  o.ID,
		TxID:        txid,
		QRCodeText:  qr,
		ExpiresAt:   exp,
		AmountCents: amountCents,
	}, nil
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
