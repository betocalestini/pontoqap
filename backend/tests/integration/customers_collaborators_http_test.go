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

	"github.com/google/uuid"

	"github.com/store-platform/store/internal/app"
	"github.com/store-platform/store/internal/customers"
	"github.com/store-platform/store/internal/identity"
	identitypostgres "github.com/store-platform/store/internal/identity/postgres"
	"github.com/store-platform/store/internal/platform/config"
	"github.com/store-platform/store/tests/testdb"
)

func TestBlockedStoreCustomerCannotLogin(t *testing.T) {
	testdb.MigrateUp(t)
	pool := testdb.Pool(t)
	ctx := context.Background()
	if err := testdb.Reset(ctx, pool); err != nil {
		t.Fatal(err)
	}
	email := testdb.UniqueEmail(t, "blocked-login")
	cust, err := testdb.SeedCustomer(ctx, pool, email, "Bloqueado")
	if err != nil {
		t.Fatal(err)
	}
	mgr, _ := testdb.SeedManager(ctx, pool, testdb.UniqueEmail(t, "mgr"))
	_ = testdb.ApproveCustomer(ctx, pool, cust.ID, mgr.UserID, 100_000)
	custSvc := customers.NewService(pool, nil)
	if err := custSvc.Block(ctx, cust.ID, "teste"); err != nil {
		t.Fatal(err)
	}

	secret := "test-session-secret-min-16"
	idSvc := identity.NewService(identitypostgres.NewRepository(pool), time.Hour, 8*time.Hour, secret)
	cfg := config.Config{
		AppEnv: "test",
		Security: config.SecurityConfig{SessionSecret: secret, AdminMFARequired: false},
		Session: config.SessionConfig{
			StoreCookie: "store_session",
			AdminCookie: "admin_session",
			StoreTTL:    time.Hour,
			AdminTTL:    8 * time.Hour,
		},
		HTTP:     config.HTTPConfig{CORSOrigins: []string{"*"}},
	}
	handler := app.NewRouter(cfg, pool, idSvc, nil, slog.Default())

	body, _ := json.Marshal(map[string]string{
		"email": email, "password": "password123", "audience": "store",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-App-Audience", "store")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d %s", rec.Code, rec.Body.String())
	}
}

func TestBlockedStoreCustomerMeCartForbidden(t *testing.T) {
	testdb.MigrateUp(t)
	pool := testdb.Pool(t)
	ctx := context.Background()
	if err := testdb.Reset(ctx, pool); err != nil {
		t.Fatal(err)
	}
	email := testdb.UniqueEmail(t, "blocked-cart")
	cust, err := testdb.SeedCustomer(ctx, pool, email, "Bloqueado")
	if err != nil {
		t.Fatal(err)
	}
	mgr, _ := testdb.SeedManager(ctx, pool, testdb.UniqueEmail(t, "mgr2"))
	_ = testdb.ApproveCustomer(ctx, pool, cust.ID, mgr.UserID, 100_000)

	secret := "test-session-secret-min-16"
	idSvc := identity.NewService(identitypostgres.NewRepository(pool), time.Hour, 8*time.Hour, secret)
	cfg := config.Config{
		AppEnv: "test",
		Security: config.SecurityConfig{SessionSecret: secret, AdminMFARequired: false},
		Session: config.SessionConfig{
			StoreCookie: "store_session",
			AdminCookie: "admin_session",
			StoreTTL:    time.Hour,
			AdminTTL:    8 * time.Hour,
		},
		HTTP:     config.HTTPConfig{CORSOrigins: []string{"*"}},
	}
	handler := app.NewRouter(cfg, pool, idSvc, nil, slog.Default())

	loginBody, _ := json.Marshal(map[string]string{
		"email": email, "password": "password123", "audience": "store",
	})
	loginReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")
	loginReq.Header.Set("X-App-Audience", "store")
	loginRec := httptest.NewRecorder()
	handler.ServeHTTP(loginRec, loginReq)
	if loginRec.Code != http.StatusOK {
		t.Fatalf("login before block: %d %s", loginRec.Code, loginRec.Body.String())
	}
	var loginRes map[string]any
	if err := json.Unmarshal(loginRec.Body.Bytes(), &loginRes); err != nil {
		t.Fatal(err)
	}
	token, _ := loginRes["access_token"].(string)
	if token == "" {
		t.Fatal("expected access_token in login response")
	}

	custSvc := customers.NewService(pool, nil)
	if err := custSvc.Block(ctx, cust.ID, "inadimplência"); err != nil {
		t.Fatal(err)
	}

	cartReq := httptest.NewRequest(http.MethodGet, "/api/v1/me/cart", nil)
	cartReq.Header.Set("X-App-Audience", "store")
	cartReq.Header.Set("Authorization", "Bearer "+token)
	cartRec := httptest.NewRecorder()
	handler.ServeHTTP(cartRec, cartReq)
	if cartRec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 on cart after block, got %d %s", cartRec.Code, cartRec.Body.String())
	}
}

