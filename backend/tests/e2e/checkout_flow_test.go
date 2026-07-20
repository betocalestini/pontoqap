package e2e_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/store-platform/store/internal/app"
	"github.com/store-platform/store/internal/identity"
	identitypostgres "github.com/store-platform/store/internal/identity/postgres"
	"github.com/store-platform/store/internal/jobs"
	"github.com/store-platform/store/internal/platform/config"
	"github.com/store-platform/store/tests/testdb"
)

func TestHTTPCheckoutFlow(t *testing.T) {
	testdb.MigrateUp(t)
	pool := testdb.Pool(t)
	ctx := context.Background()
	if err := testdb.Reset(ctx, pool); err != nil {
		t.Fatal(err)
	}

	mgr, err := testdb.SeedManager(ctx, pool, testdb.UniqueEmail(t, "mgr"))
	if err != nil {
		t.Fatal(err)
	}
	prod, err := testdb.SeedProduct(ctx, pool, "Macarrão", "MAC-1", 1200)
	if err != nil {
		t.Fatal(err)
	}
	setupE2EInventory(t, ctx, pool, prod.SKUID, mgr.UserID)

	cfg := config.Config{
		AppEnv:   "development",
		LogLevel: "error",
		HTTP: config.HTTPConfig{
			CORSOrigins: []string{"http://localhost"},
		},
		Security: config.SecurityConfig{
			SessionSecret: "test-session-secret-min-16",
			CSRFSecret:    "test-csrf-secret-min-16-chars",
			EncryptionKey: "test-encryption-key-32-bytes!!",
		},
		Session: config.SessionConfig{
			StoreCookie: "store_session",
			AdminCookie: "admin_session",
			StoreTTL:    time.Hour,
			AdminTTL:    time.Hour,
		},
		Payments: config.PaymentsConfig{
			WebhookSecret: "test-webhook-secret",
		},
		App:      config.AppConfig{StoreWebURL: "http://localhost"},
		Customer: config.CustomerConfig{DefaultCreditLimitCents: 100_000},
	}
	idRepo := identitypostgres.NewRepository(pool)
	idSvc := identity.NewService(idRepo, cfg.Session.StoreTTL, cfg.Session.AdminTTL, cfg.Security.SessionSecret)
	jobRepo := jobs.NewRepository(pool)
	verifySvc := identity.NewVerificationService(pool, jobRepo, cfg.App, cfg.Customer)
	handler := app.NewRouter(cfg, pool, idSvc, verifySvc, slog.Default())
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	client := &http.Client{Timeout: 10 * time.Second}

	// Admin: aprovar cliente após registro na loja
	email := testdb.UniqueEmail(t, "cli")
	registerStoreCustomer(t, client, server.URL, email)
	verifyStoreCustomer(t, ctx, pool, email, 100_000)

	storeCookie := login(t, client, server.URL, email, "password123", "store")
	cartBody, _ := json.Marshal(map[string]any{"sku_id": prod.SKUID.String(), "quantity": 3})
	cartRes := doStoreJSON(t, client, http.MethodPost, server.URL+"/api/v1/me/cart/items", storeCookie, cartBody)
	if cartRes.StatusCode != http.StatusOK {
		t.Fatalf("cart: %d", cartRes.StatusCode)
	}
	cartRes.Body.Close()

	req, _ := http.NewRequest(http.MethodPost, server.URL+"/api/v1/me/cart/checkout", nil)
	req.Header.Set("Idempotency-Key", "e2e-checkout-1")
	req.Header.Set("X-App-Audience", "store")
	req.AddCookie(storeCookie)
	chRes, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if chRes.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(chRes.Body)
		t.Fatalf("checkout status=%d body=%s", chRes.StatusCode, body)
	}
	chRes.Body.Close()

	pubRes, err := client.Get(server.URL + "/api/v1/catalog/products")
	if err != nil {
		t.Fatal(err)
	}
	if pubRes.StatusCode != http.StatusOK {
		t.Fatalf("catalog: %d", pubRes.StatusCode)
	}
	pubRes.Body.Close()
}

func login(t *testing.T, client *http.Client, baseURL, email, password, audience string) *http.Cookie {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"email": email, "password": password, "audience": audience})
	req, _ := http.NewRequest(http.MethodPost, baseURL+"/api/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-App-Audience", audience)
	res, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res.Body)
		t.Fatalf("login %s: %d %s", audience, res.StatusCode, b)
	}
	for _, c := range res.Cookies() {
		if c.Name == "store_session" || c.Name == "admin_session" {
			return c
		}
	}
	t.Fatal("session cookie not found")
	return nil
}

func doAdminJSON(t *testing.T, client *http.Client, method, url string, cookie *http.Cookie, body []byte) *http.Response {
	return doJSONWithAudience(t, client, method, url, cookie, body, "admin")
}

func doStoreJSON(t *testing.T, client *http.Client, method, url string, cookie *http.Cookie, body []byte) *http.Response {
	return doJSONWithAudience(t, client, method, url, cookie, body, "store")
}

func doJSONWithAudience(t *testing.T, client *http.Client, method, url string, cookie *http.Cookie, body []byte, audience string) *http.Response {
	t.Helper()
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}
	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		t.Fatal(err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if audience != "" {
		req.Header.Set("X-App-Audience", audience)
	}
	if cookie != nil {
		req.AddCookie(cookie)
	}
	res, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func doJSON(t *testing.T, client *http.Client, method, url string, cookie *http.Cookie, body []byte) *http.Response {
	return doJSONWithAudience(t, client, method, url, cookie, body, "")
}
