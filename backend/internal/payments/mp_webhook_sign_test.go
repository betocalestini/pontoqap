package payments

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"
	"time"

	mpwebhook "github.com/mercadopago/sdk-go/pkg/webhook"
)

func TestSignMercadoPagoOrderWebhookHeaderRoundTrip(t *testing.T) {
	secret := "test-webhook-secret-homolog"
	dataID := "ORDTST01KY81SELFTEST"
	reqID := "req-selftest-1"
	ts := time.Now().UTC()
	sig := SignMercadoPagoOrderWebhookHeader(secret, reqID, dataID, ts)
	if err := mpwebhook.ValidateSignature(sig, reqID, strings.ToLower(dataID), secret); err != nil {
		t.Fatalf("ValidateSignature: %v", err)
	}
}

func TestSignMercadoPagoOrderWebhookHeaderUpperEqualsLower(t *testing.T) {
	secret := "test-webhook-secret-homolog"
	reqID := "req-case-1"
	ts := time.UnixMilli(1_700_000_000_000).UTC()
	upper := SignMercadoPagoOrderWebhookHeader(secret, reqID, "ORDTST01ABC", ts)
	lower := SignMercadoPagoOrderWebhookHeader(secret, reqID, "ordtst01abc", ts)
	if upper != lower {
		t.Fatalf("uppercase and lowercase data.id must produce the same x-signature\nupper=%s\nlower=%s", upper, lower)
	}
}

func TestValidateMercadoPagoWebhookSignatureAcceptsOriginalCasing(t *testing.T) {
	secret := "test-webhook-secret-homolog"
	dataID := "ORDTST01ORIGINALCASE"
	reqID := "req-original-1"
	tsStr := "1700000000100"
	manifest := "id:" + dataID + ";request-id:" + reqID + ";ts:" + tsStr + ";"
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(manifest))
	xSig := "ts=" + tsStr + ",v1=" + hex.EncodeToString(mac.Sum(nil))
	body := []byte(`{"type":"order","data":{"id":"` + dataID + `"}}`)
	used, err := validateMercadoPagoWebhookSignature(xSig, reqID, dataID, body, secret)
	if err != nil {
		t.Fatalf("expected original casing to validate: %v", err)
	}
	if used != dataID {
		t.Fatalf("used=%q want original %q", used, dataID)
	}
}

func TestValidateMercadoPagoWebhookSignatureAcceptsLowerCasing(t *testing.T) {
	secret := "test-webhook-secret-homolog"
	dataID := "ORDTST01LOWERCASE"
	reqID := "req-lower-1"
	ts := time.Now().UTC()
	sig := SignMercadoPagoOrderWebhookHeader(secret, reqID, dataID, ts)
	body := []byte(`{"type":"order","data":{"id":"` + dataID + `"}}`)
	used, err := validateMercadoPagoWebhookSignature(sig, reqID, dataID, body, secret)
	if err != nil {
		t.Fatalf("expected lowercasing path: %v", err)
	}
	if used != strings.ToLower(dataID) {
		t.Fatalf("used=%q", used)
	}
}

func TestMercadoPagoWebhookSignatureDataIDCandidates(t *testing.T) {
	body := []byte(`{"type":"order","data":{"id":"ORDTST01FROMBODY"}}`)
	got := mercadoPagoWebhookSignatureDataIDCandidates("ORDTST01FROMQUERY", body)
	if len(got) != 2 || got[0] != "ordtst01fromquery" || got[1] != "ORDTST01FROMQUERY" {
		t.Fatalf("got %#v", got)
	}
}

func TestIsMercadoPagoLookupableOrderID(t *testing.T) {
	if !isMercadoPagoLookupableOrderID("ORDTST01ABC") {
		t.Fatal("ORDTST should be lookupable")
	}
	if isMercadoPagoLookupableOrderID("123456") {
		t.Fatal("simulator fake id must not be lookupable")
	}
}
