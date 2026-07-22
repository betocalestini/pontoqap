package payments

import (
	"testing"
)

func TestParseMercadoPagoOrderWebhook(t *testing.T) {
	_, err := parseMercadoPagoOrderWebhook([]byte(`{"type":"payment"}`))
	if err == nil {
		t.Fatal("expected error for wrong type")
	}
	p, err := parseMercadoPagoOrderWebhook([]byte(`{"type":"order","data":{"id":"ORD1"}}`))
	if err != nil {
		t.Fatal(err)
	}
	if p.Data.ID != "ORD1" {
		t.Fatalf("id: %q", p.Data.ID)
	}
}
