package payments

import (
	"encoding/json"
	"fmt"
	"strings"

	mpwebhook "github.com/mercadopago/sdk-go/pkg/webhook"
)

type mercadoPagoOrderWebhook struct {
	Type   string `json:"type"`
	Action string `json:"action"`
	Data   struct {
		ID string `json:"id"`
	} `json:"data"`
}

// mercadoPagoWebhookDataID returns the data.id from query (preferred) or body.
// MP signs using the query parameter when present; some proxies strip the query string.
func mercadoPagoWebhookDataID(queryDataID string, body []byte) string {
	queryDataID = strings.TrimSpace(queryDataID)
	if queryDataID != "" {
		return queryDataID
	}
	var probe struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &probe); err != nil {
		return ""
	}
	return strings.TrimSpace(probe.Data.ID)
}

// mercadoPagoWebhookSignatureDataIDCandidates returns data.id variants for HMAC.
// Official Orders docs require lowercase; sdk-go PR claims MP may sign with original casing.
func mercadoPagoWebhookSignatureDataIDCandidates(queryDataID string, body []byte) []string {
	raw := mercadoPagoWebhookDataID(queryDataID, body)
	if raw == "" {
		return []string{""}
	}
	lower := strings.ToLower(raw)
	if lower == raw {
		return []string{raw}
	}
	return []string{lower, raw}
}

// validateMercadoPagoWebhookSignature tries each data.id casing until HMAC matches.
// Also retries without request-id (MP docs: omit missing pairs) in case a proxy rewrites x-request-id.
func validateMercadoPagoWebhookSignature(xSignature, xRequestID, queryDataID string, body []byte, secret string) (usedDataID string, err error) {
	candidates := mercadoPagoWebhookSignatureDataIDCandidates(queryDataID, body)
	reqIDs := []string{strings.TrimSpace(xRequestID)}
	if reqIDs[0] != "" {
		reqIDs = append(reqIDs, "") // try omitting request-id from manifest
	}
	var lastErr error
	for _, id := range candidates {
		for _, rid := range reqIDs {
			if err := mpwebhook.ValidateSignature(xSignature, rid, id, secret); err == nil {
				return id, nil
			} else {
				lastErr = err
			}
		}
	}
	return candidates[0], lastErr
}

func mercadoPagoWebhookApplicationID(body []byte) string {
	var probe struct {
		ApplicationID json.Number `json:"application_id"`
	}
	if err := json.Unmarshal(body, &probe); err != nil {
		return ""
	}
	return strings.TrimSpace(probe.ApplicationID.String())
}

// isMercadoPagoLookupableOrderID reports whether orderID can be fetched via GET /v1/orders/{id}.
// The MP panel simulator uses a fake numeric id (e.g. "123456") that must not enqueue settlement jobs.
func isMercadoPagoLookupableOrderID(orderID string) bool {
	id := strings.TrimSpace(orderID)
	if id == "" {
		return false
	}
	return strings.HasPrefix(strings.ToUpper(id), "ORD")
}

func parseMercadoPagoOrderWebhook(body []byte) (mercadoPagoOrderWebhook, error) {
	var p mercadoPagoOrderWebhook
	if err := json.Unmarshal(body, &p); err != nil {
		return mercadoPagoOrderWebhook{}, err
	}
	if p.Type != "order" {
		return mercadoPagoOrderWebhook{}, fmt.Errorf("tipo de notificação inválido")
	}
	if p.Data.ID == "" {
		return mercadoPagoOrderWebhook{}, fmt.Errorf("data.id ausente")
	}
	return p, nil
}
