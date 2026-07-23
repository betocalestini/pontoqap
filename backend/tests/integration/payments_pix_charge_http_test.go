package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/store-platform/store/internal/billing"
	"github.com/store-platform/store/tests/testdb"
)

func TestStoreInstallmentPixChargeHTTP(t *testing.T) {
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
	billSvc := billing.NewService(pool, nil, "", nil)
	if err := testdb.ActivateSingleInstallmentPlan(ctx, billSvc, invID, cust.ID, cust.UserID); err != nil {
		t.Fatal(err)
	}
	instID, err := testdb.FirstOpenInstallmentID(ctx, pool, invID)
	if err != nil {
		t.Fatal(err)
	}

	handler := newIntegrationHandler(t, pool)
	token := storeLoginToken(t, handler, cust.Email, "password123")

	req := httptest.NewRequest(http.MethodPost, "/api/v1/me/installments/"+instID.String()+"/pix-charge", nil)
	req.Header.Set("X-App-Audience", "store")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK && rec.Code != http.StatusCreated {
		t.Fatalf("pix-charge: %d %s", rec.Code, rec.Body.String())
	}
	var out struct {
		ID          string `json:"id"`
		QRCodeText  string `json:"qr_code_text"`
		AmountCents int64  `json:"amount_cents"`
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

	// Segunda chamada com parcela pix_active deve reutilizar a mesma cobrança.
	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/me/installments/"+instID.String()+"/pix-charge", nil)
	req2.Header.Set("X-App-Audience", "store")
	req2.Header.Set("Authorization", "Bearer "+token)
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK && rec2.Code != http.StatusCreated {
		t.Fatalf("pix-charge reuse: %d %s", rec2.Code, rec2.Body.String())
	}
	var out2 struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(rec2.Body).Decode(&out2); err != nil {
		t.Fatal(err)
	}
	if out2.ID != out.ID {
		t.Fatalf("expected same charge id %s got %s", out.ID, out2.ID)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/me/installments/"+instID.String()+"/pix-charge", nil)
	getReq.Header.Set("X-App-Audience", "store")
	getReq.Header.Set("Authorization", "Bearer "+token)
	getRec := httptest.NewRecorder()
	handler.ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("get pix-charge: %d %s", getRec.Code, getRec.Body.String())
	}

	_, err = pool.Exec(ctx, `
		UPDATE payment_charges SET expires_at = NOW() - INTERVAL '1 minute'
		WHERE id = $1 AND status = 'pending'
	`, out.ID)
	if err != nil {
		t.Fatal(err)
	}
	req3 := httptest.NewRequest(http.MethodPost, "/api/v1/me/installments/"+instID.String()+"/pix-charge", nil)
	req3.Header.Set("X-App-Audience", "store")
	req3.Header.Set("Authorization", "Bearer "+token)
	rec3 := httptest.NewRecorder()
	handler.ServeHTTP(rec3, req3)
	if rec3.Code != http.StatusOK && rec3.Code != http.StatusCreated {
		t.Fatalf("pix-charge after expiry: %d %s", rec3.Code, rec3.Body.String())
	}
	var out3 struct {
		ID         string `json:"id"`
		QRCodeText string `json:"qr_code_text"`
	}
	if err := json.NewDecoder(rec3.Body).Decode(&out3); err != nil {
		t.Fatal(err)
	}
	if out3.ID == out.ID {
		t.Fatal("expected new charge after pix expiration")
	}
	if out3.QRCodeText == "" {
		t.Fatal("expected qr on reissue")
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

func storeLoginToken(t *testing.T, handler http.Handler, email, password string) string {
	t.Helper()
	loginBody, _ := json.Marshal(map[string]string{
		"email":    email,
		"password": password,
		"audience": "store",
	})
	loginReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")
	loginReq.Header.Set("X-App-Audience", "store")
	loginRec := httptest.NewRecorder()
	handler.ServeHTTP(loginRec, loginReq)
	if loginRec.Code != http.StatusOK {
		t.Fatalf("store login: %d %s", loginRec.Code, loginRec.Body.String())
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
