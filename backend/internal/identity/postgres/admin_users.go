package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/store-platform/store/internal/identity"
)

type AdminUsersRepository struct {
	pool *pgxpool.Pool
}

func NewAdminUsersRepository(pool *pgxpool.Pool) *AdminUsersRepository {
	return &AdminUsersRepository{pool: pool}
}

func (r *AdminUsersRepository) ListStaff(ctx context.Context) ([]identity.StaffUserSummary, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT u.id, u.name, u.email, u.status, u.mfa_enabled, u.last_login_at, u.created_at
		FROM users u
		WHERE (
			u.status = 'invited'
			AND NOT EXISTS (SELECT 1 FROM customers c WHERE c.user_id = u.id)
		) OR EXISTS (
			SELECT 1 FROM user_roles ur
			JOIN roles ro ON ro.id = ur.role_id
			WHERE ur.user_id = u.id AND ro.code <> 'customer'
		)
		ORDER BY u.created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []identity.StaffUserSummary
	for rows.Next() {
		var s identity.StaffUserSummary
		if err := rows.Scan(&s.ID, &s.Name, &s.Email, &s.Status, &s.MFAEnabled, &s.LastLoginAt, &s.CreatedAt); err != nil {
			return nil, err
		}
		roles, err := r.pool.Query(ctx, `
			SELECT ro.code FROM user_roles ur JOIN roles ro ON ro.id = ur.role_id WHERE ur.user_id = $1
		`, s.ID)
		if err != nil {
			return nil, err
		}
		for roles.Next() {
			var code string
			if err := roles.Scan(&code); err != nil {
				roles.Close()
				return nil, err
			}
			s.Roles = append(s.Roles, code)
		}
		roles.Close()
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *AdminUsersRepository) GetStaffByID(ctx context.Context, userID uuid.UUID) (*identity.StaffUserSummary, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT u.id, u.name, u.email, u.status, u.mfa_enabled, u.last_login_at, u.created_at
		FROM users u
		WHERE u.id = $1
		AND (
			(u.status = 'invited' AND NOT EXISTS (SELECT 1 FROM customers c WHERE c.user_id = u.id))
			OR EXISTS (
				SELECT 1 FROM user_roles ur JOIN roles ro ON ro.id = ur.role_id
				WHERE ur.user_id = u.id AND ro.code <> 'customer'
			)
		)
	`, userID)
	var s identity.StaffUserSummary
	err := row.Scan(&s.ID, &s.Name, &s.Email, &s.Status, &s.MFAEnabled, &s.LastLoginAt, &s.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	roles, err := r.pool.Query(ctx, `
		SELECT ro.code FROM user_roles ur JOIN roles ro ON ro.id = ur.role_id WHERE ur.user_id = $1
	`, userID)
	if err != nil {
		return nil, err
	}
	defer roles.Close()
	for roles.Next() {
		var code string
		if err := roles.Scan(&code); err != nil {
			return nil, err
		}
		s.Roles = append(s.Roles, code)
	}
	return &s, roles.Err()
}

func (r *AdminUsersRepository) ListInternalRoles(ctx context.Context) ([]identity.RoleInfo, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, code, name FROM roles
		WHERE code IN ('system_admin', 'manager', 'inventory_operator', 'finance_operator')
		ORDER BY code
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []identity.RoleInfo
	for rows.Next() {
		var ri identity.RoleInfo
		if err := rows.Scan(&ri.ID, &ri.Code, &ri.Name); err != nil {
			return nil, err
		}
		out = append(out, ri)
	}
	return out, rows.Err()
}

func (r *AdminUsersRepository) RoleCodeByID(ctx context.Context, roleID uuid.UUID) (string, error) {
	var code string
	err := r.pool.QueryRow(ctx, `SELECT code FROM roles WHERE id = $1`, roleID).Scan(&code)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", errNotFoundIdentity()
	}
	return code, err
}

func (r *AdminUsersRepository) FindUserIDByEmailTx(ctx context.Context, tx pgx.Tx, email string) (*uuid.UUID, string, error) {
	var id uuid.UUID
	var status string
	err := tx.QueryRow(ctx, `SELECT id, status FROM users WHERE LOWER(email) = LOWER($1)`, email).Scan(&id, &status)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, "", nil
	}
	if err != nil {
		return nil, "", err
	}
	return &id, status, nil
}

func (r *AdminUsersRepository) UserHasCustomerRecordTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID) (bool, error) {
	var exists bool
	err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM customers WHERE user_id = $1)`, userID).Scan(&exists)
	return exists, err
}

