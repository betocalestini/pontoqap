package payments

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// SignMercadoPagoOrderWebhookHeader builds an x-signature value accepted by mpwebhook.ValidateSignature.
// Alphanumeric Order data.id is lowercased in the manifest (MP Orders webhook docs).
func SignMercadoPagoOrderWebhookHeader(secret, xRequestID, dataID string, ts time.Time) string {
	secret = strings.TrimSpace(secret)
	dataID = strings.ToLower(strings.TrimSpace(dataID))
	xRequestID = strings.TrimSpace(xRequestID)
	tsMs := ts.UnixMilli()
	tsStr := strconv.FormatInt(tsMs, 10)
	manifest := fmt.Sprintf("id:%s;request-id:%s;ts:%s;", dataID, xRequestID, tsStr)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(manifest))
	return fmt.Sprintf("ts=%s,v1=%s", tsStr, hex.EncodeToString(mac.Sum(nil)))
}
