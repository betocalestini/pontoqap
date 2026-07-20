package identity

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID                  uuid.UUID
	Name                string
	Email               string
	Phone               string
	PasswordHash        string
	Status              string
	MFAEnabled          bool
	MFASecret           string
	FailedLoginAttempts int
	LockedUntil         *time.Time
	LastLoginAt         *time.Time
}

type Session struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	TokenHash string
	Audience  string
	ExpiresAt time.Time
	RevokedAt *time.Time
}

type AuthUser struct {
	User        User
	Permissions []string
	Roles       []string
	CustomerID  *uuid.UUID
}

type Repository interface {
	FindUserByEmail(ctx context.Context, email string) (*User, error)
	FindUserByID(ctx context.Context, id uuid.UUID) (*User, error)
	UpdateUserLoginState(ctx context.Context, userID uuid.UUID, failed int, lockedUntil *time.Time, lastLogin *time.Time) error
	ListUserPermissions(ctx context.Context, userID uuid.UUID) ([]string, error)
	ListUserRoles(ctx context.Context, userID uuid.UUID) ([]string, error)
	CreateSession(ctx context.Context, s Session, ip, userAgent string) error
	FindSessionByTokenHash(ctx context.Context, tokenHash string) (*Session, error)
	FindSessionByID(ctx context.Context, sessionID uuid.UUID) (*Session, error)
	RevokeSession(ctx context.Context, sessionID uuid.UUID) error
	RevokeUserSessions(ctx context.Context, userID uuid.UUID, audience string) error
	UpdateMFA(ctx context.Context, userID uuid.UUID, secret string, enabled bool) error
	FindCustomerIDByUser(ctx context.Context, userID uuid.UUID) (*uuid.UUID, error)
	IsCustomerBlocked(ctx context.Context, userID uuid.UUID) (bool, error)
	EnsureBootstrapManager(ctx context.Context, email, name, passwordHash string) error
}
