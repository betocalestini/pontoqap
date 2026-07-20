package customers

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/store-platform/store/internal/identity/security"
	platformerrors "github.com/store-platform/store/internal/platform/errors"
)

type Service struct {
	pool *pgxpool.Pool
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

type Customer struct {
	ID                   uuid.UUID  `json:"id"`
	UserID               uuid.UUID  `json:"user_id"`
	Name                 string     `json:"name"`
	Email                string     `json:"email"`
	Document             string     `json:"document,omitempty"`
	Status               string     `json:"status"`
	CreditLimitCents     int64      `json:"credit_limit_cents"`
	CurrentExposureCents int64      `json:"current_exposure_cents"`
	ApprovedAt           *time.Time `json:"approved_at,omitempty"`
}

type RegisterInput struct {
	Name     string
	Email    string
	Password string
	Phone    string
	Document string
}

func (s *Service) Register(ctx context.Context, in RegisterInput) (*Customer, error) {
	hash, err := security.HashPassword(in.Password)
	if err != nil {
		return nil, err
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	userID := uuid.New()
	_, err = tx.Exec(ctx, `
		INSERT INTO users (id, name, email, phone, password_hash, status)
		VALUES ($1, $2, $3, NULLIF($4,''), $5, 'active')
	`, userID, in.Name, in.Email, in.Phone, hash)
	if err != nil {
		return nil, err
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO user_roles (user_id, role_id)
		VALUES ($1, 'a0000000-0000-4000-8000-000000000003')
	`, userID)
	if err != nil {
		return nil, err
	}
	customerID := uuid.New()
	_, err = tx.Exec(ctx, `
		INSERT INTO customers (id, user_id, document, status)
		VALUES ($1, $2, NULLIF($3,''), 'pending')
	`, customerID, userID, in.Document)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return s.GetByID(ctx, customerID)
}

func (s *Service) List(ctx context.Context) ([]Customer, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT c.id, c.user_id, u.name, u.email, COALESCE(c.document,''), c.status,
		       c.credit_limit_cents, c.current_exposure_cents, c.approved_at
		FROM customers c
		JOIN users u ON u.id = c.user_id
		ORDER BY c.created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Customer
	for rows.Next() {
		var c Customer
		if err := rows.Scan(&c.ID, &c.UserID, &c.Name, &c.Email, &c.Document, &c.Status,
			&c.CreditLimitCents, &c.CurrentExposureCents, &c.ApprovedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*Customer, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT c.id, c.user_id, u.name, u.email, COALESCE(c.document,''), c.status,
		       c.credit_limit_cents, c.current_exposure_cents, c.approved_at
		FROM customers c
		JOIN users u ON u.id = c.user_id
		WHERE c.id = $1
	`, id)
	var c Customer
	err := row.Scan(&c.ID, &c.UserID, &c.Name, &c.Email, &c.Document, &c.Status,
		&c.CreditLimitCents, &c.CurrentExposureCents, &c.ApprovedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (s *Service) Approve(ctx context.Context, customerID, managerID uuid.UUID, limitCents int64) error {
	now := time.Now()
	tag, err := s.pool.Exec(ctx, `
		UPDATE customers SET status = 'approved', credit_limit_cents = $3,
		       approved_by = $2, approved_at = $4, updated_at = NOW()
		WHERE id = $1 AND status = 'pending'
	`, customerID, managerID, limitCents, now)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("customer not pending")
	}
	_, err = s.pool.Exec(ctx, `
		INSERT INTO customer_limit_history (customer_id, previous_limit_cents, new_limit_cents, reason, changed_by)
		VALUES ($1, 0, $2, 'Aprovação inicial', $3)
	`, customerID, limitCents, managerID)
	return err
}

func (s *Service) ChangeLimit(ctx context.Context, customerID, managerID uuid.UUID, newLimit int64, reason string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var prev int64
	if err := tx.QueryRow(ctx, `SELECT credit_limit_cents FROM customers WHERE id = $1 FOR UPDATE`, customerID).Scan(&prev); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `UPDATE customers SET credit_limit_cents = $2, updated_at = NOW() WHERE id = $1`, customerID, newLimit); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO customer_limit_history (customer_id, previous_limit_cents, new_limit_cents, reason, changed_by)
		VALUES ($1, $2, $3, $4, $5)
	`, customerID, prev, newLimit, reason, managerID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *Service) AvailableLimit(c Customer) int64 {
	available := c.CreditLimitCents - c.CurrentExposureCents
	if available < 0 {
		return 0
	}
	return available
}

type AppError struct {
	Code    string
	Message string
	Status  int
}

func (e *AppError) Error() string { return e.Message }

func ErrNotApproved() error {
	return &AppError{Code: platformerrors.CodeForbidden, Message: "Cliente não aprovado", Status: 403}
}

func ErrInsufficientLimit() error {
	return &AppError{Code: platformerrors.CodeInsufficientLimit, Message: "Limite de crédito insuficiente", Status: 422}
}
