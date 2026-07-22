package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/pquerna/otp/totp"

	"github.com/store-platform/store/tests/testdb"
)

func TestAdminMFASetupAndVerifyHTTP(t *testing.T) {
	testdb.MigrateUp(t)
	pool := testdb.Pool(t)
	ctx := context.Background()
	if err := testdb.Reset(ctx, pool); err != nil {
		t.Fatal(err)
	}

	mgr, err := testdb.SeedManager(ctx, pool, testdb.UniqueEmail(t, "mgr-mfa"))
	if err != nil {
		t.Fatal(err)
	}
	handler := newIntegrationHandler(t, pool)
	token := adminLoginToken(t, handler, mgr.Email, "password123")

	setupReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/mfa/setup", nil)
	setupReq.Header.Set("X-App-Audience", "admin")
	setupReq.Header.Set("Authorization", "Bearer "+token)
	setupRec := httptest.NewRecorder()
	handler.ServeHTTP(setupRec, setupReq)
	if setupRec.Code != http.StatusOK {
		t.Fatalf("mfa setup: %d %s", setupRec.Code, setupRec.Body.String())
	}
	var setup struct {
		Secret string `json:"secret"`
	}
	if err := json.NewDecoder(setupRec.Body).Decode(&setup); err != nil || setup.Secret == "" {
		t.Fatalf("setup response: %v secret=%q", err, setup.Secret)
	}

	code, err := totp.GenerateCode(setup.Secret, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	verifyBody, _ := json.Marshal(map[string]string{"code": code})
	verifyReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/mfa/verify", bytes.NewReader(verifyBody))
	verifyReq.Header.Set("Content-Type", "application/json")
	verifyReq.Header.Set("X-App-Audience", "admin")
	verifyReq.Header.Set("Authorization", "Bearer "+token)
	verifyRec := httptest.NewRecorder()
	handler.ServeHTTP(verifyRec, verifyReq)
	if verifyRec.Code != http.StatusNoContent {
		t.Fatalf("mfa verify: %d %s", verifyRec.Code, verifyRec.Body.String())
	}

	meReq := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	meReq.Header.Set("X-App-Audience", "admin")
	meReq.Header.Set("Authorization", "Bearer "+token)
	meRec := httptest.NewRecorder()
	handler.ServeHTTP(meRec, meReq)
	if meRec.Code != http.StatusOK {
		t.Fatalf("me: %d %s", meRec.Code, meRec.Body.String())
	}
	var me struct {
		MFAEnabled bool `json:"mfa_enabled"`
	}
	if err := json.NewDecoder(meRec.Body).Decode(&me); err != nil {
		t.Fatal(err)
	}
	if !me.MFAEnabled {
		t.Fatal("expected mfa_enabled true after verify")
	}
}
