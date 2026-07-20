package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/store-platform/store/internal/app"
	"github.com/store-platform/store/internal/identity"
	identitypostgres "github.com/store-platform/store/internal/identity/postgres"
	"github.com/store-platform/store/internal/platform/config"
	"github.com/store-platform/store/tests/testdb"
)

func TestAdminInvitationAcceptAndLogin(t *testing.T) {
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

	secret := "test-session-secret-min-16"
	idSvc := identity.NewService(identitypostgres.NewRepository(pool), time.Hour, 8*time.Hour, secret)
	cfg := config.Config{
		AppEnv: "test",
		App: config.AppConfig{
			AdminWebURL: "http://localhost:5174",
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
	handler := app.NewRouter(cfg, pool, idSvc, nil, slog.Default())

	adminToken := adminLoginToken(t, handler, adminEmail, "password123")

	inviteEmail := testdb.UniqueEmail(t, "invited-mgr")
	cust, err := testdb.SeedCustomer(ctx, pool, inviteEmail, "Novo Gerente")
	if err != nil {
		t.Fatal(err)
	}
	mgr, _ := testdb.SeedManager(ctx, pool, testdb.UniqueEmail(t, "mgr"))
	if err := testdb.ApproveCustomer(ctx, pool, cust.ID, mgr.UserID, 10_000); err != nil {
		t.Fatal(err)
	}

	inviteBody, _ := json.Marshal(map[string]string{
		"email":   inviteEmail,
		"name":    "Novo Gerente",
		"role_id": "a0000000-0000-4000-8000-000000000002",
	})
	invReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/users/invitations", bytes.NewReader(inviteBody))
	invReq.Header.Set("Content-Type", "application/json")
	invReq.Header.Set("X-App-Audience", "admin")
	invReq.Header.Set("Authorization", "Bearer "+adminToken)
	invRec := httptest.NewRecorder()
	handler.ServeHTTP(invRec, invReq)
	if invRec.Code != http.StatusNoContent {
		t.Fatalf("invite status %d body %s", invRec.Code, invRec.Body.String())
	}

	var tokenHash string
	err = pool.QueryRow(ctx, `
		SELECT token_hash FROM admin_invitations
		WHERE LOWER(email) = LOWER($1) AND accepted_at IS NULL
		ORDER BY created_at DESC LIMIT 1
	`, inviteEmail).Scan(&tokenHash)
	if err != nil {
		t.Fatal(err)
	}
	rawToken := inviteTokenFromOutbox(t, ctx, pool, inviteEmail)
	_ = tokenHash

	acceptBody, _ := json.Marshal(map[string]string{
		"token":    rawToken,
		"password": "password123",
		"name":     "Novo Gerente",
	})
	accReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/accept-invitation", bytes.NewReader(acceptBody))
	accReq.Header.Set("Content-Type", "application/json")
	accRec := httptest.NewRecorder()
	handler.ServeHTTP(accRec, accReq)
	if accRec.Code != http.StatusNoContent {
		t.Fatalf("accept status %d body %s", accRec.Code, accRec.Body.String())
	}

	_ = adminLoginToken(t, handler, inviteEmail, "password123")

	var hasCustomerRole bool
	err = pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM user_roles ur JOIN roles ro ON ro.id = ur.role_id
			WHERE ur.user_id = (SELECT user_id FROM customers WHERE id = $1) AND ro.code = 'customer'
		)
	`, cust.ID).Scan(&hasCustomerRole)
	if err != nil || !hasCustomerRole {
		t.Fatalf("customer role preserved: err=%v has=%v", err, hasCustomerRole)
	}
}

func inviteTokenFromOutbox(t *testing.T, ctx context.Context, pool *pgxpool.Pool, email string) string {
	t.Helper()
	var payload []byte
	err := pool.QueryRow(ctx, `
		SELECT payload FROM outbox_events
		WHERE event_type = 'admin.invitation_sent'
		ORDER BY created_at DESC LIMIT 1
	`).Scan(&payload)
	if err != nil {
		t.Fatal(err)
	}
	var p struct {
		To        string `json:"to"`
		InviteURL string `json:"invite_url"`
	}
	if err := json.Unmarshal(payload, &p); err != nil {
		t.Fatal(err)
	}
	if !strings.EqualFold(p.To, email) {
		t.Fatalf("outbox email %q want %q", p.To, email)
	}
	u, err := url.Parse(p.InviteURL)
	if err != nil {
		t.Fatal(err)
	}
	token := u.Query().Get("token")
	if token == "" {
		t.Fatal("missing token in invite_url")
	}
	return token
}

func adminLoginToken(t *testing.T, handler http.Handler, email, password string) string {
	t.Helper()
	loginBody, _ := json.Marshal(map[string]string{
		"email":    email,
		"password": password,
		"audience": "admin",
	})
	loginReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")
	loginReq.Header.Set("X-App-Audience", "admin")
	loginRec := httptest.NewRecorder()
	handler.ServeHTTP(loginRec, loginReq)
	if loginRec.Code != http.StatusOK {
		t.Fatalf("login status %d body %s", loginRec.Code, loginRec.Body.String())
	}
	var loginRes map[string]any
	if err := json.Unmarshal(loginRec.Body.Bytes(), &loginRes); err != nil {
		t.Fatal(err)
	}
	token, _ := loginRes["access_token"].(string)
	if token == "" {
		t.Fatal("missing access_token")
	}
	return token
}