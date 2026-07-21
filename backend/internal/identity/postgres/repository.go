package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/store-platform/store/internal/identity"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) FindUserByEmail(ctx context.Context, email string) (*identity.User, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, name, email, COALESCE(phone,''), password_hash, status,
		       mfa_enabled, COALESCE(mfa_secret,''), failed_login_attempts, locked_until, last_login_at
		FROM users WHERE LOWER(email) = LOWER($1)
	`, email)
	return scanUser(row)
}

func (r *Repository) FindUserByID(ctx context.Context, id uuid.UUID) (*identity.User, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, name, email, COALESCE(phone,''), password_hash, status,
		       mfa_enabled, COALESCE(mfa_secret,''), failed_login_attempts, locked_until, last_login_at
		FROM users WHERE id = $1
	`, id)
	return scanUser(row)
}

func scanUser(row pgx.Row) (*identity.User, error) {
	var u identity.User
	err := row.Scan(&u.ID, &u.Name, &u.Email, &u.Phone, &u.PasswordHash, &u.Status,
		&u.MFAEnabled, &u.MFASecret, &u.FailedLoginAttempts, &u.LockedUntil, &u.LastLoginAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repository) UpdateUserLoginState(ctx context.Context, userID uuid.UUID, failed int, lockedUntil *time.Time, lastLogin *time.Time) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE users SET failed_login_attempts = $2, locked_until = $3, last_login_at = $4, updated_at = NOW()
		WHERE id = $1
	`, userID, failed, lockedUntil, lastLogin)
	return err
}

func (r *Repository) ListUserPermissions(ctx context.Context, userID uuid.UUID) ([]string, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT DISTINCT p.code FROM permissions p
		JOIN role_permissions rp ON rp.permission_id = p.id
		JOIN user_roles ur ON ur.role_id = rp.role_id
		WHERE ur.user_id = $1
		ORDER BY p.code
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var codes []string
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			return nil, err
		}
		codes = append(codes, c)
	}
	return codes, rows.Err()
}

func (r *Repository) ListUserRoles(ctx context.Context, userID uuid.UUID) ([]string, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT r.code FROM roles r
		JOIN user_roles ur ON ur.role_id = r.id
		WHERE ur.user_id = $1
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var codes []string
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			return nil, err
		}
		codes = append(codes, c)
	}
	return codes, rows.Err()
}

func (r *Repository) CreateSession(ctx context.Context, s identity.Session, ip, userAgent string) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO sessions (id, user_id, token_hash, audience, expires_at, ip_address, user_agent)
		VALUES ($1, $2, $3, $4, $5, NULLIF($6,'')::inet, NULLIF($7,''))
	`, s.ID, s.UserID, s.TokenHash, s.Audience, s.ExpiresAt, ip, userAgent)
	return err
}

func (r *Repository) FindSessionByTokenHash(ctx context.Context, tokenHash string) (*identity.Session, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, user_id, token_hash, audience, expires_at, revoked_at
		FROM sessions WHERE token_hash = $1
	`, tokenHash)
	var s identity.Session
	err := row.Scan(&s.ID, &s.UserID, &s.TokenHash, &s.Audience, &s.ExpiresAt, &s.RevokedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *Repository) FindSessionByID(ctx context.Context, sessionID uuid.UUID) (*identity.Session, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, user_id, token_hash, audience, expires_at, revoked_at
		FROM sessions WHERE id = $1
	`, sessionID)
	var s identity.Session
	err := row.Scan(&s.ID, &s.UserID, &s.TokenHash, &s.Audience, &s.ExpiresAt, &s.RevokedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *Repository) RevokeSession(ctx context.Context, sessionID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE sessions SET revoked_at = NOW() WHERE id = $1`, sessionID)
	return err
}

func (r *Repository) RevokeUserSessions(ctx context.Context, userID uuid.UUID, audience string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE sessions SET revoked_at = NOW()
		WHERE user_id = $1 AND audience = $2 AND revoked_at IS NULL
	`, userID, audience)
	return err
}

func (r *Repository) UpdateMFA(ctx context.Context, userID uuid.UUID, secret string, enabled bool) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE users SET mfa_secret = $2, mfa_enabled = $3, updated_at = NOW() WHERE id = $1
	`, userID, secret, enabled)
	return err
}

func (r *Repository) FindCustomerIDByUser(ctx context.Context, userID uuid.UUID) (*uuid.UUID, error) {
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, `SELECT id FROM customers WHERE user_id = $1`, userID).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &id, nil
}

func (r *Repository) IsCustomerBlocked(ctx context.Context, userID uuid.UUID) (bool, error) {
	var status string
	err := r.pool.QueryRow(ctx, `SELECT status FROM customers WHERE user_id = $1`, userID).Scan(&status)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return status == "blocked", nil
}

func (r *Repository) UpdateUserProfile(ctx context.Context, userID uuid.UUID, name, phone string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE users SET name = $2, phone = NULLIF($3,''), updated_at = NOW() WHERE id = $1
	`, userID, name, phone)
	return err
}

func (r *Repository) GetCustomerDocument(ctx context.Context, customerID uuid.UUID) (string, error) {
	var doc *string
	err := r.pool.QueryRow(ctx, `SELECT document FROM customers WHERE id = $1`, customerID).Scan(&doc)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	if doc == nil {
		return "", nil
	}
	return *doc, nil
}

func (r *Repository) UpdateCustomerDocument(ctx context.Context, customerID uuid.UUID, document string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE customers SET document = NULLIF($2,''), updated_at = NOW() WHERE id = $1
	`, customerID, document)
	return err
}

var _ identity.Repository = (*Repository)(nil)
