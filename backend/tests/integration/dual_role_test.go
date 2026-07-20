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

func TestAssignStaffRoleFromCustomerDualLogin(t *testing.T) {
	testdb.MigrateUp(t)
	pool := testdb.Pool(t)
	ctx := context.Background()
	if err := testdb.Reset(ctx, pool); err != nil {
		t.Fatal(err)
	}

	adminEmail := testdb.UniqueEmail(t, "sysadmin")
	_, err := testdb.SeedSystemAdmin(ctx, pool, adminEmail)
	if err != nil {
		t.Fatal(err)
	}

	handler := dualRoleHandler(t, pool)
	adminToken := adminLoginToken(t, handler, adminEmail, "password123")

	email := testdb.UniqueEmail(t, "dual")
	cust, err := testdb.SeedCustomer(ctx, pool, email, "Dual User")
	if err != nil {
		t.Fatal(err)
	}
	mgr, _ := testdb.SeedManager(ctx, pool, testdb.UniqueEmail(t, "m"))
	if err := testdb.ApproveCustomer(ctx, pool, cust.ID, mgr.UserID, 50_000); err != nil {
		t.Fatal(err)
	}

	body, _ := json.Marshal(map[string]string{
		"role_id":  "a0000000-0000-4000-8000-000000000002",
		"password": "password123",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/customers/"+cust.ID.String()+"/staff-role", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-App-Audience", "admin")
	req.Header.Set("Authorization", "Bearer "+adminToken)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("assign staff role: %d %s", rec.Code, rec.Body.String())
	}

	_ = adminLoginToken(t, handler, email, "password123")
	storeLoginOK(t, handler, email, "password123")
}

func TestSuspendedStaffStillLogsIntoStore(t *testing.T) {
	testdb.MigrateUp(t)
	pool := testdb.Pool(t)
	ctx := context.Background()
	if err := testdb.Reset(ctx, pool); err != nil {
		t.Fatal(err)
	}

	adminEmail := testdb.UniqueEmail(t, "sysadmin")
	_, err := testdb.SeedSystemAdmin(ctx, pool, adminEmail)
	if err != nil {
		t.Fatal(err)
	}
	handler := dualRoleHandler(t, pool)
	adminToken := adminLoginToken(t, handler, adminEmail, "password123")

	email := testdb.UniqueEmail(t, "susp")
	cust, err := testdb.SeedCustomer(ctx, pool, email, "Suspended Staff")
	if err != nil {
		t.Fatal(err)
	}
	mgr, _ := testdb.SeedManager(ctx, pool, testdb.UniqueEmail(t, "m2"))
	if err := testdb.ApproveCustomer(ctx, pool, cust.ID, mgr.UserID, 10_000); err != nil {
		t.Fatal(err)
	}

	assignBody, _ := json.Marshal(map[string]string{
		"role_id":  "a0000000-0000-4000-8000-000000000002",
		"password": "password123",
	})
	assignReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/customers/"+cust.ID.String()+"/staff-role", bytes.NewReader(assignBody))
	assignReq.Header.Set("Content-Type", "application/json")
	assignReq.Header.Set("X-App-Audience", "admin")
	assignReq.Header.Set("Authorization", "Bearer "+adminToken)
	assignRec := httptest.NewRecorder()
	handler.ServeHTTP(assignRec, assignReq)
	if assignRec.Code != http.StatusNoContent {
		t.Fatalf("assign: %d %s", assignRec.Code, assignRec.Body.String())
	}

	statusBody, _ := json.Marshal(map[string]string{
		"status":   "suspended",
		"password": "password123",
	})
	statusReq := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/users/"+cust.UserID.String()+"/status", bytes.NewReader(statusBody))
	statusReq.Header.Set("Content-Type", "application/json")
	statusReq.Header.Set("X-App-Audience", "admin")
	statusReq.Header.Set("Authorization", "Bearer "+adminToken)
	statusRec := httptest.NewRecorder()
	handler.ServeHTTP(statusRec, statusReq)
	if statusRec.Code != http.StatusNoContent {
		t.Fatalf("suspend: %d %s", statusRec.Code, statusRec.Body.String())
	}

	loginBody, _ := json.Marshal(map[string]string{
		"email":    email,
		"password": "password123",
		"audience": "admin",
	})
	loginReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")
	loginReq.Header.Set("X-App-Audience", "admin")
	loginRec := httptest.NewRecorder()
	handler.ServeHTTP(loginRec, loginReq)
	if loginRec.Code != http.StatusForbidden {
		t.Fatalf("admin login when suspended want 403 got %d %s", loginRec.Code, loginRec.Body.String())
	}

	storeLoginOK(t, handler, email, "password123")
}

func TestInvitationRequiresStoreCustomer(t *testing.T) {
	testdb.MigrateUp(t)
	pool := testdb.Pool(t)
	ctx := context.Background()
	if err := testdb.Reset(ctx, pool); err != nil {
		t.Fatal(err)
	}

	adminEmail := testdb.UniqueEmail(t, "sysadmin")
	_, err := testdb.SeedSystemAdmin(ctx, pool, adminEmail)
	if err != nil {
		t.Fatal(err)
	}
	handler := dualRoleHandler(t, pool)
	adminToken := adminLoginToken(t, handler, adminEmail, "password123")

	inviteEmail := testdb.UniqueEmail(t, "no-customer")
	inviteBody, _ := json.Marshal(map[string]string{
		"email":   inviteEmail,
		"name":    "Sem Loja",
		"role_id": "a0000000-0000-4000-8000-000000000002",
	})
	invReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/users/invitations", bytes.NewReader(inviteBody))
	invReq.Header.Set("Content-Type", "application/json")
	invReq.Header.Set("X-App-Audience", "admin")
	invReq.Header.Set("Authorization", "Bearer "+adminToken)
	invRec := httptest.NewRecorder()
	handler.ServeHTTP(invRec, invReq)
	if invRec.Code != http.StatusConflict {
		t.Fatalf("invite without customer want 409 got %d %s", invRec.Code, invRec.Body.String())
	}
}

func dualRoleHandler(t *testing.T, pool *pgxpool.Pool) http.Handler {
	t.Helper()
	secret := "test-session-secret-min-16"
	idSvc := identity.NewService(identitypostgres.NewRepository(pool), time.Hour, 8*time.Hour, secret)
	cfg := config.Config{
		AppEnv: "test",
		App: config.AppConfig{
			AdminWebURL: "http://localhost:5174",
			StoreWebURL: "http://localhost:5173",
		},
		Security: config.SecurityConfig{
			SessionSecret:    secret,
			AdminMFARequired: false,
		},
		Session: config.SessionConfig{
			StoreTTL: time.Hour,
			AdminTTL: 8 * time.Hour,
		},
		HTTP: config.HTTPConfig{CORSOrigins: []string{"*"}},
	}
	return app.NewRouter(cfg, pool, idSvc, nil, slog.Default())
}

func storeLoginOK(t *testing.T, handler http.Handler, email, password string) {
	t.Helper()
	loginBody, _ := json.Marshal(map[string]string{
		"email":    email,
		"password": password,
		"audience": "store",
	})
	loginReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")
	loginRec := httptest.NewRecorder()
	handler.ServeHTTP(loginRec, loginReq)
	if loginRec.Code != http.StatusOK {
		t.Fatalf("store login: %d %s", loginRec.Code, loginRec.Body.String())
	}
}
