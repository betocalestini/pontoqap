package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/store-platform/store/internal/app"
	"github.com/store-platform/store/internal/billing"
	"github.com/store-platform/store/internal/catalog"
	"github.com/store-platform/store/internal/customers"
	"github.com/store-platform/store/internal/identity"
	identitypostgres "github.com/store-platform/store/internal/identity/postgres"
	"github.com/store-platform/store/internal/inventory"
	"github.com/store-platform/store/internal/platform/config"
	"github.com/store-platform/store/internal/sales"
	"github.com/store-platform/store/tests/testdb"
)

func TestAdminCancelOrderRequiresPassword(t *testing.T) {
	testdb.MigrateUp(t)
	pool := testdb.Pool(t)
	ctx := context.Background()
	if err := testdb.Reset(ctx, pool); err != nil {
		t.Fatal(err)
	}

	email := testdb.UniqueEmail(t, "cancel-order")
	cust, err := testdb.SeedCustomer(ctx, pool, email, "Comprador")
	if err != nil {
		t.Fatal(err)
	}
	mgr, err := testdb.SeedManager(ctx, pool, testdb.UniqueEmail(t, "mgr-cancel"))
	if err != nil {
		t.Fatal(err)
	}
	if err := testdb.ApproveCustomer(ctx, pool, cust.ID, mgr.UserID, 500_000); err != nil {
		t.Fatal(err)
	}

	invSvc := inventory.NewService(pool)
	catSvc := catalog.NewService(pool)
	billSvc := billing.NewService(pool, nil, "http://localhost:5173", nil)
	custSvc := customers.NewService(pool, nil)
	salesSvc := sales.NewService(pool, invSvc, billSvc, catSvc, custSvc, nil)

	prod, err := testdb.SeedProduct(ctx, pool, "Prod Cancel", "SKU-CAN", 2000)
	if err != nil {
		t.Fatal(err)
	}
	if err := invSvc.RegisterEntry(ctx, prod.SKUID, 5, mgr.UserID, "entrada", 0, 0); err != nil {
		t.Fatal(err)
	}
	if _, err := salesSvc.AddCartItem(ctx, cust.ID, prod.SKUID, 1); err != nil {
		t.Fatal(err)
	}
	order, err := salesSvc.Checkout(ctx, cust.ID, "cancel-order-key", cust.UserID)
	if err != nil {
		t.Fatal(err)
	}

	secret := "test-session-secret-min-16"
	idSvc := identity.NewService(identitypostgres.NewRepository(pool), time.Hour, 8*time.Hour, secret, nil)
	cfg := config.Config{
		AppEnv:   "test",
		Security: config.SecurityConfig{SessionSecret: secret, AdminMFARequired: false},
		Session:  config.SessionConfig{StoreTTL: time.Hour, AdminTTL: 8 * time.Hour},
		HTTP:     config.HTTPConfig{CORSOrigins: []string{"*"}},
	}
	handler := app.NewRouter(cfg, pool, idSvc, nil, slog.Default())
	token := adminLoginToken(t, handler, mgr.Email, "password123")

	cancelURL := "/api/v1/admin/orders/" + order.ID.String() + "/cancel"

	req := httptest.NewRequest(http.MethodPost, cancelURL, bytes.NewReader([]byte(`{"password":"wrong"}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-App-Audience", "admin")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("want 403 got %d body %s", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodPost, cancelURL, bytes.NewReader([]byte(`{"password":"password123"}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-App-Audience", "admin")
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200 got %d body %s", rec.Code, rec.Body.String())
	}
	var cancelled struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &cancelled); err != nil {
		t.Fatal(err)
	}
	if cancelled.Status != "cancelled" {
		t.Fatalf("want cancelled got %s", cancelled.Status)
	}
}
