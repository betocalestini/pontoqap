package integration_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/store-platform/store/tests/testdb"
)

func TestAdminPixChargeHTTP(t *testing.T) {
	testdb.MigrateUp(t)
	pool := testdb.Pool(t)
	ctx := context.Background()
	if err := testdb.Reset(ctx, pool); err != nil {
		t.Fatal(err)
	}

	mgr, _ := testdb.SeedManager(ctx, pool, testdb.UniqueEmail(t, "mgr-pix"))
	cust, _ := testdb.SeedCustomer(ctx, pool, testdb.UniqueEmail(t, "c-pix"), "Cliente")
	_ = testdb.ApproveCustomer(ctx, pool, cust.ID, mgr.UserID, 50_000)
	invID := seedClosedInvoice(t, ctx, pool, cust.ID, 2026, 7, 4500)

	handler := newIntegrationHandler(t, pool)
	token := adminLoginToken(t, handler, mgr.Email, "password123")

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/invoices/"+invID.String()+"/pix-charge", nil)
	req.Header.Set("X-App-Audience", "admin")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK && rec.Code != http.StatusCreated {
		t.Fatalf("pix-charge: %d %s", rec.Code, rec.Body.String())
	}
	var out struct {
		ID           string `json:"id"`
		QRCodeText   string `json:"qr_code_text"`
		AmountCents  int64  `json:"amount_cents"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if out.AmountCents != 4500 {
		t.Fatalf("amount: got %d want 4500", out.AmountCents)
	}
	if out.QRCodeText == "" {
		t.Fatal("expected qr_code_text")
	}

	// idempotent second call
	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/admin/invoices/"+invID.String()+"/pix-charge", nil)
	req2.Header.Set("X-App-Audience", "admin")
	req2.Header.Set("Authorization", "Bearer "+token)
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK && rec2.Code != http.StatusCreated {
		t.Fatalf("pix-charge retry: %d %s", rec2.Code, rec2.Body.String())
	}
}

func TestAdminPixChargeNotFoundInvoice(t *testing.T) {
	testdb.MigrateUp(t)
	pool := testdb.Pool(t)
	ctx := context.Background()
	if err := testdb.Reset(ctx, pool); err != nil {
		t.Fatal(err)
	}
	mgr, _ := testdb.SeedManager(ctx, pool, testdb.UniqueEmail(t, "mgr-pix-404"))
	handler := newIntegrationHandler(t, pool)
	token := adminLoginToken(t, handler, mgr.Email, "password123")

	fakeID := "00000000-0000-4000-8000-000000000099"
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/invoices/"+fakeID+"/pix-charge", nil)
	req.Header.Set("X-App-Audience", "admin")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound && rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 404/400, got %d %s", rec.Code, rec.Body.String())
	}
}
