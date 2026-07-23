// mp-webhook-check assina um POST Order, envia para a API local, ou verifica uma captura forense.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"

	"github.com/store-platform/store/internal/payments"
)

func main() {
	_ = godotenv.Load()
	_ = godotenv.Load("../.env")

	baseURL := flag.String("url", "http://localhost:8080", "API base URL (sem path)")
	dataID := flag.String("data-id", "ORDTST01HOMOLOGCHECK", "data.id query + manifest")
	post := flag.Bool("post", false, "enviar POST para o webhook (senão só imprime headers)")
	verifyCapture := flag.String("verify-capture", "", "caminho JSON de captura (ex. /tmp/mp-webhook-last.json)")
	flag.Parse()

	secret := strings.TrimSpace(os.Getenv("MERCADO_PAGO_WEBHOOK_SECRET"))
	if secret == "" {
		fmt.Fprintln(os.Stderr, "MERCADO_PAGO_WEBHOOK_SECRET vazio")
		os.Exit(1)
	}

	if path := strings.TrimSpace(*verifyCapture); path != "" {
		raw, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		var cap payments.MercadoPagoWebhookCapture
		if err := json.Unmarshal(raw, &cap); err != nil {
			fmt.Fprintln(os.Stderr, "capture JSON inválido:", err)
			os.Exit(1)
		}
		matches, all := payments.VerifyMercadoPagoWebhookCapture(cap, secret)
		fmt.Print(payments.FormatManifestTrials(matches, all))
		if len(matches) == 0 {
			os.Exit(2)
		}
		return
	}

	reqID := uuid.NewString()
	ts := time.Now().UTC()
	sig := payments.SignMercadoPagoOrderWebhookHeader(secret, reqID, *dataID, ts)
	body, _ := json.Marshal(map[string]any{
		"type":   "order",
		"action": "order.processed",
		"data":   map[string]string{"id": *dataID},
	})

	if !*post {
		fmt.Printf("x-request-id: %s\n", reqID)
		fmt.Printf("x-signature: %s\n", sig)
		fmt.Printf("query: data.id=%s\n", *dataID)
		fmt.Println("OK (assinatura gerada; use -post para testar a API, -verify-capture para forense)")
		return
	}

	u := strings.TrimRight(*baseURL, "/") + "/api/v1/webhooks/mercado-pago/orders?data.id=" + *dataID
	req, err := http.NewRequest(http.MethodPost, u, bytes.NewReader(body))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-request-id", reqID)
	req.Header.Set("x-signature", sig)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	fmt.Printf("HTTP %d\n", resp.StatusCode)
	if len(b) > 0 {
		fmt.Println(string(b))
	}
	if resp.StatusCode != http.StatusOK {
		os.Exit(1)
	}
}
