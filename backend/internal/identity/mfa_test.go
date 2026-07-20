package identity_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pquerna/otp/totp"

	"github.com/store-platform/store/internal/identity"
	"github.com/store-platform/store/internal/identity/security"
)

type mfaRepo struct {
	user identity.User
}

func (m *mfaRepo) FindUserByEmail(ctx context.Context, email string) (*identity.User, error) {
	return &m.user, nil
}
func (m *mfaRepo) FindUserByID(ctx context.Context, id uuid.UUID) (*identity.User, error) {
	return &m.user, nil
}
func (m *mfaRepo) UpdateUserLoginState(ctx context.Context, userID uuid.UUID, failed int, locked *time.Time, last *time.Time) error {
	return nil
}
func (m *mfaRepo) RevokeUserSessions(ctx context.Context, userID uuid.UUID, audience string) error {
	return nil
}
func (m *mfaRepo) CreateSession(ctx context.Context, s identity.Session, ip, ua string) error {
	return nil
}
func (m *mfaRepo) FindSessionByTokenHash(ctx context.Context, hash string) (*identity.Session, error) {
	return nil, nil
}
func (m *mfaRepo) FindSessionByID(ctx context.Context, sessionID uuid.UUID) (*identity.Session, error) {
	return nil, nil
}
func (m *mfaRepo) RevokeSession(ctx context.Context, id uuid.UUID) error { return nil }
func (m *mfaRepo) ListUserPermissions(ctx context.Context, userID uuid.UUID) ([]string, error) {
	return []string{"products.read"}, nil
}
func (m *mfaRepo) ListUserRoles(ctx context.Context, userID uuid.UUID) ([]string, error) {
	return []string{"manager"}, nil
}
func (m *mfaRepo) FindCustomerIDByUser(ctx context.Context, userID uuid.UUID) (*uuid.UUID, error) {
	return nil, nil
}
func (m *mfaRepo) IsCustomerBlocked(ctx context.Context, userID uuid.UUID) (bool, error) {
	return false, nil
}
func (m *mfaRepo) UpdateMFA(ctx context.Context, userID uuid.UUID, secret string, enabled bool) error {
	m.user.MFASecret = secret
	m.user.MFAEnabled = enabled
	return nil
}

func TestAdminLoginRequiresMFACodeWhenEnabled(t *testing.T) {
	secret := "JBSWY3DPEHPK3PXP"
	hash, err := security.HashPassword("pass")
	if err != nil {
		t.Fatal(err)
	}
	repo := &mfaRepo{user: identity.User{
		ID: uuid.New(), Email: "m@test.local", Status: "active",
		PasswordHash: hash,
		MFAEnabled:   true, MFASecret: secret,
	}}
	svc := identity.NewService(repo, time.Hour, time.Hour, "test-session-secret-min-16")
	ctx := context.Background()

	res, err := svc.Login(ctx, identity.LoginInput{Email: "m@test.local", Password: "pass", Audience: "admin"})
	if err != nil || !res.MFARequired {
		t.Fatalf("expected mfa required, got %+v err=%v", res, err)
	}
	code, err := totp.GenerateCode(secret, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	res, err = svc.Login(ctx, identity.LoginInput{Email: "m@test.local", Password: "pass", Audience: "admin", MFACode: code})
	if err != nil || res.SessionToken == "" {
		t.Fatalf("login with totp: %v %+v", err, res)
	}
}
