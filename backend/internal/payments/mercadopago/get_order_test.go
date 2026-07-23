package mercadopago_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/store-platform/store/internal/payments/mercadopago"
)

func TestGatewayFetchOrder(t *testing.T) {
	const orderJSON = `{
	  "id": "ORD99",
	  "external_reference": "INSTALLMENT-x",
	  "total_amount": "350.00",
	  "transactions": {
	    "payments": [{
	      "id": "PAY99",
	      "status": "processed",
	      "status_detail": "accredited",
	      "amount": "350.00",
	      "payment_method": { "type": "bank_transfer" }
	    }]
	  }
	}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/orders/ORD99" {
			t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(orderJSON))
	}))
	defer srv.Close()

	cfg := mercadopago.Config{
		BaseURL:        srv.URL,
		AccessToken:    "test-token",
		RequestTimeout: 5 * time.Second,
	}
	gw, err := mercadopago.NewGateway(cfg, nil)
	if err != nil {
		t.Fatal(err)
	}
	order, err := gw.FetchOrder(context.Background(), "ORD99")
	if err != nil {
		t.Fatal(err)
	}
	if order.ID != "ORD99" || order.TotalAmountCents != 35000 {
		t.Fatalf("order: %+v", order)
	}
}
