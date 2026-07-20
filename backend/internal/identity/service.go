package identity

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/pquerna/otp/totp"

	"github.com/store-platform/store/internal/identity/security"
	platformerrors "github.com/store-platform/store/internal/platform/errors"
)

const maxFailedAttempts = 5
const lockDuration = 15 * time.Minute

type Service struct {
	repo       Repository
	storeTTL   time.Duration
	adminTTL   time.Duration
}

func NewService(repo Repository, storeTTL, adminTTL time.Duration) *Service {
	return &Service{repo: repo, storeTTL: storeTTL, adminTTL: adminTTL}
}

type LoginInput struct {
	Email     string
	Password  string
	Audience  string
	MFACode   string
	IP        string
	UserAgent string
}

type LoginResult struct {
	SessionToken string
	User         AuthUser
	MFARequired  bool
}

func (s *Service) Login(ctx context.Context, in LoginInput) (*LoginResult, error) {
	user, err := s.repo.FindUserByEmail(ctx, in.Email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errInvalidCredentials()
	}
	if user.Status == "blocked" {
		return nil, errForbidden("Conta bloqueada")
	}
	if user.LockedUntil != nil && user.LockedUntil.After(time.Now()) {
		return nil, errAccountLocked()
	}
	ok, err := security.VerifyPassword(user.PasswordHash, in.Password)
	if err != nil || !ok {
		failed := user.FailedLoginAttempts + 1
		var locked *time.Time
		if failed >= maxFailedAttempts {
			t := time.Now().Add(lockDuration)
			locked = &t
			failed = 0
		}
		_ = s.repo.UpdateUserLoginState(ctx, user.ID, failed, locked, user.LastLoginAt)
		return nil, errInvalidCredentials()
	}

	if in.Audience == "admin" {
		roles, _ := s.repo.ListUserRoles(ctx, user.ID)
		if !hasAdminRole(roles) {
			return nil, errForbidden("Acesso não permitido ao painel administrativo")
		}
		if user.MFAEnabled {
			if in.MFACode == "" {
				return &LoginResult{MFARequired: true}, nil
			}
			if !totp.Validate(in.MFACode, user.MFASecret) {
				return nil, errInvalidCredentials()
			}
		}
	}

	now := time.Now()
	_ = s.repo.UpdateUserLoginState(ctx, user.ID, 0, nil, &now)
	_ = s.repo.RevokeUserSessions(ctx, user.ID, in.Audience)

	token, tokenHash, err := newSessionToken()
	if err != nil {
		return nil, err
	}
	ttl := s.storeTTL
	if in.Audience == "admin" {
		ttl = s.adminTTL
	}
	sess := Session{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: tokenHash,
		Audience:  in.Audience,
		ExpiresAt: now.Add(ttl),
	}
	if err := s.repo.CreateSession(ctx, sess, in.IP, in.UserAgent); err != nil {
		return nil, err
	}

	authUser, err := s.buildAuthUser(ctx, *user)
	if err != nil {
		return nil, err
	}
	return &LoginResult{SessionToken: token, User: authUser}, nil
}

func (s *Service) AuthenticateSession(ctx context.Context, token, expectedAudience string) (*AuthUser, error) {
	hash := security.HashToken(token)
	sess, err := s.repo.FindSessionByTokenHash(ctx, hash)
	if err != nil || sess == nil {
		return nil, errUnauthorized()
	}
	if sess.RevokedAt != nil || sess.ExpiresAt.Before(time.Now()) {
		return nil, errUnauthorized()
	}
	if sess.Audience != expectedAudience {
		return nil, errUnauthorized()
	}
	user, err := s.repo.FindUserByID(ctx, sess.UserID)
	if err != nil || user == nil || user.Status != "active" {
		return nil, errUnauthorized()
	}
	authUser, err := s.buildAuthUser(ctx, *user)
	if err != nil {
		return nil, err
	}
	return &authUser, nil
}

func (s *Service) Logout(ctx context.Context, token string) error {
	hash := security.HashToken(token)
	sess, err := s.repo.FindSessionByTokenHash(ctx, hash)
	if err != nil || sess == nil {
		return nil
	}
	return s.repo.RevokeSession(ctx, sess.ID)
}

func (s *Service) SetupMFA(ctx context.Context, userID uuid.UUID) (secret string, uri string, err error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "Store Platform",
		AccountName: userID.String(),
	})
	if err != nil {
		return "", "", err
	}
	if err := s.repo.UpdateMFA(ctx, userID, key.Secret(), false); err != nil {
		return "", "", err
	}
	return key.Secret(), key.URL(), nil
}

func (s *Service) VerifyMFA(ctx context.Context, userID uuid.UUID, code string) error {
	user, err := s.repo.FindUserByID(ctx, userID)
	if err != nil || user == nil {
		return errNotFound()
	}
	if !totp.Validate(code, user.MFASecret) {
		return errInvalidCredentials()
	}
	return s.repo.UpdateMFA(ctx, userID, user.MFASecret, true)
}

func (s *Service) buildAuthUser(ctx context.Context, user User) (AuthUser, error) {
	perms, err := s.repo.ListUserPermissions(ctx, user.ID)
	if err != nil {
		return AuthUser{}, err
	}
	roles, err := s.repo.ListUserRoles(ctx, user.ID)
	if err != nil {
		return AuthUser{}, err
	}
	cid, _ := s.repo.FindCustomerIDByUser(ctx, user.ID)
	return AuthUser{
		User:        user,
		Permissions: perms,
		Roles:       roles,
		CustomerID:  cid,
	}, nil
}

func newSessionToken() (raw, hash string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", "", err
	}
	raw = base64.RawURLEncoding.EncodeToString(b)
	hash = security.HashToken(raw)
	return raw, hash, nil
}

func hasAdminRole(roles []string) bool {
	for _, r := range roles {
		if r == "manager" || r == "system_admin" {
			return true
		}
	}
	return false
}

type AppError struct {
	Code    string
	Message string
	Status  int
}

func (e *AppError) Error() string { return e.Message }

func errInvalidCredentials() error {
	return &AppError{Code: platformerrors.CodeInvalidCredentials, Message: "E-mail ou senha inválidos", Status: 401}
}

func errUnauthorized() error {
	return &AppError{Code: platformerrors.CodeUnauthorized, Message: "Não autenticado", Status: 401}
}

func errForbidden(msg string) error {
	return &AppError{Code: platformerrors.CodeForbidden, Message: msg, Status: 403}
}

func errAccountLocked() error {
	return &AppError{Code: platformerrors.CodeAccountLocked, Message: "Conta temporariamente bloqueada", Status: 423}
}

func errNotFound() error {
	return &AppError{Code: platformerrors.CodeNotFound, Message: "Não encontrado", Status: 404}
}

func AsAppError(err error) *AppError {
	var ae *AppError
	if errors.As(err, &ae) {
		return ae
	}
	return nil
}
