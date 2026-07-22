package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	"github.com/store-platform/store/tests/testdb"
)

func TestAdminRevokeInvitation(t *testing.T) {
	testdb.MigrateUp(t)
	pool := testdb.Pool(t)
	ctx := context.Background()
	if err := testdb.Reset(ctx, pool); err != nil {
		t.Fatal(err)
	}

	admin, _ := testdb.SeedSystemAdmin(ctx, pool, testdb.UniqueEmail(t, "sys-revoke"))
	inviteEmail := testdb.UniqueEmail(t, "pending-inv")
	cust, _ := testdb.SeedCustomer(ctx, pool, inviteEmail, "Convidado")
	mgr, _ := testdb.SeedManager(ctx, pool, testdb.UniqueEmail(t, "mgr-revoke"))
	_ = testdb.ApproveCustomer(ctx, pool, cust.ID, mgr.UserID, 10_000)

	handler := newIntegrationHandler(t, pool)
	adminToken := adminLoginToken(t, handler, admin.Email, "password123")

	inviteBody, _ := json.Marshal(map[string]string{
		"email":   inviteEmail,
		"name":    "Convidado",
		"role_id": "a0000000-0000-4000-8000-000000000002",
	})
	invReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/users/invitations", bytes.NewReader(inviteBody))
	invReq.Header.Set("Content-Type", "application/json")
	invReq.Header.Set("X-App-Audience", "admin")
	invReq.Header.Set("Authorization", "Bearer "+adminToken)
	invRec := httptest.NewRecorder()
	handler.ServeHTTP(invRec, invReq)
	if invRec.Code != http.StatusNoContent {
		t.Fatalf("invite: %d %s", invRec.Code, invRec.Body.String())
	}

	var invID uuid.UUID
	if err := pool.QueryRow(ctx, `
		SELECT id FROM admin_invitations
		WHERE LOWER(email) = LOWER($1) AND revoked_at IS NULL
		ORDER BY created_at DESC LIMIT 1
	`, inviteEmail).Scan(&invID); err != nil {
		t.Fatal(err)
	}

	revReq := httptest.NewRequest(http.MethodPost, "/api/v1/admin/users/invitations/"+invID.String()+"/revoke", nil)
	revReq.Header.Set("X-App-Audience", "admin")
	revReq.Header.Set("Authorization", "Bearer "+adminToken)
	revRec := httptest.NewRecorder()
	handler.ServeHTTP(revRec, revReq)
	if revRec.Code != http.StatusNoContent {
		t.Fatalf("revoke: %d %s", revRec.Code, revRec.Body.String())
	}

	var revoked bool
	_ = pool.QueryRow(ctx, `SELECT revoked_at IS NOT NULL FROM admin_invitations WHERE id = $1`, invID).Scan(&revoked)
	if !revoked {
		t.Fatal("invitation should be revoked")
	}
}

func TestSystemAdminRevokeUserSessions(t *testing.T) {
	testdb.MigrateUp(t)
	pool := testdb.Pool(t)
	ctx := context.Background()
	if err := testdb.Reset(ctx, pool); err != nil {
		t.Fatal(err)
	}

	admin, _ := testdb.SeedSystemAdmin(ctx, pool, testdb.UniqueEmail(t, "sys-sess"))
	target, _ := testdb.SeedManager(ctx, pool, testdb.UniqueEmail(t, "mgr-sess"))
	handler := newIntegrationHandler(t, pool)
	_ = adminLoginToken(t, handler, target.Email, "password123")

	adminToken := adminLoginToken(t, handler, admin.Email, "password123")
	body, _ := json.Marshal(map[string]string{"password": "password123"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/users/"+target.UserID.String()+"/sessions/revoke", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-App-Audience", "admin")
	req.Header.Set("Authorization", "Bearer "+adminToken)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("revoke sessions: %d %s", rec.Code, rec.Body.String())
	}
}
