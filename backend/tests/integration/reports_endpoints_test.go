package integration_test

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/store-platform/store/tests/testdb"
)

func TestReportsEndpointsSmoke(t *testing.T) {
	testdb.MigrateUp(t)
	pool := testdb.Pool(t)
	ctx := context.Background()
	if err := testdb.Reset(ctx, pool); err != nil {
		t.Fatal(err)
	}

	mgr, err := testdb.SeedManager(ctx, pool, testdb.UniqueEmail(t, "mgr-rpt-smoke"))
	if err != nil {
		t.Fatal(err)
	}
	handler := newIntegrationHandler(t, pool)
	token := adminLoginToken(t, handler, mgr.Email, "password123")

	now := time.Now().UTC()
	y, m := now.Year(), int(now.Month())

	paths := []string{
		"/api/v1/admin/reports/top-products?limit=5",
		"/api/v1/admin/reports/top-customers?limit=5",
		"/api/v1/admin/reports/inventory/position",
		"/api/v1/admin/reports/inventory/movements",
		"/api/v1/admin/reports/receivables/invoices?year=" + itoa(y) + "&month=" + itoa(m),
		"/api/v1/admin/reports/customers/exposure",
		"/api/v1/admin/reports/forecast?limit=5",
	}
	for _, path := range paths {
		rec := adminGET(t, handler, token, path)
		if rec.Code != http.StatusOK {
			t.Fatalf("%s: status %d body %s", path, rec.Code, rec.Body.String())
		}
	}

	csvPath := "/api/v1/admin/reports/sales/orders/export.csv?year=" + itoa(y) + "&month=" + itoa(m)
	csvRec := adminGET(t, handler, token, csvPath)
	if csvRec.Code != http.StatusOK {
		t.Fatalf("csv export: %d %s", csvRec.Code, csvRec.Body.String())
	}
	ct := csvRec.Header().Get("Content-Type")
	if !strings.Contains(ct, "text/csv") && !strings.Contains(ct, "application/octet-stream") {
		t.Fatalf("csv content-type: %q", ct)
	}
	if csvRec.Body.Len() == 0 {
		t.Fatal("csv body empty")
	}
}

func TestReportsReceivablesAfterClose(t *testing.T) {
	testdb.MigrateUp(t)
	pool := testdb.Pool(t)
	ctx := context.Background()
	if err := testdb.Reset(ctx, pool); err != nil {
		t.Fatal(err)
	}
	mgr, _ := testdb.SeedManager(ctx, pool, testdb.UniqueEmail(t, "mgr-rec"))
	cust, _ := testdb.SeedCustomer(ctx, pool, testdb.UniqueEmail(t, "c-rec"), "Cliente")
	_ = testdb.ApproveCustomer(ctx, pool, cust.ID, mgr.UserID, 50_000)
	_ = seedClosedInvoice(t, ctx, pool, cust.ID, 2026, 4, 3200)

	handler := newIntegrationHandler(t, pool)
	token := adminLoginToken(t, handler, mgr.Email, "password123")
	rec := adminGET(t, handler, token, "/api/v1/admin/reports/receivables/invoices?year=2026&month=4")
	if rec.Code != http.StatusOK {
		t.Fatalf("receivables: %d %s", rec.Code, rec.Body.String())
	}
	var out struct {
		Items []json.RawMessage `json:"items"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if len(out.Items) < 1 {
		t.Fatal("expected at least one receivable row")
	}
}
