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

func TestStoreLoginReturnsJWTAndProtectsRoutes(t *testing.T) {
	testdb.MigrateUp(t)
	pool := testdb.Pool(t)
	ctx := context.Background()
	if err := testdb.Reset(ctx, pool); err != nil {
		t.Fatal(err)
	}

	email := testdb.UniqueEmail(t, "jwt-store")
	cust, err := testdb.SeedCustomer(ctx, pool, email, "Cliente JWT")
	if err != nil {
		t.Fatal(err)
	}
	mgr, _ := testdb.SeedManager(ctx, pool, testdb.UniqueEmail(t, "mgr-jwt"))
	_ = testdb.ApproveCustomer(ctx, pool, cust.ID, mgr.UserID, 100_000)

	secret := "test-session-secret-min-16"
	idSvc := identity.NewService(identitypostgres.NewRepository(pool), time.Hour, 8*time.Hour, secret)
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
		"audience": "store",
	})
	loginReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")
	loginReq.Header.Set("X-App-Audience", "store")
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

	meReq := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	meReq.Header.Set("X-App-Audience", "store")
	meReq.Header.Set("Authorization", "Bearer "+token)
	meRec := httptest.NewRecorder()
	handler.ServeHTTP(meRec, meReq)
	if meRec.Code != http.StatusOK {
		t.Fatalf("me should accept bearer: %s", meRec.Body.String())
	}

	cartReq := httptest.NewRequest(http.MethodGet, "/api/v1/me/cart", nil)
	cartReq.Header.Set("X-App-Audience", "store")
	cartReq.Header.Set("Authorization", "Bearer "+token)
	cartRec := httptest.NewRecorder()
	handler.ServeHTTP(cartRec, cartReq)
	if cartRec.Code == http.StatusUnauthorized {
		t.Fatalf("cart route should accept bearer: %s", cartRec.Body.String())
	}

	logoutReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	logoutReq.Header.Set("X-App-Audience", "store")
	logoutReq.Header.Set("Authorization", "Bearer "+token)
	logoutRec := httptest.NewRecorder()
	handler.ServeHTTP(logoutRec, logoutReq)
	if logoutRec.Code != http.StatusNoContent {
		t.Fatalf("logout status %d", logoutRec.Code)
	}

	meRec2 := httptest.NewRecorder()
	handler.ServeHTTP(meRec2, meReq)
	if meRec2.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 after logout, got %d", meRec2.Code)
	}
}
