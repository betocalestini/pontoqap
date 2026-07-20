package http_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	"github.com/store-platform/store/internal/identity"
	identityhttp "github.com/store-platform/store/internal/identity/transport/http"
)

func TestRequirePermissionDeniesWithoutPermission(t *testing.T) {
	user := &identity.AuthUser{
		User:        identity.User{ID: uuid.New()},
		Permissions: []string{"products.read"},
	}
	handler := identityhttp.RequirePermission("products.write")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	req = req.WithContext(identityhttp.ContextWithAuthUser(context.Background(), user))
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestRequirePermissionAllowsWithPermission(t *testing.T) {
	user := &identity.AuthUser{
		User:        identity.User{ID: uuid.New()},
		Permissions: []string{"products.write"},
	}
	called := false
	handler := identityhttp.RequirePermission("products.write")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	req = req.WithContext(identityhttp.ContextWithAuthUser(context.Background(), user))
	handler.ServeHTTP(rec, req)
	if !called || rec.Code != http.StatusOK {
		t.Fatalf("expected handler to run, code=%d", rec.Code)
	}
}