func TestAdminUpdateCollaboratorCategoryNotFound(t *testing.T) {
	testdb.MigrateUp(t)
	pool := testdb.Pool(t)
	ctx := context.Background()
	if err := testdb.Reset(ctx, pool); err != nil {
		t.Fatal(err)
	}
	email := testdb.UniqueEmail(t, "admin-cat")
	_, err := testdb.SeedManager(ctx, pool, email)
	if err != nil {
		t.Fatal(err)
	}

	secret := "test-session-secret-min-16"
	idSvc := identity.NewService(identitypostgres.NewRepository(pool), time.Hour, 8*time.Hour, secret)
	cfg := config.Config{
		AppEnv: "test",
		Security: config.SecurityConfig{SessionSecret: secret, AdminMFARequired: false},
		Session: config.SessionConfig{
			StoreCookie: "store_session",
			AdminCookie: "admin_session",
			StoreTTL:    time.Hour,
			AdminTTL:    8 * time.Hour,
		},
		HTTP:     config.HTTPConfig{CORSOrigins: []string{"*"}},
	}
	handler := app.NewRouter(cfg, pool, idSvc, nil, slog.Default())

	loginBody, _ := json.Marshal(map[string]string{
		"email": email, "password": "password123", "audience": "admin",
	})
	loginReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")
	loginReq.Header.Set("X-App-Audience", "admin")
	loginRec := httptest.NewRecorder()
	handler.ServeHTTP(loginRec, loginReq)
	if loginRec.Code != http.StatusOK {
		t.Fatalf("admin login: %d %s", loginRec.Code, loginRec.Body.String())
	}
	var loginRes map[string]any
	_ = json.Unmarshal(loginRec.Body.Bytes(), &loginRes)
	token, _ := loginRes["access_token"].(string)

	patchBody, _ := json.Marshal(map[string]string{"name": "X"})
	id := uuid.New().String()
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/collaborator-categories/"+id, bytes.NewReader(patchBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-App-Audience", "admin")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d %s", rec.Code, rec.Body.String())
	}
}

func TestUnblockRestoresPriorStatus(t *testing.T) {
	testdb.MigrateUp(t)
	pool := testdb.Pool(t)
	ctx := context.Background()
	if err := testdb.Reset(ctx, pool); err != nil {
		t.Fatal(err)
	}
	email := testdb.UniqueEmail(t, "unblock")
	cust, err := testdb.SeedCustomer(ctx, pool, email, "Pendente")
	if err != nil {
		t.Fatal(err)
	}
	custSvc := customers.NewService(pool, nil)
	if err := custSvc.Block(ctx, cust.ID, "review"); err != nil {
		t.Fatal(err)
	}
	if err := custSvc.Unblock(ctx, cust.ID); err != nil {
		t.Fatal(err)
	}
	var status string
	if err := pool.QueryRow(ctx, `SELECT status FROM customers WHERE id = $1`, cust.ID).Scan(&status); err != nil {
		t.Fatal(err)
	}
	if status != "pending" {
		t.Fatalf("expected status pending after unblock, got %q", status)
	}
}
