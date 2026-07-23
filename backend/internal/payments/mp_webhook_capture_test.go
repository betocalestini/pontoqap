package payments

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestVerifyMercadoPagoWebhookCaptureHitsLowercase(t *testing.T) {
	secret := "forensic-secret-64chars-abcdefghijklmnopqrstuvwxyz012345"
	dataID := "ORDTST01FORENSIC"
	reqID := "req-forensic-1"
	ts := time.UnixMilli(1_700_000_000_000).UTC()
	sig := SignMercadoPagoOrderWebhookHeader(secret, reqID, dataID, ts)

	cap := MercadoPagoWebhookCapture{
		DataIDQuery: dataID,
		DataIDBody:  dataID,
		XSignature:  sig,
		XRequestID:  reqID,
	}
	matches, all := VerifyMercadoPagoWebhookCapture(cap, secret)
	if len(matches) == 0 {
		t.Fatalf("expected HIT, trials=%s", FormatManifestTrials(matches, all))
	}
}

func TestVerifyMercadoPagoWebhookCaptureHitsOriginalCasing(t *testing.T) {
	secret := "forensic-secret-64chars-abcdefghijklmnopqrstuvwxyz012345"
	dataID := "ORDTST01ORIGINAL"
	reqID := "req-forensic-2"
	tsStr := "1700000000200"
	manifest := "id:" + dataID + ";request-id:" + reqID + ";ts:" + tsStr + ";"
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(manifest))
	sig := "ts=" + tsStr + ",v1=" + hex.EncodeToString(mac.Sum(nil))

	cap := MercadoPagoWebhookCapture{
		DataIDQuery: dataID,
		XSignature:  sig,
		XRequestID:  reqID,
	}
	matches, all := VerifyMercadoPagoWebhookCapture(cap, secret)
	if len(matches) == 0 {
		t.Fatalf("expected HIT on original casing, trials=%s", FormatManifestTrials(matches, all))
	}
	foundOriginal := false
	for _, m := range matches {
		if m.DataID == dataID {
			foundOriginal = true
		}
	}
	if !foundOriginal {
		t.Fatalf("expected match with original data.id, got %#v", matches)
	}
}

func TestVerifyMercadoPagoWebhookCaptureMissWrongSecret(t *testing.T) {
	secret := "forensic-secret-64chars-abcdefghijklmnopqrstuvwxyz012345"
	dataID := "ORDTST01MISS"
	reqID := "req-forensic-3"
	ts := time.UnixMilli(1_700_000_000_300).UTC()
	sig := SignMercadoPagoOrderWebhookHeader(secret, reqID, dataID, ts)

	cap := MercadoPagoWebhookCapture{
		DataIDQuery: dataID,
		XSignature:  sig,
		XRequestID:  reqID,
	}
	matches, _ := VerifyMercadoPagoWebhookCapture(cap, "wrong-secret")
	if len(matches) != 0 {
		t.Fatalf("expected NONE, got %#v", matches)
	}
}

func TestWriteMercadoPagoWebhookCapture(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cap.json")
	cap := BuildMercadoPagoWebhookCapture(
		"data.id=ORD1&type=order",
		"ORD1",
		"ts=1,v1=abc",
		"rid-1",
		"example.com",
		12,
		[]byte(`{"type":"order","application_id":"3962","data":{"id":"ORD1"}}`),
	)
	if err := WriteMercadoPagoWebhookCapture(path, cap); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}
	if cap.ApplicationID != "3962" {
		t.Fatalf("application_id=%q", cap.ApplicationID)
	}
	if cap.DataIDBody != "ORD1" {
		t.Fatalf("data_id_body=%q", cap.DataIDBody)
	}
}
