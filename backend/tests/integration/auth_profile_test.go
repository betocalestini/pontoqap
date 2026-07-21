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

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/store-platform/store/internal/app"
	"github.com/store-platform/store/internal/identity"
	identitypostgres "github.com/store-platform/store/internal/identity/postgres"
	"github.com/store-platform/store/internal/platform/config"
	"github.com/store-platform/store/tests/testdb"
)

func testAPIHandler(t *testing.T, pool *pgxpool.Pool) http.Handler {
	t.Helper()
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
		HTTP: config.HTTPConfig{CORSOrigins: []string{"*"}},
	}
	return app.NewRouter(cfg, pool, idSvc, nil, slog.Default())
}

func TestPatchAuthMeStoreProfile(t *testing.T) {
	testdb.MigrateUp(t)
	pool := testdb.Pool(t)
	ctx := context.Background()
	if err := testdb.Reset(ctx, pool); err != nil {
		t.Fatal(err)
	}
	email := testdb.UniqueEmail(t, "profile")
	cust, err := testdb.SeedCustomer(ctx, pool, email, "Antes")
	if err != nil {
		t.Fatal(err)
	}
	mgr, _ := testdb.SeedManager(ctx, pool, testdb.UniqueEmail(t, "mgr"))
	_ = testdb.ApproveCustomer(ctx, pool, cust.ID, mgr.UserID, 100_000)

	handler := testAPIHandler(t, pool)

	loginBody, _ := json.Marshal(map[string]string{
		"email": email, "password": "password123", "audience": "store",
	})
	loginReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")
	loginReq.Header.Set("X-App-Audience", "store")
	loginRec := httptest.NewRecorder()
	handler.ServeHTTP(loginRec, loginReq)
	if loginRec.Code != http.StatusOK {
		t.Fatalf("login %d %s", loginRec.Code, loginRec.Body.String())
	}
	var cookie string
	for _, c := range loginRec.Result().Cookies() {
		if c.Name == "store_session" {
			cookie = c.Name + "=" + c.Value
			break
		}
	}
	if cookie == "" {
		t.Fatal("expected store_session cookie")
	}

	patchBody, _ := json.Marshal(map[string]string{
		"name":  "Cliente Atualizado",
		"phone": "11999998888",
	})
	patchReq := httptest.NewRequest(http.MethodPatch, "/api/v1/auth/me", bytes.NewReader(patchBody))
	patchReq.Header.Set("Content-Type", "application/json")
	patchReq.Header.Set("Cookie", cookie)
	patchReq.Header.Set("X-App-Audience", "store")
	patchRec := httptest.NewRecorder()
	handler.ServeHTTP(patchRec, patchReq)
	if patchRec.Code != http.StatusOK {
		t.Fatalf("patch %d %s", patchRec.Code, patchRec.Body.String())
	}
	var me map[string]any
	if err := json.Unmarshal(patchRec.Body.Bytes(), &me); err != nil {
		t.Fatal(err)
	}
	if me["name"] != "Cliente Atualizado" || me["phone"] != "11999998888" {
		t.Fatalf("unexpected me %+v", me)
	}
}
