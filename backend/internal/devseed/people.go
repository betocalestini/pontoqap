package devseed

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/store-platform/store/internal/identity/security"
)

type People struct {
	ManagerUserID uuid.UUID
	CustomerIDs   []uuid.UUID
	CustomerUsers []uuid.UUID
}

func seedPeople(ctx context.Context, pool *pgxpool.Pool, cfg Config) (*People, error) {
	hash, err := security.HashPassword(cfg.Password)
	if err != nil {
		return nil, err
	}

	var managerUserID uuid.UUID
	out := &People{}

	for _, spec := range staffSpecs(cfg.Domain) {
		userID, custID, err := insertDualRoleUser(ctx, pool, spec.Name, spec.Email, hash, spec.RoleID)
		if err != nil {
			return nil, err
		}
		if spec.RoleCode == "manager" && managerUserID == uuid.Nil {
			managerUserID = userID
		}
		out.CustomerIDs = append(out.CustomerIDs, custID)
		out.CustomerUsers = append(out.CustomerUsers, userID)
	}
	if managerUserID == uuid.Nil {
		return nil, fmt.Errorf("nenhum gerente seed criado")
	}
	out.ManagerUserID = managerUserID

	dir := ResolveDataDir(cfg)
	customerRows, err := loadCustomersCSV(dir, cfg.Domain)
	if err != nil {
		return nil, err
	}
	for _, row := range customerRows {
		userID, custID, err := insertCustomer(ctx, pool, row.Name, row.Email, hash, row.CreditLimitCents, row.Collaborator, managerUserID)
		if err != nil {
			return nil, err
		}
		out.CustomerIDs = append(out.CustomerIDs, custID)
		out.CustomerUsers = append(out.CustomerUsers, userID)
	}
	// Clientes extras sintéticos além do CSV (SEED_CUSTOMERS / -customers).
	target := cfg.Customers
	if target == 0 {
		target = len(customerRows)
	} else if target < len(customerRows) {
		target = len(customerRows)
	}
	for i := len(customerRows); i < target; i++ {
		email := fmt.Sprintf("demo-cliente-%03d@%s", i+1, cfg.Domain)
		name := fmt.Sprintf("Cliente Demo %03d", i+1)
		limit := int64(250_000 + (i%25)*50_000)
		collab := i%7 == 0
		userID, custID, err := insertCustomer(ctx, pool, name, email, hash, limit, collab, managerUserID)
		if err != nil {
			return nil, err
		}
		out.CustomerIDs = append(out.CustomerIDs, custID)
		out.CustomerUsers = append(out.CustomerUsers, userID)
	}
	return out, nil
}

func insertDualRoleUser(ctx context.Context, pool *pgxpool.Pool, name, email, hash string, staffRoleID uuid.UUID) (userID, customerID uuid.UUID, err error) {
	if userID, customerID, ok, err := lookupCustomerUserByEmail(ctx, pool, email); err != nil {
		return uuid.Nil, uuid.Nil, err
	} else if ok {
		return userID, customerID, nil
	}
	userID = uuid.New()
	customerID = uuid.New()
	tx, err := pool.Begin(ctx)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		INSERT INTO users (id, name, email, password_hash, status, email_verified_at)
		VALUES ($1, $2, $3, $4, 'active', NOW())
	`, userID, name, email, hash)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	_, err = tx.Exec(ctx, `INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2), ($1, $3)`, userID, roleCustomer, staffRoleID)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO customers (id, user_id, status, credit_limit_cents, approved_by, approved_at)
		VALUES ($1, $2, 'approved', $3, $4, NOW())
	`, customerID, userID, int64(50_000), userID)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	return userID, customerID, nil
}

func insertCustomer(ctx context.Context, pool *pgxpool.Pool, name, email, hash string, limit int64, collaborator bool, approvedBy uuid.UUID) (userID, customerID uuid.UUID, err error) {
	if userID, customerID, ok, err := lookupCustomerUserByEmail(ctx, pool, email); err != nil {
		return uuid.Nil, uuid.Nil, err
	} else if ok {
		return userID, customerID, nil
	}
	userID = uuid.New()
	customerID = uuid.New()
	tx, err := pool.Begin(ctx)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		INSERT INTO users (id, name, email, password_hash, status, email_verified_at)
		VALUES ($1, $2, $3, $4, 'active', NOW())
	`, userID, name, email, hash)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	_, err = tx.Exec(ctx, `INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2)`, userID, roleCustomer)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	var collabID *uuid.UUID
	if collaborator {
		collabID = &collaboratorCategory
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO customers (id, user_id, status, credit_limit_cents, collaborator_category_id, approved_by, approved_at)
		VALUES ($1, $2, 'approved', $3, $4, $5, NOW())
	`, customerID, userID, limit, collabID, approvedBy)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	return userID, customerID, nil
}

func resolveActorID(ctx context.Context, pool *pgxpool.Pool, people *People) (uuid.UUID, error) {
	var id uuid.UUID
	err := pool.QueryRow(ctx, `
		SELECT u.id FROM users u
		JOIN user_roles ur ON ur.user_id = u.id
		JOIN roles r ON r.id = ur.role_id
		WHERE r.code = 'system_admin'
		ORDER BY u.created_at ASC
		LIMIT 1
	`).Scan(&id)
	if err == nil {
		return id, nil
	}
	return people.ManagerUserID, nil
}

func lookupCustomerUserByEmail(ctx context.Context, pool *pgxpool.Pool, email string) (userID, customerID uuid.UUID, ok bool, err error) {
	err = pool.QueryRow(ctx, `
		SELECT u.id, c.id FROM users u
		JOIN customers c ON c.user_id = u.id
		WHERE u.email = $1
	`, email).Scan(&userID, &customerID)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, uuid.Nil, false, nil
	}
	if err != nil {
		return uuid.Nil, uuid.Nil, false, err
	}
	return userID, customerID, true, nil
}