func (r *AdminUsersRepository) ReplaceUserRoleTx(ctx context.Context, tx pgx.Tx, userID, roleID uuid.UUID) error {
	_, err := tx.Exec(ctx, `
		DELETE FROM user_roles ur USING roles ro
		WHERE ur.user_id = $1 AND ur.role_id = ro.id AND ro.code <> 'customer'
	`, userID)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2)`, userID, roleID)
	return err
}

func (r *AdminUsersRepository) UserHasCustomerRoleTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID) (bool, error) {
	var exists bool
	err := tx.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM user_roles ur JOIN roles ro ON ro.id = ur.role_id
			WHERE ur.user_id = $1 AND ro.code = 'customer'
		)
	`, userID).Scan(&exists)
	return exists, err
}

func (r *AdminUsersRepository) FindInvitation(ctx context.Context, id uuid.UUID) (*identity.AdminInvitationRecord, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, email, role_id, expires_at, invited_by, accepted_at, revoked_at
		FROM admin_invitations WHERE id = $1
	`, id)
	return scanInvitation(row)
}

func (r *AdminUsersRepository) FindInvitationByTokenHash(ctx context.Context, hash string) (*identity.AdminInvitationRecord, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, email, role_id, expires_at, invited_by, accepted_at, revoked_at
		FROM admin_invitations WHERE token_hash = $1
	`, hash)
	return scanInvitation(row)
}

func scanInvitation(row pgx.Row) (*identity.AdminInvitationRecord, error) {
	var inv identity.AdminInvitationRecord
	err := row.Scan(&inv.ID, &inv.Email, &inv.RoleID, &inv.ExpiresAt, &inv.InvitedBy, &inv.AcceptedAt, &inv.RevokedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

func (r *AdminUsersRepository) RevokeInvitation(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE admin_invitations SET revoked_at = NOW() WHERE id = $1`, id)
	return err
}

func (r *AdminUsersRepository) ReplaceUserRole(ctx context.Context, userID, roleID uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if err := r.ReplaceUserRoleTx(ctx, tx, userID, roleID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *AdminUsersRepository) UpdateUserStatus(ctx context.Context, userID uuid.UUID, status string) error {
	_, err := r.pool.Exec(ctx, `UPDATE users SET status = $2, updated_at = NOW() WHERE id = $1`, userID, status)
	return err
}

func (r *AdminUsersRepository) CountActiveSystemAdmins(ctx context.Context, excludeUserID uuid.UUID) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(DISTINCT u.id) FROM users u
		JOIN user_roles ur ON ur.user_id = u.id
		JOIN roles ro ON ro.id = ur.role_id AND ro.code = 'system_admin'
		WHERE u.status = 'active' AND u.id <> $1
	`, excludeUserID).Scan(&n)
	return n, err
}

func (r *AdminUsersRepository) ExistsSystemAdmin(ctx context.Context) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM user_roles ur
			JOIN roles ro ON ro.id = ur.role_id AND ro.code = 'system_admin'
		)
	`).Scan(&exists)
	return exists, err
}

func (r *AdminUsersRepository) CreateBootstrapSystemAdmin(ctx context.Context, email, name, passwordHash string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	userID := uuid.New()
	_, err = tx.Exec(ctx, `
		INSERT INTO users (id, name, email, password_hash, status)
		VALUES ($1, $2, $3, $4, 'active')
	`, userID, name, email, passwordHash)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO user_roles (user_id, role_id)
		VALUES ($1, 'a0000000-0000-4000-8000-000000000001')
	`, userID)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func errNotFoundIdentity() error {
	return pgx.ErrNoRows
}
