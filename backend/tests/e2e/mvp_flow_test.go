package e2e_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/store-platform/store/tests/testdb"
)

// TestMVPBillingPixAndDashboard cobre fechamento de período, Pix sandbox e dashboard admin.
func TestMVPBillingPixAndDashboard(t *testing.T) {
	testdb.MigrateUp(t)
	pool := testdb.Pool(t)
	ctx := context.Background()
	if err := testdb.Reset(ctx, pool); err != nil {
		t.Fatal(err)
	}

	mgr, err := testdb.SeedManager(ctx, pool, testdb.UniqueEmail(t, "mgr"))
	if err != nil {
		t.Fatal(err)
	}
	prod, err := testdb.SeedProduct(ctx, pool, "Arroz", "ARR-1", 500)
	if err != nil {
		t.Fatal(err)
	}
	setupE2EInventory(t, ctx, pool, prod.SKUID, mgr.UserID)

	server, client := newE2EServer(t, pool)
	email := testdb.UniqueEmail(t, "cli")
	registerStoreCustomer(t, client, server.URL, email)
	verifyStoreCustomer(t, ctx, pool, email, 50_000)
	adminCookie := loginAdminCookie(t, client, server.URL, mgr.Email, "password123")

	storeToken := login(t, client, server.URL, email, "password123", "store")
	cartBody, _ := json.Marshal(map[string]any{"sku_id": prod.SKUID.String(), "quantity": 2})
	cartRes := doStoreJSON(t, client, http.MethodPost, server.URL+"/api/v1/me/cart/items", storeToken, cartBody)
	if cartRes.StatusCode != http.StatusOK {
		t.Fatalf("cart: %d", cartRes.StatusCode)
	}
	cartRes.Body.Close()

	chReq, _ := http.NewRequest(http.MethodPost, server.URL+"/api/v1/me/cart/checkout", nil)
	chReq.Header.Set("Idempotency-Key", "mvp-flow-1")
	chReq.Header.Set("X-App-Audience", "store")
	chReq.Header.Set("Authorization", "Bearer "+storeToken)
	chRes, err := client.Do(chReq)
	if err != nil {
		t.Fatal(err)
	}
	if chRes.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(chRes.Body)
		t.Fatalf("checkout: %d %s", chRes.StatusCode, body)
	}
	chRes.Body.Close()

	now := time.Now()
	closeBody, _ := json.Marshal(map[string]any{
		"year":   now.Year(),
		"month":  int(now.Month()),
		"reason": "fechamento e2e fluxo MVP",
	})
	closeRes := doAdminJSON(t, client, http.MethodPost, server.URL+"/api/v1/admin/billing/close", adminCookie, closeBody)
	if closeRes.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(closeRes.Body)
		t.Fatalf("close billing: %d %s", closeRes.StatusCode, body)
	}
	closeRes.Body.Close()

	invRes := doStoreJSON(t, client, http.MethodGet, server.URL+"/api/v1/me/invoices", storeToken, nil)
	if invRes.StatusCode != http.StatusOK {
		t.Fatalf("invoices: %d", invRes.StatusCode)
	}
	var invList struct {
		Items []struct {
			ID string `json:"id"`
		} `json:"items"`
	}
	if err := json.NewDecoder(invRes.Body).Decode(&invList); err != nil {
		t.Fatal(err)
	}
	invRes.Body.Close()
	if len(invList.Items) == 0 {
		t.Fatal("expected invoice after close")
	}
	invID := invList.Items[0].ID

	pixRes := doStoreJSON(t, client, http.MethodPost, server.URL+"/api/v1/me/invoices/"+invID+"/pix-charge", storeToken, nil)
	if pixRes.StatusCode != http.StatusOK && pixRes.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(pixRes.Body)
		t.Fatalf("pix charge: %d %s", pixRes.StatusCode, body)
	}
	var charge struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(pixRes.Body).Decode(&charge); err != nil {
		t.Fatal(err)
	}
	pixRes.Body.Close()

	simReq, _ := http.NewRequest(http.MethodPost, server.URL+"/api/v1/dev/pix/simulate/"+charge.ID, nil)
	simRes, err := client.Do(simReq)
	if err != nil {
		t.Fatal(err)
	}
	if simRes.StatusCode != http.StatusOK && simRes.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(simRes.Body)
		t.Fatalf("simulate pix: %d %s", simRes.StatusCode, body)
	}
	simRes.Body.Close()

	dashRes := doAdminJSON(t, client, http.MethodGet, server.URL+"/api/v1/admin/reports/dashboard", adminCookie, nil)
	if dashRes.StatusCode != http.StatusOK {
		t.Fatalf("dashboard: %d", dashRes.StatusCode)
	}
	dashRes.Body.Close()
}
