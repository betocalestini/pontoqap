package payments

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// MercadoPagoWebhookCapture is a forensic dump of one inbound Order webhook (no secret).
type MercadoPagoWebhookCapture struct {
	RawQuery           string          `json:"raw_query"`
	DataIDQuery        string          `json:"data_id_query"`
	DataIDBody         string          `json:"data_id_body"`
	XSignature         string          `json:"x_signature"`
	XRequestID         string          `json:"x_request_id"`
	Host               string          `json:"host"`
	ContentLength      int64           `json:"content_length"`
	ApplicationID      string          `json:"application_id"`
	Body               json.RawMessage `json:"body"`
}

// DefaultMercadoPagoWebhookCapturePath is where the API writes the last capture in debug mode.
const DefaultMercadoPagoWebhookCapturePath = "/tmp/mp-webhook-last.json"

// BuildMercadoPagoWebhookCapture extracts forensic fields from an inbound webhook request.
func BuildMercadoPagoWebhookCapture(rawQuery, dataIDQuery, xSignature, xRequestID, host string, contentLength int64, body []byte) MercadoPagoWebhookCapture {
	return MercadoPagoWebhookCapture{
		RawQuery:      rawQuery,
		DataIDQuery:   strings.TrimSpace(dataIDQuery),
		DataIDBody:    mercadoPagoWebhookDataID("", body),
		XSignature:    strings.TrimSpace(xSignature),
		XRequestID:    strings.TrimSpace(xRequestID),
		Host:          host,
		ContentLength: contentLength,
		ApplicationID: mercadoPagoWebhookApplicationID(body),
		Body:          append(json.RawMessage(nil), body...),
	}
}

// WriteMercadoPagoWebhookCapture writes the capture JSON to path (dev/debug only).
func WriteMercadoPagoWebhookCapture(path string, cap MercadoPagoWebhookCapture) error {
	if path == "" {
		path = DefaultMercadoPagoWebhookCapturePath
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil && !os.IsExist(err) {
		// /tmp may not need mkdir; ignore if parent is root-like
		_ = err
	}
	b, err := json.MarshalIndent(cap, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o600)
}

// ManifestTrial is one HMAC candidate tried against a captured x-signature.
type ManifestTrial struct {
	DataID    string `json:"data_id"`
	RequestID string `json:"request_id"`
	TS        string `json:"ts"`
	Manifest  string `json:"manifest"`
	Match     bool   `json:"match"`
}

// VerifyMercadoPagoWebhookCapture tries known manifest variants against the capture and secret.
// Returns matching trials (may be empty) and all trials attempted.
func VerifyMercadoPagoWebhookCapture(cap MercadoPagoWebhookCapture, secret string) (matches []ManifestTrial, all []ManifestTrial) {
	secret = strings.TrimSpace(secret)
	ts, v1 := parseXSignatureParts(cap.XSignature)
	if ts == "" || v1 == "" {
		return nil, nil
	}

	dataIDs := uniqueNonEmpty(
		strings.ToLower(cap.DataIDQuery),
		cap.DataIDQuery,
		strings.ToLower(cap.DataIDBody),
		cap.DataIDBody,
	)
	if len(dataIDs) == 0 {
		dataIDs = []string{""}
	}
	reqIDs := uniqueNonEmpty(cap.XRequestID, "")
	if len(reqIDs) == 0 {
		reqIDs = []string{""}
	}

	for _, dataID := range dataIDs {
		for _, reqID := range reqIDs {
			manifest := buildWebhookManifest(dataID, reqID, ts)
			mac := hmac.New(sha256.New, []byte(secret))
			mac.Write([]byte(manifest))
			got := hex.EncodeToString(mac.Sum(nil))
			trial := ManifestTrial{
				DataID:    dataID,
				RequestID: reqID,
				TS:        ts,
				Manifest:  manifest,
				Match:     hmac.Equal([]byte(got), []byte(v1)),
			}
			all = append(all, trial)
			if trial.Match {
				matches = append(matches, trial)
			}
		}
	}
	return matches, all
}

func parseXSignatureParts(header string) (ts, v1 string) {
	for _, part := range strings.Split(header, ",") {
		rawKey, rawValue, ok := strings.Cut(part, "=")
		if !ok {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(rawKey))
		value := strings.TrimSpace(rawValue)
		switch key {
		case "ts":
			ts = value
		case "v1":
			v1 = value
		}
	}
	return ts, v1
}

func buildWebhookManifest(dataID, requestID, ts string) string {
	parts := make([]string, 0, 3)
	if dataID != "" {
		parts = append(parts, "id:"+dataID)
	}
	if requestID != "" {
		parts = append(parts, "request-id:"+requestID)
	}
	parts = append(parts, "ts:"+ts)
	return strings.Join(parts, ";") + ";"
}

func uniqueNonEmpty(vals ...string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(vals))
	for _, v := range vals {
		// allow explicit empty once (omit pair from manifest)
		if seen[v] {
			continue
		}
		seen[v] = true
		out = append(out, v)
	}
	return out
}

// FormatManifestTrials is a human-readable summary for CLI output.
func FormatManifestTrials(matches, all []ManifestTrial) string {
	var b strings.Builder
	b.WriteString("trials=" + strconv.Itoa(len(all)) + " matches=" + strconv.Itoa(len(matches)) + "\n")
	for i, t := range all {
		status := "MISS"
		if t.Match {
			status = "HIT"
		}
		fmt.Fprintf(&b, "%d. %s data_id=%q request_id=%q manifest=%q\n", i+1, status, t.DataID, t.RequestID, t.Manifest)
	}
	if len(matches) == 0 {
		b.WriteString("verdict: NONE — secret no .env nao assina este POST (ops: reset secret app correta modo teste)\n")
	} else {
		b.WriteString("verdict: HIT — ajustar validateMercadoPagoWebhookSignature para a variante HIT\n")
	}
	return b.String()
}
