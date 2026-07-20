package identity_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/store-platform/store/internal/identity"
	"github.com/store-platform/store/internal/identity/security"
	platformerrors "github.com/store-platform/store/internal/platform/errors"
)

type mockRepo struct {
	user       *identity.User
	perms      []string
	roles      []string
	sessions   []identity.Session
	loginFails int
}

func (m *mockRepo) FindUserByEmail(ctx context.Context, email string) (*identity.User, error) {
	return m.user, nil
}
func (m *mockRepo) FindUserByID(ctx context.Context, id uuid.UUID) (*identity.User, error) {
	if m.user != nil && m.user.ID == id {
		return m.user, nil
	}
	return nil, nil
}
func (m *mockRepo) UpdateUserLoginState(ctx context.Context, userID uuid.UUID, failed int, lockedUntil *time.Time, lastLogin *time.Time) error {
	m.loginFails = failed
	if m.user != nil {
		m.user.FailedLoginAttempts = failed
		m.user.LockedUntil = lockedUntil
	}
	return nil
}
func (m *mockRepo) ListUserPermissions(ctx context.Context, userID uuid.UUID) ([]string, error) {
	return m.perms, nil
}
func (m *mockRepo) ListUserRoles(ctx context.Context, userID uuid.UUID) ([]string, error) {
	return m.roles, nil
}
func (m *mockRepo) CreateSession(ctx context.Context, s identity.Session, ip, userAgent string) error {
	m.sessions = append(m.sessions, s)
	return nil
}
func (m *mockRepo) FindSessionByTokenHash(ctx context.Context, tokenHash string) (*identity.Session, error) {
	for i := range m.sessions {
		if m.sessions[i].TokenHash == tokenHash && m.sessions[i].RevokedAt == nil {
			s := m.sessions[i]
			return &s, nil
		}
	}
	return nil, nil
}
func (m *mockRepo) RevokeSession(ctx context.Context, sessionID uuid.UUID) error { return nil }
func (m *mockRepo) RevokeUserSessions(ctx context.Context, userID uuid.UUID, audience string) error {
	return nil
}
func (m *mockRepo) UpdateMFA(ctx context.Context, userID uuid.UUID, secret string, enabled bool) error {
	return nil
}
func (m *mockRepo) FindCustomerIDByUser(ctx context.Context, userID uuid.UUID) (*uuid.UUID, error) {
	return nil, nil
}
func (m *mockRepo) EnsureBootstrapManager(ctx context.Context, email, name, passwordHash string) error {
	return nil
}

func TestLoginSuccessStore(t *testing.T) {
	hash, _ := security.HashPassword("secret")
	repo := &mockRepo{
		user: &identity.User{
			ID:           uuid.New(),
			Email:        "c@test.local",
			PasswordHash: hash,
			Status:       "active",
		},
		roles: []string{"customer"},
	}
	svc := identity.NewService(repo, time.Hour, time.Hour)
	res, err := svc.Login(context.Background(), identity.LoginInput{
		Email: "c@test.local", Password: "secret", Audience: "store",
	})
	if err != nil || res == nil || res.SessionToken == "" {
		t.Fatalf("login failed: %v %+v", err, res)
	}
}

func TestLoginInvalidPasswordIncrementsFailures(t *testing.T) {
	hash, _ := security.HashPassword("secret")
	repo := &mockRepo{
		user: &identity.User{
			ID: uuid.New(), Email: "c@test.local", PasswordHash: hash, Status: "active",
		},
	}
	svc := identity.NewService(repo, time.Hour, time.Hour)
	_, err := svc.Login(context.Background(), identity.LoginInput{
		Email: "c@test.local", Password: "wrong", Audience: "store",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	ae := identity.AsAppError(err)
	if ae == nil || ae.Code != platformerrors.CodeInvalidCredentials {
		t.Fatalf("expected invalid credentials, got %v", err)
	}
	if repo.loginFails != 1 {
		t.Fatalf("expected 1 failed attempt, got %d", repo.loginFails)
	}
}

func TestLoginAdminRequiresManagerRole(t *testing.T) {
	hash, _ := security.HashPassword("secret")
	repo := &mockRepo{
		user: &identity.User{
			ID: uuid.New(), Email: "c@test.local", PasswordHash: hash, Status: "active",
		},
		roles: []string{"customer"},
	}
	svc := identity.NewService(repo, time.Hour, time.Hour)
	_, err := svc.Login(context.Background(), identity.LoginInput{
		Email: "c@test.local", Password: "secret", Audience: "admin",
	})
	ae := identity.AsAppError(err)
	if ae == nil || ae.Code != platformerrors.CodeForbidden {
		t.Fatalf("expected forbidden, got %v", err)
	}
}
