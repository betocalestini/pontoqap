package integration_test

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/store-platform/store/internal/app"
	"github.com/store-platform/store/internal/identity"
	identitypostgres "github.com/store-platform/store/internal/identity/postgres"
	"github.com/store-platform/store/internal/platform/config"
	"github.com/store-platform/store/tests/testdb"
)

func TestDashboardSeriesMonths(t *testing.T) {
	testdb.MigrateUp(t)
	pool := testdb.Pool(t)
	ctx := context.Background()
	if err := testdb.Reset(ctx, pool); err != nil {
		t.Fatal(err)
	}

	mgr, err := testdb.SeedManager(ctx, pool, testdb.UniqueEmail(t, "mgr-series"))
	if err != nil {
		t.Fatal(err)
	}
	cust, err := testdb.SeedCustomer(ctx, pool, testdb.UniqueEmail(t, "buy-series"), "Comprador")
	if err != nil {
		t.Fatal(err)
	}
	if err := testdb.ApproveCustomer(ctx, pool, cust.ID, mgr.UserID, 100_000); err != nil {
		t.Fatal(err)
	}

	now := time.Now().UTC()
	orderID := uuid.New()
	if _, err := pool.Exec(ctx, `
		INSERT INTO orders (id, order_number, customer_id, status, subtotal_cents, discount_cents, total_cents, idempotency_key, confirmed_at)
		VALUES ($1, $2, $3, 'confirmed', 3000, 0, 3000, $4, $5)
	`, orderID, "SER-001", cust.ID, orderID.String(), now); err != nil {
		t.Fatal(err)
	}

	secret := "test-session-secret-min-16"
	idSvc := identity.NewService(identitypostgres.NewRepository(pool), time.Hour, 8*time.Hour, secret)
	cfg := config.Config{
		AppEnv:   "test",
		Security: config.SecurityConfig{SessionSecret: secret, AdminMFARequired: false},
		Session:  config.SessionConfig{StoreTTL: time.Hour, AdminTTL: 8 * time.Hour},
		HTTP:     config.HTTPConfig{CORSOrigins: []string{"*"}},
	}
	handler := app.NewRouter(cfg, pool, idSvc, nil, slog.Default())
	token := adminLoginToken(t, handler, mgr.Email, "password123")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/reports/dashboard/series?months=6", nil)
	req.Header.Set("X-App-Audience", "admin")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("series: %d %s", rec.Code, rec.Body.String())
	}
	var body struct {
		Items []struct {
			Year        int   `json:"year"`
			Month       int   `json:"month"`
			SalesCents  int64 `json:"sales_cents"`
		} `json:"items"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if len(body.Items) != 6 {
		t.Fatalf("want 6 months got %d", len(body.Items))
	}
	found := false
	y, m := now.Year(), int(now.Month())
	for _, pt := range body.Items {
		if pt.Year == y && pt.Month == m && pt.SalesCents >= 3000 {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected sales in current month in series: %+v", body.Items)
	}
}
