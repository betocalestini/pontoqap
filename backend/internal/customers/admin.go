package customers

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type UpdateCustomerInput struct {
	Name                   *string
	Phone                  *string
	Document               *string
	CollaboratorCategoryID *uuid.UUID
	ClearCollaborator      bool
}

func (s *Service) Update(ctx context.Context, customerID uuid.UUID, in UpdateCustomerInput) (*Customer, error) {
	c, err := s.GetByID(ctx, customerID)
	if err != nil || c == nil {
		return nil, err
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	name := c.Name
	phone := c.Phone
	if in.Name != nil {
		name = *in.Name
	}
	if in.Phone != nil {
		phone = *in.Phone
	}
	_, err = tx.Exec(ctx, `
		UPDATE users SET name = $2, phone = NULLIF($3,''), updated_at = NOW()
		WHERE id = $1
	`, c.UserID, name, phone)
	if err != nil {
		return nil, err
	}

	doc := c.Document
	if in.Document != nil {
		doc = *in.Document
	}
	catID := c.CollaboratorCategoryID
	if in.ClearCollaborator {
		catID = nil
	} else if in.CollaboratorCategoryID != nil {
		if err := s.AssertCollaboratorCategoryAssignable(ctx, *in.CollaboratorCategoryID); err != nil {
			return nil, err
		}
		catID = in.CollaboratorCategoryID
	}
	_, err = tx.Exec(ctx, `
		UPDATE customers SET document = NULLIF($2,''), collaborator_category_id = $3, updated_at = NOW()
		WHERE id = $1
	`, customerID, doc, catID)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return s.GetByID(ctx, customerID)
}

func (s *Service) Block(ctx context.Context, customerID uuid.UUID, reason string) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE customers SET status_before_block = status, status = 'blocked', blocked_reason = NULLIF($2,''), updated_at = NOW()
		WHERE id = $1 AND status != 'blocked'
	`, customerID, reason)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		var status string
		_ = s.pool.QueryRow(ctx, `SELECT status FROM customers WHERE id = $1`, customerID).Scan(&status)
		if status == "blocked" {
			return nil
		}
		return fmt.Errorf("customer not found or cannot block")
	}
	return nil
}

func (s *Service) Unblock(ctx context.Context, customerID uuid.UUID) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE customers SET status = COALESCE(status_before_block, 'approved'), status_before_block = NULL,
			blocked_reason = NULL, updated_at = NOW()
		WHERE id = $1 AND status = 'blocked'
	`, customerID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("customer not blocked")
	}
	return nil
}

func (s *Service) EnsureNotBlocked(c *Customer) error {
	if c == nil {
		return ErrNotApproved()
	}
	if c.Status == "blocked" {
		return ErrBlocked()
	}
	if c.Status != "approved" {
		return ErrNotApproved()
	}
	return nil
}

func ErrBlocked() error {
	return &AppError{Code: "CUSTOMER_BLOCKED", Message: "Cliente bloqueado", Status: 403}
}

func scanCustomer(row pgx.Row) (*Customer, error) {
	var c Customer
	var catID *uuid.UUID
	var catName *string
	var blockedReason *string
	var phone string
	err := row.Scan(&c.ID, &c.UserID, &c.Name, &c.Email, &phone, &c.Document, &c.Status,
		&c.CreditLimitCents, &c.CurrentExposureCents, &c.ApprovedAt, &c.EmailVerified,
		&catID, &catName, &blockedReason,
		&c.OpenInvoicesCount, &c.OverdueInvoicesCount)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	c.Phone = phone
	c.CollaboratorCategoryID = catID
	if catName != nil {
		c.CollaboratorCategoryName = *catName
	}
	if blockedReason != nil {
		c.BlockedReason = *blockedReason
	}
	return &c, nil
}

func (s *Service) attachStaffRoles(ctx context.Context, c *Customer) error {
	if c == nil {
		return nil
	}
	rows, err := s.pool.Query(ctx, `
		SELECT ro.code FROM user_roles ur
		JOIN roles ro ON ro.id = ur.role_id
		WHERE ur.user_id = $1 AND ro.code <> 'customer'
		ORDER BY ro.code
	`, c.UserID)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			return err
		}
		c.StaffRoles = append(c.StaffRoles, code)
	}
	return rows.Err()
}

const customerSelectSQL = `
	SELECT c.id, c.user_id, u.name, u.email, COALESCE(u.phone,''), COALESCE(c.document,''), c.status,
	       c.credit_limit_cents, c.current_exposure_cents, c.approved_at,
	       (u.email_verified_at IS NOT NULL),
	       c.collaborator_category_id, cc.name, c.blocked_reason,
	       COALESCE(inv.open_invoices_count, 0), COALESCE(inv.overdue_invoices_count, 0)
	FROM customers c
	JOIN users u ON u.id = c.user_id
	LEFT JOIN collaborator_categories cc ON cc.id = c.collaborator_category_id
	LEFT JOIN LATERAL (
		SELECT
			COUNT(*) FILTER (
				WHERE i.status IN ('open', 'overdue') AND i.total_cents > i.paid_cents
			)::int AS open_invoices_count,
			COUNT(*) FILTER (WHERE i.status = 'overdue')::int AS overdue_invoices_count
		FROM invoices i
		WHERE i.customer_id = c.id
	) inv ON true
`
