package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"

	"github.com/store-platform/store/internal/app"
	"github.com/store-platform/store/internal/billing"
	"github.com/store-platform/store/internal/identity"
	identitypostgres "github.com/store-platform/store/internal/identity/postgres"
	"github.com/store-platform/store/internal/platform/config"
	"github.com/store-platform/store/tests/testdb"
)

func seedClosedInvoice(t *testing.T, ctx context.Context, pool *pgxpool.Pool, custID uuid.UUID, year, month int, total int64) uuid.UUID {
	t.Helper()
	svc := billing.NewService(pool, nil, "")
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	at := time.Date(year, time.Month(month), 15, 12, 0, 0, 0, time.UTC)
	periodID, err := svc.EnsureOpenPeriodTx(ctx, tx, custID, at)
	if err != nil {
		t.Fatal(err)
	}
	orderID := uuid.New()
	ordNum := "T-" + orderID.String()[:8]
	if _, err := tx.Exec(ctx, `
		INSERT INTO orders (id, order_number, customer_id, status, subtotal_cents, total_cents, idempotency_key, confirmed_at)
		VALUES ($1, $2, $3, 'confirmed', $4, $4, $5, NOW())
	`, orderID, ordNum, custID, total, orderID.String()); err != nil {
		t.Fatal(err)
	}
	if err := svc.AddOrderEntryTx(ctx, tx, custID, orderID, total, at); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	inv, err := svc.ClosePeriod(ctx, periodID)
	if err != nil {
		t.Fatal(err)
	}
	return inv.ID
}

func seedOpenPeriodWithEntry(t *testing.T, ctx context.Context, pool *pgxpool.Pool, custID uuid.UUID, year, month int, total int64) uuid.UUID {
	t.Helper()
	svc := billing.NewService(pool, nil, "")
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	at := time.Date(year, time.Month(month), 10, 12, 0, 0, 0, time.UTC)
	periodID, err := svc.EnsureOpenPeriodTx(ctx, tx, custID, at)
	if err != nil {
		t.Fatal(err)
	}
	orderID := uuid.New()
	ordNum := "T-" + orderID.String()[:8]
	if _, err := tx.Exec(ctx, `
		INSERT INTO orders (id, order_number, customer_id, status, subtotal_cents, total_cents, idempotency_key, confirmed_at)
		VALUES ($1, $2, $3, 'confirmed', $4, $4, $5, NOW())
	`, orderID, ordNum, custID, total, orderID.String()); err != nil {
		t.Fatal(err)
	}
	if err := svc.AddOrderEntryTx(ctx, tx, custID, orderID, total, at); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	return periodID
}

