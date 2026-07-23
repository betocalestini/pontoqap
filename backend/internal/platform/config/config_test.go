package config_test

import (
	"os"
	"testing"

	"github.com/store-platform/store/internal/platform/config"
)

func TestNormalizePaymentProvider(t *testing.T) {
	if got := config.NormalizePaymentProvider("mercado_pago"); got != "mercadopago" {
		t.Fatalf("got %q", got)
	}
}

func TestLoadRejectsTestAutoApproveInProduction(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("SESSION_SECRET", "long-enough-secret-key")
	t.Setenv("MERCADO_PAGO_TEST_AUTO_APPROVE", "true")
	t.Setenv("MERCADO_PAGO_ENVIRONMENT", "test")
	t.Setenv("PAYMENT_PROVIDER", "sandbox")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLoadRejectsWebhookDebugInProduction(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("SESSION_SECRET", "long-enough-secret-key")
	t.Setenv("PAYMENT_PROVIDER", "sandbox")
	t.Setenv("MERCADO_PAGO_WEBHOOK_DEBUG", "true")
	t.Setenv("MERCADO_PAGO_TEST_AUTO_APPROVE", "false")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error")
	}
	os.Unsetenv("MERCADO_PAGO_WEBHOOK_DEBUG")
}

func TestLoadRejectsTestAutoApproveOutsideTestEnvironment(t *testing.T) {
	t.Setenv("APP_ENV", "development")
	t.Setenv("SESSION_SECRET", "long-enough-secret-key")
	t.Setenv("MERCADO_PAGO_TEST_AUTO_APPROVE", "true")
	t.Setenv("MERCADO_PAGO_ENVIRONMENT", "production")
	t.Setenv("PAYMENT_PROVIDER", "sandbox")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLoadAllowsTestAutoApproveInDev(t *testing.T) {
	t.Setenv("APP_ENV", "development")
	t.Setenv("SESSION_SECRET", "long-enough-secret-key")
	t.Setenv("MERCADO_PAGO_TEST_AUTO_APPROVE", "true")
	t.Setenv("MERCADO_PAGO_ENVIRONMENT", "test")
	t.Setenv("PAYMENT_PROVIDER", "sandbox")

	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.Payments.MercadoPago.TestAutoApprove {
		t.Fatal("expected TestAutoApprove true")
	}
	os.Unsetenv("MERCADO_PAGO_TEST_AUTO_APPROVE")
}

func TestLoadRejectsEmptyMercadoPagoToken(t *testing.T) {
	t.Setenv("APP_ENV", "development")
	t.Setenv("SESSION_SECRET", "long-enough-secret-key")
	t.Setenv("PAYMENT_PROVIDER", "mercadopago")
	t.Setenv("MERCADO_PAGO_ACCESS_TOKEN", "")
	t.Setenv("MERCADO_PAGO_ENVIRONMENT", "test")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLoadAcceptsMercadoPagoTESTToken(t *testing.T) {
	t.Setenv("APP_ENV", "development")
	t.Setenv("SESSION_SECRET", "long-enough-secret-key")
	t.Setenv("PAYMENT_PROVIDER", "mercado_pago")
	t.Setenv("MERCADO_PAGO_ENVIRONMENT", "test")
	t.Setenv("MERCADO_PAGO_ACCESS_TOKEN", "TEST-example-token")
	t.Setenv("MERCADO_PAGO_TEST_AUTO_APPROVE", "true")

	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Payments.MercadoPago.AccessToken != "TEST-example-token" {
		t.Fatalf("token %q", cfg.Payments.MercadoPago.AccessToken)
	}
	os.Unsetenv("MERCADO_PAGO_TEST_AUTO_APPROVE")
}

func TestLoadAcceptsMercadoPagoAPPUSRToken(t *testing.T) {
	t.Setenv("APP_ENV", "development")
	t.Setenv("SESSION_SECRET", "long-enough-secret-key")
	t.Setenv("PAYMENT_PROVIDER", "mercadopago")
	t.Setenv("MERCADO_PAGO_ENVIRONMENT", "test")
	t.Setenv("MERCADO_PAGO_ACCESS_TOKEN", "APP_USR-example-token")

	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Payments.MercadoPago.AccessToken != "APP_USR-example-token" {
		t.Fatalf("token %q", cfg.Payments.MercadoPago.AccessToken)
	}
}

func TestLoadDoesNotRequireMercadoPagoTokenForSandbox(t *testing.T) {
	t.Setenv("APP_ENV", "development")
	t.Setenv("SESSION_SECRET", "long-enough-secret-key")
	t.Setenv("PAYMENT_PROVIDER", "sandbox")
	t.Setenv("MERCADO_PAGO_ACCESS_TOKEN", "")

	_, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
}
