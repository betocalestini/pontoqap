package integration_test

import (
	"bytes"
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
)

const integrationSessionSecret = "test-session-secret-min-16"

func integrationTestConfig() config.Config {
	return config.Config{
		AppEnv: "test",
		App: config.AppConfig{
			AdminWebURL: "http://localhost:5174",
			StoreWebURL: "http://localhost:5173",
		},
		Security: config.SecurityConfig{
			SessionSecret:    integrationSessionSecret,
			AdminMFARequired: false,
		},
		Payments: config.PaymentsConfig{
			WebhookSecret: "test-webhook-secret",
		},
		Session: config.SessionConfig{
			StoreTTL: time.Hour,
			AdminTTL: 8 * time.Hour,
		},
		HTTP: config.HTTPConfig{CORSOrigins: []string{"*"}},
	}
}

func newIntegrationHandler(t *testing.T, pool *pgxpool.Pool) http.Handler {
	t.Helper()
	cfg := integrationTestConfig()
	idSvc := identity.NewService(identitypostgres.NewRepository(pool), cfg.Session.StoreTTL, cfg.Session.AdminTTL, cfg.Security.SessionSecret, nil)
	return app.NewRouter(cfg, pool, idSvc, nil, slog.Default())
}

func adminGET(t *testing.T, handler http.Handler, token, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.Header.Set("X-App-Audience", "admin")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func httptestNewJSONRequest(method, path string, body []byte) *http.Request {
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	return req
}

func httptestNewRecorder(handler http.Handler, req *http.Request) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}
