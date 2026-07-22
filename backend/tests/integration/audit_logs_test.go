package integration_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/google/uuid"

	"github.com/store-platform/store/tests/testdb"
)

func TestAuditLogsListAfterManualClose(t *testing.T) {
	testdb.MigrateUp(t)
	pool := testdb.Pool(t)
	ctx := context.Background()
	if err := testdb.Reset(ctx, pool); err != nil {
		t.Fatal(err)
	}

	admin, err := testdb.SeedSystemAdmin(ctx, pool, testdb.UniqueEmail(t, "audit-admin"))
	if err != nil {
		t.Fatal(err)
	}
	mgr, _ := testdb.SeedManager(ctx, pool, testdb.UniqueEmail(t, "mgr-audit"))
	cust, _ := testdb.SeedCustomer(ctx, pool, testdb.UniqueEmail(t, "c-audit"), "Cliente")
	_ = testdb.ApproveCustomer(ctx, pool, cust.ID, mgr.UserID, 50_000)
	_ = seedOpenPeriodWithEntry(t, ctx, pool, cust.ID, 2026, 6, 1500)

	handler := newIntegrationHandler(t, pool)
	adminToken := adminLoginToken(t, handler, admin.Email, "password123")

	closeBody := []byte(`{"year":2026,"month":6,"reason":"teste auditoria listagem"}`)
	closeReq := httptestNewJSONRequest(http.MethodPost, "/api/v1/admin/billing/close", closeBody)
	closeReq.Header.Set("X-App-Audience", "admin")
	closeReq.Header.Set("Authorization", "Bearer "+adminToken)
	closeRec := httptestNewRecorder(handler, closeReq)
	if closeRec.Code != http.StatusOK {
		t.Fatalf("close: %d %s", closeRec.Code, closeRec.Body.String())
	}

	rec := adminGET(t, handler, adminToken, "/api/v1/admin/audit/logs?action=billing.close_manual&limit=10")
	if rec.Code != http.StatusOK {
		t.Fatalf("audit logs: %d %s", rec.Code, rec.Body.String())
	}
	var out struct {
		Items []struct {
			Action       string    `json:"action"`
			ActorUserID  uuid.UUID `json:"actor_user_id"`
		} `json:"items"`
		Total int `json:"total"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if out.Total < 1 || len(out.Items) < 1 {
		t.Fatal("expected audit log entries")
	}
	found := false
	for _, it := range out.Items {
		if it.Action == "billing.close_manual" && it.ActorUserID == admin.UserID {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected billing.close_manual by system admin")
	}
}

func TestManagerCannotListAuditLogs(t *testing.T) {
	testdb.MigrateUp(t)
	pool := testdb.Pool(t)
	ctx := context.Background()
	if err := testdb.Reset(ctx, pool); err != nil {
		t.Fatal(err)
	}
	mgr, _ := testdb.SeedManager(ctx, pool, testdb.UniqueEmail(t, "mgr-no-audit"))
	handler := newIntegrationHandler(t, pool)
	token := adminLoginToken(t, handler, mgr.Email, "password123")
	rec := adminGET(t, handler, token, "/api/v1/admin/audit/logs")
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}