func TestListInvoicesAdminAfterClose(t *testing.T) {
	testdb.MigrateUp(t)
	pool := testdb.Pool(t)
	ctx := context.Background()
	if err := testdb.Reset(ctx, pool); err != nil {
		t.Fatal(err)
	}
	mgr, _ := testdb.SeedManager(ctx, pool, testdb.UniqueEmail(t, "mgr"))
	cust, _ := testdb.SeedCustomer(ctx, pool, testdb.UniqueEmail(t, "c"), "Cliente Lista")
	_ = testdb.ApproveCustomer(ctx, pool, cust.ID, mgr.UserID, 50_000)
	invID := seedClosedInvoice(t, ctx, pool, cust.ID, 2026, 3, 3200)

	svc := billing.NewService(pool, nil, "")
	items, total, err := svc.ListInvoicesAdmin(ctx, billing.AdminInvoiceFilter{Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if total < 1 || len(items) < 1 {
		t.Fatalf("expected invoices, got total=%d len=%d", total, len(items))
	}
	found := false
	for _, row := range items {
		if row.ID == invID {
			found = true
			if row.TotalCents != 3200 || row.CustomerName != "Cliente Lista" {
				t.Fatalf("unexpected row %+v", row)
			}
		}
	}
	if !found {
		t.Fatal("invoice not in admin list")
	}
}

func TestAddInvoiceAdjustmentUpdatesTotal(t *testing.T) {
	testdb.MigrateUp(t)
	pool := testdb.Pool(t)
	ctx := context.Background()
	if err := testdb.Reset(ctx, pool); err != nil {
		t.Fatal(err)
	}
	mgr, _ := testdb.SeedManager(ctx, pool, testdb.UniqueEmail(t, "mgr"))
	cust, _ := testdb.SeedCustomer(ctx, pool, testdb.UniqueEmail(t, "c"), "Cliente Adj")
	_ = testdb.ApproveCustomer(ctx, pool, cust.ID, mgr.UserID, 50_000)
	invID := seedClosedInvoice(t, ctx, pool, cust.ID, 2026, 4, 5000)

	svc := billing.NewService(pool, nil, "")
	detail, err := svc.AddInvoiceAdjustment(ctx, invID, mgr.UserID, billing.AdjustmentTypeCredit, 500, "desconto comercial")
	if err != nil {
		t.Fatal(err)
	}
	if detail.TotalCents != 4500 || detail.AdjustmentCents != -500 {
		t.Fatalf("expected total 4500 adjustment -500, got total=%d adj=%d", detail.TotalCents, detail.AdjustmentCents)
	}
}

func TestClosePeriodsHTTPRequiresReason(t *testing.T) {
	testdb.MigrateUp(t)
	pool := testdb.Pool(t)
	ctx := context.Background()
	if err := testdb.Reset(ctx, pool); err != nil {
		t.Fatal(err)
	}
	mgr, _ := testdb.SeedManager(ctx, pool, testdb.UniqueEmail(t, "mgr"))
	secret := "test-session-secret-min-16"
	idSvc := identity.NewService(identitypostgres.NewRepository(pool), time.Hour, 8*time.Hour, secret)
	cfg := config.Config{
		AppEnv: "test",
		Security: config.SecurityConfig{SessionSecret: secret, AdminMFARequired: false},
		Session:  config.SessionConfig{StoreTTL: time.Hour, AdminTTL: 8 * time.Hour},
		HTTP:     config.HTTPConfig{CORSOrigins: []string{"*"}},
	}
	handler := app.NewRouter(cfg, pool, idSvc, nil, slog.Default())
	mgrEmail := mgr.Email
	token := adminLoginToken(t, handler, mgrEmail, "password123")

	body, _ := json.Marshal(map[string]any{"year": 2026, "month": 3})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/billing/close", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-App-Audience", "admin")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d %s", rec.Code, rec.Body.String())
	}
}

func TestClosePeriodsHTTPWritesAuditLog(t *testing.T) {
	testdb.MigrateUp(t)
	pool := testdb.Pool(t)
	ctx := context.Background()
	if err := testdb.Reset(ctx, pool); err != nil {
		t.Fatal(err)
	}
	mgr, _ := testdb.SeedManager(ctx, pool, testdb.UniqueEmail(t, "mgr-close"))
	cust, _ := testdb.SeedCustomer(ctx, pool, testdb.UniqueEmail(t, "c"), "Cliente Close")
	_ = testdb.ApproveCustomer(ctx, pool, cust.ID, mgr.UserID, 50_000)
	_ = seedOpenPeriodWithEntry(t, ctx, pool, cust.ID, 2026, 5, 1000)

	secret := "test-session-secret-min-16"
	idSvc := identity.NewService(identitypostgres.NewRepository(pool), time.Hour, 8*time.Hour, secret)
	cfg := config.Config{
		AppEnv: "test",
		Security: config.SecurityConfig{SessionSecret: secret, AdminMFARequired: false},
		Session:  config.SessionConfig{StoreTTL: time.Hour, AdminTTL: 8 * time.Hour},
		HTTP:     config.HTTPConfig{CORSOrigins: []string{"*"}},
	}
	handler := app.NewRouter(cfg, pool, idSvc, nil, slog.Default())
	token := adminLoginToken(t, handler, mgr.Email, "password123")

	body, _ := json.Marshal(map[string]any{
		"year":   2026,
		"month":  5,
		"reason": "fechamento homologação teste",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/billing/close", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-App-Audience", "admin")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d %s", rec.Code, rec.Body.String())
	}
	var n int
	if err := pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM audit_logs WHERE action = 'billing.close_manual' AND actor_user_id = $1
	`, mgr.UserID).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n < 1 {
		t.Fatal("expected audit log for manual close")
	}
}
