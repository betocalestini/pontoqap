package mercadopago_test

import (
	"testing"

	"github.com/store-platform/store/internal/payments/mercadopago"
)

func TestBuildPayerAPRO(t *testing.T) {
	cfg := mercadopago.Config{Environment: "test", TestAutoApprove: true}
	p := mercadopago.BuildPayerForTest(cfg, "a@b.com")
	if p.Email != "test_user_br@testuser.com" || p.FirstName != "APRO" {
		t.Fatalf("got %+v", p)
	}
}

func TestBuildPayerNormal(t *testing.T) {
	cfg := mercadopago.Config{Environment: "test", TestAutoApprove: false}
	p := mercadopago.BuildPayerForTest(cfg, "a@b.com")
	if p.Email != "a@b.com" || p.FirstName != "" {
		t.Fatalf("got %+v", p)
	}
}
