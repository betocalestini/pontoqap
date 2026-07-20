package integration_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/store-platform/store/internal/billing"
	"github.com/store-platform/store/internal/identity"
	"github.com/store-platform/store/internal/identity/postgres"
	"github.com/store-platform/store/internal/identity/security"
	"github.com/store-platform/store/tests/testdb"
)

func TestIdentityLoginCreatesSession(t *testing.T) {
	testdb.MigrateUp(t)
	pool := testdb.Pool(t)
	ctx := context.Background()
	if err := testdb.Reset(ctx, pool); err != nil {
		t.Fatal(err)
	}

	email := testdb.UniqueEmail(t, "user")
	hash, _ := security.HashPassword("pass")
	userID := uuid.New()
	_, err := pool.Exec(ctx, `
		INSERT INTO users (id, name, email, password_hash, status) VALUES ($1,'U',$2,$3,'active')
	`, userID, email, hash)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = pool.Exec(ctx, `INSERT INTO user_roles (user_id, role_id) VALUES ($1,'a0000000-0000-4000-8000-000000000003')`, userID)

	repo := postgres.NewRepository(pool)
	svc := identity.NewService(repo, time.Hour, time.Hour, "test-session-secret-min-16")
	res, err := svc.Login(ctx, identity.LoginInput{Email: email, Password: "pass", Audience: "store"})
	if err != nil || res.SessionToken == "" {
		t.Fatalf("login: %v", err)
	}
	auth, err := svc.AuthenticateSession(ctx, res.SessionToken, "store")
	if err != nil || auth.User.Email != email {
		t.Fatalf("session: %v %+v", err, auth)
	}
}

func TestBillingEnsureOpenPeriod(t *testing.T) {
	testdb.MigrateUp(t)
	pool := testdb.Pool(t)
	ctx := context.Background()
	if err := testdb.Reset(ctx, pool); err != nil {
		t.Fatal(err)
	}

	mgr, _ := testdb.SeedManager(ctx, pool, testdb.UniqueEmail(t, "mgr"))
	cust, _ := testdb.SeedCustomer(ctx, pool, testdb.UniqueEmail(t, "c"), "C")
	_ = testdb.ApproveCustomer(ctx, pool, cust.ID, mgr.UserID, 10_000)

	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)

	svc := billing.NewService(pool, nil, "")
	now := time.Date(2026, 3, 15, 12, 0, 0, 0, time.UTC)
	pid, err := svc.EnsureOpenPeriodTx(ctx, tx, cust.ID, now)
	if err != nil || pid == uuid.Nil {
		t.Fatalf("period: %v", err)
	}
	pid2, err := svc.EnsureOpenPeriodTx(ctx, tx, cust.ID, now)
	if err != nil || pid2 != pid {
		t.Fatalf("expected same period id")
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
}
