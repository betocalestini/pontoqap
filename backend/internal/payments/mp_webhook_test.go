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

func TestMercadoPagoWebhookDataID(t *testing.T) {
	body := []byte(`{"type":"order","data":{"id":"ORDTST01"}}`)
	if got := mercadoPagoWebhookDataID("from-query", body); got != "from-query" {
		t.Fatalf("query wins: %q", got)
	}
	if got := mercadoPagoWebhookDataID("", body); got != "ORDTST01" {
		t.Fatalf("body fallback: %q", got)
	}
}
