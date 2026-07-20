package identity

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type AdminUsersRepository interface {
	ListStaff(ctx context.Context) ([]StaffUserSummary, error)
	GetStaffByID(ctx context.Context, userID uuid.UUID) (*StaffUserSummary, error)
	ListInternalRoles(ctx context.Context) ([]RoleInfo, error)
	RoleCodeByID(ctx context.Context, roleID uuid.UUID) (string, error)
	FindUserIDByEmailTx(ctx context.Context, tx pgx.Tx, email string) (*uuid.UUID, string, error)
	UserHasCustomerRoleTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID) (bool, error)
	UserHasCustomerRecordTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID) (bool, error)
	ReplaceUserRoleTx(ctx context.Context, tx pgx.Tx, userID, roleID uuid.UUID) error
	FindInvitation(ctx context.Context, id uuid.UUID) (*AdminInvitationRecord, error)
	FindInvitationByTokenHash(ctx context.Context, hash string) (*AdminInvitationRecord, error)
	RevokeInvitation(ctx context.Context, id uuid.UUID) error
	ReplaceUserRole(ctx context.Context, userID, roleID uuid.UUID) error
	UpdateUserStatus(ctx context.Context, userID uuid.UUID, status string) error
	CountActiveSystemAdmins(ctx context.Context, excludeUserID uuid.UUID) (int, error)
	ExistsSystemAdmin(ctx context.Context) (bool, error)
	CreateBootstrapSystemAdmin(ctx context.Context, email, name, passwordHash string) error
}

type AdminInvitationRecord struct {
	ID         uuid.UUID
	Email      string
	RoleID     uuid.UUID
	ExpiresAt  time.Time
	InvitedBy  uuid.UUID
	AcceptedAt *time.Time
	RevokedAt  *time.Time
}
