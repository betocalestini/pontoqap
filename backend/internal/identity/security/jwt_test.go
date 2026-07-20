package security_test

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/store-platform/store/internal/identity/security"
)

func TestIssueAndParseAdminToken(t *testing.T) {
	secret := "test-session-secret-min-16"
	userID := uuid.New()
	sessionID := uuid.New()
	exp := time.Now().Add(8 * time.Hour)

	raw, err := security.IssueAdminToken(secret, userID, sessionID, exp)
	if err != nil {
		t.Fatal(err)
	}
	gotUser, gotSession, err := security.ParseAdminToken(secret, raw)
	if err != nil {
		t.Fatal(err)
	}
	if gotUser != userID || gotSession != sessionID {
		t.Fatalf("got user %v session %v", gotUser, gotSession)
	}
}

func TestParseAdminTokenExpired(t *testing.T) {
	secret := "test-session-secret-min-16"
	userID := uuid.New()
	sessionID := uuid.New()
	exp := time.Now().Add(-time.Minute)

	raw, err := security.IssueAdminToken(secret, userID, sessionID, exp)
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = security.ParseAdminToken(secret, raw)
	if err == nil {
		t.Fatal("expected expired token error")
	}
}

func TestParseAdminTokenWrongSecret(t *testing.T) {
	userID := uuid.New()
	sessionID := uuid.New()
	raw, err := security.IssueAdminToken("secret-one-min-16-chars", userID, sessionID, time.Now().Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = security.ParseAdminToken("secret-two-min-16-chars", raw)
	if err == nil {
		t.Fatal("expected invalid signature")
	}
}
