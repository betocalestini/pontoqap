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

func TestAdminLoginReturnsJWTAndProtectsRoutes(t *testing.T) {
	testdb.MigrateUp(t)
	pool := testdb.Pool(t)
	ctx := context.Background()
	if err := testdb.Reset(ctx, pool); err != nil {
		t.Fatal(err)
	}

	email := testdb.UniqueEmail(t, "jwt-admin")
	mgr, err := testdb.SeedManager(ctx, pool, email)
	if err != nil {
		t.Fatal(err)
	}
	_ = mgr

	secret := "test-session-secret-min-16"
	idSvc := identity.NewService(identitypostgres.NewRepository(pool), time.Hour, 8*time.Hour, secret, nil)
	cfg := config.Config{
		AppEnv: "test",
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
	if loginRec.Code != http.StatusOK {
		t.Fatalf("login status %d body %s", loginRec.Code, loginRec.Body.String())
	}
	var loginRes map[string]any
	if err := json.Unmarshal(loginRec.Body.Bytes(), &loginRes); err != nil {
		t.Fatal(err)
	}
	token, _ := loginRes["access_token"].(string)
	if token == "" {
		t.Fatal("expected access_token in login response")
	}

	adminReq := httptest.NewRequest(http.MethodGet, "/api/v1/admin/categories", nil)
	adminReq.Header.Set("X-App-Audience", "admin")
	adminReq.Header.Set("Authorization", "Bearer "+token)
	adminRec := httptest.NewRecorder()
	handler.ServeHTTP(adminRec, adminReq)
	if adminRec.Code == http.StatusUnauthorized {
		t.Fatalf("admin route should accept bearer: %s", adminRec.Body.String())
	}

	logoutReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	logoutReq.Header.Set("X-App-Audience", "admin")
	logoutReq.Header.Set("Authorization", "Bearer "+token)
	logoutRec := httptest.NewRecorder()
	handler.ServeHTTP(logoutRec, logoutReq)
	if logoutRec.Code != http.StatusNoContent {
		t.Fatalf("logout status %d", logoutRec.Code)
	}

	adminRec2 := httptest.NewRecorder()
	handler.ServeHTTP(adminRec2, adminReq)
	if adminRec2.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 after logout, got %d", adminRec2.Code)
	}
}
