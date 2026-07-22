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

func TestReportsDashboardMatchesSalesTotal(t *testing.T) {
	testdb.MigrateUp(t)
	pool := testdb.Pool(t)
	ctx := context.Background()
	if err := testdb.Reset(ctx, pool); err != nil {
		t.Fatal(err)
	}

	mgr, err := testdb.SeedManager(ctx, pool, testdb.UniqueEmail(t, "mgr-rpt"))
	if err != nil {
		t.Fatal(err)
	}
	cust, err := testdb.SeedCustomer(ctx, pool, testdb.UniqueEmail(t, "buy-rpt"), "Comprador")
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
		VALUES ($1, $2, $3, 'confirmed', 5000, 0, 5000, $4, $5)
	`, orderID, "RPT-001", cust.ID, orderID.String(), now); err != nil {
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

	y, m := now.Year(), int(now.Month())

	dashReq := httptest.NewRequest(http.MethodGet, "/api/v1/admin/reports/dashboard", nil)
	dashReq.Header.Set("X-App-Audience", "admin")
	dashReq.Header.Set("Authorization", "Bearer "+token)
	dashRec := httptest.NewRecorder()
	handler.ServeHTTP(dashRec, dashReq)
	if dashRec.Code != http.StatusOK {
		t.Fatalf("dashboard: %d %s", dashRec.Code, dashRec.Body.String())
	}
	var dash struct {
		SalesMonthCents int64 `json:"sales_month_cents"`
		OrdersMonth     int64 `json:"orders_month"`
	}
	if err := json.NewDecoder(dashRec.Body).Decode(&dash); err != nil {
		t.Fatal(err)
	}

	salesURL := "/api/v1/admin/reports/sales/orders?year=" + itoa(y) + "&month=" + itoa(m)
	salesReq := httptest.NewRequest(http.MethodGet, salesURL, nil)
	salesReq.Header.Set("X-App-Audience", "admin")
	salesReq.Header.Set("Authorization", "Bearer "+token)
	salesRec := httptest.NewRecorder()
	handler.ServeHTTP(salesRec, salesReq)
	if salesRec.Code != http.StatusOK {
		t.Fatalf("sales: %d %s", salesRec.Code, salesRec.Body.String())
	}
	var sales struct {
		Summary struct {
			TotalSalesCents int64 `json:"total_sales_cents"`
			OrderCount      int   `json:"order_count"`
		} `json:"summary"`
	}
	if err := json.NewDecoder(salesRec.Body).Decode(&sales); err != nil {
		t.Fatal(err)
	}

	if dash.SalesMonthCents != sales.Summary.TotalSalesCents {
		t.Fatalf("sales mismatch dashboard=%d report=%d", dash.SalesMonthCents, sales.Summary.TotalSalesCents)
	}
	if int(dash.OrdersMonth) != sales.Summary.OrderCount {
		t.Fatalf("orders count mismatch dashboard=%d report=%d", dash.OrdersMonth, sales.Summary.OrderCount)
	}
}

func TestReportsDashboardEmptyMonthReturnsArraysNotNull(t *testing.T) {
	testdb.MigrateUp(t)
	pool := testdb.Pool(t)
	ctx := context.Background()
	if err := testdb.Reset(ctx, pool); err != nil {
		t.Fatal(err)
	}

	mgr, err := testdb.SeedManager(ctx, pool, testdb.UniqueEmail(t, "mgr-rpt-empty"))
	if err != nil {
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

	now := time.Now()
	futureMonth := int(now.Month()) + 1
	year := now.Year()
	if futureMonth > 12 {
		futureMonth = 1
		year++
	}

	url := "/api/v1/admin/reports/dashboard?year=" + itoa(year) + "&month=" + itoa(futureMonth)
	req := httptest.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("X-App-Audience", "admin")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("dashboard: %d %s", rec.Code, rec.Body.String())
	}

	var raw map[string]json.RawMessage
	if err := json.NewDecoder(rec.Body).Decode(&raw); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"top_products", "top_customers"} {
		v, ok := raw[key]
		if !ok {
			t.Fatalf("missing %s", key)
		}
		if string(v) == "null" {
			t.Fatalf("%s must not be null", key)
		}
		var arr []json.RawMessage
		if err := json.Unmarshal(v, &arr); err != nil {
			t.Fatalf("%s: %v", key, err)
		}
		if len(arr) != 0 {
			t.Fatalf("%s: expected empty array, got %d items", key, len(arr))
		}
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [12]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}
