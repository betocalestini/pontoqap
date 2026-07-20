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

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/store-platform/store/internal/app"
	"github.com/store-platform/store/internal/identity"
	identitypostgres "github.com/store-platform/store/internal/identity/postgres"
	"github.com/store-platform/store/internal/inventory"
	"github.com/store-platform/store/internal/platform/config"
)

func newE2EServer(t *testing.T, pool *pgxpool.Pool) (*httptest.Server, *http.Client) {
	t.Helper()
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
	}
	idRepo := identitypostgres.NewRepository(pool)
	idSvc := identity.NewService(idRepo, cfg.Session.StoreTTL, cfg.Session.AdminTTL)
	handler := app.NewRouter(cfg, pool, idSvc, slog.Default())
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return server, &http.Client{Timeout: 15 * time.Second}
}

func setupE2EInventory(t *testing.T, ctx context.Context, pool *pgxpool.Pool, skuID uuid.UUID, managerID uuid.UUID) {
	t.Helper()
	inv := inventory.NewService(pool)
	if err := inv.RegisterEntry(ctx, skuID, 20, managerID, "entrada"); err != nil {
		t.Fatal(err)
	}
}

func registerStoreCustomer(t *testing.T, client *http.Client, baseURL, email string) {
	t.Helper()
	regBody, _ := json.Marshal(map[string]string{
		"name": "Cliente E2E", "email": email, "password": "password123",
	})
	regRes, err := client.Post(baseURL+"/api/v1/customers/register", "application/json", bytes.NewReader(regBody))
	if err != nil || regRes.StatusCode != http.StatusCreated {
		t.Fatalf("register: %v status=%d", err, regRes.StatusCode)
	}
	regRes.Body.Close()
}

func findCustomerID(t *testing.T, client *http.Client, baseURL string, adminCookie *http.Cookie, email string) string {
	t.Helper()
	var custList struct {
		Items []struct {
			ID    string `json:"id"`
			Email string `json:"email"`
		} `json:"items"`
	}
	listRes := doAdminJSON(t, client, http.MethodGet, baseURL+"/api/v1/admin/customers", adminCookie, nil)
	if listRes.StatusCode != http.StatusOK {
		t.Fatalf("list customers: %d", listRes.StatusCode)
	}
	_ = json.NewDecoder(listRes.Body).Decode(&custList)
	listRes.Body.Close()
	for _, c := range custList.Items {
		if c.Email == email {
			return c.ID
		}
	}
	t.Fatal("customer not found")
	return ""
}

func approveCustomer(t *testing.T, client *http.Client, baseURL string, adminCookie *http.Cookie, customerID string, limit int64) {
	t.Helper()
	apBody, _ := json.Marshal(map[string]int64{"credit_limit_cents": limit})
	apRes := doAdminJSON(t, client, http.MethodPatch, baseURL+"/api/v1/admin/customers/"+customerID+"/approve", adminCookie, apBody)
	if apRes.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(apRes.Body)
		t.Fatalf("approve: %d %s", apRes.StatusCode, body)
	}
	apRes.Body.Close()
}
