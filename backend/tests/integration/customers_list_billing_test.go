package integration_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/store-platform/store/internal/app"
	"github.com/store-platform/store/internal/identity"
	identitypostgres "github.com/store-platform/store/internal/identity/postgres"
	"github.com/store-platform/store/internal/platform/config"
	"github.com/store-platform/store/tests/testdb"
	"log/slog"
	"time"
)

func TestAdminListCustomersIncludesInvoiceCounts(t *testing.T) {
	testdb.MigrateUp(t)
	pool := testdb.Pool(t)
	ctx := context.Background()
	if err := testdb.Reset(ctx, pool); err != nil {
		t.Fatal(err)
	}
	email := testdb.UniqueEmail(t, "cust-bill")
	cust, err := testdb.SeedCustomer(ctx, pool, email, "Cliente Billing")
	if err != nil {
		t.Fatal(err)
	}
	_ = cust

	secret := "test-session-secret-min-16"
	idSvc := identity.NewService(identitypostgres.NewRepository(pool), time.Hour, 8*time.Hour, secret, nil)
	cfg := config.Config{
		AppEnv: "test",
		Security: config.SecurityConfig{SessionSecret: secret, AdminMFARequired: false},
		Session:  config.SessionConfig{StoreTTL: time.Hour, AdminTTL: 8 * time.Hour},
		HTTP:     config.HTTPConfig{CORSOrigins: []string{"*"}},
	}
	handler := app.NewRouter(cfg, pool, idSvc, nil, slog.Default())

	mgrEmail := testdb.UniqueEmail(t, "mgr-list")
	_, err = testdb.SeedManager(ctx, pool, mgrEmail)
	if err != nil {
		t.Fatal(err)
	}
	token := adminLoginToken(t, handler, mgrEmail, "password123")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/customers", nil)
	req.Header.Set("X-App-Audience", "admin")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
	var body struct {
		Items []struct {
			Email                string `json:"email"`
			OpenInvoicesCount    int    `json:"open_invoices_count"`
			OverdueInvoicesCount int    `json:"overdue_invoices_count"`
		} `json:"items"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	var found bool
	for _, it := range body.Items {
		if it.Email == email {
			found = true
			if it.OpenInvoicesCount != 0 || it.OverdueInvoicesCount != 0 {
				t.Fatalf("expected zero invoice counts, got open=%d overdue=%d", it.OpenInvoicesCount, it.OverdueInvoicesCount)
			}
		}
	}
	if !found {
		t.Fatal("seeded customer not in list")
	}
}
