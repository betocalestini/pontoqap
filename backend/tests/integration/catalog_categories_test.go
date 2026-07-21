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

	"github.com/google/uuid"
	"github.com/store-platform/store/internal/app"
	"github.com/store-platform/store/internal/catalog"
	"github.com/store-platform/store/internal/identity"
	identitypostgres "github.com/store-platform/store/internal/identity/postgres"
	"github.com/store-platform/store/internal/inventory"
	"github.com/store-platform/store/internal/platform/config"
	"github.com/store-platform/store/tests/testdb"
)

func TestAdminDeleteProductCategoryUnlinksProducts(t *testing.T) {
	testdb.MigrateUp(t)
	pool := testdb.Pool(t)
	ctx := context.Background()
	if err := testdb.Reset(ctx, pool); err != nil {
		t.Fatal(err)
	}

	mgr, err := testdb.SeedManager(ctx, pool, testdb.UniqueEmail(t, "mgr-cat"))
	if err != nil {
		t.Fatal(err)
	}

	catSvc := catalog.NewService(pool)
	invSvc := inventory.NewService(pool)
	cat, err := catSvc.CreateCategory(ctx, catalog.CreateCategoryInput{Name: "Teste Cat", Slug: "teste-cat-del"})
	if err != nil {
		t.Fatal(err)
	}

	prod, err := testdb.SeedProduct(ctx, pool, "Com Cat", "SKU-CAT-DEL", 1000)
	if err != nil {
		t.Fatal(err)
	}
	_, err = pool.Exec(ctx, `UPDATE products SET category_id = $2 WHERE id = $1`, prod.ProductID, cat.ID)
	if err != nil {
		t.Fatal(err)
	}
	_ = invSvc

	secret := "test-session-secret-min-16"
	idSvc := identity.NewService(identitypostgres.NewRepository(pool), time.Hour, 8*time.Hour, secret)
	cfg := config.Config{
		AppEnv:   "test",
		Security: config.SecurityConfig{SessionSecret: secret, AdminMFARequired: false},
		Session:  config.SessionConfig{StoreTTL: time.Hour, AdminTTL: 8 * time.Hour},
		HTTP:     config.HTTPConfig{CORSOrigins: []string{"*"}},
	}
	handler := app.NewRouter(cfg, pool, idSvc, nil, slog.Default())
	token := adminLoginToken(t, handler, mgr.Email, "password123")

	delURL := "/api/v1/admin/categories/" + cat.ID.String()
	req := httptest.NewRequest(http.MethodDelete, delURL, nil)
	req.Header.Set("X-App-Audience", "admin")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("delete status %d body %s", rec.Code, rec.Body.String())
	}
	var delRes struct {
		ProductsUnlinked int `json:"products_unlinked"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &delRes); err != nil {
		t.Fatal(err)
	}
	if delRes.ProductsUnlinked != 1 {
		t.Fatalf("want 1 unlinked got %d", delRes.ProductsUnlinked)
	}

	var catID *uuid.UUID
	if err := pool.QueryRow(ctx, `SELECT category_id FROM products WHERE id = $1`, prod.ProductID).Scan(&catID); err != nil {
		t.Fatal(err)
	}
	if catID != nil {
		t.Fatal("expected null category_id")
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/admin/categories", nil)
	listReq.Header.Set("X-App-Audience", "admin")
	listReq.Header.Set("Authorization", "Bearer "+token)
	listRec := httptest.NewRecorder()
	handler.ServeHTTP(listRec, listReq)
	var listBody struct {
		Items []struct {
			ID string `json:"id"`
		} `json:"items"`
	}
	_ = json.Unmarshal(listRec.Body.Bytes(), &listBody)
	for _, it := range listBody.Items {
		if it.ID == cat.ID.String() {
			t.Fatal("category still listed")
		}
	}
}

func TestAdminPatchProductCategory(t *testing.T) {
	testdb.MigrateUp(t)
	pool := testdb.Pool(t)
	ctx := context.Background()
	if err := testdb.Reset(ctx, pool); err != nil {
		t.Fatal(err)
	}
	mgr, _ := testdb.SeedManager(ctx, pool, testdb.UniqueEmail(t, "mgr-cat2"))
	catSvc := catalog.NewService(pool)
	cat, err := catSvc.CreateCategory(ctx, catalog.CreateCategoryInput{Name: "Antiga", Slug: "antiga"})
	if err != nil {
		t.Fatal(err)
	}

	secret := "test-session-secret-min-16"
	idSvc := identity.NewService(identitypostgres.NewRepository(pool), time.Hour, 8*time.Hour, secret)
	cfg := config.Config{
		AppEnv:   "test",
		Security: config.SecurityConfig{SessionSecret: secret, AdminMFARequired: false},
		Session:  config.SessionConfig{StoreTTL: time.Hour, AdminTTL: 8 * time.Hour},
		HTTP:     config.HTTPConfig{CORSOrigins: []string{"*"}},
	}
	handler := app.NewRouter(cfg, pool, idSvc, nil, slog.Default())
	token := adminLoginToken(t, handler, mgr.Email, "password123")

	patchBody, _ := json.Marshal(map[string]any{"name": "Nova", "active": false})
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/categories/"+cat.ID.String(), bytes.NewReader(patchBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-App-Audience", "admin")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("patch %d %s", rec.Code, rec.Body.String())
	}
	var updated struct {
		Name   string `json:"name"`
		Slug   string `json:"slug"`
		Active bool   `json:"active"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &updated); err != nil {
		t.Fatal(err)
	}
	if updated.Name != "Nova" || updated.Slug != "antiga" || updated.Active {
		t.Fatalf("unexpected %+v", updated)
	}
}
