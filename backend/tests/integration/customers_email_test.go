package integration_test

import (
	"context"
	"encoding/json"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/store-platform/store/internal/customers"
	"github.com/store-platform/store/internal/identity"
	identitypostgres "github.com/store-platform/store/internal/identity/postgres"
	"github.com/store-platform/store/internal/jobs"
	"github.com/store-platform/store/internal/platform/config"
	"github.com/store-platform/store/tests/testdb"
)

func TestCustomerRegisterVerifyAndLogin(t *testing.T) {
	testdb.MigrateUp(t)
	pool := testdb.Pool(t)
	ctx := context.Background()
	if err := testdb.Reset(ctx, pool); err != nil {
		t.Fatal(err)
	}

	jobRepo := jobs.NewRepository(pool)
	verify := identity.NewVerificationService(pool, jobRepo, config.AppConfig{StoreWebURL: "http://localhost:5173"}, config.CustomerConfig{DefaultCreditLimitCents: 25_000})
	custSvc := customers.NewService(pool, verify)

	email := testdb.UniqueEmail(t, "reg")
	c, err := custSvc.Register(ctx, customers.RegisterInput{
		Name: "Novo Cliente", Email: email, Password: "password123",
	})
	if err != nil || c == nil {
		t.Fatalf("register: %v", err)
	}

	var payload []byte
	err = pool.QueryRow(ctx, `
		SELECT payload FROM outbox_events WHERE event_type = 'user.verification_requested' ORDER BY created_at DESC LIMIT 1
	`).Scan(&payload)
	if err != nil {
		t.Fatal(err)
	}
	var p struct {
		VerifyURL string `json:"verify_url"`
	}
	if err := json.Unmarshal(payload, &p); err != nil {
		t.Fatal(err)
	}
	u, err := url.Parse(p.VerifyURL)
	if err != nil {
		t.Fatal(err)
	}
	token := u.Query().Get("token")
	if token == "" {
		t.Fatal("missing token in verify_url")
	}
	if err := verify.VerifyEmail(ctx, token); err != nil {
		t.Fatal(err)
	}

	got, err := custSvc.GetByID(ctx, c.ID)
	if err != nil || got.Status != "approved" || got.CreditLimitCents != 25_000 {
		t.Fatalf("after verify: %+v err=%v", got, err)
	}

	idSvc := identity.NewService(identitypostgres.NewRepository(pool), time.Hour, time.Hour, "test-session-secret-min-16", nil)
	res, err := idSvc.Login(ctx, identity.LoginInput{Email: email, Password: "password123", Audience: "store"})
	if err != nil || res == nil || res.SessionToken == "" {
		t.Fatalf("login after verify: %v", err)
	}
}

func TestExtractTokenFromURL(t *testing.T) {
	u, _ := url.Parse("http://localhost/verificar-email?token=abc123")
	if strings.TrimSpace(u.Query().Get("token")) != "abc123" {
		t.Fatal("parse")
	}
}
