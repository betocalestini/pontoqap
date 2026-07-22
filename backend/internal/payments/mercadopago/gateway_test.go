package mercadopago_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/store-platform/store/internal/payments/mercadopago"
)

func TestGatewayCreatePixCharge(t *testing.T) {
	const orderJSON = `{
	  "id": "ORD01TEST",
	  "transactions": {
	    "payments": [{
	      "id": "PAY01",
	      "reference_id": "REF-PIX-123",
	      "date_of_expiration": "2030-01-15T12:00:00Z",
	      "payment_method": {
	        "qr_code": "00020126580014br.gov.bcb.pix",
	        "reference_id": "REF-PIX-123"
	      }
	    }]
	  }
	}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/orders" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Fatalf("missing bearer token")
		}
		if r.Header.Get("X-Idempotency-Key") == "" {
			t.Fatalf("missing idempotency key")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(orderJSON))
	}))
	defer srv.Close()

	cfg := mercadopago.Config{
		BaseURL:        srv.URL,
		AccessToken:    "test-token",
		PixExpiration:  "PT24H",
		RequestTimeout: 5 * time.Second,
	}
	gw, err := mercadopago.NewGateway(cfg, nil)
	if err != nil {
		t.Fatal(err)
	}
	res, err := gw.CreatePixCharge(context.Background(), mercadopago.PixChargeInput{
		InvoiceID:         uuid.New(),
		AmountCents:       35050,
		ExternalReference: "INV-2026-000123",
		PayerEmail:        "cliente@exemplo.com",
		IdempotencyKey:    uuid.NewString(),
		ExpirationISO:     "PT24H",
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.ExternalID != "ORD01TEST" {
		t.Fatalf("external id: %q", res.ExternalID)
	}
	if res.QRCodeText == "" {
		t.Fatal("expected qr code")
	}
	if res.AmountCents != 35050 {
		t.Fatalf("amount: %d", res.AmountCents)
	}
}

func TestParseOrderWebhookPayload(t *testing.T) {
	body, _ := json.Marshal(map[string]any{
		"type":   "order",
		"action": "order.action_required",
		"data":   map[string]string{"id": "ORD01"},
	})
	p, err := mercadopago.ParseOrderWebhookPayload(body)
	if err != nil {
		t.Fatal(err)
	}
	if p.Data.ID != "ORD01" {
		t.Fatalf("data.id: %q", p.Data.ID)
	}
}
