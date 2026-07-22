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
	"github.com/store-platform/store/internal/identity"
	identitypostgres "github.com/store-platform/store/internal/identity/postgres"
	"github.com/store-platform/store/internal/platform/config"
	"github.com/store-platform/store/tests/testdb"
)

func TestManagerListCustomersWithoutUsersManage(t *testing.T) {
	testdb.MigrateUp(t)
	pool := testdb.Pool(t)
	ctx := context.Background()
	if err := testdb.Reset(ctx, pool); err != nil {
		t.Fatal(err)
	}
	_, _ = testdb.SeedCustomer(ctx, pool, testdb.UniqueEmail(t, "c"), "C")

	secret := "test-session-secret-min-16"
	idSvc := identity.NewService(identitypostgres.NewRepository(pool), time.Hour, 8*time.Hour, secret, nil)
	cfg := config.Config{
		AppEnv:   "test",
		Security: config.SecurityConfig{SessionSecret: secret, AdminMFARequired: false},
		Session:  config.SessionConfig{StoreTTL: time.Hour, AdminTTL: 8 * time.Hour},
		HTTP:     config.HTTPConfig{CORSOrigins: []string{"*"}},
	}
	handler := app.NewRouter(cfg, pool, idSvc, nil, slog.Default())

	mgrEmail := testdb.UniqueEmail(t, "mgr-rbac")
	_, err := testdb.SeedManager(ctx, pool, mgrEmail)
	if err != nil {
		t.Fatal(err)
	}
	token := adminLoginToken(t, handler, mgrEmail, "password123")

	custReq := httptest.NewRequest(http.MethodGet, "/api/v1/admin/customers", nil)
	custReq.Header.Set("X-App-Audience", "admin")
	custReq.Header.Set("Authorization", "Bearer "+token)
	custRec := httptest.NewRecorder()
	handler.ServeHTTP(custRec, custReq)
	if custRec.Code != http.StatusOK {
		t.Fatalf("customers status %d body %s", custRec.Code, custRec.Body.String())
	}

	rolesReq := httptest.NewRequest(http.MethodGet, "/api/v1/admin/roles", nil)
	rolesReq.Header.Set("X-App-Audience", "admin")
	rolesReq.Header.Set("Authorization", "Bearer "+token)
	rolesRec := httptest.NewRecorder()
	handler.ServeHTTP(rolesRec, rolesReq)
	if rolesRec.Code != http.StatusForbidden {
		t.Fatalf("roles want 403 got %d body %s", rolesRec.Code, rolesRec.Body.String())
	}
}

func TestAdminCannotChangeOwnStaffRole(t *testing.T) {
	testdb.MigrateUp(t)
	pool := testdb.Pool(t)
	ctx := context.Background()
	if err := testdb.Reset(ctx, pool); err != nil {
		t.Fatal(err)
	}
	adminEmail := testdb.UniqueEmail(t, "self-role")
	admin, err := testdb.SeedSystemAdmin(ctx, pool, adminEmail)
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
	token := adminLoginToken(t, handler, adminEmail, "password123")

	body, _ := json.Marshal(map[string]string{
		"role_id":  "a0000000-0000-4000-8000-000000000002",
		"password": "password123",
	})
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/users/"+admin.UserID.String()+"/role", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-App-Audience", "admin")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("want 403 got %d body %s", rec.Code, rec.Body.String())
	}
	_ = admin
}
